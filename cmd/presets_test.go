package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresetsCommand(t *testing.T) {
	// Save original values to restore after test
	origShowAdvanced := showAdvanced
	origShowAll := showAll
	defer func() {
		showAdvanced = origShowAdvanced
		showAll = origShowAll
	}()

	tests := []struct {
		name         string
		args         []string
		showAdvanced bool
		showAll      bool
		contains     []string
		notContains  []string
	}{
		{
			name:         "basic presets",
			args:         []string{},
			showAdvanced: false,
			showAll:      false,
			contains: []string{
				"Available basic filter presets:",
				"Usage example:",
				"To see advanced presets",
			},
			notContains: []string{
				"Pattern types:",
			},
		},
		{
			name:         "advanced presets",
			args:         []string{"--advanced"},
			showAdvanced: true,
			showAll:      false,
			contains: []string{
				"Available advanced filter presets:",
				"Pattern types:",
				"simple:",
				"optional:",
				"exclude:",
				"json:",
				"regex:",
			},
			notContains: []string{
				"To see advanced presets",
			},
		},
		{
			name:         "all presets",
			args:         []string{"--all"},
			showAdvanced: false,
			showAll:      true,
			contains: []string{
				"Available filter presets (basic and advanced):",
				"Pattern types:",
			},
			notContains: []string{
				"To see advanced presets",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set flags
			showAdvanced = tt.showAdvanced
			showAll = tt.showAll

			// Create a buffer to capture output
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			unifiedPresetsCmd.Run(unifiedPresetsCmd, tt.args)

			// Close the write end of the pipe to flush the buffer
			w.Close()
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			io.Copy(&buf, r)
			output := buf.String()

			// Verify output contains expected strings
			for _, s := range tt.contains {
				assert.Contains(t, output, s)
			}

			// Verify output does not contain unexpected strings
			for _, s := range tt.notContains {
				assert.NotContains(t, output, s)
			}
		})
	}
}

func TestPresetsCommandFlags(t *testing.T) {
	// Test that the flags are properly registered
	advancedFlag := unifiedPresetsCmd.Flags().Lookup("advanced")
	assert.NotNil(t, advancedFlag)
	assert.Equal(t, "advanced", advancedFlag.Name)
	assert.Equal(t, "false", advancedFlag.DefValue)

	allFlag := unifiedPresetsCmd.Flags().Lookup("all")
	assert.NotNil(t, allFlag)
	assert.Equal(t, "all", allFlag.Name)
	assert.Equal(t, "false", allFlag.DefValue)
}
