package aws

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/kzcat/ekslogs/pkg/log"
)

func TestListClustersError(t *testing.T) {
	client := &EKSLogsClient{
		eksClient: &MockEKSClient{
			ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
				return nil, fmt.Errorf("test error")
			},
		},
		verbose: false,
	}

	clusters, err := client.ListClusters(context.TODO())
	if err == nil {
		t.Fatalf("ListClusters() expected error, got nil")
	}
	if len(clusters) != 0 {
		t.Errorf("expected 0 clusters, got %v", clusters)
	}
}

func TestGetClusterInfoError(t *testing.T) {
	client := &EKSLogsClient{
		eksClient: &MockEKSClient{
			DescribeClusterFunc: func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
				return nil, fmt.Errorf("test error")
			},
		},
		verbose: false,
	}

	info, err := client.GetClusterInfo(context.TODO(), "test-cluster")
	if err == nil {
		t.Fatalf("GetClusterInfo() expected error, got nil")
	}
	if info != nil {
		t.Errorf("expected nil info, got %v", info)
	}
}

func TestGetLogGroupsError(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return nil, fmt.Errorf("test error")
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	logGroups, err := client.GetLogGroups(context.TODO(), "my-cluster")
	if err == nil {
		t.Fatalf("GetLogGroups() expected error, got nil")
	}
	if len(logGroups) != 0 {
		t.Errorf("expected 0 log groups, got %v", logGroups)
	}
}

func TestGetLogsWithNoLogGroups(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []types.LogGroup{},
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

	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, nil, 10, printFunc)
	if err == nil {
		t.Fatalf("GetLogs() expected error, got nil")
	}
	if len(receivedLogs) != 0 {
		t.Errorf("expected 0 log entries, got %d", len(receivedLogs))
	}
}

func TestTailLogsWithNoLogGroups(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
				LogGroups: []types.LogGroup{},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := client.TailLogs(ctx, "my-cluster", []string{"api"}, nil, 1*time.Second, false)
	if err == nil {
		t.Fatalf("TailLogs() expected error, got nil")
	}
}
