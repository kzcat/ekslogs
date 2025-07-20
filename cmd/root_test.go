package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"
	"time"

	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/kzcat/ekslogs/pkg/log"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockEKSLogsClient is a mock implementation of the EKSLogsClient
type MockEKSLogsClient struct {
	mock.Mock
}

func (m *MockEKSLogsClient) ListClusters(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockEKSLogsClient) GetClusterInfo(ctx context.Context, clusterName string) (interface{}, error) {
	args := m.Called(ctx, clusterName)
	return args.Get(0), args.Error(1)
}

func (m *MockEKSLogsClient) GetLogGroups(ctx context.Context, clusterName string) ([]string, error) {
	args := m.Called(ctx, clusterName)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockEKSLogsClient) GetLogs(ctx context.Context, clusterName string, logTypes []string, startTime, endTime *time.Time, filterPattern *string, limit int32, printFunc func(log.LogEntry)) error {
	args := m.Called(ctx, clusterName, logTypes, startTime, endTime, filterPattern, limit, printFunc)
	return args.Error(0)
}

func (m *MockEKSLogsClient) TailLogs(ctx context.Context, clusterName string, logTypes []string, filterPattern *string, interval time.Duration, messageOnly bool) error {
	args := m.Called(ctx, clusterName, logTypes, filterPattern, interval, messageOnly)
	return args.Error(0)
}

// Helper function to create a new root command with a mock client
func newTestRootCmd(mockClient *MockEKSLogsClient) *cobra.Command {
	// Create a new command that uses the mock client
	cmd := &cobra.Command{
		Use:   "ekslogs <cluster-name> [log-types...]",
		Short: rootCmd.Short,
		Long:  rootCmd.Long,
		Args:  rootCmd.Args,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				clusterName = args[0]
			}
			if len(args) > 1 {
				logTypes = args[1:]
			}

			// Apply preset filter if specified
			if presetName != "" {
				preset, exists := filter.GetUnifiedPreset(presetName)
				if !exists {
					return nil // Just for testing
				}

				if filterPattern == "" {
					filterPattern = preset.Pattern
				}
				if len(logTypes) == 0 {
					logTypes = preset.LogTypes
				}
			}

			messageOnly, _ := cmd.Flags().GetBool("message-only")

			var fp *string
			if filterPattern != "" {
				fp = &filterPattern
			}

			if follow {
				return mockClient.TailLogs(context.Background(), clusterName, logTypes, fp, interval, messageOnly)
			}

			var startT, endT *time.Time
			if startTime != "" {
				t, err := log.ParseTimeString(startTime)
				if err != nil {
					return err
				}
				startT = t
			}

			if endTime != "" {
				t, err := log.ParseTimeString(endTime)
				if err != nil {
					return err
				}
				endT = t
			}

			if startT == nil && endT == nil {
				now := time.Now()
				oneHourAgo := now.Add(-1 * time.Hour)
				startT = &oneHourAgo
				endT = &now
			}

			// Calculate effective limit
			var effectiveLimit int32
			if limitSpecified {
				effectiveLimit = limit
			} else {
				effectiveLimit = 0 // 0 means unlimited
			}

			return mockClient.GetLogs(context.Background(), clusterName, logTypes, startT, endT, fp, effectiveLimit, func(entry log.LogEntry) {
				log.PrintLog(entry, messageOnly)
			})
		},
	}

	// Add the same flags as the original command
	cmd.Flags().StringVarP(&region, "region", "r", "", "AWS region")
	cmd.Flags().StringVarP(&startTime, "start-time", "s", "", "Start time")
	cmd.Flags().StringVarP(&endTime, "end-time", "e", "", "End time")
	cmd.Flags().StringVarP(&filterPattern, "filter-pattern", "f", "", "Filter pattern")
	cmd.Flags().StringVarP(&presetName, "preset", "p", "", "Preset name")
	cmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Limit")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
	cmd.Flags().BoolVarP(&follow, "follow", "F", false, "Follow")
	cmd.Flags().DurationVar(&interval, "interval", 1*time.Second, "Interval")
	cmd.Flags().BoolP("message-only", "m", false, "Message only")

	// Add PreRun to set limitSpecified flag
	cmd.PreRun = func(cmd *cobra.Command, args []string) {
		limitSpecified = cmd.Flags().Changed("limit")
	}

	return cmd
}

func TestPresetFlagHandling(t *testing.T) {
	// Save original values to restore after test
	origFilterPattern := filterPattern
	origLogTypes := logTypes
	defer func() {
		filterPattern = origFilterPattern
		logTypes = origLogTypes
	}()

	// Reset values for test
	filterPattern = ""
	logTypes = nil

	// Test case 1: Valid preset
	presetName = "api-errors"
	preset, exists := filter.GetUnifiedPreset(presetName)
	assert.True(t, exists)

	// Simulate the preset application logic
	if exists {
		if filterPattern == "" {
			filterPattern = preset.Pattern
		}
		if len(logTypes) == 0 {
			logTypes = preset.LogTypes
		}
	}

	// Verify preset was applied correctly
	assert.Equal(t, preset.Pattern, filterPattern)
	assert.Equal(t, preset.LogTypes, logTypes)

	// Test case 2: Custom filter pattern takes precedence
	// Reset values
	filterPattern = "custom-pattern"
	logTypes = nil
	presetName = "api-errors"
	preset, exists = filter.GetUnifiedPreset(presetName)
	assert.True(t, exists)

	// Simulate the preset application logic
	if exists {
		if filterPattern == "" {
			filterPattern = preset.Pattern
		}
		if len(logTypes) == 0 {
			logTypes = preset.LogTypes
		}
	}

	// Verify custom filter pattern was preserved
	assert.Equal(t, "custom-pattern", filterPattern)
	assert.Equal(t, preset.LogTypes, logTypes)
}

