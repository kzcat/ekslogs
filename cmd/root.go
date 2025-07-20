package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/fatih/color"
	"github.com/kzcat/ekslogs/pkg/aws"
	"github.com/kzcat/ekslogs/pkg/filter"
	"github.com/kzcat/ekslogs/pkg/log"
	"github.com/spf13/cobra"
)

var (
	version        = "dev"
	commit         = "none"
	date           = "unknown"
	clusterName    string
	region         string
	logTypes       []string
	startTime      string
	endTime        string
	filterPattern  string
	presetName     string
	limit          int32
	limitSpecified bool // 制限が明示的に指定されたかどうか
	verbose        bool
	follow         bool
	interval       time.Duration

	// Execute is the function that executes the root command
	// It can be replaced in tests
	Execute = executeRoot
)

var rootCmd = &cobra.Command{
	Use:   "ekslogs <cluster-name> [log-types...]",
	Short: "A CLI tool for retrieving and monitoring EKS cluster Control Plane logs.",
	Long: `A fast and intuitive CLI tool for retrieving and monitoring Amazon EKS cluster Control Plane logs.

Features:
- Retrieve various EKS Control Plane log types
- Real-time log monitoring (tail functionality)
- Time range specification (absolute and relative)
- Log filtering with pattern matching
- Colored output support
- Preset filters for common use cases

Log types: api, audit, auth, kcm, ccm, scheduler (or sched)
If no log types are specified, all available log types will be retrieved.
Run 'ekslogs logtypes' for more detailed information about available log types.`,
	Example: `  ekslogs my-cluster                         # Get all logs from past hour
  ekslogs my-cluster api audit -F -f "error" # Monitor API/audit errors in real-time
  ekslogs my-cluster -s "-1h" -e "now"       # Get logs from specific time range
  ekslogs my-cluster -p api-errors -F        # Monitor API errors in real-time using preset`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName = args[0]
		if len(args) > 1 {
			logTypes = args[1:]
		}

		// Apply preset filter if specified
		if presetName != "" {
			preset, exists := filter.GetUnifiedPreset(presetName)
			if !exists {
				return fmt.Errorf("preset filter '%s' not found. Run 'ekslogs presets' to see available presets", presetName)
			}

			// Apply preset filter pattern if no custom filter pattern is provided
			if filterPattern == "" {
				filterPattern = preset.Pattern
				if verbose {
					if preset.Advanced {
						fmt.Printf("Using preset filter pattern: %s (type: %s)\n", filterPattern, preset.PatternType)
					} else {
						fmt.Printf("Using preset filter pattern: %s\n", filterPattern)
					}
				}
			}

			// Apply preset log types if no custom log types are provided
			if len(logTypes) == 0 {
				logTypes = preset.LogTypes
				if verbose {
					fmt.Printf("Using preset log types: %s\n", strings.Join(logTypes, ", "))
				}
			}
		}

		if region == "" {
			cfg, err := config.LoadDefaultConfig(context.TODO())
			if err == nil && cfg.Region != "" {
				region = cfg.Region
			} else {
				region = "us-east-1"
			}
		}

		client, err := aws.NewEKSLogsClient(region, verbose)
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}

		ctx := context.Background()

		clusterInfo, err := client.GetClusterInfo(ctx, clusterName)
		if err != nil {
			return fmt.Errorf("failed to get cluster info: %w", err)
		}

		messageOnly, err := cmd.Flags().GetBool("message-only")
		if err != nil {
			return err
		}

		if verbose {
			color.Cyan("=== EKS Control Plane Logs CLI ===")
			color.Cyan("Cluster: %s", clusterName)
			color.Cyan("Region: %s", region)
			if len(logTypes) > 0 {
				color.Cyan("Log Types: %v", logTypes)
			} else {
				color.Cyan("Log Types: all")
			}
			color.Cyan("Cluster Status: %s", string(clusterInfo.Status))
			color.Green("Cluster found")
		}

		var fp *string
		if filterPattern != "" {
			fp = &filterPattern
		}

		printLogEntry := func(entry log.LogEntry) {
			log.PrintLog(entry, messageOnly)
		}

		if follow {
			ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
			defer cancel()

			return client.TailLogs(ctx, clusterName, logTypes, fp, interval, messageOnly)
		}

		var startT, endT *time.Time

		if startTime != "" {
			t, err := log.ParseTimeString(startTime)
			if err != nil {
				return fmt.Errorf("failed to parse start time: %w", err)
			}
			startT = t
		}

		if endTime != "" {
			t, err := log.ParseTimeString(endTime)
			if err != nil {
				return fmt.Errorf("failed to parse end time: %w", err)
			}
			endT = t
		}

		if startT == nil && endT == nil {
			now := time.Now()
			oneHourAgo := now.Add(-1 * time.Hour)
			startT = &oneHourAgo
			endT = &now
		}

		// ユーザーが明示的に制限を指定した場合のみ制限を適用
		var effectiveLimit int32
		if limitSpecified {
			effectiveLimit = limit
		} else {
			effectiveLimit = 0 // 0は無制限を意味する
		}

		err = client.GetLogs(ctx, clusterName, logTypes, startT, endT, fp, effectiveLimit, printLogEntry)
		if err != nil {
			return err
		}

		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Print version information about the ekslogs CLI tool, including version number, commit hash, and build date.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ekslogs version %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built at: %s\n", date)
	},
}

