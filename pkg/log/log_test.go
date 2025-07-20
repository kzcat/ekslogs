package log

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseTimeString(t *testing.T) {
	tests := []struct {
		name      string
		timeStr   string
		wantError bool
	}{
		{
			name:      "empty string",
			timeStr:   "",
			wantError: false,
		},
		{
			name:      "RFC3339 format",
			timeStr:   "2024-01-01T12:00:00Z",
			wantError: false,
		},
		{
			name:      "invalid format",
			timeStr:   "2024-01-01",
			wantError: true,
		},
		{
			name:      "relative time in hours",
			timeStr:   "-1h",
			wantError: false,
		},
		{
			name:      "relative time in minutes",
			timeStr:   "-30m",
			wantError: false,
		},
		{
			name:      "relative time in seconds",
			timeStr:   "-60s",
			wantError: false,
		},
		{
			name:      "relative time in days",
			timeStr:   "-2d",
			wantError: false,
		},
		{
			name:      "invalid relative time",
			timeStr:   "-1x",
			wantError: true,
		},
		{
			name:      "invalid number in relative time",
			timeStr:   "-xh",
			wantError: true,
		},
		// "now" is not supported in the current implementation
		//{
		//	name:      "now keyword",
		//	timeStr:   "now",
		//	wantError: false,
		//},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseTimeString(tt.timeStr)
			if tt.wantError && err == nil {
				t.Errorf("ParseTimeString(%q) expected error, got nil", tt.timeStr)
			}
			if !tt.wantError && err != nil {
				t.Errorf("ParseTimeString(%q) unexpected error: %v", tt.timeStr, err)
			}
			if tt.timeStr == "" && result != nil {
				t.Errorf("ParseTimeString(%q) expected nil result, got %v", tt.timeStr, result)
			}
		})
	}
}

func TestNormalizeLogType(t *testing.T) {
	tests := []struct {
		name     string
		logType  string
		expected string
	}{
		{
			name:     "short name - api",
			logType:  "api",
			expected: "api",
		},
		{
			name:     "short name - audit",
			logType:  "audit",
			expected: "audit",
		},
		{
			name:     "alias - auth",
			logType:  "auth",
			expected: "authenticator",
		},
		{
			name:     "long name - kube-controller-manager",
			logType:  "kube-controller-manager",
			expected: "kcm",
		},
		{
			name:     "alias - controller",
			logType:  "controller",
			expected: "kcm",
		},
		{
			name:     "alias - cloud",
			logType:  "cloud",
			expected: "ccm",
		},
		{
			name:     "alias - sched",
			logType:  "sched",
			expected: "scheduler",
		},
		{
			name:     "unknown type",
			logType:  "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeLogType(tt.logType)
			if result != tt.expected {
				t.Errorf("NormalizeLogType(%q) = %q, expected %q", tt.logType, result, tt.expected)
			}
		})
	}
}

func TestGetLogTypeDescription(t *testing.T) {
	tests := []struct {
		name              string
		availableLogTypes []string
		expected          string
	}{
		{
			name:              "single log type",
			availableLogTypes: []string{"api"},
			expected:          "api (kube-apiserver)",
		},
		{
			name:              "multiple log types",
			availableLogTypes: []string{"api", "audit", "kcm"},
			expected:          "api (kube-apiserver), audit (kube-apiserver-audit), kcm (kubeControllerManager, kube-controller-manager, controller)",
		},
		{
			name:              "unknown log type",
			availableLogTypes: []string{"unknown"},
			expected:          "unknown",
		},
		{
			name:              "empty list",
			availableLogTypes: []string{},
			expected:          "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetLogTypeDescription(tt.availableLogTypes)
			if result != tt.expected {
				t.Errorf("GetLogTypeDescription(%v) = %q, expected %q", tt.availableLogTypes, result, tt.expected)
			}
		})
	}
}

func TestExtractLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "empty message",
			message:  "",
			expected: "",
		},
		{
			name:     "kubernetes info log",
			message:  "I0719 06:09:10.476002 1 controller.go:123] Starting controller",
			expected: "info",
		},
		{
			name:     "kubernetes warning log",
			message:  "W0719 06:09:10.476002 1 controller.go:123] Warning message",
			expected: "warning",
		},
		{
			name:     "kubernetes error log",
			message:  "E0719 06:09:10.476002 1 controller.go:123] Error occurred",
			expected: "error",
		},
		{
			name:     "kubernetes fatal log",
			message:  "F0719 06:09:10.476002 1 controller.go:123] Fatal error",
			expected: "fatal",
		},
		{
			name:     "json info log",
			message:  `{"level":"info","msg":"Starting controller"}`,
			expected: "info",
		},
		{
			name:     "json warning log",
			message:  `{"level":"warning","msg":"Warning message"}`,
			expected: "warning",
		},
		{
			name:     "json error log",
			message:  `{"level":"error","msg":"Error occurred"}`,
			expected: "error",
		},
		{
			name:     "unknown format",
			message:  "Starting controller",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractLogLevel(tt.message)
			if result != tt.expected {
				t.Errorf("ExtractLogLevel(%q) = %q, expected %q", tt.message, result, tt.expected)
			}
		})
	}
}

