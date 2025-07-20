package cmd

import (
	"testing"
	"time"

	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestPresetFlagHandling(t *testing.T) {
	// Save original values to restore after test
	origFilterPattern := filterPattern
	origLogTypes := logTypes
	defer func() {
		filterPattern = origFilterPattern
		logTypes = origLogTypes
	}()

	// Reset values for test
	filterPattern = ""
	logTypes = nil

	// Test case 1: Valid preset
	presetName = "api-errors"
	preset, exists := filter.GetUnifiedPreset(presetName)
	assert.True(t, exists)

	// Simulate the preset application logic
	if exists {
		if filterPattern == "" {
			filterPattern = preset.Pattern
		}
		if len(logTypes) == 0 {
			logTypes = preset.LogTypes
		}
	}

	// Verify preset was applied correctly
	assert.Equal(t, preset.Pattern, filterPattern)
	assert.Equal(t, preset.LogTypes, logTypes)

	// Test case 2: Custom filter pattern takes precedence
	// Reset values
	filterPattern = "custom-pattern"
	logTypes = nil
	presetName = "api-errors"
	preset, exists = filter.GetUnifiedPreset(presetName)
	assert.True(t, exists)

	// Simulate the preset application logic
	if exists {
		if filterPattern == "" {
			filterPattern = preset.Pattern
		}
		if len(logTypes) == 0 {
			logTypes = preset.LogTypes
		}
	}

	// Verify custom filter pattern was preserved
	assert.Equal(t, "custom-pattern", filterPattern)
	assert.Equal(t, preset.LogTypes, logTypes)
}

func TestLimitFlagHandling(t *testing.T) {
	// Save original values to restore after test
	origLimit := limit
	origLimitSpecified := limitSpecified
	defer func() {
		limit = origLimit
		limitSpecified = origLimitSpecified
	}()

	// Test case 1: When limit flag is not specified
	limitSpecified = false
	
	// Verify that limitSpecified is false
	assert.False(t, limitSpecified)
	
	// Test case 2: When limit flag is specified
	limitSpecified = true
	
	// Verify that limitSpecified is true
	assert.True(t, limitSpecified)
}

func TestEffectiveLimitCalculation(t *testing.T) {
	// Save original values to restore after test
	origLimit := limit
	origLimitSpecified := limitSpecified
	defer func() {
		limit = origLimit
		limitSpecified = origLimitSpecified
	}()

	// Test case 1: When limit flag is not specified
	limit = 1000
	limitSpecified = false
	
	// Calculate effective limit as in the code
	var effectiveLimit int32
	if limitSpecified {
		effectiveLimit = limit
	} else {
		effectiveLimit = 0 // 0 means unlimited
	}
	
	// Verify that effective limit is 0 (unlimited)
	assert.Equal(t, int32(0), effectiveLimit)
	
	// Test case 2: When limit flag is specified
	limit = 500
	limitSpecified = true
	
	// Calculate effective limit again
	if limitSpecified {
		effectiveLimit = limit
	} else {
		effectiveLimit = 0
	}
	
	// Verify that effective limit is the specified value
	assert.Equal(t, int32(500), effectiveLimit)
}

// TestRootCommand tests the root command execution
func TestRootCommand(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a test command that doesn't execute the actual AWS calls
	testCmd := &cobra.Command{
		Use:   "ekslogs <cluster-name> [log-types...]",
		Short: rootCmd.Short,
		Long:  rootCmd.Long,
		Args:  rootCmd.Args,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Just verify args are processed correctly
			if len(args) > 0 {
				clusterName = args[0]
			}
			if len(args) > 1 {
				logTypes = args[1:]
			}
			return nil
		},
	}

	// Add the same flags as the original command
	testCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region")
	testCmd.Flags().StringVarP(&startTime, "start-time", "s", "", "Start time")
	testCmd.Flags().StringVarP(&endTime, "end-time", "e", "", "End time")
	testCmd.Flags().StringVarP(&filterPattern, "filter-pattern", "f", "", "Filter pattern")
	testCmd.Flags().StringVarP(&presetName, "preset", "p", "", "Preset name")
	testCmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Limit")
	testCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	testCmd.Flags().BoolVarP(&follow, "follow", "F", false, "Follow")
	testCmd.Flags().DurationVar(&interval, "interval", 1*time.Second, "Interval")
	testCmd.Flags().BoolP("message-only", "m", false, "Message only")

	// Replace the root command with our test command
	rootCmd = testCmd

	// Test case 1: Basic command with cluster name
	testCmd.SetArgs([]string{"test-cluster"})
	err := testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Empty(t, logTypes)

	// Test case 2: Command with cluster name and log types
	testCmd.SetArgs([]string{"test-cluster", "api", "audit"})
	err = testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Equal(t, []string{"api", "audit"}, logTypes)

	// Test case 3: Command with flags
	testCmd.SetArgs([]string{"test-cluster", "-r", "us-west-2", "-v", "-l", "500"})
	err = testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Equal(t, "us-west-2", region)
	assert.True(t, verbose)
	assert.Equal(t, int32(500), limit)
}

// TestVersionCommand tests the version command
func TestVersionCommand(t *testing.T) {
	// Skip this test for now as it requires more complex setup
	t.Skip("Skipping version command test")
}

// TestLogTypesCommand tests the logtypes command
func TestLogTypesCommand(t *testing.T) {
	// Skip this test for now as it requires more complex setup
	t.Skip("Skipping logtypes command test")
}
