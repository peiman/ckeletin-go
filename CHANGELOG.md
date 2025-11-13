# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Test Quality Improvements**:
  - Added `testify/assert` and `testify/require` for cleaner test assertions
  - Created `internal/testutil` package with standardized platform skip helpers:
    - `SkipOnWindows(t)` - Skip tests on Windows
    - `SkipOnWindowsWithReason(t, reason)` - Skip with custom message
    - `SkipOnNonWindows(t)` - Skip on non-Windows platforms
    - `SkipOnPlatform(t, platform)` - Skip on specific platform
  - Added `getExitCode()` helper in integration tests to eliminate duplicate error handling code
  - Created comprehensive `docs/testing-guide.md`:
    - Test patterns (table-driven, SETUP-EXECUTION-ASSERTION)
    - Platform-specific testing guidelines
    - Anti-patterns to avoid
    - Integration testing patterns
    - ADR-003 compliance examples
  - **Integrated race detection into quality checks**: `task check` now runs `test:race` automatically to catch concurrency bugs
  - **Achieved 100% coverage on cmd/ package**: Added test for `runPing` wrapper function (cmd/ping_test.go:222-256)
  - **Added ADR-000 task naming validator**: Automatically enforces `action:target` pattern
    - New task: `task validate:task-naming` (integrated into `task check`)
    - Provides smart suggestions for violations (e.g., "fuzz" → "Did you mean 'test:fuzz'?")
    - Detects typos using pattern matching (e.g., "tset:race" → "Did you mean 'test:race'?")
    - Prevents naming drift and guides contributors to correct patterns
  - **Added fuzz testing for config parsing**:
    - `FuzzValidateConfigValue` in internal/config/limits_test.go - tests string limits, nested structures, type conversions
    - `FuzzValidate` in internal/config/validator/validator_test.go - end-to-end YAML parsing with malformed input
    - `FuzzFindUnknownKeys` in internal/config/validator/validator_test.go - recursive key traversal with special characters
    - New task commands: `task test:fuzz`, `task test:fuzz:config`, `task test:fuzz:validator`
    - Configurable duration: `task test:fuzz FUZZTIME=1h` (default: 10s per function)
    - Not included in `task check` (manual/exploratory testing only)

- **Automated scaffold initialization** (`task init`):
  - Single command to customize module path and binary name: `task init name=myapp module=github.com/myuser/myapp`
  - Automatically updates 40+ files including all Go imports, configs, and templates
  - Cross-platform pure Go implementation (Windows/Linux/macOS)
  - Includes formatting and validation
  - Comprehensive integration test ensures reliability
  - Eliminates 15-20 minutes of manual find/replace work
  - Example: `task init name=mycompany-cli module=github.com/mycompany/mycompany-cli`
- **ADR-009: Layered Architecture Pattern**:
  - Documents 4-layer architecture (Entry → Command → Business Logic → Infrastructure)
  - Explains dependency rules and framework independence
  - Automated enforcement via go-arch-lint
  - New task: `task validate:layering` checks layer dependencies
  - Integrated into `task check` quality pipeline
  - Configuration: `.go-arch-lint.yml` defines components and allowed dependencies
  - Prevents architectural drift through CI validation
  - Note: Adding new business logic packages requires updating `.go-arch-lint.yml`
- **ADR-010: Package Organization Strategy**:
  - Documents CLI-first package organization (no `pkg/`, all implementation in `internal/`)
  - Explains why ckeletin-go is a CLI application, not a library
  - Automated enforcement via validation script
  - New task: `task validate:package-organization` checks directory structure
  - Integrated into `task check` quality pipeline
  - Validates: No Go code in `pkg/`, only `main.go` at root, all packages in expected locations
  - Prevents accidental public API surface expansion
  - Complements ADR-009 layering rules
- **ADR-001 & ADR-002 Extensions - Implementation Patterns**:
  - Extended ADR-001 with two implementation patterns:
    - **Command Metadata Pattern**: Declarative command definitions in `internal/config/commands/`, factory functions for construction
    - **Executor Pattern**: Business logic in `internal/*/executor.go` with `Execute()` method for framework independence
  - Extended ADR-002 with Type-Safe Config Consumption pattern:
    - Commands use `getConfigValueWithFlags[T]()` helper with generated constants
    - Config passed as typed structs to executors (e.g., `ping.Config`)
    - New task: `task validate:config-consumption` enforces pattern
    - Prevents direct `viper.Get*()` calls in command files (except whitelisted: helpers.go, root.go, flags.go)
  - Removed 3 TODO markers from ARCHITECTURE.md, replaced with ADR pattern references
  - Added validation to Pattern Enforcement table

