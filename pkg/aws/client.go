package aws

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	cwt "github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	ekstypes "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/fatih/color"
	"github.com/kzcat/ekslogs/pkg/log"
)

// EKSAPI defines the interface for the EKS client.
type EKSAPI interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	DescribeCluster(ctx context.Context, params *eks.DescribeClusterInput, optFns ...func(*eks.Options)) (*eks.DescribeClusterOutput, error)
}

// CloudWatchLogsAPI defines the interface for the CloudWatch Logs client.
type CloudWatchLogsAPI interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeLogStreams(ctx context.Context, params *cloudwatchlogs.DescribeLogStreamsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogStreamsOutput, error)
	FilterLogEvents(ctx context.Context, params *cloudwatchlogs.FilterLogEventsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.FilterLogEventsOutput, error)
}

type EKSLogsClient struct {
	logsClient CloudWatchLogsAPI
	eksClient  EKSAPI
	region     string
	verbose    bool
}

func NewEKSLogsClient(region string, verbose bool) (*EKSLogsClient, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &EKSLogsClient{
		logsClient: cloudwatchlogs.NewFromConfig(cfg),
		eksClient:  eks.NewFromConfig(cfg),
		region:     region,
		verbose:    verbose,
	}, nil
}

func (c *EKSLogsClient) ListClusters(ctx context.Context) ([]string, error) {
	resp, err := c.eksClient.ListClusters(ctx, &eks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %w", err)
	}
	return resp.Clusters, nil
}

func (c *EKSLogsClient) GetClusterInfo(ctx context.Context, clusterName string) (*ekstypes.Cluster, error) {
	resp, err := c.eksClient.DescribeCluster(ctx, &eks.DescribeClusterInput{
		Name: aws.String(clusterName),
	})
	if err != nil {
		// If cluster not found, suggest available clusters
		if strings.Contains(err.Error(), "ResourceNotFoundException") {
			clusters, listErr := c.ListClusters(ctx)
			if listErr == nil && len(clusters) > 0 {
				return nil, fmt.Errorf("cluster '%s' not found. Available clusters: %v", clusterName, clusters)
			}
		}
		return nil, err
	}
	return resp.Cluster, nil
}

func (c *EKSLogsClient) GetLogGroups(ctx context.Context, clusterName string) ([]string, error) {
	prefix := fmt.Sprintf("/aws/eks/%s/cluster", clusterName)

	resp, err := c.logsClient.DescribeLogGroups(ctx, &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log groups: %w", err)
	}

	var logGroups []string
	for _, lg := range resp.LogGroups {
		if lg.LogGroupName != nil {
			logGroups = append(logGroups, *lg.LogGroupName)
		}
	}

	return logGroups, nil
}

