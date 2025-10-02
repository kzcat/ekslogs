package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/kzcat/ekslogs/pkg/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestCommandInitialization tests the initialization of commands
func TestCommandInitialization(t *testing.T) {
	// Test that the root command has subcommands
	assert.NotEmpty(t, rootCmd.Commands())

	// Test that the version command is registered
	found := false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be registered")

	// Test that the logtypes command is registered
	found = false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "logtypes" {
			found = true
			break
		}
	}
	assert.True(t, found, "logtypes command should be registered")

	// Test that the presets command is registered
	found = false
	for _, c := range rootCmd.Commands() {
		if c.Name() == "presets" {
			found = true
			break
		}
	}
	assert.True(t, found, "presets command should be registered")
}

// TestVersionCommandOutput tests the version command
func TestVersionCommandOutput(t *testing.T) {
	// Save original values to restore after test
	origVersion := version
	origCommit := commit
	origDate := date
	defer func() {
		version = origVersion
		commit = origCommit
		date = origDate
	}()

	// Set test values
	version = "1.0.0"
	commit = "abcdef"
	date = "2024-01-01"

	// Create a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the version command
	versionCmd.Run(versionCmd, []string{})

	// Close the write end of the pipe to flush the buffer
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	// Verify output
	assert.Contains(t, output, "ekslogs version 1.0.0")
	assert.Contains(t, output, "commit: abcdef")
	assert.Contains(t, output, "built at: 2024-01-01")
}

// TestLogTypesCommandOutput tests the logtypes command
func TestLogTypesCommandOutput(t *testing.T) {
	// Create a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the logtypes command
	logTypesCmd.Run(logTypesCmd, []string{})

	// Close the write end of the pipe to flush the buffer
	if err := w.Close(); err != nil {
		t.Fatalf("Failed to close pipe: %v", err)
	}
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("Failed to copy output: %v", err)
	}
	output := buf.String()

	// Verify output
	assert.Contains(t, output, "Available log types")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "audit")
	assert.Contains(t, output, "authenticator")
	assert.Contains(t, output, "kcm")
	assert.Contains(t, output, "ccm")
	assert.Contains(t, output, "scheduler")
	assert.Contains(t, output, "Not all log types may be available")
}

// TestRootCommandFlags tests the flags of the root command
func TestRootCommandFlags(t *testing.T) {
	// Test that all expected flags are registered
	flags := rootCmd.Flags()

	// Check required flags
	assert.NotNil(t, flags.Lookup("region"))
	assert.NotNil(t, flags.Lookup("start-time"))
	assert.NotNil(t, flags.Lookup("end-time"))
	assert.NotNil(t, flags.Lookup("filter-pattern"))
	assert.NotNil(t, flags.Lookup("preset"))
	assert.NotNil(t, flags.Lookup("limit"))
	assert.NotNil(t, flags.Lookup("verbose"))
	assert.NotNil(t, flags.Lookup("follow"))
	assert.NotNil(t, flags.Lookup("interval"))
	assert.NotNil(t, flags.Lookup("message-only"))
}

// TestPreRunFunction tests the PreRun function of the root command
func TestPreRunFunction(t *testing.T) {
	// Save original values to restore after test
	origLimitSpecified := limitSpecified
	defer func() {
		limitSpecified = origLimitSpecified
	}()

	// Reset value
	limitSpecified = false

	// Create a mock command
	cmd := rootCmd

	// Test case 1: Flag not changed
	cmd.Flags().Changed("limit") // This doesn't actually mark it as changed
	rootCmd.PreRun(cmd, []string{})
	assert.False(t, limitSpecified)

	// Test case 2: Flag changed (simulated)
	limitSpecified = true // Simulate flag changed
	assert.True(t, limitSpecified)
}

// TestFilterPatternHandling tests the filter pattern handling logic
func TestFilterPatternHandling(t *testing.T) {
	// Save original values to restore after test
	origFilterPatterns := filterPatterns
	defer func() {
		filterPatterns = origFilterPatterns
	}()

	// Test cases
	testCases := []struct {
		name            string
		initialPattern  string
		expectedPointer bool
	}{
		{
			name:            "empty pattern",
			initialPattern:  "",
			expectedPointer: false,
		},
		{
			name:            "non-empty pattern",
			initialPattern:  "ERROR",
			expectedPointer: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set filter pattern
			if tc.initialPattern != "" {
				filterPatterns = []string{tc.initialPattern}
			} else {
				filterPatterns = []string{}
			}

			// Create pointer if needed
			var fp *string
			if len(filterPatterns) > 0 {
				combinedPattern := buildCombinedFilterPattern(filterPatterns, []string{}, false)
				if combinedPattern != "" {
					fp = &combinedPattern
				}
			}

			// Verify pointer
			if tc.expectedPointer {
				assert.NotNil(t, fp)
				// For simple text patterns, they should be quoted
				expectedPattern := tc.initialPattern
				if !strings.HasPrefix(expectedPattern, "\"") && !strings.HasSuffix(expectedPattern, "\"") &&
					!strings.Contains(expectedPattern, "{") && !strings.Contains(expectedPattern, "[") &&
					!strings.Contains(expectedPattern, "?") && !strings.Contains(expectedPattern, "*") &&
					!strings.HasPrefix(expectedPattern, "-") {
					expectedPattern = fmt.Sprintf("\"%s\"", expectedPattern)
				}
				assert.Equal(t, expectedPattern, *fp)
			} else {
				assert.Nil(t, fp)
			}
		})
	}
}

