# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Added centralized configuration system:
  - Created `internal/config/registry.go` as single source of truth for all configuration
  - Implemented automatic documentation generation with `docs config` command
  - Added task commands for generating both Markdown and YAML documentation
  - Added check-defaults.sh script to detect direct viper.SetDefault() calls
- Improved environment variable handling:
  - Added EnvPrefix() function to generate consistent environment variable prefixes
  - Created ConfigPaths() helper to centralize all config file paths/names
  - Ensured binary name changes automatically update environment variable prefixes
- Enhanced YAML configuration handling:
  - Improved YAML generation with proper nested structure
  - Updated documentation to use consistent YAML format
  - Added comprehensive tests to verify YAML structure
- Enhanced developer guidance:
  - Updated testing rules with clear table-driven test examples
  - Added explicit test phase separation guidelines (setup, execution, assertion)
  - Improved workflow rules for quality checks
- Implemented Options Pattern for command configuration:
  - Added `DocsConfig` with functional options in the docs command
  - Created `WithOutputFormat` and `WithOutputFile` configuration options
  - Updated tests to use the new pattern and improve coverage
- Enhanced UI testing capabilities:
  - Added program factory pattern to `DefaultUIRunner` for better testability
  - Created `NewDefaultUIRunner` and `NewTestUIRunner` factory functions
  - Added special test mode to `RunUI` to simulate successful execution
  - Improved UI test coverage with comprehensive test cases

### Changed

- Improved test coverage:
  - Added tests for config file path handling in root.go
  - Enhanced command testing for edge cases
  - Achieved higher coverage across core components
  - Converted tests to table-driven format with phase separation
  - Added clear setup, execution, and assertion phases for all tests
  - Increased UI package coverage from 73.2% to 76.6%
  - Improved RunUI coverage from 33.3% to 53.3%
- Migrated Cursor rules architecture:
  - Transitioned from monolithic `dot.cursorrules` file to modular `.cursor/rules/` directory
  - Created separate `.mdc` files for each rule category
  - Improved organization with targeted rule files by domain
  - Made rules more discoverable with specific file names
- Refactored Viper initialization to use a centralized, idiomatic Cobra/Viper pattern with `PersistentPreRunE` and command inheritance.
- Introduced `setupCommandConfig` helper for consistent command configuration across all commands.
- Added `getConfigValue[T]` generic helper for type-safe and simplified configuration retrieval.
  - Enhanced to support string, bool, int, float64, and string slice types
  - Added comprehensive tests for all type handling scenarios
  - Improved flag overriding behavior with proper type conversions
- Removed redundant per-command Viper initialization logic.
- Updated the `ping` command and template to follow the new pattern.
- Improved documentation and code comments to reinforce centralized configuration management.
- Enhanced test coverage for the new configuration pattern and helpers.

### Security

- Added explicit permissions to GitHub Actions workflows:
  - Limited permissions to minimum required for each job
  - Added separate permission blocks for build and release jobs
  - Fixed CodeQL security warning about missing permissions
- Fixed environment variable injection vulnerability in GitHub Actions:
  - Added proper environment variable sanitization and quoting
  - Used GitHub Environment File syntax instead of direct variable setting
  - Improved handling of user-controlled input in workflow files
  - Fixed CodeQL security warning about environment injection

## [0.5.0] - 2025-04-22

### Added

- Added this CHANGELOG.md file to track changes between releases
- Added comprehensive dependency management system:
  - New Taskfile tasks: `deps:verify`, `deps:outdated`, and `deps:check`
  - Integrated dependency verification in pre-commit hooks
  - Dependency checks included in the CI pipeline
  - New section in README about dependency management
- Added project specification as `.cursorrules`:
  - Comprehensive project guidelines in LLM-friendly format
  - Documentation of commit conventions and changelog requirements
  - Explicit coding standards and implementation patterns
  - Clear collaboration and quality requirements
- Renamed `.cursorrules` to `dot.cursorrules` for better usability:
  - Added documentation in README about Cursor AI integration
  - Added `.cursorrules` to `.gitignore` for customization flexibility
  - Users can now copy the template and adapt it to their needs
- Enhanced git commit convention documentation in `dot.cursorrules`:
  - Added specific instructions for AI assistants on how to present commit messages
- Improved binary name handling:
  - Updated completion command to use binaryName variable
  - Added clear documentation about BINARY_NAME in Taskfile.yml
  - Added explanatory comments in .gitignore
  - Enhanced README with "Single Source of Truth" section

### Changed

- Updated Go version from 1.23.3 to 1.24.0
- Updated CI workflow to use Go 1.24.x
- Enhanced `task check` to include dependency verification
- Updated all outdated dependencies to their latest versions:
  - bubbletea: v1.2.4 → v1.3.4
  - lipgloss: v1.0.0 → v1.1.0
  - zerolog: v1.33.0 → v1.34.0
  - cobra: v1.8.1 → v1.9.1
  - viper: v1.19.0 → v1.20.1

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

[Unreleased]: https://github.com/peiman/ckeletin-go/compare/v0.5.0...HEAD
[0.5.0]: https://github.com/peiman/ckeletin-go/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/peiman/ckeletin-go/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/peiman/ckeletin-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/peiman/ckeletin-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/peiman/ckeletin-go/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/peiman/ckeletin-go/releases/tag/v0.0.1
