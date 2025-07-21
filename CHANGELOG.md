# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- Enhanced color output for different log types
- New `--color` option to control color output (auto, always, never)
- Automatic color detection based on terminal or pipe output

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
