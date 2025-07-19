package filter

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUnifiedPreset(t *testing.T) {
	tests := []struct {
		name          string
		presetName    string
		expectExists  bool
		expectPattern string
		expectLogTypes []string
	}{
		{
			name:          "existing preset",
			presetName:    "api-errors",
			expectExists:  true,
			expectPattern: "ERROR",
			expectLogTypes: []string{"api"},
		},
		{
			name:          "non-existing preset",
			presetName:    "non-existing",
			expectExists:  false,
			expectPattern: "",
			expectLogTypes: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset, exists := GetUnifiedPreset(tt.presetName)
			assert.Equal(t, tt.expectExists, exists)
			if exists {
				assert.Equal(t, tt.expectPattern, preset.Pattern)
				assert.Equal(t, tt.expectLogTypes, preset.LogTypes)
			}
		})
	}
}

func TestListUnifiedPresets(t *testing.T) {
	presets := ListUnifiedPresets()
	assert.NotEmpty(t, presets)
	
	// Check that all defined presets are returned
	expectedCount := len(UnifiedPresets)
	assert.Equal(t, expectedCount, len(presets))
	
	// Check that the list contains expected preset names
	sort.Strings(presets)
	assert.Contains(t, presets, "api-errors")
	assert.Contains(t, presets, "auth-failures")
	assert.Contains(t, presets, "audit-privileged")
}

func TestListBasicPresets(t *testing.T) {
	presets := ListBasicPresets()
	assert.NotEmpty(t, presets)
	
	// Check that only basic presets are returned
	for _, name := range presets {
		preset, exists := GetUnifiedPreset(name)
		assert.True(t, exists)
		assert.False(t, preset.Advanced)
	}
}

func TestListAdvancedPresets(t *testing.T) {
	presets := ListAdvancedPresets()
	assert.NotEmpty(t, presets)
	
	// Check that only advanced presets are returned
	for _, name := range presets {
		preset, exists := GetUnifiedPreset(name)
		assert.True(t, exists)
		assert.True(t, preset.Advanced)
	}
}
