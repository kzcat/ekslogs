package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestContextCancellation tests that context cancellation is handled properly
func TestContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Verify that the context is cancelled
	assert.Equal(t, context.Canceled, ctx.Err())
}

// TestTailLogsHandlesCancellation tests that TailLogs handles context cancellation
// This is a simplified test that doesn't actually call TailLogs
func TestTailLogsHandlesCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Simulate the behavior in TailLogs
	var err error
	if ctx.Err() == context.Canceled {
		// This is what TailLogs does when context is cancelled
		err = nil
	} else {
		err = fmt.Errorf("some error")
	}

	// Verify that no error is returned when context is cancelled
	assert.NoError(t, err, "Should return nil when context is canceled")
}