### Changed

- **Test code improvements**:
  - Replaced `panic()` with proper error handling in `TestMain` (integration tests)
  - Removed commented-out test code in `internal/logger/logger_test.go`
  - Standardized platform-specific skip patterns across test files
  - Updated `internal/config/security_test.go` and `internal/config/validator/validator_test.go` to use `testutil` helpers
- Updated Go version from 1.24 to 1.25 in go.mod and CI workflow
- Updated CI test matrix to test against Go 1.24.x and 1.25.x
- **Task naming pattern refactoring** (ADR-000 compliance):
  - Refactored all task names to follow `action:target[:subvariant]` pattern
  - Updated all documentation (ADRs, README.md, CONTRIBUTING.md, cmd/README.md) to use new task names
  - Examples: `check-defaults` → `validate:defaults`, `deps:check` → `check:deps`, `release:test` → `test:release`
  - Improved discoverability with consistent action-based grouping (all `check:*`, `validate:*`, `generate:*`, etc.)
- **README.md refactoring** (SSOT/DRY compliance):
  - Documented new `task doctor` command for environment diagnostics
  - Documented `config validate` command with exit codes and security features
  - Updated Architecture section to emphasize project vision and principles
  - Simplified task lists to reference `task --list` and Taskfile.yml (avoiding duplication)
  - Simplified configuration section to reference ADR-002 and generated docs
  - Added references to ADR-000 for task-based workflow documentation
  - Updated module path customization to reflect auto-detection
  - Applied SSOT/DRY principles throughout - README now references canonical sources instead of duplicating content

### Security

- Added explicit permissions to lint-workflows.yml following principle of least privilege (CodeQL security recommendation)

### Fixed

- **ADR-000 compliance in CI**:
  - Updated test-matrix job in `.github/workflows/ci.yml` to use `task` commands instead of direct `go` commands
  - Changed `go build -v ./...` → `task build`
  - Changed `go test -v -race ./...` → `task test:race` (Unix/macOS)
  - Changed `go test -v ./...` → `task test` (Windows)
  - Ensures CI and local development use identical commands (Single Source of Truth)

- **ADR-006 logging compliance**:
  - Added missing `component` field to 5 log statements in `internal/ui/ui.go` (lines 51, 66, 79, 87, 103)
  - All logs in internal/ui package now consistently include component field for improved filtering and debugging

- Fixed CI build failure by updating Taskfile to install golangci-lint v2 (required for .golangci.yml v2 configuration)
- Fixed Go standard library vulnerabilities (GO-2025-4007, GO-2025-4009, GO-2025-4010, GO-2025-4011) by upgrading to Go 1.25
- **Configuration discovery improvements**:
  - Config search now includes current directory (`./.ckeletin-go.yaml`) with higher priority than home directory
  - Application no longer requires `$HOME` environment variable (gracefully falls back to current directory search only)
  - Updated configuration search documentation to match actual implementation
- **Runtime log-level adjustment now functional**:
  - `logger.SetConsoleLevel()` and `logger.SetFileLevel()` now rebuild the logger stack to apply level changes
  - Added comprehensive tests verifying runtime level changes affect actual log filtering
  - Logger now stores writer references for dynamic rebuilding
- **Configuration validation**:
  - Fixed incorrect config keys in example files (`docs/examples/advanced-config.yaml`, `docs/examples/production-config.yaml`)
  - Changed `app.docs.format` → `app.docs.output_format` and `app.docs.output` → `app.docs.output_file`
  - Added validation script (`scripts/validate-example-configs.sh`) to catch config key errors
- **Documentation path sanitization**:
  - Generated documentation now uses `~` notation instead of exposing user-specific paths like `/Users/username`
  - Added `sanitizeConfigPath()` helper to replace home directories with tilde notation
  - Prevents user-specific information leakage in version-controlled documentation
- **Exit code documentation**:
  - Updated `config validate` command documentation to accurately reflect exit codes (0 for valid, 1 for errors or warnings)

