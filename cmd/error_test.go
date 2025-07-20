package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandling(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a test command that returns an error
	testCmd := &cobra.Command{
		Use:   "ekslogs <cluster-name> [log-types...]",
		Short: rootCmd.Short,
		Long:  rootCmd.Long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.New("test error")
		},
	}

	// Replace the root command with our test command
	rootCmd = testCmd

	// Create a buffer to capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Execute the command (this should call os.Exit(1) in a real scenario)
	// But we're testing the error handling before that point
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Equal(t, "test error", err.Error())

	// Close the write end of the pipe to flush the buffer
	w.Close()
	os.Stderr = oldStderr

	// Read the output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// The error should be printed to stderr
	// But since we're not actually executing the full Execute() function
	// which would call os.Exit(1), we don't see the color.Red output
	// This is just testing that the error is properly returned from rootCmd.Execute()
	assert.NotEmpty(t, output)
}

func TestInvalidArguments(t *testing.T) {
	// Test with no arguments (should fail)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "requires at least 1 arg")

	// Test with invalid flag
	rootCmd.SetArgs([]string{"test-cluster", "--invalid-flag"})
	err = rootCmd.Execute()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown flag")
}
