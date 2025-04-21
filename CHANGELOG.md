# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added this CHANGELOG.md file to track changes between releases

## [0.4.0] - 2025-01-05

### Added

- Added conventions documentation (CONVENSIONS.md)
- Enhanced test coverage for ping command

### Changed

- Extracted and enhanced Ping Command logic for better maintainability
- Updated README to include instructions for changing module name

### Fixed

- Fixed CI workflow to require version prefix 'v' for proper Go modules versioning

## [0.3.0] - 2024-12-09

### Added

- Support for dynamic binary name via ldflags
- Shell completion generation command
- Enhanced 'ping' command with better UI and configuration management

### Changed

- Refactored internal structure for better testability and reliability
- Updated CI and release workflows
- Improved README with detailed project introduction, features, and usage instructions

### Fixed

- Added missing comments in main.go

## [0.2.0] - 2024-11-29

### Added

- Release workflow for automated builds
- Improved test coverage

### Changed

- Renamed taskfile.yml to Taskfile.yml (standard naming convention)
- Refactored CI configuration
- Enhanced logging system
- Modularized ping command structure

### Fixed

- Fixed multiple CI release workflow issues
- Fixed CI app name variable handling
- Fixed version tag interpretation in CI

## [0.1.0] - 2024-11-27

### Added

- Initial project structure with Go modules
- Command-line interface using Cobra and Viper
- Basic ping command implementation
- Configuration management
- Structured logging with Zerolog
- Basic test framework
- CI/CD setup with GitHub Actions
- Documentation in README.md

### Fixed

- ESC and CTRL-C key handling for properly exiting the program

## [0.0.1] - 2024-07-30

### Added

- Initial commit with basic project setup
- Configuration handling
- Test coverage setup
- Error logging improvements

[Unreleased]: https://github.com/peiman/ckeletin-go/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/peiman/ckeletin-go/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/peiman/ckeletin-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/peiman/ckeletin-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/peiman/ckeletin-go/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/peiman/ckeletin-go/releases/tag/v0.0.1
