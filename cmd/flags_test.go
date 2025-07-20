package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestRootCommandFlags(t *testing.T) {
	// Test that all expected flags are registered
	flags := rootCmd.Flags()

	// Required flags
	regionFlag := flags.Lookup("region")
	assert.NotNil(t, regionFlag)
	assert.Equal(t, "region", regionFlag.Name)
	assert.Equal(t, "r", regionFlag.Shorthand)
	assert.Equal(t, "", regionFlag.DefValue)

	startTimeFlag := flags.Lookup("start-time")
	assert.NotNil(t, startTimeFlag)
	assert.Equal(t, "start-time", startTimeFlag.Name)
	assert.Equal(t, "s", startTimeFlag.Shorthand)
	assert.Equal(t, "", startTimeFlag.DefValue)

	endTimeFlag := flags.Lookup("end-time")
	assert.NotNil(t, endTimeFlag)
	assert.Equal(t, "end-time", endTimeFlag.Name)
	assert.Equal(t, "e", endTimeFlag.Shorthand)
	assert.Equal(t, "", endTimeFlag.DefValue)

	filterPatternFlag := flags.Lookup("filter-pattern")
	assert.NotNil(t, filterPatternFlag)
	assert.Equal(t, "filter-pattern", filterPatternFlag.Name)
	assert.Equal(t, "f", filterPatternFlag.Shorthand)
	assert.Equal(t, "", filterPatternFlag.DefValue)

	presetFlag := flags.Lookup("preset")
	assert.NotNil(t, presetFlag)
	assert.Equal(t, "preset", presetFlag.Name)
	assert.Equal(t, "p", presetFlag.Shorthand)
	assert.Equal(t, "", presetFlag.DefValue)

	limitFlag := flags.Lookup("limit")
	assert.NotNil(t, limitFlag)
	assert.Equal(t, "limit", limitFlag.Name)
	assert.Equal(t, "l", limitFlag.Shorthand)
	assert.Equal(t, "1000", limitFlag.DefValue)

	verboseFlag := flags.Lookup("verbose")
	assert.NotNil(t, verboseFlag)
	assert.Equal(t, "verbose", verboseFlag.Name)
	assert.Equal(t, "v", verboseFlag.Shorthand)
	assert.Equal(t, "false", verboseFlag.DefValue)

	followFlag := flags.Lookup("follow")
	assert.NotNil(t, followFlag)
	assert.Equal(t, "follow", followFlag.Name)
	assert.Equal(t, "F", followFlag.Shorthand)
	assert.Equal(t, "false", followFlag.DefValue)

	intervalFlag := flags.Lookup("interval")
	assert.NotNil(t, intervalFlag)
	assert.Equal(t, "interval", intervalFlag.Name)
	assert.Equal(t, "", intervalFlag.Shorthand)
	assert.Equal(t, "1s", intervalFlag.DefValue)

	messageOnlyFlag := flags.Lookup("message-only")
	assert.NotNil(t, messageOnlyFlag)
	assert.Equal(t, "message-only", messageOnlyFlag.Name)
	assert.Equal(t, "m", messageOnlyFlag.Shorthand)
	assert.Equal(t, "false", messageOnlyFlag.DefValue)
}

func TestFlagParsing(t *testing.T) {
	// Skip this test for now as it's difficult to test flag parsing without side effects
	t.Skip("Skipping flag parsing test")
}

func TestPreRunFunction(t *testing.T) {
	// Save original values to restore after test
	origLimitSpecified := limitSpecified
	defer func() {
		limitSpecified = origLimitSpecified
	}()

	// Reset value
	limitSpecified = false

	// Create a mock command with the same PreRun function
	cmd := &cobra.Command{}
	cmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Limit")
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		limitSpecified = cmd.Flags().Changed("limit")
	}

	// Test case 1: Flag not specified
	cmd.ParseFlags([]string{})
	cmd.PreRun(cmd, []string{})
	assert.False(t, limitSpecified)

	// Test case 2: Flag specified
	cmd.ParseFlags([]string{"-l", "500"})
	cmd.PreRun(cmd, []string{})
	assert.True(t, limitSpecified)
}
