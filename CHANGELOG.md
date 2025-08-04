# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.1.10] - 2025-08-04

### Added
- Support for multiple `-F` flags with AND condition (all patterns must match)
- Support for multiple `-I` flags with OR condition (any pattern excludes logs)
- Enhanced filter pattern handling with string arrays instead of single strings
- New helper functions `buildCombinedFilterPattern()` and `processFilterPattern()`

### Improved
- More flexible and powerful log filtering capabilities
- Better CloudWatch Logs filter syntax generation for complex patterns
- Comprehensive test coverage for multiple filter pattern scenarios
- Updated documentation with detailed examples of multiple filter usage

### Examples
- `ekslogs my-cluster -F "error" -F "timeout"` - Include logs with both error AND timeout
- `ekslogs my-cluster -I "debug" -I "info"` - Exclude logs with debug OR info
- `ekslogs my-cluster -F "error" -F "warning" -I "debug" -I "info"` - Include errors AND warnings, exclude debug OR info

### Backward Compatibility
- Fully backward compatible with existing single filter pattern usage

## [0.1.9] - 2025-08-04

### Added
- New `-I, --ignore-filter-pattern` flag for excluding logs matching specific patterns
- Support for combining include (`-F`) and exclude (`-I`) filter patterns
- Enhanced filter pattern processing with automatic quoting for simple text searches

### Improved
- More flexible log filtering capabilities
- Better CloudWatch Logs filter pattern handling
- Comprehensive test coverage for filter pattern combinations

### Examples
- `ekslogs my-cluster -F "volume" -I "health"` - Include volume logs but exclude health checks
- `ekslogs my-cluster -F "error" -I "healthcheck"` - Include error logs but exclude healthcheck errors
- `ekslogs my-cluster -I "debug"` - Exclude all debug logs

## [0.1.7] - 2025-07-28

### Fixed
- Fixed Homebrew formula to include proper version information in build
- Homebrew users will now see correct version, commit, and build date information

## [0.1.6] - 2025-07-28

### Added
- Enhanced color output for different log types
- New `--color` option to control color output (auto, always, never)
- Automatic color detection based on terminal or pipe output
- Color support for message-only mode (`-m` option)

### Improved
- Enhanced error message highlighting in audit logs
- Bold formatting for error messages, status codes, and failure states
- Added highlighting for error reasons and status fields

### Fixed
- Fixed duplicate log output in follow mode (`-f` flag)
- Added thread-safe timestamp tracking to prevent race conditions
- Implemented deduplication mechanism to prevent duplicate log entries
- Fixed JSON parsing for audit logs to properly highlight escaped error messages

## [0.1.5] - 2025-07-20

### Added
- Homebrew installation support
- Documentation for setting up Homebrew tap

## [0.1.4] - 2025-07-20

### Changed
- Swapped `-f` and `-F` flags for better usability:
  - `-f` is now used for follow mode (tail functionality)
  - `-F` is now used for filter pattern

## [0.1.3] - 2025-07-20

### Fixed
- Further improved Ctrl+C handling to ensure help messages are never displayed when exiting
- Added top-level signal handler for more robust interrupt handling

## [0.1.2] - 2025-07-20

### Fixed
- Fixed Ctrl+C handling in follow mode to gracefully exit without showing help message
- Improved error handling for context cancellation

### Changed
- Translated all Japanese comments to English for better maintainability
- Improved code documentation and consistency

## [0.1.1] - 2025-07-20

### Fixed
- Fixed checksum issues with go install

## [0.1.0] - 2025-07-20

### Added
- Initial release of ekslogs CLI tool
- Support for retrieving various EKS Control Plane log types
- Real-time log monitoring (tail functionality)
- Time range specification (absolute and relative)
- Log filtering with pattern matching
- Colored output support
- Preset filters for common use cases
