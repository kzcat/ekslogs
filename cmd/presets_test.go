package cmd

import (
	"testing"

	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/stretchr/testify/assert"
)

func TestPresetsCommandExists(t *testing.T) {
	// Check that the presets command is registered
	cmd, _, err := rootCmd.Find([]string{"presets"})
	assert.NoError(t, err)
	assert.NotNil(t, cmd)
	assert.Equal(t, "presets", cmd.Name())
}

func TestPresetsData(t *testing.T) {
	// Verify that presets data is available
	presets := filter.ListUnifiedPresets()
	assert.NotEmpty(t, presets)
	
	// Check specific presets
	preset, exists := filter.GetUnifiedPreset("api-errors")
	assert.True(t, exists)
	assert.Equal(t, "ERROR", preset.Pattern)
	assert.Equal(t, []string{"api"}, preset.LogTypes)
	
	preset, exists = filter.GetUnifiedPreset("auth-failures")
	assert.True(t, exists)
	assert.Equal(t, "unauthorized", preset.Pattern)
	assert.Equal(t, []string{"authenticator", "api"}, preset.LogTypes)
}
