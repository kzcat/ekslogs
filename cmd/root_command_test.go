package cmd

import (
	"testing"

	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/kzcat/ekslogs/pkg/log"
	"github.com/stretchr/testify/assert"
)

// TestPresetApplication tests the preset application logic
func TestPresetApplication(t *testing.T) {
	// Save original values to restore after test
	origPresetName := presetName
	origFilterPattern := filterPattern
	origLogTypes := logTypes
	defer func() {
		presetName = origPresetName
		filterPattern = origFilterPattern
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
			filterPattern = tc.initialFilter
			logTypes = tc.initialTypes

			// Get the preset
			preset, exists := filter.GetUnifiedPreset(presetName)
			assert.True(t, exists)

			// Apply preset filter pattern if no custom filter pattern is provided
			if filterPattern == "" {
				filterPattern = preset.Pattern
			}

			// Apply preset log types if no custom log types are provided
			if len(logTypes) == 0 {
				logTypes = preset.LogTypes
			}

			// Verify results
			assert.Equal(t, tc.expectedFilter, filterPattern)
			assert.Equal(t, tc.expectedTypes, logTypes)
		})
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

// TestEffectiveLimitWithFlags tests the effective limit calculation with flags
func TestEffectiveLimitWithFlags(t *testing.T) {
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
