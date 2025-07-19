package log

import (
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
			name:     "unknown",
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

func TestExtractLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "info log",
			message:  "I0719 06:09:10.476002 1 controller.go:123] Starting controller",
			expected: "info",
		},
		{
			name:     "warning log",
			message:  "W0719 06:09:10.476002 1 controller.go:123] Warning message",
			expected: "warning",
		},
		{
			name:     "error log",
			message:  "E0719 06:09:10.476002 1 controller.go:123] Error occurred",
			expected: "error",
		},
		{
			name:     "fatal log",
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
		{
			name:     "empty message",
			message:  "",
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

func TestPrintLog(t *testing.T) {
	tests := []struct {
		name        string
		logEntry    LogEntry
		messageOnly bool
		wantOutput  string
	}{
		{
			name: "standard output",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "info",
				Component: "kube-apiserver",
				Message:   "test message",
			},
			messageOnly: false,
			wantOutput:  "2024-01-01T12:00:00Z [info] [kube-apiserver] test message\n",
		},
		{
			name: "message only output",
			logEntry: LogEntry{
				Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
				Level:     "warning",
				Component: "scheduler",
				Message:   "warning message",
			},
			messageOnly: true,
			wantOutput:  "warning message\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Call PrintLog
			PrintLog(tt.logEntry, tt.messageOnly)

			// Read captured output
			w.Close()
			out, _ := io.ReadAll(r)
			actualOutput := string(out)

			// Restore stdout
			os.Stdout = oldStdout

			// Compare with expected output
			if !strings.Contains(actualOutput, tt.logEntry.Message) {
				t.Errorf("PrintLog() output = %q, should contain %q", actualOutput, tt.logEntry.Message)
			}
		})
	}
}