func (c *EKSLogsClient) GetLogs(ctx context.Context, clusterName string, logTypes []string, startTime, endTime *time.Time, filterPattern *string, limit int32, printFunc func(log.LogEntry)) error {
	logGroups, err := c.GetLogGroups(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get log groups: %w\nPlease check your AWS credentials and permissions", err)
	}

	if len(logGroups) == 0 {
		return fmt.Errorf(`no log groups found for cluster '%s'. Please ensure:
  1. The cluster exists in the specified region
  2. Control plane logging is enabled for the cluster (check EKS console -> cluster -> Logging tab)
  3. You have the required permissions (logs:DescribeLogGroups, logs:FilterLogEvents, eks:DescribeCluster)
  4. Try using the -v flag for more detailed output`, clusterName)
	}

	var normalizedLogTypes []string
	for _, logType := range logTypes {
		normalizedLogTypes = append(normalizedLogTypes, log.NormalizeLogType(logType))
	}

	if len(logTypes) > 0 {
		availableLogTypes, err := c.getAvailableLogTypes(ctx, logGroups)
		if err != nil {
			return fmt.Errorf("failed to get available log types: %w", err)
		}

		var validLogTypes []string
		for _, logType := range normalizedLogTypes {
			if contains(availableLogTypes, logType) {
				validLogTypes = append(validLogTypes, logType)
			}
		}

		if len(validLogTypes) == 0 {
			sort.Strings(availableLogTypes)
			return fmt.Errorf(`no logs found for specified types: %v
Available log types for cluster '%s': %s
Run 'ekslogs logtypes' for more information about available log types`,
				logTypes, clusterName, log.GetLogTypeDescription(availableLogTypes))
		}

		logGroups = c.filterLogGroupsByTypes(ctx, logGroups, validLogTypes)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(logGroups)) // Buffer for errors

	for _, logGroup := range logGroups {
		wg.Add(1)
		go func(lg string) {
			defer wg.Done()

			var currentLogStreamNames []string
			var getLogsErr error

			if len(logTypes) > 0 {
				availableLogTypes, err := c.getAvailableLogTypes(ctx, []string{lg})
				if err != nil {
					errChan <- fmt.Errorf("warning: failed to get available log types for log group '%s': %v", lg, err)
					return
				}

				var validLogTypes []string
				for _, logType := range normalizedLogTypes {
					if contains(availableLogTypes, logType) {
						validLogTypes = append(validLogTypes, logType)
					}
				}

				currentLogStreamNames, getLogsErr = c.getLogStreamsForTypes(ctx, lg, validLogTypes)
				if getLogsErr != nil {
					errChan <- fmt.Errorf("warning: failed to get log streams for log group '%s': %v", lg, getLogsErr)
					return
				}
			} else {
				resp, err := c.logsClient.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
					LogGroupName: aws.String(lg),
					Limit:        aws.Int32(50), // CloudWatch Logs limit
					OrderBy:      cwt.OrderBy("LastEventTime"),
					Descending:   aws.Bool(true),
				})
				if err != nil {
					errChan <- fmt.Errorf("warning: failed to describe log streams for log group '%s': %v", lg, err)
					return
				}
				for _, stream := range resp.LogStreams {
					if stream.LogStreamName != nil {
						currentLogStreamNames = append(currentLogStreamNames, *stream.LogStreamName)
					}
				}
			}

			input := &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String(lg),
				Limit:        aws.Int32(limit),
			}
			if len(currentLogStreamNames) > 0 {
				input.LogStreamNames = currentLogStreamNames
			}

			if startTime != nil {
				input.StartTime = aws.Int64(startTime.UnixMilli())
			}
			if endTime != nil {
				input.EndTime = aws.Int64(endTime.UnixMilli())
			}
			if filterPattern != nil {
				input.FilterPattern = filterPattern
			}

			// Use pagination to retrieve all log events
			var nextToken *string
			var totalEvents int32 = 0
			var pageCount = 0

			// Set a reasonable page size for each API call
			pageSize := int32(1000)
			input.Limit = aws.Int32(pageSize)

			if c.verbose {
				fmt.Printf("Retrieving logs from %s\n", lg)
				fmt.Printf("Start time: %v\n", startTime)
				fmt.Printf("End time: %v\n", endTime)
				fmt.Printf("Limit: %d\n", limit)
			}

			for {
				pageCount++
				if nextToken != nil {
					input.NextToken = nextToken
				}

				resp, err := c.logsClient.FilterLogEvents(ctx, input)
				if err != nil {
					errChan <- fmt.Errorf("warning: failed to get logs from log group '%s': %v", lg, err)
					return
				}

				if c.verbose {
					fmt.Printf("Page %d, Events in response: %d, HasNextToken: %v\n",
						pageCount, len(resp.Events), resp.NextToken != nil)
				}

				eventsProcessed := 0

				for _, event := range resp.Events {
					if event.Timestamp != nil && event.LogStreamName != nil && event.Message != nil {
						// Check if we've reached the overall limit (limit=0 means unlimited)
						if limit > 0 && totalEvents >= limit {
							break
						}

						entry := log.LogEntry{
							Timestamp: time.UnixMilli(*event.Timestamp),
							Level:     log.ExtractLogLevel(*event.Message),
							Component: log.ExtractComponentFromStreamName(*event.LogStreamName),
							Message:   *event.Message,
							LogGroup:  lg,
							LogStream: *event.LogStreamName,
						}
						printFunc(entry) // Call the print function directly

						// Increment the counters
						totalEvents++
						eventsProcessed++
					}
				}

				// Check if we've reached the overall limit (limit=0 means unlimited)
				if limit > 0 && totalEvents >= limit {
					break
				}

				// If no more pages, break the loop
				if resp.NextToken == nil || len(resp.Events) == 0 {
					break
				}

				// Otherwise, continue with the next page
				nextToken = resp.NextToken
			}
		}(logGroup)
	}

	wg.Wait()
	close(errChan)

	var collectedErrors []error
	for err := range errChan {
		if err != nil {
			collectedErrors = append(collectedErrors, err)
		}
	}
	if len(collectedErrors) > 0 {
		return fmt.Errorf("encountered errors during log retrieval: %v", collectedErrors)
	}

	return nil
}

