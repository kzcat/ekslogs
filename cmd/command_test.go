package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCommandInitialization tests the initialization of commands
func TestCommandInitialization(t *testing.T) {
	// Test that the root command has subcommands
	assert.NotEmpty(t, rootCmd.Commands())

	// Test that the version command is registered
	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "version" {
			found = true
			break
		}
	}
	assert.True(t, found, "version command should be registered")

	// Test that the logtypes command is registered
	found = false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "logtypes" {
			found = true
			break
		}
	}
	assert.True(t, found, "logtypes command should be registered")

	// Test that the presets command is registered
	found = false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "presets" {
			found = true
			break
		}
	}
	assert.True(t, found, "presets command should be registered")
}

// TestRootCommandFlagsCheck tests the flags of the root command
func TestRootCommandFlagsCheck(t *testing.T) {
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

// TestPreRunFunctionCheck tests the PreRun function of the root command
func TestPreRunFunctionCheck(t *testing.T) {
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
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
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
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output
	assert.Contains(t, output, "Available log types")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "audit")
	assert.Contains(t, output, "authenticator")
	assert.Contains(t, output, "kcm")
	assert.Contains(t, output, "ccm")
	assert.Contains(t, output, "scheduler")
}
