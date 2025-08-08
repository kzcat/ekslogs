package log

import (
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"golang.org/x/term"
	"os"
	"regexp"
	"sort"
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
	// Use golang.org/x/term to properly detect terminal
	return term.IsTerminal(int(file.Fd()))
}

// LogColorizer provides rich color formatting for logs
type LogColorizer struct {
	config *ColorConfig
}

// NewLogColorizer creates a new LogColorizer
func NewLogColorizer(config *ColorConfig) *LogColorizer {
	// Force color output when ColorModeAlways is set
	switch config.Mode {
	case ColorModeAlways:
		color.NoColor = false
	case ColorModeNever:
		color.NoColor = true
	case ColorModeAuto:
		// Let the color package handle detection automatically
	}

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

	// Highlight CRD names and API groups
	crdPattern := regexp.MustCompile(`([a-zA-Z0-9-]+\.[a-zA-Z0-9.-]+\.(com|io|sh|aws|k8s\.aws))`)
	message = crdPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgMagenta, color.Bold).Sprint(s)
	})

	// Highlight Kubernetes resource types in messages
	k8sResourcePattern := regexp.MustCompile(`\b(CRD|CustomResourceDefinition|OpenAPI|spec|controller|webhook|admission)\b`)
	message = k8sResourcePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgYellow).Sprint(s)
	})

	// Highlight file paths and line numbers
	filePathPattern := regexp.MustCompile(`([a-zA-Z0-9_-]+\.go):(\d+)`)
	message = filePathPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgHiBlack).Sprint(s)
	})

	// Highlight success messages
	successPattern := regexp.MustCompile(`(success|successfully|created|updated|deleted|Updating)`)
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
	message := entry.Message

	if strings.HasPrefix(strings.TrimSpace(message), "{") {
		// Try to parse the JSON
		var auditData map[string]interface{}
		err := json.Unmarshal([]byte(message), &auditData)
		if err == nil {
			// Create a new colored version of the message
			coloredMessage := lc.colorizeAuditJSON(auditData)
			if coloredMessage != "" {
				return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, coloredMessage)
			}
		}
	}

	return fmt.Sprintf("%s [%s] [%s] %s", timestamp, level, component, message)
}

