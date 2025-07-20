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
- Preset filters for common use cases

## Installation

### Using Go Install
```bash
go install github.com/kzcat/ekslogs@latest
```

### From Source
```bash
git clone https://github.com/kzcat/ekslogs.git
cd ekslogs
go build
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

### Using Filter Presets

The tool comes with predefined filter presets for common use cases:

```bash
# List available presets
ekslogs presets

# Show advanced presets
ekslogs presets --advanced

# Use a preset filter
ekslogs my-cluster -p api-errors

# Monitor API errors in real-time
ekslogs my-cluster -p api-errors -F
```

### Common Filter Preset Examples

| Preset                   | Description                                   | Log Types                |
| ------------------------ | --------------------------------------------- | ------------------------ |
| api-errors               | API server errors                             | api                      |
| audit-privileged         | Privileged operations in audit logs           | audit                    |
| auth-failures            | Authentication failures                       | authenticator, api       |
| network-issues           | Network related issues                        | api, kcm, ccm            |
| scheduler-issues         | Scheduler issues                              | scheduler                |
| critical-api-errors      | Critical API server errors (excluding warnings)| api                     |
| memory-pressure          | Memory pressure and OOM events                | api, kcm                 |
| network-timeouts         | Network timeout issues                        | api, kcm, ccm            |

### Output Formatting and Filtering
```bash
# Output only the message part
ekslogs my-cluster -m

# Filter and process logs with grep
ekslogs my-cluster | grep "ERROR"

# Filter and process audit logs
ekslogs my-cluster audit -m | jq '[.verb, .requestURI]'
```

## Advanced Usage Examples

### Monitoring Authentication Issues

```bash
# Monitor authentication failures in real-time
ekslogs my-cluster -p auth-failures -F

# Monitor advanced authentication issues
ekslogs my-cluster -p auth-issues-adv -F
```

### Investigating Network Problems

```bash
# Check for network issues in the last 3 hours
ekslogs my-cluster -p network-issues -s "-3h"

# Monitor network timeouts in real-time
ekslogs my-cluster -p network-timeouts -F
```

### Debugging Scheduler Problems

```bash
# Check for pod scheduling failures
ekslogs my-cluster -p pod-scheduling-failures

# Monitor scheduler issues in real-time
ekslogs my-cluster -p scheduler-issues -F
```

### Security Auditing

```bash
# Check for privileged admin actions
ekslogs my-cluster -p privileged-admin-actions

# Monitor security events in real-time
ekslogs my-cluster -p security-events -F
```

## Options

| Option             | Short | Description                                                     | Default      |
| ------------------ | ----- | --------------------------------------------------------------- | ------------ |
| `--region`         | `-r`  | AWS region                                                      | Auto-detect from AWS config, fallback to us-east-1 |
| `--start-time`     | `-s`  | Start time (RFC3339 format or relative: -1h, -15m, -30s, -2d)   | 1 hour ago   |
| `--end-time`       | `-e`  | End time (RFC3339 format or relative: -1h, -15m, -30s, -2d)     | Current time |
| `--filter-pattern` | `-f`  | Log filter pattern                                              | -            |
| `--preset`         | `-p`  | Use filter preset (run 'ekslogs presets' to list available presets) | -         |
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

## Troubleshooting

### No logs found

If you receive a message that no logs were found, check the following:

1. Ensure that Control Plane logging is enabled for your EKS cluster
2. Verify that you have the required IAM permissions
3. Check that the specified time range contains logs
4. Try using the `-v` flag for verbose output to see more details

### Authentication errors

If you encounter authentication errors:

1. Verify that your AWS credentials are properly configured
2. Check that your IAM role or user has the required permissions
3. Try specifying the region explicitly with the `-r` flag

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Development Setup

1. Clone the repository
   ```bash
   git clone https://github.com/kzcat/ekslogs.git
   cd ekslogs
   ```

2. Install required tools
   ```bash
   make install-tools
   ```

3. Install pre-commit hooks
   ```bash
   make install-hooks
   ```

### Development Workflow

The project includes a Makefile with useful commands:

```bash
# Build the binary
make build

# Run tests
make test

# Generate test coverage report
make coverage

# Format code
make fmt

# Run linters
make lint

# Clean up build artifacts
make clean

# Show all available commands
make help
```

### Pre-commit Hooks

The project uses pre-commit hooks to ensure code quality. The hooks run:

1. `gofmt` to format code
2. `go vet` for static analysis
3. `go test` to run tests
4. `golangci-lint` for comprehensive linting (if installed)

These hooks run automatically when you commit changes.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