func (c *EKSLogsClient) getAvailableLogTypes(ctx context.Context, logGroups []string) ([]string, error) {
	logTypeSet := make(map[string]bool)

	for _, logGroup := range logGroups {
		resp, err := c.logsClient.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(logGroup),
			Limit:        aws.Int32(50), // Check first 50 streams
		})
		if err != nil {
			continue // Continue processing other log groups even if there's an error
		}

		for _, stream := range resp.LogStreams {
			if stream.LogStreamName != nil {
				logType := log.ExtractLogTypeFromStreamName(*stream.LogStreamName)
				if logType != "" {
					logTypeSet[logType] = true
				}
			}
		}
	}

	var logTypes []string
	for logType := range logTypeSet {
		logTypes = append(logTypes, logType)
	}

	return logTypes, nil
}

func (c *EKSLogsClient) filterLogGroupsByTypes(ctx context.Context, logGroups []string, logTypes []string) []string {
	return logGroups
}

func (c *EKSLogsClient) getLogStreamsForTypes(ctx context.Context, logGroup string, logTypes []string) ([]string, error) {
	resp, err := c.logsClient.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName: aws.String(logGroup),
		Limit:        aws.Int32(50), // Maximum limit for CloudWatch Logs
		OrderBy:      cwt.OrderBy("LastEventTime"),
		Descending:   aws.Bool(true), // Get the most recent streams first
	})
	if err != nil {
		return nil, err
	}

	var matchingStreams []string
	for _, stream := range resp.LogStreams {
		if stream.LogStreamName != nil {
			streamLogType := log.ExtractLogTypeFromStreamName(*stream.LogStreamName)
			if contains(logTypes, streamLogType) {
				matchingStreams = append(matchingStreams, *stream.LogStreamName)
			}
		}
	}

	return matchingStreams, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (c *EKSLogsClient) TailLogs(ctx context.Context, clusterName string, logTypes []string, filterPattern *string, interval time.Duration, messageOnly bool, colorConfig *log.ColorConfig) error {
	logGroups, err := c.GetLogGroups(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get log groups: %w\nPlease check your AWS credentials and permissions", err)
	}

	if len(logGroups) == 0 {
		return fmt.Errorf(`no log groups found for cluster '%s'. Please ensure:
  1. The cluster exists in the specified region
  2. Control plane logging is enabled for the cluster (check EKS console -> cluster -> Logging tab)
  3. You have the required permissions (logs:DescribeLogGroups, logs:FilterLogEvents, eks:DescribeCluster)
  4. Try using the -v flag for more detailed output`, clusterName)
	}

	lastTimestamp := time.Now().Add(-1 * time.Minute) // Start from 1 minute ago
	var mu sync.Mutex                                 // Mutex to protect lastTimestamp and prevent duplicate prints
	seenEntries := make(map[string]bool)              // Track seen log entries to prevent duplicates

	if c.verbose {
		fmt.Printf("Starting tail mode with interval: %v\n", interval)
		fmt.Printf("Initial start time: %v\n", lastTimestamp)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// When Ctrl+C is pressed, exit gracefully without error
			if ctx.Err() == context.Canceled {
				return nil
			}
			return ctx.Err()
		case <-ticker.C:
			now := time.Now()

			printAndTrackTimestamp := func(entry log.LogEntry) {
				mu.Lock()
				defer mu.Unlock()

				// Create a unique key for this log entry to prevent duplicates
				entryKey := fmt.Sprintf("%d-%s-%s", entry.Timestamp.UnixNano(), entry.LogStream, entry.Message)

				// Skip if we've already seen this entry
				if seenEntries[entryKey] {
					return
				}

				// Only print entries newer than our last timestamp
				if entry.Timestamp.After(lastTimestamp) {
					log.PrintLog(entry, messageOnly, colorConfig)
					seenEntries[entryKey] = true
					lastTimestamp = entry.Timestamp
				}
			}

			err := c.GetLogs(ctx, clusterName, logTypes, &lastTimestamp, &now, filterPattern, 100, printAndTrackTimestamp)
			if err != nil {
				// If context was cancelled during GetLogs execution, exit gracefully
				if ctx.Err() == context.Canceled {
					return nil
				}
				color.Red("Log retrieval error: %v", err)
				continue
			}

			// Clean up old entries from the seen map to prevent memory growth
			mu.Lock()
			if len(seenEntries) > 1000 {
				seenEntries = make(map[string]bool)
			}
			mu.Unlock()
		}
	}
}