var logTypesCmd = &cobra.Command{
	Use:   "logtypes",
	Short: "Show detailed information about available log types",
	Long: `Show detailed information about available log types for EKS Control Plane logs.

Each log type corresponds to a specific component of the EKS Control Plane.
You can specify one or more log types when retrieving logs to focus on specific components.

Examples:
  ekslogs my-cluster api audit     # Get logs from API server and audit logs
  ekslogs my-cluster auth          # Get authentication logs
  ekslogs my-cluster scheduler     # Get scheduler logs`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Available log types for EKS Control Plane logs:")
		fmt.Println()
		fmt.Println("  api           - API Server logs (kube-apiserver)")
		fmt.Println("  audit         - Audit logs (kube-apiserver-audit)")
		fmt.Println("  authenticator - Authentication logs (aws-iam-authenticator)")
		fmt.Println("                  Alias: auth")
		fmt.Println("  kcm           - Kube Controller Manager logs (kube-controller-manager)")
		fmt.Println("                  Aliases: controller, kube-controller-manager")
		fmt.Println("  ccm           - Cloud Controller Manager logs (cloud-controller-manager)")
		fmt.Println("                  Aliases: cloud, cloud-controller-manager")
		fmt.Println("  scheduler     - Scheduler logs (kube-scheduler)")
		fmt.Println("                  Alias: sched")
		fmt.Println()
		fmt.Println("If no log types are specified, all available log types will be retrieved.")
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(logTypesCmd)

	rootCmd.Flags().StringVarP(&region, "region", "r", "", "AWS region")
	rootCmd.Flags().StringVarP(&startTime, "start-time", "s", "", "Start time (RFC3339 format or relative: -1h, -15m, -30s, -2d)")
	rootCmd.Flags().StringVarP(&endTime, "end-time", "e", "", "End time (RFC3339 format or relative: -1h, -15m, -30s, -2d)")
	rootCmd.Flags().StringVarP(&filterPattern, "filter-pattern", "f", "", "Log filter pattern")
	rootCmd.Flags().StringVarP(&presetName, "preset", "p", "", "Use filter preset (run 'ekslogs presets' to list available presets)")
	rootCmd.Flags().Int32VarP(&limit, "limit", "l", 1000, "Maximum number of logs to retrieve")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVarP(&follow, "follow", "F", false, "Continuously monitor logs (tail mode)")
	rootCmd.Flags().DurationVar(&interval, "interval", 1*time.Second, "Update interval for tail mode")
	rootCmd.Flags().BoolP("message-only", "m", false, "Output only the log message")

	// フラグが指定されたかどうかを確認するためのPreRunを追加
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		limitSpecified = cmd.Flags().Changed("limit")
	}
}

func executeRoot() {
	if err := rootCmd.Execute(); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
}
