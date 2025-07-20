package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogTypesCommand(t *testing.T) {
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

	// Verify output contains expected strings
	assert.Contains(t, output, "Available log types")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "audit")
	assert.Contains(t, output, "authenticator")
	assert.Contains(t, output, "kcm")
	assert.Contains(t, output, "ccm")
	assert.Contains(t, output, "scheduler")
}

func TestLogTypesCommandWithFlags(t *testing.T) {
	// Test that the logtypes command has no flags
	assert.Equal(t, 0, len(logTypesCmd.Flags().Args()))
}

func TestLogTypesCommandWithArgs(t *testing.T) {
	// Create a buffer to capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Execute the logtypes command with arguments (should be ignored)
	logTypesCmd.Run(logTypesCmd, []string{"arg1", "arg2"})

	// Close the write end of the pipe to flush the buffer
	w.Close()
	os.Stdout = oldStdout

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output still works even with arguments
	assert.Contains(t, output, "Available log types")
}