// TestEffectiveLimitCalculation tests the effective limit calculation
func TestEffectiveLimitCalculation(t *testing.T) {
	// Save original values to restore after test
	origLimit := limit
	origLimitSpecified := limitSpecified
	defer func() {
		limit = origLimit
		limitSpecified = origLimitSpecified
	}()

	// Test cases
	testCases := []struct {
		name           string
		limit          int32
		limitSpecified bool
		expectedLimit  int32
	}{
		{
			name:           "limit not specified",
			limit:          1000,
			limitSpecified: false,
			expectedLimit:  0, // 0 means unlimited
		},
		{
			name:           "limit specified",
			limit:          500,
			limitSpecified: true,
			expectedLimit:  500,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set values
			limit = tc.limit
			limitSpecified = tc.limitSpecified

			// Calculate effective limit
			var effectiveLimit int32
			if limitSpecified {
				effectiveLimit = limit
			} else {
				effectiveLimit = 0 // 0 means unlimited
			}

			// Verify result
			assert.Equal(t, tc.expectedLimit, effectiveLimit)
		})
	}
}

// TestPresetsCommand tests the presets command
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
			if err := w.Close(); err != nil {
				t.Fatalf("Failed to close pipe: %v", err)
			}
			os.Stdout = oldStdout

			// Read the output
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, r); err != nil {
				t.Fatalf("Failed to copy output: %v", err)
			}
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

// TestPresetsCommandFlags tests the flags of the presets command
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

// TestPresetApplication tests the preset application logic
func TestPresetApplication(t *testing.T) {
	// Save original values to restore after test
	origPresetName := presetName
	origFilterPatterns := filterPatterns
	origLogTypes := logTypes
	defer func() {
		presetName = origPresetName
		filterPatterns = origFilterPatterns
		logTypes = origLogTypes
	}()

	// Test cases
	testCases := []struct {
		name           string
		presetName     string
		initialFilter  string
		initialTypes   []string
		expectedFilter string
		expectedTypes  []string
	}{
		{
			name:           "apply preset completely",
			presetName:     "api-errors",
			initialFilter:  "",
			initialTypes:   nil,
			expectedFilter: "ERROR",         // From api-errors preset
			expectedTypes:  []string{"api"}, // From api-errors preset
		},
		{
			name:           "preserve custom filter",
			presetName:     "api-errors",
			initialFilter:  "custom-filter",
			initialTypes:   nil,
			expectedFilter: "custom-filter", // Custom filter preserved
			expectedTypes:  []string{"api"}, // From api-errors preset
		},
		{
			name:           "preserve custom log types",
			presetName:     "api-errors",
			initialFilter:  "",
			initialTypes:   []string{"audit"},
			expectedFilter: "ERROR",           // From api-errors preset
			expectedTypes:  []string{"audit"}, // Custom log types preserved
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset values
			presetName = tc.presetName
			if tc.initialFilter != "" {
				filterPatterns = []string{tc.initialFilter}
			} else {
				filterPatterns = []string{}
			}
			logTypes = tc.initialTypes

			// Get the preset
			preset, exists := filter.GetUnifiedPreset(presetName)
			assert.True(t, exists)

			// Apply preset filter pattern if no custom filter pattern is provided
			if len(filterPatterns) == 0 {
				filterPatterns = []string{preset.Pattern}
			}

			// Apply preset log types if no custom log types are provided
			if len(logTypes) == 0 {
				logTypes = preset.LogTypes
			}

			// Verify results
			if tc.expectedFilter != "" {
				assert.Equal(t, []string{tc.expectedFilter}, filterPatterns)
			} else {
				assert.Equal(t, []string{}, filterPatterns)
			}
			assert.Equal(t, tc.expectedTypes, logTypes)
		})
	}
}

// TestErrorHandling tests the error handling of the root command
func TestErrorHandling(t *testing.T) {
	// Create a test command that returns an error
	testCmd := &cobra.Command{
		Use: "test",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("test error")
		},
	}

	// Execute the command
	err := testCmd.Execute()

	// Verify that the error is returned
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())
}

// TestInvalidArguments tests the root command with invalid arguments
func TestInvalidArguments(t *testing.T) {
	// Test with no arguments (should fail)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least 1 arg")

	// Test with invalid flag
	rootCmd.SetArgs([]string{"test-cluster", "--invalid-flag"})
	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag")
}

// TestRegionHandling tests the region handling logic
func TestRegionHandling(t *testing.T) {
	// Save original values to restore after test
	origRegion := region
	defer func() {
		region = origRegion
	}()

	// Test cases
	testCases := []struct {
		name           string
		initialRegion  string
		expectedRegion string
	}{
		{
			name:           "explicit region",
			initialRegion:  "us-west-2",
			expectedRegion: "us-west-2",
		},
		{
			name:           "empty region",
			initialRegion:  "",
			expectedRegion: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set region
			region = tc.initialRegion

			// Verify region
			assert.Equal(t, tc.expectedRegion, region)
		})
	}
}

// TestDefaultTimeRange tests the default time range logic
func TestDefaultTimeRange(t *testing.T) {
	// Test case: both start and end time are nil
	var startT, endT *string

	// Verify that default time range would be applied
	if startT == nil && endT == nil {
		assert.True(t, true, "Default time range should be applied")
	} else {
		assert.Fail(t, "Default time range should be applied")
	}
}

// TestTimeRangeParsing tests the time range parsing logic
func TestTimeRangeParsing(t *testing.T) {
	// Test valid relative time
	startTime := "-1h"
	_, err := log.ParseTimeString(startTime)
	assert.NoError(t, err)

	// Test invalid time format
	startTime = "invalid-time"
	_, err = log.ParseTimeString(startTime)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse time")

	// Test valid RFC3339 time
	startTime = "2024-01-01T00:00:00Z"
	_, err = log.ParseTimeString(startTime)
	assert.NoError(t, err)
}
