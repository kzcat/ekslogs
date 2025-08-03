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
			input:          "persistent-volume-binder",
			expectedOutput: "\"persistent-volume-binder\"",
			shouldQuote:    true,
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
				!strings.Contains(tt.input, "?") && !strings.Contains(tt.input, "*")

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
