# ckeletin-go

[![Build Status](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/peiman/ckeletin-go)](https://codecov.io/gh/peiman/ckeletin-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peiman/ckeletin-go)](https://goreportcard.com/report/github.com/peiman/ckeletin-go)
[![Version](https://img.shields.io/github/v/release/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/peiman/ckeletin-go.svg)](https://pkg.go.dev/github.com/peiman/ckeletin-go)
[![License](https://img.shields.io/github/license/peiman/ckeletin-go)](LICENSE)
[![CodeQL](https://github.com/peiman/ckeletin-go/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/peiman/ckeletin-go/security/code-scanning)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](https://go.dev)

**A professional Golang CLI scaffold for building beautiful, robust, and modular command-line applications.**

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
  - [Configuration](#configuration)
    - [Configuration File](#configuration-file)
    - [Environment Variables](#environment-variables)
    - [Command-Line Flags](#command-line-flags)
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
   git clone https://github.com/yourusername/ckeletin-go.git
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

   You’ll see “Pong” printed—congratulations, you’re running the scaffold!

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

---

## Configuration

**ckeletin-go** uses Viper for flexible configuration:

### Configuration File

Default config file: `$HOME/.ckeletin-go.yaml` (or `myapp.yaml` if renamed).

Example:

```yaml
app:
  log_level: "info"
  ping:
    output_message: "Pong"
    output_color: "green"
    ui: false
```

### Environment Variables

Override any config via environment variables:

```bash
export APP_LOG_LEVEL="debug"
export APP_PING_OUTPUT_MESSAGE="Hello, World!"
export APP_PING_UI=true
```

### Command-Line Flags

Override at runtime:

```bash
./myapp ping --message "Hi there!" --color yellow --ui
```

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
- `task test`: Run tests with coverage.
- `task test:coverage-text`: Detailed coverage report.
- `task check`: All checks.
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

This follows Cobra’s best practice: each command in its own file, cleanly separated and easily testable.

### Modifying Configurations

Set new defaults in `initConfig` or in command files. Use `viper.BindPFlag()` to bind flags. Adjust config files or env vars to match your desired behavior.

### Customizing the UI

Explore the `internal/ui/` package to modify the Bubble Tea model, colors, and interactivity. Use configs to allow runtime customization of UI elements.

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
- Regularly run tests, lint, and format tasks to maintain code quality and style.

---

## Note

Keep your environment and tools updated. Embrace the structured approach offered by this scaffold, and enjoy building a professional-grade CLI with ckeletin-go!