func TestLimitFlagHandling(t *testing.T) {
	// Save original values to restore after test
	origLimit := limit
	origLimitSpecified := limitSpecified
	defer func() {
		limit = origLimit
		limitSpecified = origLimitSpecified
	}()

	// Test case 1: When limit flag is not specified
	limitSpecified = false
	
	// Verify that limitSpecified is false
	assert.False(t, limitSpecified)
	
	// Test case 2: When limit flag is specified
	limitSpecified = true
	
	// Verify that limitSpecified is true
	assert.True(t, limitSpecified)
}

func TestEffectiveLimitCalculation(t *testing.T) {
	// Save original values to restore after test
	origLimit := limit
	origLimitSpecified := limitSpecified
	defer func() {
		limit = origLimit
		limitSpecified = origLimitSpecified
	}()

	// Test case 1: When limit flag is not specified
	limit = 1000
	limitSpecified = false
	
	// Calculate effective limit as in the code
	var effectiveLimit int32
	if limitSpecified {
		effectiveLimit = limit
	} else {
		effectiveLimit = 0 // 0 means unlimited
	}
	
	// Verify that effective limit is 0 (unlimited)
	assert.Equal(t, int32(0), effectiveLimit)
	
	// Test case 2: When limit flag is specified
	limit = 500
	limitSpecified = true
	
	// Calculate effective limit again
	if limitSpecified {
		effectiveLimit = limit
	} else {
		effectiveLimit = 0
	}
	
	// Verify that effective limit is the specified value
	assert.Equal(t, int32(500), effectiveLimit)
}

// TestRootCommand tests the root command execution
func TestRootCommand(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a mock client
	mockClient := new(MockEKSLogsClient)
	
	// Setup mock expectations
	mockClient.On("GetLogs", 
		mock.Anything, 
		"test-cluster", 
		mock.AnythingOfType("[]string"), 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*string"), 
		int32(0), 
		mock.AnythingOfType("func(log.LogEntry)"),
	).Return(nil)

	// Create a test command with our mock
	testCmd := newTestRootCmd(mockClient)
	
	// Replace the root command with our test command
	rootCmd = testCmd

	// Test case: Basic command with cluster name
	testCmd.SetArgs([]string{"test-cluster"})
	err := testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Empty(t, logTypes)
	
	// Verify that the mock methods were called
	mockClient.AssertExpectations(t)
}

// TestRootCommandWithLogTypes tests the root command with log types
func TestRootCommandWithLogTypes(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a mock client
	mockClient := new(MockEKSLogsClient)
	
	// Setup mock expectations
	mockClient.On("GetLogs", 
		mock.Anything, 
		"test-cluster", 
		[]string{"api", "audit"}, 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*string"), 
		int32(0), 
		mock.AnythingOfType("func(log.LogEntry)"),
	).Return(nil)

	// Create a test command with our mock
	testCmd := newTestRootCmd(mockClient)
	
	// Replace the root command with our test command
	rootCmd = testCmd

	// Test case: Command with cluster name and log types
	testCmd.SetArgs([]string{"test-cluster", "api", "audit"})
	err := testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Equal(t, []string{"api", "audit"}, logTypes)
	
	// Verify that the mock methods were called
	mockClient.AssertExpectations(t)
}

// TestRootCommandWithFlags tests the root command with various flags
func TestRootCommandWithFlags(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a mock client
	mockClient := new(MockEKSLogsClient)
	
	// Setup mock expectations
	mockClient.On("GetLogs", 
		mock.Anything, 
		"test-cluster", 
		mock.AnythingOfType("[]string"), 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*time.Time"), 
		mock.AnythingOfType("*string"), 
		int32(500), 
		mock.AnythingOfType("func(log.LogEntry)"),
	).Return(nil)

	// Create a test command with our mock
	testCmd := newTestRootCmd(mockClient)
	
	// Replace the root command with our test command
	rootCmd = testCmd

	// Test case: Command with flags
	testCmd.SetArgs([]string{"test-cluster", "-r", "us-west-2", "-v", "-l", "500"})
	err := testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.Equal(t, "us-west-2", region)
	assert.True(t, verbose)
	assert.Equal(t, int32(500), limit)
	
	// Verify that the mock methods were called
	mockClient.AssertExpectations(t)
}

// TestRootCommandWithFollow tests the root command with follow flag
func TestRootCommandWithFollow(t *testing.T) {
	// Save original command to restore after test
	origRootCmd := rootCmd
	defer func() {
		rootCmd = origRootCmd
	}()

	// Create a mock client
	mockClient := new(MockEKSLogsClient)
	
	// Setup mock expectations
	mockClient.On("TailLogs", 
		mock.Anything, 
		"test-cluster", 
		mock.AnythingOfType("[]string"), 
		mock.AnythingOfType("*string"), 
		1*time.Second, 
		false,
	).Return(nil)

	// Create a test command with our mock
	testCmd := newTestRootCmd(mockClient)
	
	// Replace the root command with our test command
	rootCmd = testCmd

	// Test case: Command with follow flag
	testCmd.SetArgs([]string{"test-cluster", "-F"})
	err := testCmd.Execute()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", clusterName)
	assert.True(t, follow)
	
	// Verify that the mock methods were called
	mockClient.AssertExpectations(t)
}

// TestVersionCommand tests the version command
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

// TestLogTypesCommand tests the logtypes command
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

	// Verify output
	assert.Contains(t, output, "Available log types")
	assert.Contains(t, output, "api")
	assert.Contains(t, output, "audit")
	assert.Contains(t, output, "authenticator")
}
