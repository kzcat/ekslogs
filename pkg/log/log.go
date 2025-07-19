package log

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp time.Time `json:"@timestamp"`
	Level     string    `json:"level,omitempty"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
	LogGroup  string    `json:"log_group"`
	LogStream string    `json:"log_stream"`
}

func ParseTimeString(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		return nil, nil
	}

	// For relative time
	if strings.HasPrefix(timeStr, "-") {
		return parseRelativeTime(timeStr)
	}

	// For RFC3339 format
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse time '%s': expected RFC3339 format (2006-01-02T15:04:05Z) or relative format (-1h, -15m, -30s, -2d)", timeStr)
	}

	return &t, nil
}

func parseRelativeTime(relativeTime string) (*time.Time, error) {
	if relativeTime == "" {
		return nil, nil
	}

	// Check relative time pattern (e.g., -1h, -15m, -30s, -2d)
	re := regexp.MustCompile(`^-(\d+)([smhd])$`)
	matches := re.FindStringSubmatch(relativeTime)

	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid relative time format: %s (expected format: -1h, -15m, -30s, -2d)", relativeTime)
	}

	value, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid number in relative time: %s", matches[1])
	}

	unit := matches[2]
	var duration time.Duration

	switch unit {
	case "s":
		duration = time.Duration(value) * time.Second
	case "m":
		duration = time.Duration(value) * time.Minute
	case "h":
		duration = time.Duration(value) * time.Hour
	case "d":
		duration = time.Duration(value) * 24 * time.Hour
	default:
		return nil, fmt.Errorf("unsupported time unit: %s (supported: s, m, h, d)", unit)
	}

	result := time.Now().Add(-duration)
	return &result, nil
}

func NormalizeLogType(logType string) string {
	// Log type mapping (long names -> short names)
	logTypeMap := map[string]string{
		// Short names (as is)
		"api":           "api",
		"audit":         "audit",
		"auth":          "authenticator",
		"authenticator": "authenticator",
		"kcm":           "kcm",
		"ccm":           "ccm",
		"sched":         "scheduler",
		"scheduler":     "scheduler",

		// Long names (for compatibility)
		"kubeControllerManager":    "kcm",
		"cloudControllerManager":   "ccm",
		"kube-controller-manager":  "kcm",
		"cloud-controller-manager": "ccm",
		"controller":               "kcm", // Common abbreviation
		"cloud":                    "ccm", // Common abbreviation
	}

	if normalized, exists := logTypeMap[logType]; exists {
		return normalized
	}
	return logType // Return as is if not in the mapping
}

func GetLogTypeDescription(availableLogTypes []string) string {
	descriptions := map[string]string{
		"api":           "api (kube-apiserver)",
		"audit":         "audit (kube-apiserver-audit)",
		"authenticator": "authenticator (auth, authenticator)",
		"kcm":           "kcm (kubeControllerManager, kube-controller-manager, controller)",
		"ccm":           "ccm (cloudControllerManager, cloud-controller-manager, cloud)",
		"scheduler":     "scheduler (sched)",
	}

	var result []string
	for _, logType := range availableLogTypes {
		if desc, exists := descriptions[logType]; exists {
			result = append(result, desc)
		} else {
			result = append(result, logType)
		}
	}

	return strings.Join(result, ", ")
}

func ExtractLogLevel(message string) string {
	if len(message) == 0 {
		return ""
	}

	// Kubernetes log format: I0719 06:09:10.476002 ...
	if len(message) > 0 {
		switch message[0] {
		case 'I':
			return "info"
		case 'W':
			return "warning"
		case 'E':
			return "error"
		case 'F':
			return "fatal"
		}
	}

	// For JSON format logs
	if strings.Contains(message, `"level":"`) {
		if strings.Contains(message, `"level":"info"`) {
			return "info"
		} else if strings.Contains(message, `"level":"warning"`) {
			return "warning"
		} else if strings.Contains(message, `"level":"error"`) {
			return "error"
		}
	}

	return ""
}

func ExtractComponentFromStreamName(streamName string) string {
	if strings.HasPrefix(streamName, "kube-apiserver-audit-") {
		return "kube-apiserver-audit"
	} else if strings.HasPrefix(streamName, "kube-apiserver-") {
		return "kube-apiserver"
	} else if strings.HasPrefix(streamName, "authenticator-") {
		return "authenticator"
	} else if strings.HasPrefix(streamName, "kube-controller-manager-") {
		return "kube-controller-manager"
	} else if strings.HasPrefix(streamName, "cloud-controller-manager-") {
		return "cloud-controller-manager"
	} else if strings.HasPrefix(streamName, "kube-scheduler-") {
		return "kube-scheduler"
	}
	return "unknown"
}

func ExtractLogTypeFromStreamName(streamName string) string {
	// Determine log type based on EKS log stream name pattern
	if strings.HasPrefix(streamName, "kube-apiserver-audit-") {
		return "audit"
	} else if strings.HasPrefix(streamName, "kube-apiserver-") {
		return "api"
	} else if strings.HasPrefix(streamName, "authenticator-") {
		return "authenticator"
	} else if strings.HasPrefix(streamName, "kube-controller-manager-") {
		return "kcm"
	} else if strings.HasPrefix(streamName, "cloud-controller-manager-") {
		return "ccm"
	} else if strings.HasPrefix(streamName, "kube-scheduler-") {
		return "scheduler"
	}
	return ""
}

func PrintLog(log LogEntry, messageOnly bool) {
	if messageOnly {
		fmt.Println(log.Message)
		// Flush stdout to ensure immediate output when piped
		os.Stdout.Sync()
		return
	}

	// Color settings
	levelColor := color.New()
	switch log.Level {
	case "info":
		levelColor = color.New(color.FgGreen)
	case "warning":
		levelColor = color.New(color.FgYellow)
	case "error":
		levelColor = color.New(color.FgRed)
	case "fatal":
		levelColor = color.New(color.FgHiRed)
	}

	timestamp := log.Timestamp.UTC().Format(time.RFC3339)
	fmt.Printf("%s [%s] [%s] %s\n",
		timestamp,
		levelColor.SprintFunc()(log.Level),
		color.CyanString(log.Component),
		log.Message,
	)

	// Flush stdout to ensure immediate output when piped
	os.Stdout.Sync()
}