## [0.7.0] - 2025-10-29

### Added

- **GoReleaser for automated releases** (see [ADR-008](docs/adr/008-release-automation-with-goreleaser.md)):
  - Multi-platform builds: Linux, macOS, Windows (amd64 and arm64)
  - Automated GitHub releases with changelog
  - Optional Homebrew tap support (configurable via HOMEBREW_TAP_OWNER env var)
  - Checksum generation (SHA256) for build verification
  - SBOM (Software Bill of Materials) in SPDX format for security compliance
  - New Taskfile tasks: `release:check`, `release:test`, `release:build`, `release:clean`
  - Snapshot builds for local testing without git tags
  - Professional release artifacts (tar.gz, zip archives)
  - Template variables for single source of truth (module path, binary name, repo URLs)
  - Zero hardcoded values - all customization via environment variables or auto-detection
- **Dual logging system** with console and file outputs:
  - Console: User-friendly, colored INFO+ messages for developers
  - File: Detailed JSON DEBUG+ logs for debugging and auditing
  - New configuration options: `--log-console-level`, `--log-file-enabled`, `--log-file-path`, `--log-file-level`, `--log-color`
  - FilteredWriter pattern for independent per-output level control
  - Secure file permissions (0600) and automatic directory creation
  - Cleanup function to ensure proper file closure
  - Only 12% performance overhead, zero allocations
- **Log rotation with lumberjack**:
  - Automatic rotation when file exceeds max size
  - Configurable backup retention (count and age)
  - Optional gzip compression of rotated logs
  - New flags: `--log-file-max-size`, `--log-file-max-backups`, `--log-file-max-age`, `--log-file-compress`
- **Log sampling for high-volume scenarios**:
  - Reduce log volume during traffic spikes
  - Configurable burst and sampling rates
  - New flags: `--log-sampling-enabled`, `--log-sampling-initial`, `--log-sampling-thereafter`
- **Runtime log level adjustment**:
  - Change console/file log levels without restarting
  - Functions: `logger.SetConsoleLevel()`, `logger.SetFileLevel()`, `logger.GetConsoleLevel()`, `logger.GetFileLevel()`
- Performance benchmarking infrastructure with `task bench` commands and baseline documentation
- Architecture Decision Records (ADRs) documenting 7 key architectural patterns
- 20+ integration tests for error scenarios and edge cases
- SessionStart hook for automatic development tool installation

### Fixed

- Type switch bug in `boolDefault()` where zero values of int64/int32/int16/int8 incorrectly evaluated to true
- Benchmark name conversion using `strconv.Itoa()` instead of `string(rune())` for readable output

### Changed

- Test coverage improved to 89.4% overall with 40+ integration tests
- **Logger initialization** now supports dual output (console + file)
- **Backward compatibility maintained**: `--log-level` flag still works as before
- Updated ADR-006 with dual logging implementation details

### Security

- Fixed vulnerability in mapstructure dependency:
  - Upgraded `github.com/go-viper/mapstructure/v2` from v2.2.1 to v2.3.0
  - Resolves GO-2025-3787: potential information leak in logs when processing malformed data
- **Log files created with secure 0600 permissions** (owner read/write only)

## [0.6.0] - 2024-06-25

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
- Added ckeletin-go specific conventions to CONVENSIONS.md:
  - Documented separation of concerns between commands and business logic
  - Clarified centralized configuration patterns
  - Added guidelines for Options Pattern and Interface Abstractions
  - Provided error handling best practices

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
- Added `getConfigValueWithFlags[T]` generic helper for type-safe and simplified configuration retrieval.
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

[Unreleased]: https://github.com/peiman/ckeletin-go/compare/v0.7.0...HEAD
[0.7.0]: https://github.com/peiman/ckeletin-go/compare/v0.6.0...v0.7.0
[0.6.0]: https://github.com/peiman/ckeletin-go/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/peiman/ckeletin-go/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/peiman/ckeletin-go/compare/v0.3.0...v0.4.0
[0.3.0]: https://github.com/peiman/ckeletin-go/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/peiman/ckeletin-go/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/peiman/ckeletin-go/compare/v0.0.1...v0.1.0
[0.0.1]: https://github.com/peiman/ckeletin-go/releases/tag/v0.0.1