// colorizeAuditJSON applies color formatting to audit log JSON data
func (lc *LogColorizer) colorizeAuditJSON(auditData map[string]interface{}) string {
	// Create a deep copy of the audit data to modify
	coloredData := make(map[string]interface{})
	for k, v := range auditData {
		coloredData[k] = v
	}

	// Apply colors to specific fields
	if verb, ok := coloredData["verb"].(string); ok {
		// Color verbs based on their type
		verbColor := color.New(color.FgMagenta)
		switch verb {
		case "create", "update", "patch":
			verbColor = color.New(color.FgGreen, color.Bold)
		case "delete":
			verbColor = color.New(color.FgRed, color.Bold)
		case "get", "list", "watch":
			verbColor = color.New(color.FgCyan)
		}
		coloredData["verb"] = verbColor.Sprint(verb)
	}

	if uri, ok := coloredData["requestURI"].(string); ok {
		coloredData["requestURI"] = color.New(color.FgGreen).Sprint(uri)
	}

	// Handle user information
	if user, ok := coloredData["user"].(map[string]interface{}); ok {
		if username, ok := user["username"].(string); ok {
			// Color usernames based on type
			usernameColor := color.New(color.FgYellow)
			if strings.HasPrefix(username, "system:") {
				usernameColor = color.New(color.FgCyan)
			} else if strings.HasPrefix(username, "eks:") {
				usernameColor = color.New(color.FgMagenta)
			}
			user["username"] = usernameColor.Sprint(username)
		}

		// Color groups
		if groups, ok := user["groups"].([]interface{}); ok {
			coloredGroups := make([]interface{}, len(groups))
			for i, group := range groups {
				if groupStr, ok := group.(string); ok {
					groupColor := color.New(color.FgHiBlack)
					if strings.Contains(groupStr, "system:") {
						groupColor = color.New(color.FgBlue)
					}
					coloredGroups[i] = groupColor.Sprint(groupStr)
				} else {
					coloredGroups[i] = group
				}
			}
			user["groups"] = coloredGroups
		}
	}

	// Handle object reference
	if objectRef, ok := coloredData["objectRef"].(map[string]interface{}); ok {
		if resource, ok := objectRef["resource"].(string); ok {
			objectRef["resource"] = color.New(color.FgCyan, color.Bold).Sprint(resource)
		}
		if namespace, ok := objectRef["namespace"].(string); ok {
			objectRef["namespace"] = color.New(color.FgYellow).Sprint(namespace)
		}
		if name, ok := objectRef["name"].(string); ok {
			objectRef["name"] = color.New(color.FgHiCyan).Sprint(name)
		}
	}

	// Handle source IPs
	if sourceIPs, ok := coloredData["sourceIPs"].([]interface{}); ok {
		coloredIPs := make([]interface{}, len(sourceIPs))
		for i, ip := range sourceIPs {
			if ipStr, ok := ip.(string); ok {
				coloredIPs[i] = color.New(color.FgHiYellow).Sprint(ipStr)
			} else {
				coloredIPs[i] = ip
			}
		}
		coloredData["sourceIPs"] = coloredIPs
	}

	// Handle audit level
	if level, ok := coloredData["level"].(string); ok {
		levelColor := color.New(color.FgBlue)
		switch level {
		case "Request", "RequestResponse":
			levelColor = color.New(color.FgGreen)
		case "Metadata":
			levelColor = color.New(color.FgBlue)
		}
		coloredData["level"] = levelColor.Sprint(level)
	}

	// Handle response status
	if status, ok := coloredData["responseStatus"].(map[string]interface{}); ok {
		// Create a copy of the status
		coloredStatus := make(map[string]interface{})
		for k, v := range status {
			coloredStatus[k] = v
		}

		// Highlight error message
		if errorMsg, ok := coloredStatus["message"].(string); ok {
			coloredStatus["message"] = color.New(color.FgRed, color.Bold).Sprint(errorMsg)
		}

		// Highlight error reason
		if reason, ok := coloredStatus["reason"].(string); ok {
			coloredStatus["reason"] = color.New(color.FgRed).Sprint(reason)
		}

		// Highlight status field
		if statusField, ok := coloredStatus["status"].(string); ok {
			statusColor := color.New(color.FgGreen)
			if statusField == "Failure" {
				statusColor = color.New(color.FgRed, color.Bold)
			}
			coloredStatus["status"] = statusColor.Sprint(statusField)
		}

		// Highlight status code
		if code, ok := coloredStatus["code"].(float64); ok {
			codeColor := color.New(color.FgGreen)
			if code >= 400 {
				codeColor = color.New(color.FgRed, color.Bold)
			}
			coloredStatus["code"] = codeColor.Sprint(int(code))
		}

		// Replace the status with our colored version
		coloredData["responseStatus"] = coloredStatus
	}

	// Convert the colored data back to a string
	// We can't use json.Marshal because it would escape the ANSI color codes
	// Instead, we'll build a custom string representation
	return lc.formatColoredJSON(coloredData)
}

// formatColoredJSON formats a map as a JSON string, preserving ANSI color codes
func (lc *LogColorizer) formatColoredJSON(data map[string]interface{}) string {
	var parts []string

	// Sort keys for consistent output
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		v := data[k]
		formattedValue := lc.formatJSONValue(v)
		parts = append(parts, fmt.Sprintf(`"%s":%s`, k, formattedValue))
	}

	return "{" + strings.Join(parts, ",") + "}"
}

