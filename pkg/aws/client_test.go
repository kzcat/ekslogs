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
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/kzcat/ekslogs/pkg/log"
)

// MockEKSClient is a mock of the EKSAPI interface
type MockEKSClient struct {
	ListClustersFunc    func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	DescribeClusterFunc func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
}

func (m *MockEKSClient) ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
	return m.ListClustersFunc(ctx, params, optFns...)
}

func (m *MockEKSClient) DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
	return m.DescribeClusterFunc(ctx, params, optFns...)
}

// MockCloudWatchLogsClient is a mock of the CloudWatchLogsAPI interface
type MockCloudWatchLogsClient struct {
	DescribeLogGroupsFunc func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreamsFunc func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	FilterLogEventsFunc func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

func (m *MockCloudWatchLogsClient) DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	return m.DescribeLogGroupsFunc(ctx, params, optFns...)
}

func (m *MockCloudWatchLogsClient) DescribeLogStreams(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
	return m.DescribeLogStreamsFunc(ctx, params, optFns...)
}

func (m *MockCloudWatchLogsClient) FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error) {
	return m.FilterLogEventsFunc(ctx, params, optFns...)
}

// Basic client tests
func TestListClusters(t *testing.T) {
	client := &EKSLogsClient{
		eksClient: &MockEKSClient{
			ListClustersFunc: func(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
				return &eks.ListClustersOutput{
					Clusters: []string{"cluster1", "cluster2"},
				}, nil
			},
		},
		verbose: false,
	}

	clusters, err := client.ListClusters(context.TODO())
	if err != nil {
		t.Fatalf("ListClusters() error = %v", err)
	}
	if len(clusters) != 2 {
		t.Errorf("expected 2 clusters, got %v", clusters)
	}
	if clusters[0] != "cluster1" || clusters[1] != "cluster2" {
		t.Errorf("expected [cluster1 cluster2], got %v", clusters)
	}
}

func TestGetClusterInfo(t *testing.T) {
	client := &EKSLogsClient{
		eksClient: &MockEKSClient{
			DescribeClusterFunc: func(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error) {
				return &eks.DescribeClusterOutput{
					Cluster: &ekstypes.Cluster{
						Name:   aws.String("test-cluster"),
						Status: ekstypes.ClusterStatusActive,
					},
				}, nil
			},
		},
		verbose: false,
	}

	info, err := client.GetClusterInfo(context.TODO(), "test-cluster")
	if err != nil {
		t.Fatalf("GetClusterInfo() error = %v", err)
	}
	if info == nil {
		t.Fatalf("GetClusterInfo() info = nil")
	}
	if info.Status != ekstypes.ClusterStatusActive {
		t.Errorf("expected status %v, got %v", ekstypes.ClusterStatusActive, info.Status)
	}
}

func TestGetLogGroups(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogGroupsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
			return &cloudwatchlogs.DescribeLogGroupsOutput{
					LogGroups: []types.LogGroup{
						{LogGroupName: aws.String("/aws/eks/my-cluster/cluster")},
					},
				},
				nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	logGroups, err := client.GetLogGroups(context.TODO(), "my-cluster")
	if err != nil {
		t.Fatalf("GetLogGroups() error = %v", err)
	}
	if len(logGroups) != 1 {
		t.Errorf("expected 1 log group, got %v", logGroups)
	}
	if logGroups[0] != "/aws/eks/my-cluster/cluster" {
		t.Errorf("expected /aws/eks/my-cluster/cluster, got %v", logGroups[0])
	}
}

func TestGetLogs(t *testing.T) {
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
			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("test message"),
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
		t.Errorf("expected component kube-apiserver, got %s", receivedLogs[0].Component)
	}
}

// Error case tests
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

// Advanced tests
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
