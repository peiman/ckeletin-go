<div align="center">

![ckeletin-go](logo/ckeletin-go-banner.png)

**A professional Golang CLI scaffold for building beautiful, robust, and modular command-line applications.**

[![Build Status](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/peiman/ckeletin-go)](https://codecov.io/gh/peiman/ckeletin-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peiman/ckeletin-go)](https://goreportcard.com/report/github.com/peiman/ckeletin-go)
[![Version](https://img.shields.io/github/v/release/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/peiman/ckeletin-go.svg)](https://pkg.go.dev/github.com/peiman/ckeletin-go)
[![License](https://img.shields.io/github/license/peiman/ckeletin-go)](LICENSE)
[![CodeQL](https://github.com/peiman/ckeletin-go/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/peiman/ckeletin-go/security/code-scanning)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](https://go.dev)

</div>

---

## TLDR;

**Your boss needs a CLI tool by next sprint. You've never built one.**

ckeletin-go gives you production-ready infrastructure so you can focus on YOUR feature, not learning Cobra.

- **Read the code in 5 minutes** - Ultra-thin commands (~20 lines each). No framework magic to decode.
- **Ship with >80% test coverage** - Hundreds of real tests. Integration + unit. You won't break production.
- **One command setup** - `task init name=myapp module=...` updates 40+ files. Start coding in 2 minutes.
- **Learn as you build** - ADRs explain every decision. Level up while shipping.
- **Production-ready logging** - Debug issues without SSH-ing into servers.

**Quickstart:**
```bash
git clone https://github.com/peiman/ckeletin-go.git && cd ckeletin-go
task setup && task init name=myapp module=github.com/you/myapp
task build && ./myapp ping  # All tests passed ✓
```

You just built a CLI with better architecture than most production codebases.
Now add your feature and look like a senior engineer.

**Bonus:** Automatic GPL/AGPL blocking prevents license contamination (saved companies $100K+ in legal audits).

---

## Table of Contents

- [ckeletin-go](#ckeletin-go)
  - [TLDR;](#tldr)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Key Highlights](#key-highlights)
  - [Quick Start](#quick-start)
  - [Features](#features)
  - [Architecture](#architecture)
  - [Getting Started](#getting-started)
    - [Prerequisites](#prerequisites)
    - [Installation](#installation)
    - [Using the Scaffold](#using-the-scaffold)
    - [Important: Single Source of Truth for Names](#important-single-source-of-truth-for-names)
    - [Customizing the Module Path](#customizing-the-module-path)
      - [Steps to Update the Module Path](#steps-to-update-the-module-path)
  - [Configuration](#configuration)
    - [Configuration Management](#configuration-management)
    - [Adding New Configuration Options](#adding-new-configuration-options)
    - [Automatic Documentation Generation](#automatic-documentation-generation)
    - [Configuration File](#configuration-file)
    - [Environment Variables](#environment-variables)
    - [Command-Line Flags](#command-line-flags)
    - [Configuration Precedence](#configuration-precedence)
  - [Dependency Management](#dependency-management)
    - [Available Tasks](#available-tasks)
    - [Automated Checks](#automated-checks)
    - [Best Practices](#best-practices)
  - [Commands](#commands)
    - [`ping` Command](#ping-command)
      - [Usage](#usage)
      - [Flags](#flags)
      - [Examples](#examples)
  - [Development Workflow](#development-workflow)
    - [Essential Task Commands](#essential-task-commands)
    - [Development Tools \& Reproducibility](#development-tools--reproducibility)
    - [Pre-Commit Hooks with Lefthook](#pre-commit-hooks-with-lefthook)
    - [Continuous Integration](#continuous-integration)
    - [Creating Releases](#creating-releases)
  - [Customization](#customization)
    - [Changing the Program Name](#changing-the-program-name)
    - [Adding New Commands](#adding-new-commands)
    - [Command Implementation Pattern](#command-implementation-pattern)
    - [Options Pattern for Command Configuration](#options-pattern-for-command-configuration)
    - [Modifying Configurations](#modifying-configurations)
    - [Customizing the UI](#customizing-the-ui)
    - [Cursor AI Integration](#cursor-ai-integration)
  - [Tooling Best Practices](#tooling-best-practices)
  - [Contributing](#contributing)
  - [License](#license)
  - [Additional Notes](#additional-notes)
  - [Note](#note)

---

## Introduction

**ckeletin-go** is a Golang scaffold project designed to help developers create professional, robust, and beautiful CLI applications. Inspired by the idea of a "skeleton," **ckeletin** provides a strong foundation on which you can build your own tooling, utilities, and interactive experiences.

This scaffold integrates essential libraries and tools that follow best practices:

- [Cobra](https://github.com/spf13/cobra) for building flexible, modular CLI commands.
- [Viper](https://github.com/spf13/viper) for configuration management via files, environment variables, and flags.
- [Zerolog](https://github.com/rs/zerolog) for structured, leveled logging.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) for beautiful, interactive terminal UIs.
- Task automation with [Task](https://taskfile.dev/) and pre-commit hooks with [Lefthook](https://github.com/evilmartians/lefthook).
- High test coverage, code quality checks, and CI/CD pipelines using GitHub Actions and CodeQL.

Each command manages its own configuration and defaults, promoting modularity and ease of maintenance.

---

## Key Highlights

- **Readable Architecture**: Ultra-thin commands (~20 lines each) mean you can understand and modify code in minutes, not days
- **Production-Ready Testing**: >80% test coverage enforced. Integration + unit tests. CI fails if quality drops
- **One-Command Customization**: `task init` updates 40+ files automatically. No find/replace hell
- **Learn While You Build**: Comprehensive ADRs document every architectural decision. Level up your Go skills
- **Enterprise License Compliance**: Automated GPL/AGPL blocking prevents legal contamination. Build proprietary products worry-free
- **Reproducible Builds**: Pinned tool versions ensure identical results across dev/CI environments
- **Task-Based Workflow**: Single source of truth for all commands. No guessing what to run

---

## Quick Start

1. **Clone the repository**:

   ```bash
   git clone https://github.com/peiman/ckeletin-go.git
   cd ckeletin-go
   ```

2. **Set up development tools**:

   ```bash
   task setup
   ```

   Installs necessary tools and pre-commit hooks.

3. **Build and run the sample command**:

   ```bash
   task build
   ./ckeletin-go ping
   ```

   You'll see "Pong" printed—congratulations, you're running the scaffold!

---

## Features

- **Modular Command Structure**: Add, remove, or update commands without breaking the rest of the application.
- **Layered Architecture**: Enforced 4-layer architecture (Entry → Command → Business Logic → Infrastructure) with automated validation to prevent architectural drift.
- **Structured Logging**: Use Zerolog to create efficient, leveled logs. Perfect for debugging, auditing, and production use.
- **Bubble Tea UI**: Optional, interactive UI for advanced terminal applications.
- **Single-Source Configuration**: Set defaults in config files, override with env vars, and fine-tune with flags.
- **Enterprise License Compliance**:
  - **Dual-tool checking**: go-licenses (source analysis, ~2-5s) + lichen (binary analysis, ~10-15s) for defense in depth
  - **Automatic blocking**: Prevents GPL/AGPL contamination that could force your code open-source
  - **Commercial-friendly**: Default permissive-only policy (MIT, Apache-2.0, BSD) allows proprietary products
  - **CI integration**: Automated checks on every PR. Violations block merge
  - **Reproducible compliance**: Pinned tool versions ensure consistent license detection across environments
  - **Customizable policy**: Override allowed licenses via environment or `.lichen.yaml`
  - See [ADR-011](docs/adr/011-license-compliance.md) and [docs/licenses.md](docs/licenses.md) for details
- **Task Automation**: One Taskfile to define all build, test, and lint tasks.
- **High Test Coverage & Quality Checks**: >80% coverage enforced by CI. Hundreds of real tests ensure a robust codebase that meets production standards.

---

## Architecture

**ckeletin-go follows a principled architecture emphasizing:**
- **Layered architecture** - 4-layer pattern (Entry → Command → Business Logic → Infrastructure) with automated enforcement ([ADR-009](docs/adr/009-layered-architecture-pattern.md))
- **Separation of concerns** - CLI wiring separate from business logic
- **Single source of truth** - One canonical place for all configuration, tasks, and patterns
- **Type safety** - Auto-generated constants, strong typing throughout
- **Testability** - Dependency injection, high coverage standards
- **Automation** - Task-based workflow, auto-generated docs, enforced patterns

All architectural decisions are documented in **[Architecture Decision Records (ADRs)](docs/adr/)** with detailed rationale and implementation guidance.

**Key architectural patterns:**
- Ultra-thin commands (~20-30 lines) delegate to business logic ([ADR-001](docs/adr/001-ultra-thin-command-pattern.md))
- Centralized configuration registry with type-safe constants ([ADR-002](docs/adr/002-centralized-configuration-registry.md))
- Dependency injection over mocking for testability ([ADR-003](docs/adr/003-dependency-injection-over-mocking.md))
- Dual-tool license compliance checking (source + binary) ([ADR-011](docs/adr/011-license-compliance.md))
- Automated validation prevents architectural drift (`task validate:layering`)

---

## Getting Started

### Prerequisites

- **Go**: Version specified in `go.mod` (currently maintained at the latest stable release).
- **Task**: Install from [taskfile.dev](https://taskfile.dev/#/installation).
- **Git**: For version control.

### Installation

#### Option 1: Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/peiman/ckeletin-go/releases).

```bash
# Example for Linux amd64
curl -L https://github.com/peiman/ckeletin-go/releases/latest/download/ckeletin-go_linux_amd64.tar.gz | tar xz
sudo mv ckeletin-go /usr/local/bin/
```

#### Option 2: Homebrew (macOS/Linux) - If Enabled

If the maintainer has configured a Homebrew tap:

```bash
brew install peiman/tap/ckeletin-go
```

**Note**: Homebrew tap is optional and must be explicitly enabled by the project maintainer.

#### Option 3: Build from Source

```bash
git clone https://github.com/peiman/ckeletin-go.git
cd ckeletin-go
task setup
task build
```

### Using the Scaffold

1. **Initialize with your project details:**
   ```bash
   task init name=myapp module=github.com/myuser/myapp
   ```

   This single command automatically:
   - Updates the module path in `go.mod`
   - Replaces all import statements (40+ files)
   - Updates binary name in `Taskfile.yml` and `.goreleaser.yml`
   - Formats all code
   - Runs `go mod tidy`

2. **Build and test your customized CLI:**
   ```bash
   task check    # Run quality checks
   task build    # Build your binary
   ./myapp --version
   ```

That's it! Your scaffold is ready to customize.

### Important: Single Source of Truth for Names

This project uses a "single source of truth" approach for configuration:

1. **Binary Name**: Defined only in `Taskfile.yml` as `BINARY_NAME`. This is propagated through the codebase via build flags and the `binaryName` variable in `cmd/root.go`.

2. **Module Path**: Defined only in `go.mod` and referenced in `Taskfile.yml` as `MODULE_PATH`.

When customizing this project:

- Change `BINARY_NAME` in `Taskfile.yml` to your desired binary name
- Change the module path in `go.mod` to your own repository path
- Run `task build` to apply these changes throughout the codebase

This design ensures you don't need to search and replace names across multiple files.

### Customizing the Module Path

**Quick Method:** Use `task init` (see "Using the Scaffold" above).

**What It Does:** The `task init` command automatically updates:
- Module declaration in `go.mod`
- All import statements across 40+ Go files
- Binary name in `Taskfile.yml` and `.goreleaser.yml`
- Template files with the new module path

**Manual Method (Advanced):** If you need to understand what's happening or customize the process:

1. **Update `go.mod`**: Change the module path to your repository:
   ```go
   module github.com/yourusername/your-repo-name
   ```

2. **Update imports**: Search and replace `github.com/peiman/ckeletin-go` with your module path in all `.go` files

3. **Update configurations**: Change `BINARY_NAME` in `Taskfile.yml` and `project_name` in `.goreleaser.yml`

4. **Clean up**:
   ```bash
   go mod tidy
   task format
   task check
   ```

**Note:** `Taskfile.yml` automatically detects your module path using `go list -m` - it adapts to whatever is in `go.mod`.

---

## Configuration

**ckeletin-go** uses Viper for flexible configuration and also implements a centralized configuration system with a single source of truth:

### Configuration Management

Configuration options are defined in command files and self-register with the central registry.

**To add new configuration options:** See examples in `cmd/ping.go` or `cmd/docs.go`, or read [ADR-002](docs/adr/002-centralized-configuration-registry.md) for the complete pattern.

**For all available configuration options:** Run `task generate:docs:config` to generate [docs/configuration.md](docs/configuration.md).

**Rule:** Never use `viper.SetDefault()` directly. The `task validate:defaults` check enforces this.

### Automatic Documentation Generation

Generate comprehensive configuration documentation with:

```bash
task generate:docs:config
```

This creates a Markdown file at `docs/configuration.md` with:

- All available configuration options
- Default values and types
- Environment variable names
- Full descriptions

For a configuration template, run:

```bash
task generate:config:template
```

### Configuration File

Default config file: `$HOME/.ckeletin-go.yaml` (or `myapp.yaml` if renamed).

Example:

```yaml
app:
  log_level: "debug"
  ping:
    output_message: "Hello World!"
    output_color: "green"
    ui: true
```

### Environment Variables

Override any config via environment variables with automatic prefix based on binary name:

```bash
# For binary name "ckeletin-go":
export CKELETIN_GO_APP_LOG_LEVEL="debug"
export CKELETIN_GO_APP_PING_OUTPUT_MESSAGE="Hello, World!"
export CKELETIN_GO_APP_PING_UI=true

# If you renamed to "myapp":
export MYAPP_APP_LOG_LEVEL="debug"
```

### Command-Line Flags

Override at runtime:

```bash
./myapp ping --message "Hi there!" --color yellow --ui
```

### Configuration Precedence

Configuration values are resolved in this order:

1. Command-line flags (highest priority)
2. Environment variables
3. Configuration file
4. Default values (lowest priority)

---

## Dependency Management

**ckeletin-go** includes robust dependency management tools to ensure your application remains secure and up-to-date:

### Available Tasks

- `task check:deps` - Run all dependency checks
- `task check:deps:verify` - Verify dependency integrity
- `task check:deps:outdated` - Check for outdated dependencies

**See** `task --list` for complete task reference.

### Automated Checks

Dependency verification is automatically included in:

- Pre-commit hooks via Lefthook (prevents commits with corrupted dependencies)
- CI workflow via GitHub Actions (as part of `task check`)
- The comprehensive quality check command: `task check`

### Best Practices

1. Run `task check:deps` before starting a new feature
2. Update dependencies incrementally with `go get -u <package>` followed by `task tidy`
3. Always run tests after dependency updates
4. Document significant dependency changes in commit messages

---

## Commands

### `ping` Command

A sample command showing how to use Cobra, Viper, Zerolog, and Bubble Tea together.

#### Usage

```bash
./myapp ping [flags]
```

#### Flags

- `--message`: Override output message.
- `--color`: Override output color.
- `--ui`: Enable Bubble Tea UI.

#### Examples

```bash
./myapp ping
./myapp ping --message "Hello!" --color cyan
./myapp ping --ui
```

### `doctor` Command

Check your development environment and verify all required tools are installed.

#### Usage

```bash
task doctor
```

#### Checks

- Go installation and version
- Task installation and version
- All development tools (goimports, golangci-lint, gotestsum, govulncheck, lefthook, go-mod-outdated)

**If tools are missing:** Run `task setup` to install them.

### `config validate` Command

Validate configuration files for correctness, security, and completeness.

#### Usage

```bash
# Validate default config
./myapp config validate

# Validate specific file
./myapp config validate --file /path/to/config.yaml
```

#### Exit Codes

- `0` - Configuration is valid
- `1` - Configuration has errors or warnings

**Security features:** File size limits, permission checks, value validation.

---

## Development Workflow

This project follows a **task-based single source of truth** pattern where all development commands are defined in `Taskfile.yml` and used identically across local development, pre-commit hooks, and CI. See [ADR-000](docs/adr/000-task-based-single-source-of-truth.md) for complete documentation.

Tasks follow the `action:target[:subvariant]` naming pattern for consistency and discoverability.

### Essential Task Commands

This project uses [Task](https://taskfile.dev/) for all development commands.

**Core workflow:**
```bash
task doctor    # Check your development environment
task check     # Run all quality checks (mandatory before commits)
task format    # Format all Go code
task test      # Run tests with coverage
task build     # Build the binary
```

**For the complete task list:** Run `task --list` or see [Taskfile.yml](Taskfile.yml).
**For task naming pattern:** See [ADR-000](docs/adr/000-task-based-single-source-of-truth.md).

### Development Tools & Reproducibility

This project pins all tool versions for reproducible builds across environments:

**Tool Version Management:**
```bash
task doctor                    # Check environment and tool versions
task check:tools:installed     # Fast existence check (dev workflow)
task check:tools:version       # Strict version verification (CI workflow)
task check:tools:updates       # Discover available tool updates
```

**Pinned Tools** (defined in `Taskfile.yml`):
- **goimports** v0.28.0 - Code formatting
- **golangci-lint** v2.3.0 - Comprehensive linting
- **gotestsum** latest - Test runner (Go 1.25.x compatible)
- **govulncheck** latest - Security vulnerability scanning (Go 1.25.x compatible)
- **lefthook** v1.7.18 - Git hooks management
- **go-licenses** v2.0.1 - Source-based license compliance (CRITICAL: pinned for reproducibility)
- **lichen** v0.1.7 - Binary-based license compliance (CRITICAL: pinned for reproducibility)

**License Compliance Tasks:**
```bash
# Fast check during development (~2-5s)
task check:license:source

# Accurate check before releases (~10-15s)
task check:license:binary

# Both checks (CI workflow, ~15-20s)
task check:license

# Generate license artifacts
task generate:license:report   # CSV report
task generate:license:files    # Third-party license files
task generate:attribution       # NOTICE file
task generate:license           # All artifacts
```

**Why Pinned Versions Matter:**
- **Consistency**: Same lint results on dev machines and CI
- **Reproducibility**: License compliance checks produce identical results
- **Stability**: Avoid breaking changes from tool updates
- **Debugging**: Pinned versions make issues reproducible

**Dev vs CI Separation:**
- **Local**: `task check` runs fast existence checks for rapid feedback
- **CI**: `task check:ci` enforces strict version verification for release quality

### Pre-Commit Hooks with Lefthook

`task setup` installs hooks that run `format`, `lint`, `test` on commit, ensuring code quality before changes land in the repository.

### Continuous Integration

GitHub Actions runs `task check` on each commit or pull request, maintaining code standards and reliability.

### Creating Releases

This project uses [GoReleaser](https://goreleaser.com/) for automated, professional releases.

#### Release Process

1. **Prepare the release**:
   ```bash
   # Ensure all changes are committed
   git status

   # Run quality checks
   task check

   # Update CHANGELOG.md with release notes
   vim CHANGELOG.md
   ```

2. **Create and push a semantic version tag**:
   ```bash
   # Create annotated tag (e.g., v1.0.0, v1.1.0, v1.0.1)
   git tag -a v1.0.0 -m "Release v1.0.0"

   # Push the tag
   git push origin v1.0.0
   ```

3. **CI automatically**:
   - Validates the semantic version tag
   - Runs quality checks
   - Builds binaries for multiple platforms (Linux, macOS, Windows)
   - Generates checksums and SBOM
   - Creates GitHub release with changelog
   - Updates Homebrew tap

#### Local Release Testing

Test the release process locally before pushing a tag:

```bash
# Check if GoReleaser is installed
task check:release

# Build snapshot release (no tag required)
task test:release

# Test the generated binaries
./dist/ckeletin-go_linux_amd64_v1/ckeletin-go --version
```

#### Supported Platforms

- **Linux**: amd64, arm64
- **macOS**: amd64 (Intel), arm64 (Apple Silicon)
- **Windows**: amd64

#### Release Artifacts

Each release includes:
- Platform-specific binaries (archives: tar.gz, zip)
- `checksums.txt` for verification
- SBOM (Software Bill of Materials) in SPDX format
- Automated changelog from git commits

#### Enabling Homebrew Tap (Optional)

Homebrew tap is disabled by default. To enable it:

1. Create a `homebrew-tap` repository in your GitHub account
2. Add GitHub secrets to your repository:
   - `HOMEBREW_TAP_OWNER` - Your GitHub username
   - `HOMEBREW_TAP_GITHUB_TOKEN` - Token with repo access
3. Update CI workflow to pass the environment variable to GoReleaser

Once enabled, releases will automatically update your Homebrew tap.

For more details, see [ADR-008: Release Automation](docs/adr/008-release-automation-with-goreleaser.md).

---

## Customization

**In Short**:

- Change `BINARY_NAME` in `Taskfile.yml` to rename your CLI.
- Add commands using `cobra-cli`: `cobra-cli add hello`.
- Adjust configs in Viper.
- Enhance UI in `internal/ui/`.

### Changing the Program Name

In `Taskfile.yml`:

```yaml
vars:
  BINARY_NAME: myapp
```

Then:

```bash
task build
./myapp ping
```

### Adding New Commands

Add a new command following Cobra best practices: each command in its own file, cleanly separated and easily testable.

For faster development, copy and modify the template file:

```bash
cp cmd/template_command.go.example cmd/hello.go
```

Then edit the file to implement your command and its configuration options. Flags and help will be auto-generated from the options metadata.

You can also scaffold both files with Task:

```bash
task generate:command name=hello
```

### Command Implementation Pattern

The project uses an idiomatic Cobra/Viper pattern with command inheritance and self-registered configuration providers:

1. The root command's `PersistentPreRunE` initializes global configuration
2. Child commands inherit this behavior through Cobra's command chain
3. Command-specific configuration is declared in the command file (e.g., `cmd/<command>.go`) and self-registered via `config.RegisterOptionsProvider`
4. Flags and help text are auto-registered from the registry using `RegisterFlagsForPrefixWithOverrides(cmd, "app.<command>.", overrides)`

Benefits of this pattern:

- Reduces duplication across commands
- Ensures consistent configuration handling
- Maintains command independence
- Simplifies adding new commands

When implementing a new command:

1. In `cmd/<command>.go`, define `<Command>Options() []config.ConfigOption` and call `config.RegisterOptionsProvider(<Command>Options)` in `init()`
2. Still in `cmd/<command>.go`, call `RegisterFlagsForPrefixWithOverrides(cmd, "app.<command>.", overrides)` to create flags and bind them to Viper
3. Use `getConfigValueWithFlags[T](cmd, flagName, viperKey)` in your config constructor to honor flag precedence over config

### Options Pattern for Command Configuration

The project implements the Options Pattern for command configuration, with a single source of truth for defaults and docs in the options registry. Benefits:

1. **Testability**: Commands can be easily tested with different configurations
2. **Modularity**: Configuration is encapsulated in dedicated structs
3. **Type Safety**: Using strong typing and generics for configuration values
4. **Default Handling**: Consistent handling of default values from the registry
5. **Readability**: Clear separation of configuration from command logic

To implement the Options Pattern for a new command:

1. **Create a configuration struct**:

   ```go
   type CommandConfig struct {
     Option1 string
     Option2 bool
     // Add other options as needed
   }
   ```

2. **Define functional options**:

   ```go
   type CommandOption func(*CommandConfig)

   func WithOption1(value string) CommandOption {
     return func(cfg *CommandConfig) { cfg.Option1 = value }
   }

   func WithOption2(value bool) CommandOption {
     return func(cfg *CommandConfig) { cfg.Option2 = value }
   }
   ```

3. **Create a constructor**:

   ```go
   func NewCommandConfig(cmd *cobra.Command, opts ...CommandOption) CommandConfig {
     cfg := CommandConfig{
       Option1: getConfigValueWithFlags[string](cmd, "option1", "app.command.option1"),
       Option2: getConfigValueWithFlags[bool](cmd, "option2", "app.command.option2"),
     }
     for _, opt := range opts {
       opt(&cfg)
     }
     return cfg
   }
   ```

4. **Use the config in your command**:

   ```go
   func runCommand(cmd *cobra.Command, args []string) error {
     cfg := NewCommandConfig(cmd)
     // Use cfg.Option1, cfg.Option2, etc.
     return nil
   }
   ```

See `cmd/ping.go` and `cmd/template_command.go.example` for working examples of this pattern.

### Modifying Configurations

Configuration options live alongside their commands and self-register with the global registry. To add or modify configurations:

1. **Add options**: In `cmd/<command>.go`, implement `<Command>Options() []config.ConfigOption` with keys like `app.<command>.<option>`.
2. **Self-register**: In `init()`, call `config.RegisterOptionsProvider(<Command>Options)`.
3. **Flags binding**: In the same command file, call `RegisterFlagsForPrefixWithOverrides(cmd, "app.<command>.", overrides)` instead of manual flag definitions and binds.
4. **Read values**: In your config constructor, use `getConfigValueWithFlags[T](cmd, flagName, viperKey)` so CLI flags override config/env.
5. **Docs**: Run `task generate:docs:config` to regenerate documentation.

Remember: **Never** use `viper.SetDefault()` directly. Defaults are applied via the registry in `cmd/root.go`.

### What’s Framework vs. What You Should Edit

- Framework (avoid modifying unless you are changing the scaffold itself):
  - `cmd/root.go` (bootstrap, root wiring, `setupCommandConfig`, helpers like `getConfigValueWithFlags`/`getKeyValue`)
  - `cmd/flags.go` (auto-registration of flags from options)
  - `internal/config/options.go` (types), `internal/config/registry.go` (provider registry + `SetDefaults`)
  - `internal/logger/*`, `internal/ui/*`, `internal/docs/*` (shared subsystems)
  - `Taskfile.yml`, scripts under `scripts/`

- User-editable (where you implement your app):
  - New commands under `cmd/<command>.go`
  - Command options under `internal/config/<command>_options.go`
  - Documentation content under `docs/` as needed
  - Optional UI customizations under `internal/ui/*` for your app-specific UI

### Customizing the UI

Explore the `internal/ui/` package to modify the Bubble Tea model, colors, and interactivity. Use configs to allow runtime customization of UI elements.

### Cursor AI Integration

The project includes a template for Cursor AI rules in `dot.cursorrules`. This template contains detailed project specifications that help Cursor AI understand the project structure, coding conventions, and requirements.

To use it with Cursor:

1. Copy the template to a `.cursorrules` file:

```bash
cp dot.cursorrules .cursorrules
```

2. Cursor will automatically detect and use these rules to provide better code suggestions and assistance.

The template covers:

- Project structure and design principles
- Command implementation patterns
- Error handling guidelines
- Testing requirements
- Git commit conventions

You can customize the rules to match your project's specific requirements.

---

## Tooling Best Practices

**Cobra** ([repo](https://github.com/spf13/cobra), [docs](https://pkg.go.dev/github.com/spf13/cobra)):

- Keep commands small and focused.
- Each command in its own file promotes clarity and testability.
- Use `RunE` to return errors gracefully rather than exiting immediately.

**Viper** ([repo](https://github.com/spf13/viper), [docs](https://pkg.go.dev/github.com/spf13/viper)):

- Set defaults first, then allow overrides via config files, env vars, and flags.
- Keep configuration keys consistent and descriptive.
- Exploit environment variable binding and automatic environment detection for easy deployment in different environments.

**Zerolog** ([repo](https://github.com/rs/zerolog), [docs](https://pkg.go.dev/github.com/rs/zerolog)):

- Use structured logs for better machine readability.
- Set a global log level and pass context around rather than using global variables directly.
- Keep logs concise and meaningful; leverage fields to add context without cluttering messages.

**Bubble Tea** ([repo](https://github.com/charmbracelet/bubbletea), [docs](https://pkg.go.dev/github.com/charmbracelet/bubbletea)):

- Keep the TUI logic isolated in its package or command.
- Make colors, messages, and interactions configurable to adapt to user preferences.

By following these best practices, you ensure that your CLI remains maintainable, testable, and flexible enough to grow with your project's needs.

---

## Contributing

1. Fork & create a new branch.
2. Make changes, run `task check`.
3. Commit with descriptive messages.
4. Open a pull request against `main`.

---

## License

MIT License. See [LICENSE](LICENSE).

---

## Additional Notes

- `task test:coverage:text` identifies uncovered code paths for targeted testing improvements.
- Press `q` or `Ctrl-C` to exit UI mode.
- Use quotes for special chars in arguments.
- Run `go mod tidy` to keep dependencies clean.
- Run `task check:deps` regularly to ensure dependencies are up-to-date and secure.
- Run `task check:license:source` after adding dependencies to catch license violations early.
- Regularly run tests, lint, and format tasks to maintain code quality and style.
- See [Test Fixtures Documentation](testdata/README.md) for information about available test fixtures and the organized subdirectory structure.

---

## Note

Keep your environment and tools updated. Embrace the structured approach offered by this scaffold, and enjoy building a professional-grade CLI with ckeletin-go!

---