// formatJSONValue formats a value for JSON output, preserving ANSI color codes
func (lc *LogColorizer) formatJSONValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, val)
	case int:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case bool:
		return fmt.Sprintf("%t", val)
	case nil:
		return "null"
	case map[string]interface{}:
		return lc.formatColoredJSON(val)
	case []interface{}:
		var parts []string
		for _, item := range val {
			parts = append(parts, lc.formatJSONValue(item))
		}
		return "[" + strings.Join(parts, ",") + "]"
	default:
		// Fall back to standard JSON for unknown types
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Sprintf(`"%v"`, val)
		}
		return string(jsonBytes)
	}
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

	// Highlight usernames
	usernamePattern := regexp.MustCompile(`username="([^"]+)"`)
	message = usernamePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight error messages and codes (only standalone words or specific patterns)
	errorPattern := regexp.MustCompile(`\b(error|failed|failure|unable to|cannot|timeout|invalid|missing)\b|access (denied|granted)`)
	message = errorPattern.ReplaceAllStringFunc(message, func(s string) string {
		if strings.Contains(s, "granted") {
			return color.New(color.FgGreen).Sprint(s)
		}
		return color.New(color.FgRed).Sprint(s)
	})

	// Highlight AWS error codes (handle escaped quotes)
	awsErrorPattern := regexp.MustCompile(`\\"Code\\":\\"([^"]+)\\"`)
	message = awsErrorPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed, color.Bold).Sprint(s)
	})

	// Highlight AWS error types (handle escaped quotes)
	awsErrorTypePattern := regexp.MustCompile(`\\"Type\\":\\"([^"]+)\\"`)
	message = awsErrorTypePattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed).Sprint(s)
	})

	// Highlight HTTP status codes
	httpStatusPattern := regexp.MustCompile(`\b(200|201|204|400|401|403|404|500|502|503)\b`)
	message = httpStatusPattern.ReplaceAllStringFunc(message, func(s string) string {
		statusCode := s
		statusColor := color.New(color.FgGreen)
		if statusCode[0] == '4' || statusCode[0] == '5' {
			statusColor = color.New(color.FgRed, color.Bold)
		}
		return statusColor.Sprint(s)
	})

	// Highlight IP addresses and ports
	ipPattern := regexp.MustCompile(`\b(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}):(\d+)\b`)
	message = ipPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgHiYellow).Sprint(s)
	})

	// Highlight HTTP methods
	methodPattern := regexp.MustCompile(`method=(GET|POST|PUT|DELETE|PATCH)`)
	message = methodPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgMagenta).Sprint(s)
	})

	// Highlight paths
	pathPattern := regexp.MustCompile(`path=(/[^\s]*)`)
	message = pathPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgCyan).Sprint(s)
	})

	// Highlight log levels in the message
	levelPattern := regexp.MustCompile(`level=(debug|info|warning|error|fatal)`)
	message = levelPattern.ReplaceAllStringFunc(message, func(s string) string {
		levelStr := strings.Split(s, "=")[1]
		levelColor := getLevelColor(levelStr)
		return fmt.Sprintf("level=%s", levelColor.Sprint(levelStr))
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

// ColorizeMessageOnly applies color formatting to just the message part based on log type
func (lc *LogColorizer) ColorizeMessageOnly(message string, logType string, level string) string {
	if !lc.config.ShouldUseColor() {
		return message
	}

	// Apply color based on log type
	switch logType {
	case "api":
		return lc.colorizeAPIMessage(message, level)
	case "audit":
		return lc.colorizeAuditMessage(message, level)
	case "authenticator":
		return lc.colorizeAuthenticatorMessage(message, level)
	case "kcm":
		return lc.colorizeControllerManagerMessage(message, level)
	case "ccm":
		return lc.colorizeCloudControllerManagerMessage(message, level)
	case "scheduler":
		return lc.colorizeSchedulerMessage(message, level)
	default:
		return lc.colorizeDefaultMessage(message, level)
	}
}

// colorizeAPIMessage applies color formatting specific to API server messages
func (lc *LogColorizer) colorizeAPIMessage(message string, level string) string {
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

	return message
}

// colorizeAuditMessage applies color formatting specific to audit messages
func (lc *LogColorizer) colorizeAuditMessage(message string, level string) string {
	// For audit logs, try to parse the JSON and highlight specific fields
	if strings.HasPrefix(strings.TrimSpace(message), "{") {
		var auditData map[string]interface{}
		err := json.Unmarshal([]byte(message), &auditData)
		if err == nil {
			// Use the same JSON colorization as for full logs
			return lc.colorizeAuditJSON(auditData)
		}
	}

	return message
}

// colorizeAuthenticatorMessage applies color formatting specific to authenticator messages
func (lc *LogColorizer) colorizeAuthenticatorMessage(message string, level string) string {
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

	return message
}

// colorizeControllerManagerMessage applies color formatting specific to controller manager messages
func (lc *LogColorizer) colorizeControllerManagerMessage(message string, level string) string {
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

	return message
}

// colorizeCloudControllerManagerMessage applies color formatting specific to cloud controller manager messages
func (lc *LogColorizer) colorizeCloudControllerManagerMessage(message string, level string) string {
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

	return message
}

// colorizeSchedulerMessage applies color formatting specific to scheduler messages
func (lc *LogColorizer) colorizeSchedulerMessage(message string, level string) string {
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

	return message
}

// colorizeDefaultMessage applies default color formatting to messages
func (lc *LogColorizer) colorizeDefaultMessage(message string, level string) string {
	// Highlight error messages
	errorPattern := regexp.MustCompile(`(error|failed|failure|unable to|cannot|timeout)`)
	message = errorPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgRed).Sprint(s)
	})

	// Highlight success messages
	successPattern := regexp.MustCompile(`(success|successfully|created|updated|deleted)`)
	message = successPattern.ReplaceAllStringFunc(message, func(s string) string {
		return color.New(color.FgGreen).Sprint(s)
	})

	return message
}
