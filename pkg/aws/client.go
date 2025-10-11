package aws

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	limitEnabled := limit > 0
	var totalEvents atomic.Int32
	var cancelOnce sync.Once

	// Filter log groups by log types if specified
	if len(logTypes) > 0 {
		logGroups = c.filterLogGroupsByTypes(ctx, logGroups, normalizedLogTypes)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(logGroups)) // Buffer for errors

	for _, logGroup := range logGroups {
		wg.Add(1)
		go func(lg string) {
			defer wg.Done()

			if ctx.Err() != nil {
				return
			}

			var currentLogStreamNames []string
			var getLogsErr error

			if len(logTypes) > 0 {
				currentLogStreamNames, getLogsErr = c.getLogStreamsForTypes(ctx, lg, normalizedLogTypes)
				if getLogsErr != nil {
					if ctx.Err() != nil {
						return
					}
					errChan <- fmt.Errorf("warning: failed to get log streams for log group '%s': %v", lg, getLogsErr)
					return
				}
			} else {
				currentLogStreamNames, getLogsErr = c.listLogStreamNames(ctx, lg)
				if getLogsErr != nil {
					if ctx.Err() != nil {
						return
					}
					errChan <- fmt.Errorf("warning: failed to describe log streams for log group '%s': %v", lg, getLogsErr)
					return
				}
			}

			input := &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String(lg),
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
				if c.verbose {
					fmt.Printf("Applying filter pattern: '%s' to log group: %s\n", *filterPattern, lg)
				}
			}

			// Use pagination to retrieve all log events
			var nextToken *string
			var pageCount = 0

			// Set a reasonable page size for each API call
			pageSize := int32(1000)
			if limitEnabled && limit < pageSize {
				pageSize = limit
			}

			if c.verbose {
				fmt.Printf("Retrieving logs from %s\n", lg)
				fmt.Printf("Start time: %v\n", startTime)
				fmt.Printf("End time: %v\n", endTime)
				fmt.Printf("Limit: %d\n", limit)
			}

			for {
				if ctx.Err() != nil {
					return
				}

				if limitEnabled {
					remaining := limit - totalEvents.Load()
					if remaining <= 0 {
						cancelOnce.Do(cancel)
						return
					}
					if remaining < pageSize {
						input.Limit = aws.Int32(remaining)
					} else {
						input.Limit = aws.Int32(pageSize)
					}
				} else {
					input.Limit = aws.Int32(pageSize)
				}

				pageCount++
				if nextToken != nil {
					input.NextToken = nextToken
				} else {
					input.NextToken = nil
				}

				resp, err := c.logsClient.FilterLogEvents(ctx, input)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					if c.verbose {
						fmt.Printf("Error details for log group '%s': %v\n", lg, err)
						fmt.Printf("Request parameters: StartTime=%v, EndTime=%v, FilterPattern=%v\n",
							startTime, endTime, filterPattern)
					}
					errChan <- fmt.Errorf("warning: failed to get logs from log group '%s': %v", lg, err)
					return
				}

				if c.verbose {
					fmt.Printf("Page %d, Events in response: %d, HasNextToken: %v\n",
						pageCount, len(resp.Events), resp.NextToken != nil)
				}

				for _, event := range resp.Events {
					if event.Timestamp != nil && event.LogStreamName != nil && event.Message != nil {
						var newTotal int32

						entry := log.LogEntry{
							Timestamp: time.UnixMilli(*event.Timestamp),
							Level:     log.ExtractLogLevel(*event.Message),
							Component: log.ExtractComponentFromStreamName(*event.LogStreamName),
							Message:   *event.Message,
							LogGroup:  lg,
							LogStream: *event.LogStreamName,
						}

						if limitEnabled {
							newTotal = totalEvents.Add(1)
							if newTotal > limit {
								totalEvents.Add(-1)
								cancelOnce.Do(cancel)
								return
							}
						}

						printFunc(entry) // Call the print function directly

						if limitEnabled && newTotal >= limit {
							cancelOnce.Do(cancel)
							return
						}
					}
				}

				// If no more pages, break the loop
				if resp.NextToken == nil {
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

func (c *EKSLogsClient) filterLogGroupsByTypes(ctx context.Context, logGroups []string, logTypes []string) []string {
	// For EKS, all log types are in the same log group, so no filtering needed
	return logGroups
}

func (c *EKSLogsClient) listLogStreamNames(ctx context.Context, logGroup string) ([]string, error) {
	var nextToken *string
	var streamNames []string

	for {
		resp, err := c.logsClient.DescribeLogStreams(ctx, &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(logGroup),
			Limit:        aws.Int32(50), // Maximum limit for CloudWatch Logs
			OrderBy:      cwt.OrderBy("LastEventTime"),
			Descending:   aws.Bool(true), // Get the most recent streams first
			NextToken:    nextToken,
		})
		if err != nil {
			return nil, err
		}

		for _, stream := range resp.LogStreams {
			if stream.LogStreamName != nil {
				streamNames = append(streamNames, *stream.LogStreamName)
			}
		}

		if resp.NextToken == nil {
			break
		}

		nextToken = resp.NextToken
	}

	return streamNames, nil
}

func (c *EKSLogsClient) getLogStreamsForTypes(ctx context.Context, logGroup string, logTypes []string) ([]string, error) {
	streamNames, err := c.listLogStreamNames(ctx, logGroup)
	if err != nil {
		return nil, err
	}

	var matchingStreams []string
	for _, streamName := range streamNames {
		streamLogType := log.ExtractLogTypeFromStreamName(streamName)
		if contains(logTypes, streamLogType) {
			matchingStreams = append(matchingStreams, streamName)
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
	seenEntries := make(map[string]time.Time)         // Track seen log entries to prevent duplicates

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
				if _, exists := seenEntries[entryKey]; exists {
					return
				}

				// Only print entries newer than or equal to our last timestamp
				if entry.Timestamp.Before(lastTimestamp) {
					return
				}

				log.PrintLog(entry, messageOnly, colorConfig)
				seenEntries[entryKey] = entry.Timestamp
				lastTimestamp = entry.Timestamp
			}

			mu.Lock()
			start := lastTimestamp
			mu.Unlock()

			err := c.GetLogs(ctx, clusterName, logTypes, &start, &now, filterPattern, 100, printAndTrackTimestamp)
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
			if len(seenEntries) > 2000 {
				cutoff := lastTimestamp.Add(-2 * time.Minute)
				for key, ts := range seenEntries {
					if ts.Before(cutoff) {
						delete(seenEntries, key)
					}
				}
			}
			mu.Unlock()
		}
	}
}
