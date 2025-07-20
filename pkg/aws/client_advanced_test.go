package aws

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/kzcat/ekslogs/pkg/log"
)

func TestGetLogsWithFilterPattern(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("/aws/eks/my-cluster/cluster")},
					},
				},
				nil
		},
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return &cloudwatchlogs.DescribeLogStreamsOutput{
					LogStreams: []types.LogStream{
						{LogStreamName: aws.String("kube-apiserver-123")},
					},
				},
				nil
		},
		FilterLogEventsFunc: func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
			// Verify that the filter pattern is correctly passed to the API
			if params.FilterPattern == nil || *params.FilterPattern != "ERROR" {
				t.Errorf("expected FilterPattern to be 'ERROR', got %v", params.FilterPattern)
			}

			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("ERROR: test error message"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
				},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	var receivedLogs []log.LogEntry
	printFunc := func(entry log.LogEntry) {
		receivedLogs = append(receivedLogs, entry)
	}

	// Set filter pattern for the test
	filterPattern := "ERROR"

	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, &filterPattern, 0, printFunc)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	if len(receivedLogs) != 1 {
		t.Errorf("expected 1 log entry, got %d", len(receivedLogs))
	}

	if len(receivedLogs) > 0 && !strings.Contains(receivedLogs[0].Message, "ERROR") {
		t.Errorf("expected log message to contain 'ERROR', got %s", receivedLogs[0].Message)
	}
}

func TestTailLogsWithFilterPattern(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("/aws/eks/my-cluster/cluster")},
					},
				},
				nil
		},
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return &cloudwatchlogs.DescribeLogStreamsOutput{
				LogStreams: []types.LogStream{
					{LogStreamName: aws.String("kube-apiserver-123")},
				},
			}, nil
		},
		FilterLogEventsFunc: func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
			// Verify that the filter pattern is correctly passed to the API
			if params.FilterPattern == nil || *params.FilterPattern != "ERROR" {
				t.Errorf("expected FilterPattern to be 'ERROR', got %v", params.FilterPattern)
			}

			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    true, // Test with verbose mode
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Set filter pattern for the test
	filterPattern := "ERROR"

	err := client.TailLogs(ctx, "my-cluster", []string{"api"}, &filterPattern, 50*time.Millisecond, false)
	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("TailLogs() error = %v", err)
	}
}

func TestGetLogsWithSpecificLogTypes(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("/aws/eks/my-cluster/cluster")},
					},
				},
				nil
		},
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return &cloudwatchlogs.DescribeLogStreamsOutput{
					LogStreams: []types.LogStream{
						{LogStreamName: aws.String("kube-apiserver-123")},
						{LogStreamName: aws.String("kube-apiserver-audit-123")},
						{LogStreamName: aws.String("authenticator-123")},
					},
				},
				nil
		},
		FilterLogEventsFunc: func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
			// Verify that the log streams are filtered correctly
			if len(params.LogStreamNames) != 1 {
				t.Errorf("expected 1 log stream, got %d", len(params.LogStreamNames))
			}

			if len(params.LogStreamNames) > 0 && params.LogStreamNames[0] != "kube-apiserver-123" {
				t.Errorf("expected log stream 'kube-apiserver-123', got %s", params.LogStreamNames[0])
			}

			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("API server log message"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
				},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	var receivedLogs []log.LogEntry
	printFunc := func(entry log.LogEntry) {
		receivedLogs = append(receivedLogs, entry)
	}

	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, nil, 0, printFunc)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	if len(receivedLogs) != 1 {
		t.Errorf("expected 1 log entry, got %d", len(receivedLogs))
	}

	if len(receivedLogs) > 0 && receivedLogs[0].Component != "kube-apiserver" {
		t.Errorf("expected component 'kube-apiserver', got %s", receivedLogs[0].Component)
	}
}

func TestGetLogsWithDescribeLogStreamsError(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("/aws/eks/my-cluster/cluster")},
					},
				},
				nil
		},
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return nil, fmt.Errorf("describe log streams error")
		},
		FilterLogEventsFunc: func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	var receivedLogs []log.LogEntry
	printFunc := func(entry log.LogEntry) {
		receivedLogs = append(receivedLogs, entry)
	}

	// We expect an error here because no log streams will be found
	_ = client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, nil, 0, printFunc)

	// No logs should be received due to the error
	if len(receivedLogs) != 0 {
		t.Errorf("expected 0 log entries, got %d", len(receivedLogs))
	}
}
