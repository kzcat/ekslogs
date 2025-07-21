package log

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"os"
	"regexp"
	"strings"
	"time"
)

// ColorMode defines how colors should be handled
type ColorMode string

const (
	// ColorModeAuto automatically determines whether to use colors based on terminal detection
	ColorModeAuto ColorMode = "auto"
	// ColorModeAlways forces colors to be used
	ColorModeAlways ColorMode = "always"
	// ColorModeNever disables colors
	ColorModeNever ColorMode = "never"
)

// ColorConfig holds the configuration for color output
type ColorConfig struct {
	Mode ColorMode
}

// NewColorConfig creates a new ColorConfig with default settings
func NewColorConfig() *ColorConfig {
	return &ColorConfig{
		Mode: ColorModeAuto,
	}
}

// ShouldUseColor determines whether colors should be used based on the configuration
func (c *ColorConfig) ShouldUseColor() bool {
	switch c.Mode {
	case ColorModeAlways:
		return true
	case ColorModeNever:
		return false
	case ColorModeAuto:
		// Check if output is a terminal
		return isTerminal(os.Stdout)
	default:
		return false
	}
}

// isTerminal checks if the given file is a terminal
func isTerminal(file *os.File) bool {
	// Use the IsTerminal function from the color package
	return !color.NoColor
}

// LogColorizer provides rich color formatting for logs
type LogColorizer struct {
	config *ColorConfig
}

// NewLogColorizer creates a new LogColorizer
func NewLogColorizer(config *ColorConfig) *LogColorizer {
	return &LogColorizer{
		config: config,
	}
}

// ColorizeLog applies color formatting to a log entry based on its type and content
func (lc *LogColorizer) ColorizeLog(entry LogEntry) string {
	if !lc.config.ShouldUseColor() {
		// Return plain text if colors are disabled
		timestamp := entry.Timestamp.UTC().Format(time.RFC3339)
		return fmt.Sprintf("%s [%s] [%s] %s",
			timestamp,
			entry.Level,
			entry.Component,
			entry.Message,
		)
	}

	// Apply color based on log type
	switch NormalizeLogType(ExtractLogTypeFromStreamName(entry.LogStream)) {
	case "api":
		return lc.colorizeAPILog(entry)
	case "audit":
		return lc.colorizeAuditLog(entry)
	case "authenticator":
		return lc.colorizeAuthenticatorLog(entry)
	case "kcm":
		return lc.colorizeControllerManagerLog(entry)
	case "ccm":
		return lc.colorizeCloudControllerManagerLog(entry)
	case "scheduler":
		return lc.colorizeSchedulerLog(entry)
	default:
		return lc.colorizeDefaultLog(entry)
	}
}

// colorizeAPILog applies color formatting specific to API server logs
func (lc *LogColorizer) colorizeAPILog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)

	// Color the level
	levelColor := getLevelColor(entry.Level)
	level := levelColor.SprintFunc()(entry.Level)

	// Colorize specific patterns in the message
	message := entry.Message

	// Highlight error messages
	errorPattern := regexp.MustCompile(`(error|failed|failure|unable to|cannot|timeout)`)
	message = errorPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed).Sprint(s)
	})

	// Highlight resource names
	resourcePattern := regexp.MustCompile(`(pod|node|service|deployment|daemonset|statefulset|configmap|secret|namespace)/([a-zA-Z0-9-_.]+)`)
	message = resourcePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight success messages
	successPattern := regexp.MustCompile(`(success|successfully|created|updated|deleted)`)
	message = successPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgGreen).Sprint(s)
	})

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeAuditLog applies color formatting specific to audit logs
func (lc *LogColorizer) colorizeAuditLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := color.New(color.FgBlue).SprintFunc()(entry.Level)

	// For audit logs, try to parse the JSON and highlight specific fields
	var auditData map[string]interface{}
	message := entry.Message

	if strings.HasPrefix(strings.TrimSpace(message), "{") {
		err := json.Unmarshal([]byte(message), &auditData)
		if err == nil {
			// Format specific fields with colors
			if verb, ok := auditData["verb"].(string); ok {
				verbColor := color.New(color.FgMagenta).SprintFunc()
				message = strings.Replace(message, fmt.Sprintf(`"verb":"%s"`, verb),
					fmt.Sprintf(`"verb":"%s"`, verbColor(verb)), 1)
			}

			if uri, ok := auditData["requestURI"].(string); ok {
				uriColor := color.New(color.FgGreen).SprintFunc()
				message = strings.Replace(message, fmt.Sprintf(`"requestURI":"%s"`, uri),
					fmt.Sprintf(`"requestURI":"%s"`, uriColor(uri)), 1)
			}

			// Highlight user information
			if user, ok := auditData["user"].(map[string]interface{}); ok {
				if username, ok := user["username"].(string); ok {
					usernameColor := color.New(color.FgYellow).SprintFunc()
					message = strings.Replace(message, fmt.Sprintf(`"username":"%s"`, username),
						fmt.Sprintf(`"username":"%s"`, usernameColor(username)), 1)
				}
			}

			// Highlight response status
			if status, ok := auditData["responseStatus"].(map[string]interface{}); ok {
				if code, ok := status["code"].(float64); ok {
					codeStr := fmt.Sprintf("%.0f", code)
					codeColor := color.New(color.FgGreen)
					if code >= 400 {
						codeColor = color.New(color.FgRed)
					}
					message = strings.Replace(message, fmt.Sprintf(`"code":%s`, codeStr),
						fmt.Sprintf(`"code":%s`, codeColor.Sprint(codeStr)), 1)
				}
			}
		}
	}

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeAuthenticatorLog applies color formatting specific to authenticator logs
func (lc *LogColorizer) colorizeAuthenticatorLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := getLevelColor(entry.Level).SprintFunc()(entry.Level)

	message := entry.Message

	// Highlight ARNs
	arnPattern := regexp.MustCompile(`arn:aws:[a-zA-Z0-9-]+:[a-zA-Z0-9-]*:[0-9]+:[a-zA-Z0-9-:/]+`)
	message = arnPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgYellow).Sprint(s)
	})

	// Highlight access granted/denied
	accessPattern := regexp.MustCompile(`access (granted|denied)`)
	message = accessPattern.ReplaceAllStringFunc(message, func(match string) string {
		if strings.Contains(match, "granted") {
			return color.New(color.FgGreen).Sprint(match)
		}
		return color.New(color.FgRed).Sprint(match)
	})

	// Highlight usernames
	usernamePattern := regexp.MustCompile(`username="([^"]+)"`)
	message = usernamePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeControllerManagerLog applies color formatting specific to controller manager logs
