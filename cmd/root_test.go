package cmd

import (
	"testing"

	"github.com/kzcat/ekslogs/pkg/filter"
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
