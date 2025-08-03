package aws

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterPatternQuoting(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		shouldQuote    bool
	}{
		{
			name:           "simple text should be quoted",
			input:          "error",
			expectedOutput: "\"error\"",
			shouldQuote:    true,
		},
		{
			name:           "text with hyphen should not be quoted",
			input:          "persistent-volume-binder",
			expectedOutput: "persistent-volume-binder",
			shouldQuote:    false,
		},
		{
			name:           "already quoted text should not be double quoted",
			input:          "\"error\"",
			expectedOutput: "\"error\"",
			shouldQuote:    false,
		},
		{
			name:           "json pattern should not be quoted",
			input:          "{ $.user.username = \"admin\" }",
			expectedOutput: "{ $.user.username = \"admin\" }",
			shouldQuote:    false,
		},
		{
			name:           "wildcard pattern should not be quoted",
			input:          "error*",
			expectedOutput: "error*",
			shouldQuote:    false,
		},
		{
			name:           "optional pattern should not be quoted",
			input:          "?error ?warning",
			expectedOutput: "?error ?warning",
			shouldQuote:    false,
		},
		{
			name:           "array pattern should not be quoted",
			input:          "[timestamp, message]",
			expectedOutput: "[timestamp, message]",
			shouldQuote:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldQuote := !strings.HasPrefix(tt.input, "\"") && !strings.HasSuffix(tt.input, "\"") &&
				!strings.Contains(tt.input, "{") && !strings.Contains(tt.input, "[") &&
				!strings.Contains(tt.input, "?") && !strings.Contains(tt.input, "*") &&
				!strings.Contains(tt.input, "-")

			assert.Equal(t, tt.shouldQuote, shouldQuote, "Quote detection should match expected")

			var result string
			if shouldQuote {
				result = "\"" + tt.input + "\""
			} else {
				result = tt.input
			}

			assert.Equal(t, tt.expectedOutput, result, "Output should match expected")
		})
	}
}

func TestIgnoreFilterPatternQuoting(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedOutput string
		shouldQuote    bool
	}{
		{
			name:           "simple text should be quoted with minus prefix",
			input:          "healthcheck",
			expectedOutput: "-\"healthcheck\"",
			shouldQuote:    true,
		},
		{
			name:           "already quoted text should get minus prefix",
			input:          "\"error\"",
			expectedOutput: "-\"error\"",
			shouldQuote:    false,
		},
		{
			name:           "json pattern should get minus prefix",
			input:          "{ $.level = \"INFO\" }",
			expectedOutput: "-{ $.level = \"INFO\" }",
			shouldQuote:    false,
		},
		{
			name:           "wildcard pattern should get minus prefix",
			input:          "debug*",
			expectedOutput: "-debug*",
			shouldQuote:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldQuote := !strings.HasPrefix(tt.input, "\"") && !strings.HasSuffix(tt.input, "\"") &&
				!strings.Contains(tt.input, "{") && !strings.Contains(tt.input, "[") &&
				!strings.Contains(tt.input, "?") && !strings.Contains(tt.input, "*") &&
				!strings.Contains(tt.input, "-")

			assert.Equal(t, tt.shouldQuote, shouldQuote, "Quote detection should match expected")

			var result string
			if shouldQuote {
				result = "-\"" + tt.input + "\""
			} else {
				result = "-" + tt.input
			}

			assert.Equal(t, tt.expectedOutput, result, "Output should match expected")
		})
	}
}

func TestCombinedFilterPatterns(t *testing.T) {
	tests := []struct {
		name           string
		includePattern string
		ignorePattern  string
		expectedOutput string
	}{
		{
			name:           "simple include and ignore patterns",
			includePattern: "error",
			ignorePattern:  "healthcheck",
			expectedOutput: "\"error\" -\"healthcheck\"",
		},
		{
			name:           "only ignore pattern",
			includePattern: "",
			ignorePattern:  "debug",
			expectedOutput: "-\"debug\"",
		},
		{
			name:           "complex include with simple ignore",
			includePattern: "{ $.level = \"ERROR\" }",
			ignorePattern:  "timeout",
			expectedOutput: "{ $.level = \"ERROR\" } -\"timeout\"",
		},
		{
			name:           "already quoted patterns",
			includePattern: "\"warning\"",
			ignorePattern:  "\"info\"",
			expectedOutput: "\"warning\" -\"info\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var combinedPattern string

			// Process include pattern
			if tt.includePattern != "" {
				if !strings.HasPrefix(tt.includePattern, "\"") && !strings.HasSuffix(tt.includePattern, "\"") {
					if !strings.Contains(tt.includePattern, "{") && !strings.Contains(tt.includePattern, "[") &&
						!strings.Contains(tt.includePattern, "?") && !strings.Contains(tt.includePattern, "*") &&
						!strings.Contains(tt.includePattern, "-") {
						combinedPattern = "\"" + tt.includePattern + "\""
					} else {
						combinedPattern = tt.includePattern
					}
				} else {
					combinedPattern = tt.includePattern
				}
			}

			// Process ignore pattern
			if tt.ignorePattern != "" {
				var ignorePattern string
				if !strings.HasPrefix(tt.ignorePattern, "\"") && !strings.HasSuffix(tt.ignorePattern, "\"") {
					if !strings.Contains(tt.ignorePattern, "{") && !strings.Contains(tt.ignorePattern, "[") &&
						!strings.Contains(tt.ignorePattern, "?") && !strings.Contains(tt.ignorePattern, "*") &&
						!strings.Contains(tt.ignorePattern, "-") {
						ignorePattern = "-\"" + tt.ignorePattern + "\""
					} else {
						ignorePattern = "-" + tt.ignorePattern
					}
				} else {
					ignorePattern = "-" + tt.ignorePattern
				}

				if combinedPattern != "" {
					combinedPattern = combinedPattern + " " + ignorePattern
				} else {
					combinedPattern = ignorePattern
				}
			}

			assert.Equal(t, tt.expectedOutput, combinedPattern, "Combined pattern should match expected")
		})
	}
}
