# ckeletin-go

![ckeletin-go](ckeletin-go.png)

**A professional Golang CLI scaffold for building beautiful, robust, and modular command-line applications.**

[![Build Status](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/peiman/ckeletin-go)](https://codecov.io/gh/peiman/ckeletin-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peiman/ckeletin-go)](https://goreportcard.com/report/github.com/peiman/ckeletin-go)
[![Version](https://img.shields.io/github/v/release/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/peiman/ckeletin-go.svg)](https://pkg.go.dev/github.com/peiman/ckeletin-go)
[![License](https://img.shields.io/github/license/peiman/ckeletin-go)](LICENSE)
[![CodeQL](https://github.com/peiman/ckeletin-go/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/peiman/ckeletin-go/security/code-scanning)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](https://go.dev)

---

## Table of Contents

- [ckeletin-go](#ckeletin-go)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
  - [Key Highlights](#key-highlights)
  - [Quick Start](#quick-start)
  - [Features](#features)
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
    - [Taskfile Tasks](#taskfile-tasks)
    - [Pre-Commit Hooks with Lefthook](#pre-commit-hooks-with-lefthook)
    - [Continuous Integration](#continuous-integration)
  - [Customization](#customization)
    - [Changing the Program Name](#changing-the-program-name)
    - [Adding New Commands](#adding-new-commands)
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

- **Single-Source Binary Name**: Update `BINARY_NAME` in `Taskfile.yml`, and `ldflags` handles the rest. No more hunting down references.
- **Detailed Coverage Reports**: Use `task test:coverage-text` to see exactly what code paths need testing.
- **Seamless Customization**: Easily add new commands, reconfigure settings, or integrate Bubble Tea UIs.

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
- **Structured Logging**: Use Zerolog to create efficient, leveled logs. Perfect for debugging, auditing, and production use.
- **Bubble Tea UI**: Optional, interactive UI for advanced terminal applications.
- **Single-Source Configuration**: Set defaults in config files, override with env vars, and fine-tune with flags.
- **Task Automation**: One Taskfile to define all build, test, and lint tasks.
- **High Test Coverage & Quality Checks**: Ensure a robust codebase that meets production standards.

---

## Getting Started

### Prerequisites

- **Go**: 1.20+ recommended.
- **Task**: Install from [taskfile.dev](https://taskfile.dev/#/installation).
- **Git**: For version control.

### Installation

```bash
git clone https://github.com/yourusername/ckeletin-go.git
cd ckeletin-go
task setup
```

### Using the Scaffold

1. Update `module` path in `go.mod`.
2. Change `BINARY_NAME` in `Taskfile.yml` to rename your CLI (e.g., `myapp`).
3. Build and run to confirm setup:

   ```bash
   task build
   ./myapp ping
   ```

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

When you clone this repository, it's important to update the `MODULE_PATH` in the `go.mod` file to reflect your own repository path. This ensures that your module is uniquely identifiable and avoids conflicts with other projects.

#### Steps to Update the Module Path

1. **Open `go.mod`**: Locate the `go.mod` file in the root of the project.

2. **Edit the Module Path**: Change the module path to reflect your own repository. For example, if you're using GitHub, it might look like this:

   ```go
   module github.com/yourusername/your-repo-name
   ```

   If you're using another version control system, adjust the path accordingly. For example:

   ```go
   module gitlab.com/yourusername/your-repo-name
   ```

3. **Update References**: If the `MODULE_PATH` is used elsewhere in the project (e.g., in `Taskfile.yml` for build flags), update those references to match your new module path.

4. **Run `go mod tidy`**: After making changes, run `go mod tidy` to clean up any unnecessary dependencies and ensure the `go.mod` and `go.sum` files are up to date.

By following these steps, you can ensure that your version of the project is correctly configured and ready for further development or deployment.

---

## Configuration

**ckeletin-go** uses Viper for flexible configuration and also implements a centralized configuration system with a single source of truth:

### Configuration Management

All configuration options are defined in **one place**: `internal/config/registry.go`.

This registry contains:

- Default values
- Option descriptions
- Data types
- Example values
- Metadata for documentation

Benefits of this approach:

- No duplication of defaults across the codebase
- Self-documenting configuration
- Automatic validation during build
- Consistent environment variable naming

### Adding New Configuration Options

When adding a new configuration option, ONLY add it to the registry:

```go
// in internal/config/registry.go
func Registry() []ConfigOption {
    return []ConfigOption{
        // Existing options...
        {
            Key:          "app.myfeature.setting",
            DefaultValue: "default-value",
            Description:  "Description of what this setting does",
            Type:         "string",
            Required:     false,
            Example:      "example-value",
        },
    }
}
```

**Important**: Never use `viper.SetDefault()` directly in command files. Our `check-defaults` task will catch any violations of this rule.

### Automatic Documentation Generation

Generate comprehensive configuration documentation with:

```bash
task docs:config
```

This creates a Markdown file at `docs/configuration.md` with:

- All available configuration options
- Default values and types
- Environment variable names
- Full descriptions

For a configuration template, run:

```bash
task docs:config-yaml
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

- `task deps:verify`: Verifies that dependencies haven't been modified
- `task deps:outdated`: Checks for outdated dependencies
- `task deps:check`: Runs all dependency checks (verification, outdated, vulnerabilities)

### Automated Checks

Dependency verification is automatically included in:

- Pre-commit hooks via Lefthook (prevents commits with corrupted dependencies)
- CI workflow via GitHub Actions (as part of `task check`)
- The comprehensive quality check command: `task check`

### Best Practices

1. Run `task deps:check` before starting a new feature
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

---

## Development Workflow

### Taskfile Tasks

- `task setup`: Install tools.
- `task format`: Format code.
- `task lint`: Run linters.
- `task vuln`: Check for vulnerabilities.
- `task deps:verify`: Verify dependency integrity.
- `task deps:outdated`: Check for outdated dependencies.
- `task deps:check`: Run all dependency checks (verification, outdated, vulnerabilities).
- `task test`: Run tests with coverage.
- `task test:coverage-text`: Detailed coverage report.
- `task check`: All checks (format, lint, deps, tests).
- `task build`: Build the binary.
- `task run`: Run the binary.
- `task clean`: Clean artifacts.

### Pre-Commit Hooks with Lefthook

`task setup` installs hooks that run `format`, `lint`, `test` on commit, ensuring code quality before changes land in the repository.

### Continuous Integration

GitHub Actions runs `task check` on each commit or pull request, maintaining code standards and reliability.

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

Install Cobra CLI tool:

```bash
go install github.com/spf13/cobra-cli@latest
```

Add a new command:

```bash
cobra-cli add hello
```

This follows Cobra's best practice: each command in its own file, cleanly separated and easily testable.

### Modifying Configurations

Configuration has been centralized in `internal/config/registry.go` for clarity and maintainability. To add or modify configurations:

1. **Add new options to the registry**: Add a new entry to the `Registry()` function in `internal/config/registry.go`.

2. **Bind command flags**: Use `viper.BindPFlag()` in your command files to bind flags to configuration keys:

   ```go
   cmd.Flags().String("setting", "", "Description of the setting")
   viper.BindPFlag("app.myfeature.setting", cmd.Flags().Lookup("setting"))
   ```

3. **Access configuration**: Use Viper's `Get*` methods in your code:

   ```go
   value := viper.GetString("app.myfeature.setting")
   ```

4. **Generate documentation**: After adding new configuration options, regenerate the documentation:

   ```bash
   task docs
   ```

Remember: **Never** use `viper.SetDefault()` directly. The `check-defaults` task will flag any violations.

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

- `task test:coverage-text` identifies uncovered code paths for targeted testing improvements.
- Press `q` or `Ctrl-C` to exit UI mode.
- Use quotes for special chars in arguments.
- Run `go mod tidy` to keep dependencies clean.
- Run `task deps:check` regularly to ensure dependencies are up-to-date and secure.
- Regularly run tests, lint, and format tasks to maintain code quality and style.
- See [Test Fixtures Documentation](docs/test-fixtures.md) for information about available test fixtures.

---

## Note

Keep your environment and tools updated. Embrace the structured approach offered by this scaffold, and enjoy building a professional-grade CLI with ckeletin-go!

---