func TestExtractComponentFromStreamName(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		expected   string
	}{
		{
			name:       "kube-apiserver-audit",
			streamName: "kube-apiserver-audit-123456",
			expected:   "kube-apiserver-audit",
		},
		{
			name:       "kube-apiserver",
			streamName: "kube-apiserver-123456",
			expected:   "kube-apiserver",
		},
		{
			name:       "authenticator",
			streamName: "authenticator-123456",
			expected:   "authenticator",
		},
		{
			name:       "kube-controller-manager",
			streamName: "kube-controller-manager-123456",
			expected:   "kube-controller-manager",
		},
		{
			name:       "cloud-controller-manager",
			streamName: "cloud-controller-manager-123456",
			expected:   "cloud-controller-manager",
		},
		{
			name:       "kube-scheduler",
			streamName: "kube-scheduler-123456",
			expected:   "kube-scheduler",
		},
		{
			name:       "unknown",
			streamName: "unknown-123456",
			expected:   "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractComponentFromStreamName(tt.streamName)
			if result != tt.expected {
				t.Errorf("ExtractComponentFromStreamName(%q) = %q, expected %q", tt.streamName, result, tt.expected)
			}
		})
	}
}

func TestExtractLogTypeFromStreamName(t *testing.T) {
	tests := []struct {
		name       string
		streamName string
		expected   string
	}{
		{
			name:       "audit log",
			streamName: "kube-apiserver-audit-123456",
			expected:   "audit",
		},
		{
			name:       "api log",
			streamName: "kube-apiserver-123456",
			expected:   "api",
		},
		{
			name:       "authenticator log",
			streamName: "authenticator-123456",
			expected:   "authenticator",
		},
		{
			name:       "kcm log",
			streamName: "kube-controller-manager-123456",
			expected:   "kcm",
		},
		{
			name:       "ccm log",
			streamName: "cloud-controller-manager-123456",
			expected:   "ccm",
		},
		{
			name:       "scheduler log",
			streamName: "kube-scheduler-123456",
			expected:   "scheduler",
		},
		{
			name:       "unknown log",
			streamName: "unknown-123456",
			expected:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractLogTypeFromStreamName(tt.streamName)
			if result != tt.expected {
				t.Errorf("ExtractLogTypeFromStreamName(%q) = %q, expected %q", tt.streamName, result, tt.expected)
			}
		})
	}
}

func TestPrintLog(t *testing.T) {
	// Save original stdout
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
	}()

	tests := []struct {
		name        string
		logEntry    LogEntry
		messageOnly bool
		contains    []string
	}{
		{
			name: "message only",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "info",
				Component: "kube-apiserver",
				Message:   "Test message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: true,
			contains:    []string{"Test message"},
		},
		{
			name: "full log - info level",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "info",
				Component: "kube-apiserver",
				Message:   "Info message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: false,
			contains:    []string{"2024-01-01T12:00:00Z", "info", "kube-apiserver", "Info message"},
		},
		{
			name: "full log - warning level",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "warning",
				Component: "kube-apiserver",
				Message:   "Warning message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: false,
			contains:    []string{"2024-01-01T12:00:00Z", "warning", "kube-apiserver", "Warning message"},
		},
		{
			name: "full log - error level",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "error",
				Component: "kube-apiserver",
				Message:   "Error message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: false,
			contains:    []string{"2024-01-01T12:00:00Z", "error", "kube-apiserver", "Error message"},
		},
		{
			name: "full log - fatal level",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "fatal",
				Component: "kube-apiserver",
				Message:   "Fatal message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: false,
			contains:    []string{"2024-01-01T12:00:00Z", "fatal", "kube-apiserver", "Fatal message"},
		},
		{
			name: "full log - unknown level",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "",
				Component: "kube-apiserver",
				Message:   "Unknown level message",
				LogGroup:  "/aws/eks/test/cluster",
				LogStream: "kube-apiserver-123456",
			},
			messageOnly: false,
			contains:    []string{"2024-01-01T12:00:00Z", "kube-apiserver", "Unknown level message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a pipe to capture stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call the function
			PrintLog(tt.logEntry, tt.messageOnly)

			// Close the write end of the pipe to flush the buffer
			w.Close()

			// Read the output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Check if the output contains the expected strings
			for _, s := range tt.contains {
				if !strings.Contains(output, s) {
					t.Errorf("PrintLog() output does not contain %q, got: %q", s, output)
				}
			}
		})
	}
}
