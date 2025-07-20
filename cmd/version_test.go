package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionCommand(t *testing.T) {
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

func TestVersionCommandWithFlags(t *testing.T) {
	// Test that the version command has no flags
	assert.Equal(t, 0, len(versionCmd.Flags().Args()))
}

func TestVersionCommandWithArgs(t *testing.T) {
	// Create a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the version command with arguments (should be ignored)
	versionCmd.Run(versionCmd, []string{"arg1", "arg2"})

	// Close the write end of the pipe to flush the buffer
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output still works even with arguments
	assert.Contains(t, output, "ekslogs version")
}
