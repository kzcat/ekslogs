package main

import (
	"os"
	"testing"

	"github.com/kzcat/ekslogs/cmd"
	"github.com/stretchr/testify/assert"
)

// TestMainWithArgs tests the main function with various arguments
func TestMainWithArgs(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	oldExecute := cmd.Execute
	defer func() {
		os.Args = oldArgs
		cmd.Execute = oldExecute
	}()

	// Replace cmd.Execute with a mock function
	executeCalled := false
	cmd.Execute = func() {
		executeCalled = true
	}

	// Test with version command
	os.Args = []string{"ekslogs", "version"}

	// Call main() - this should call our mock Execute function
	main()

	// Verify that Execute was called
	assert.True(t, executeCalled, "Execute should have been called")
}

// TestMainWithoutArgs tests the main function without arguments
func TestMainWithoutArgs(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	oldExecute := cmd.Execute
	defer func() {
		os.Args = oldArgs
		cmd.Execute = oldExecute
	}()

	// Replace cmd.Execute with a mock function
	executeCalled := false
	cmd.Execute = func() {
		executeCalled = true
	}

	// Test with no arguments
	os.Args = []string{"ekslogs"}

	// Call main() - this should call our mock Execute function
	main()

	// Verify that Execute was called
	assert.True(t, executeCalled, "Execute should have been called")
}

// TestMainWithInvalidArgs tests the main function with invalid arguments
func TestMainWithInvalidArgs(t *testing.T) {
	// Save original args and restore after test
	oldArgs := os.Args
	oldExecute := cmd.Execute
	defer func() {
		os.Args = oldArgs
		cmd.Execute = oldExecute
	}()

	// Replace cmd.Execute with a mock function
	executeCalled := false
	cmd.Execute = func() {
		executeCalled = true
	}

	// Test with invalid arguments
	os.Args = []string{"ekslogs", "--invalid-flag"}

	// Call main() - this should call our mock Execute function
	main()

	// Verify that Execute was called
	assert.True(t, executeCalled, "Execute should have been called")
}
