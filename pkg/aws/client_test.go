package aws

import (
	"context"
	"fmt"
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
	DescribeLogGroupsFunc  func(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreamsFunc func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	FilterLogEventsFunc    func(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
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

	if len(clusters) != 2 || clusters[0] != "cluster1" {
		t.Errorf("expected 2 clusters, got %v", clusters)
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

	if *info.Name != "test-cluster" {
		t.Errorf("expected cluster name 'test-cluster', got '%s'", *info.Name)
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

	if len(logGroups) != 1 || logGroups[0] != "/aws/eks/my-cluster/cluster" {
		t.Errorf("expected 1 log group, got %v", logGroups)
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
							Message:       aws.String("test log message"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
					},
				},
				nil
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
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	if len(receivedLogs) != 1 || receivedLogs[0].Message != "test log message" {
		t.Errorf("expected 1 log entry, got %v", receivedLogs)
	}
}

func TestTailLogs(t *testing.T) {
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
							Message:       aws.String("tail log message"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
					},
				},
				nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Run for a short duration
	defer cancel()

	err := client.TailLogs(ctx, "my-cluster", []string{"api"}, nil, 1*time.Second, false)
	if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
		t.Fatalf("TailLogs() error = %v", err)
	}
}

func TestGetLogsWithPagination(t *testing.T) {
	// Test case for pagination with multiple pages of log events
	var filterLogEventsCallCount int
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
			filterLogEventsCallCount++

			// First page has events and a next token
			if filterLogEventsCallCount == 1 {
				return &cloudwatchlogs.FilterLogEventsOutput{
					Events: []types.FilteredLogEvent{
						{
							Timestamp:     aws.Int64(time.Now().UnixMilli()),
							Message:       aws.String("page 1 log message 1"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
						{
							Timestamp:     aws.Int64(time.Now().UnixMilli()),
							Message:       aws.String("page 1 log message 2"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
					},
					NextToken: aws.String("next-token-1"),
				}, nil
			}

			// Second page has events and a next token
			if filterLogEventsCallCount == 2 {
				return &cloudwatchlogs.FilterLogEventsOutput{
					Events: []types.FilteredLogEvent{
						{
							Timestamp:     aws.Int64(time.Now().UnixMilli()),
							Message:       aws.String("page 2 log message 1"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
						{
							Timestamp:     aws.Int64(time.Now().UnixMilli()),
							Message:       aws.String("page 2 log message 2"),
							LogStreamName: aws.String("kube-apiserver-123"),
						},
					},
					NextToken: aws.String("next-token-2"),
				}, nil
			}

			// Third page has events but no next token (last page)
			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("page 3 log message 1"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
				},
				NextToken: nil, // No more pages
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

	// Test with no limit (should get all logs)
	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, nil, 0, printFunc)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	// Should have received all 5 log entries from all 3 pages
	if len(receivedLogs) != 5 {
		t.Errorf("expected 5 log entries, got %d", len(receivedLogs))
	}

	// Verify that FilterLogEvents was called 3 times (for all 3 pages)
	if filterLogEventsCallCount != 3 {
		t.Errorf("expected FilterLogEvents to be called 3 times, got %d", filterLogEventsCallCount)
	}
}

func TestGetLogsWithLimit(t *testing.T) {
	// Test case for applying limits correctly
	var filterLogEventsCallCount int
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
			filterLogEventsCallCount++

			// First page has many events
			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message 1"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message 2"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message 3"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message 4"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message 5"),
						LogStreamName: aws.String("kube-apiserver-123"),
					},
				},
				NextToken: aws.String("next-token"), // More pages available
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	// Test with a limit of 3 logs
	var receivedLogs []log.LogEntry
	printFunc := func(entry log.LogEntry) {
		receivedLogs = append(receivedLogs, entry)
	}

	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, nil, nil, nil, 3, printFunc)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	// Should have received exactly 3 log entries (due to limit)
	if len(receivedLogs) != 3 {
		t.Errorf("expected 3 log entries, got %d", len(receivedLogs))
	}

	// Verify that FilterLogEvents was called only once
	// (we don't need to fetch more pages after reaching the limit)
	if filterLogEventsCallCount != 1 {
		t.Errorf("expected FilterLogEvents to be called once, got %d", filterLogEventsCallCount)
	}
}

func TestGetLogsWithTimeRange(t *testing.T) {
	// Test case for applying time range correctly
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
			// Verify that the time range parameters are correctly passed to the API
			if params.StartTime == nil {
				t.Errorf("expected StartTime to be set")
			}
			if params.EndTime == nil {
				t.Errorf("expected EndTime to be set")
			}

			return &cloudwatchlogs.FilterLogEventsOutput{
				Events: []types.FilteredLogEvent{
					{
						Timestamp:     aws.Int64(time.Now().UnixMilli()),
						Message:       aws.String("log message within time range"),
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

	// Set time range for the test
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now()

	err := client.GetLogs(context.TODO(), "my-cluster", []string{"api"}, &startTime, &endTime, nil, 0, printFunc)
	if err != nil {
		t.Fatalf("GetLogs() error = %v", err)
	}

	if len(receivedLogs) != 1 {
		t.Errorf("expected 1 log entry, got %d", len(receivedLogs))
	}
}

func TestGetAvailableLogTypes(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return &cloudwatchlogs.DescribeLogStreamsOutput{
				LogStreams: []types.LogStream{
					{LogStreamName: aws.String("kube-apiserver-123")},
					{LogStreamName: aws.String("kube-apiserver-audit-123")},
					{LogStreamName: aws.String("authenticator-123")},
					{LogStreamName: aws.String("kube-controller-manager-123")},
					{LogStreamName: aws.String("cloud-controller-manager-123")},
					{LogStreamName: aws.String("kube-scheduler-123")},
					{LogStreamName: aws.String("unknown-123")},
				},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	logGroups := []string{"/aws/eks/my-cluster/cluster"}
	logTypes, err := client.getAvailableLogTypes(context.TODO(), logGroups)
	if err != nil {
		t.Fatalf("getAvailableLogTypes() error = %v", err)
	}

	// Check that all expected log types are present
	expectedLogTypes := []string{"api", "audit", "authenticator", "kcm", "ccm", "scheduler"}
	for _, expected := range expectedLogTypes {
		found := false
		for _, actual := range logTypes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected log type %s not found in result: %v", expected, logTypes)
		}
	}
}

func TestFilterLogGroupsByTypes(t *testing.T) {
	client := &EKSLogsClient{
		verbose: false,
	}

	logGroups := []string{"/aws/eks/my-cluster/cluster"}
	logTypes := []string{"api", "audit"}

	result := client.filterLogGroupsByTypes(context.TODO(), logGroups, logTypes)

	// Currently, this function just returns the input logGroups
	if len(result) != len(logGroups) {
		t.Errorf("filterLogGroupsByTypes() returned %d log groups, expected %d", len(result), len(logGroups))
	}

	for i, lg := range logGroups {
		if result[i] != lg {
			t.Errorf("filterLogGroupsByTypes() returned %s at index %d, expected %s", result[i], i, lg)
		}
	}
}

func TestGetLogStreamsForTypes(t *testing.T) {
	mockLogsClient := &MockCloudWatchLogsClient{
		DescribeLogStreamsFunc: func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
			return &cloudwatchlogs.DescribeLogStreamsOutput{
				LogStreams: []types.LogStream{
					{LogStreamName: aws.String("kube-apiserver-123")},
					{LogStreamName: aws.String("kube-apiserver-audit-123")},
					{LogStreamName: aws.String("authenticator-123")},
					{LogStreamName: aws.String("kube-controller-manager-123")},
					{LogStreamName: aws.String("cloud-controller-manager-123")},
					{LogStreamName: aws.String("kube-scheduler-123")},
					{LogStreamName: aws.String("unknown-123")},
				},
			}, nil
		},
	}

	client := &EKSLogsClient{
		logsClient: mockLogsClient,
		verbose:    false,
	}

	// Test case 1: Filter for specific log types
	logTypes := []string{"api", "audit"}
	streams, err := client.getLogStreamsForTypes(context.TODO(), "/aws/eks/my-cluster/cluster", logTypes)
	if err != nil {
		t.Fatalf("getLogStreamsForTypes() error = %v", err)
	}

	// Should return only streams for the specified log types
	expectedStreams := []string{"kube-apiserver-123", "kube-apiserver-audit-123"}
	if len(streams) != len(expectedStreams) {
		t.Errorf("getLogStreamsForTypes() returned %d streams, expected %d", len(streams), len(expectedStreams))
	}

	for _, expected := range expectedStreams {
		found := false
		for _, actual := range streams {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected stream %s not found in result: %v", expected, streams)
		}
	}

	// Test case 2: Error handling
	mockLogsClient.DescribeLogStreamsFunc = func(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error) {
		return nil, fmt.Errorf("test error")
	}

	_, err = client.getLogStreamsForTypes(context.TODO(), "/aws/eks/my-cluster/cluster", logTypes)
	if err == nil {
		t.Errorf("getLogStreamsForTypes() expected error, got nil")
	}
}
