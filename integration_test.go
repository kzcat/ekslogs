package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPresetsCommandIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary for testing
	cmd := exec.Command("go", "build", "-o", "ekslogs_test")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove("ekslogs_test")

	// Run the presets command
	cmd = exec.Command("./ekslogs_test", "presets")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	assert.NoError(t, err)

	// Check output
	output := stdout.String()
	assert.Contains(t, output, "Available basic filter presets:")
	assert.Contains(t, output, "api-errors")
	assert.Contains(t, output, "auth-failures")
}

func TestHelpWithPresetFlag(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Build the binary for testing
	cmd := exec.Command("go", "build", "-o", "ekslogs_test")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build test binary: %v", err)
	}
	defer os.Remove("ekslogs_test")

	// Run the help command
	cmd = exec.Command("./ekslogs_test", "--help")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err = cmd.Run()
	assert.NoError(t, err)

	// Check that help output includes the preset flag
	output := stdout.String()
	assert.Contains(t, output, "--preset")
	assert.Contains(t, output, "-p")
}