func (lc *LogColorizer) colorizeControllerManagerLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := getLevelColor(entry.Level).SprintFunc()(entry.Level)

	message := entry.Message

	// Highlight controller names
	controllerPattern := regexp.MustCompile(`\b([a-zA-Z0-9-]+)_controller\b`)
	message = controllerPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgMagenta).Sprint(s)
	})

	// Highlight resource names
	resourcePattern := regexp.MustCompile(`(pod|node|service|deployment|daemonset|statefulset|configmap|secret|namespace)/([a-zA-Z0-9-_.]+)`)
	message = resourcePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight error messages
	errorPattern := regexp.MustCompile(`(error|failed|failure|unable to|cannot|timeout)`)
	message = errorPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed).Sprint(s)
	})

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeCloudControllerManagerLog applies color formatting specific to cloud controller manager logs
func (lc *LogColorizer) colorizeCloudControllerManagerLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := getLevelColor(entry.Level).SprintFunc()(entry.Level)

	message := entry.Message

	// Highlight AWS resource IDs
	awsResourcePattern := regexp.MustCompile(`\b(vpc-|subnet-|sg-|i-|vol-|rtb-|igw-|nat-|eni-|eip-|acl-)[a-f0-9]+\b`)
	message = awsResourcePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight controller names
	controllerPattern := regexp.MustCompile(`\b([a-zA-Z0-9-]+)_controller\b`)
	message = controllerPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgMagenta).Sprint(s)
	})

	// Highlight error messages
	errorPattern := regexp.MustCompile(`(error|failed|failure|unable to|cannot|timeout)`)
	message = errorPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed).Sprint(s)
	})

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeSchedulerLog applies color formatting specific to scheduler logs
func (lc *LogColorizer) colorizeSchedulerLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := getLevelColor(entry.Level).SprintFunc()(entry.Level)

	message := entry.Message

	// Highlight scheduling related keywords
	schedPattern := regexp.MustCompile(`\b(schedule|scheduling|scheduled|unschedulable|predicates|priorities|binding|bound)\b`)
	message = schedPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgMagenta).Sprint(s)
	})

	// Highlight pod names
	podPattern := regexp.MustCompile(`pod/([a-zA-Z0-9-_.]+)`)
	message = podPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight node names
	nodePattern := regexp.MustCompile(`node/([a-zA-Z0-9-_.]+)`)
	message = nodePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgYellow).Sprint(s)
	})

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeDefaultLog applies default color formatting to logs
func (lc *LogColorizer) colorizeDefaultLog(entry LogEntry) string {
	timestamp := color.New(color.FgHiBlack).SprintFunc()(entry.Timestamp.UTC().Format(time.RFC3339))
	component := color.New(color.FgGreen).SprintFunc()(entry.Component)
	level := getLevelColor(entry.Level).SprintFunc()(entry.Level)

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, entry.Message)
}

// getLevelColor returns the appropriate color for a log level
func getLevelColor(level string) *color.Color {
	switch strings.ToLower(level) {
	case "info":
		return color.New(color.FgBlue)
	case "warning", "warn":
		return color.New(color.FgYellow)
	case "error", "err":
		return color.New(color.FgRed)
	case "fatal", "crit":
		return color.New(color.FgHiRed)
	default:
		return color.New()
	}
}
