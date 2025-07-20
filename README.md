# ekslogs

[![CI](https://github.com/kzcat/ekslogs/workflows/CI/badge.svg)](https://github.com/kzcat/ekslogs/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/kzcat/ekslogs)](https://goreportcard.com/report/github.com/kzcat/ekslogs)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A fast and intuitive CLI tool for retrieving and monitoring Amazon EKS cluster Control Plane logs.

## Features

- Retrieve various EKS Control Plane log types
- Real-time log monitoring (tail functionality)
- Time range specification (absolute and relative)
- Log filtering with pattern matching
- Colored output support

## Installation

### Using Go Install
```bash
go install github.com/kzcat/ekslogs@latest
```

## Available Log Types

Run `ekslogs logtypes` for detailed information about available log types.

| Log Type      | Description                       | Aliases                                  |
| ------------- | --------------------------------- | ---------------------------------------- |
| api           | API Server logs                   | -                                        |
| audit         | Audit logs                        | -                                        |
| authenticator | Authentication logs               | auth                                     |
| kcm           | Kube Controller Manager logs      | controller, kube-controller-manager      |
| ccm           | Cloud Controller Manager logs     | cloud, cloud-controller-manager          |
| scheduler     | Scheduler logs                    | sched                                    |

## Usage

### Basic Usage
```bash
# Get logs from the past 1 hour
ekslogs my-cluster

# Get specific log types
ekslogs my-cluster api audit

# Specify time range (absolute)
ekslogs my-cluster -s "2024-01-01T00:00:00Z" -e "2024-01-01T23:59:59Z"

# Specify time range (relative)
ekslogs my-cluster -s "-1h" -e "now"
```

### Real-time Monitoring (tail functionality)
```bash
# Monitor logs in real-time
ekslogs my-cluster -F

# Monitor only error logs
ekslogs my-cluster -F -f "ERROR"

# Monitor specific log types
ekslogs my-cluster api audit -F

# Specify update interval (default: 1 second)
ekslogs my-cluster -F --interval 10s
```

### Output Formatting and Filtering
```bash
# Output only the message part
ekslogs my-cluster -m

# Filter and process logs with grep
ekslogs my-cluster | grep "ERROR"

# Filter and process audit logs
ekslogs my-cluster audit -m | jq '[.verb, .requestURI]'
```

## Options

| Option             | Short | Description                                                     | Default      |
| ------------------ | ----- | --------------------------------------------------------------- | ------------ |
| `--region`         | `-r`  | AWS region                                                      | Auto-detect from AWS config, fallback to us-east-1 |
| `--start-time`     | `-s`  | Start time (RFC3339 format or relative: -1h, -15m, -30s, -2d)   | 1 hour ago   |
| `--end-time`       | `-e`  | End time (RFC3339 format or relative: -1h, -15m, -30s, -2d)     | Current time |
| `--filter-pattern` | `-f`  | Log filter pattern                                              | -            |
| `--limit`          | `-l`  | Maximum number of logs to retrieve                              | 1000         |
| `--message-only`   | `-m`  | Output only the log message                                     | false        |
| `--verbose`        | `-v`  | Verbose output                                                  | false        |
| `--follow`         | `-F`  | Real-time monitoring                                            | false        |
| `--interval`       | -     | Update interval for tail mode                                   | 1s           |

## Commands

| Command    | Description                                      |
| ---------- | ------------------------------------------------ |
| `logtypes` | Show detailed information about available log types |
| `presets`  | List available filter presets                    |
| `version`  | Print version information                        |
| `help`     | Help about any command                           |

## Required Permissions

- `logs:DescribeLogGroups`
- `logs:FilterLogEvents`
- `eks:DescribeCluster`
