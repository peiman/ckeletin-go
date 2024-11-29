# ckeletin-go

[![Build Status](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/peiman/ckeletin-go)](https://codecov.io/gh/peiman/ckeletin-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peiman/ckeletin-go)](https://goreportcard.com/report/github.com/peiman/ckeletin-go)
[![Version](https://img.shields.io/github/v/release/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/peiman/ckeletin-go.svg)](https://pkg.go.dev/github.com/peiman/ckeletin-go)
[![License](https://img.shields.io/github/license/peiman/ckeletin-go)](LICENSE)
[![CodeQL](https://github.com/peiman/ckeletin-go/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/peiman/ckeletin-go/security/code-scanning)
[![Made with Go](https://img.shields.io/badge/made%20with-Go-brightgreen.svg)](https://go.dev)

**A professional Golang CLI scaffold for building beautiful, robust, and modular command-line applications.**

---

## Table of Contents

- [ckeletin-go](#ckeletin-go)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
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
  - [Contributing](#contributing)
  - [License](#license)
  - [Additional Notes](#additional-notes)
  - [Note](#note)

---

## Introduction

**ckeletin-go** is a Golang scaffold project designed to help developers create professional, robust, and beautiful command-line interface (CLI) applications. The name **ckeletin** is a playful twist on the word "skeleton," reflecting the project's role as a foundational structure for your CLI applications.

This scaffold integrates essential libraries and tools that follow best practices in Go development, including proper logging, configuration management, error handling, testing, and task automation. It emphasizes modularity by having each command manage its own configurations and defaults, promoting separation of concerns and ease of maintenance.

---

## Features

- **Modular Command Structure with Cobra and Viper**
  - Each command is self-contained, handling its own configurations and defaults.
  - Easily add new commands and flags.
  - Manage commands and configurations seamlessly.

- **Structured Logging with Zerolog**
  - Efficient, structured logging with configurable log levels.
  - Centralized logging initialization ensures consistent logging behavior.

- **Beautiful Terminal UI with Bubble Tea**
  - Create interactive and aesthetically pleasing interfaces.
  - Commands can optionally enable UI features, enhancing user experience.
  - Customize UI elements like colors and messages.

- **Task Automation with Taskfile**
  - Define and automate development tasks.
  - Single source of truth for tasks, used in pre-commit hooks and CI pipelines.
  - Ensures consistent development workflow and environment.

- **Robust Error Handling**
  - Follow Go best practices for error propagation and handling.
  - Provide meaningful error messages to users.
  - Gracefully handle errors within commands.

- **Testing and Code Quality**
  - Test-driven development with high test coverage.
  - Tests are designed to be self-contained, ensuring configurations are properly initialized.
  - Static code analysis and vulnerability checks.

- **Configurability and Extensibility**
  - Commands manage their own configurations, enhancing modularity.
  - Customize command behaviors, output messages, colors, and log levels.
  - Easily change the program name; changes propagate throughout the application.

---

## Getting Started

### Prerequisites

- **Go**: Version **1.20** or higher (recommended **1.23.x**). You can download it from [golang.org](https://golang.org/dl/).

- **Task**: Task is a task runner similar to Make, but written in Go. Install it by following the instructions at [taskfile.dev](https://taskfile.dev/#/installation).

- **Git**: Git is required for version control and to clone the repository.

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/yourusername/ckeletin-go.git
   cd ckeletin-go
   ```

2. **Set Up Development Tools**

   Install all the necessary development tools using Task:

   ```bash
   task setup
   ```

   This will install Go tools and set up the Git pre-commit hooks automatically.

### Using the Scaffold

To use **ckeletin-go** as a starting point for your own CLI application:

1. **Rename the Module Path**

   - Update the `module` path in `go.mod` to your own project's path.
   - Replace all occurrences of `github.com/yourusername/ckeletin-go` with your module path.

2. **Change the Program Name**

   - Update the `BINARY_NAME` variable in `Taskfile.yml`.
   - Modify the `ProgramName` constant in the configuration or main package.
   - Use Go's `ldflags` in `Taskfile.yml` to inject the program name at build time.
   - The name change will propagate throughout the application.

3. **Customize the `ping` Command**

   - Modify the `ping` command to suit your needs.
   - Use it as a template to add new commands.
   - Ensure new commands manage their own configurations and defaults within their files.

4. **Run the Application**

   ```bash
   task run
   ```

---

## Configuration

**ckeletin-go** uses Viper for configuration management, allowing you to configure the application via configuration files, environment variables, and flags. Each command handles its own configurations, enhancing modularity and separation of concerns.

### Configuration File

Create a configuration file in one of the supported formats (YAML, JSON, TOML) and place it in one of the default paths (e.g., `$HOME/.ckeletin-go.yaml`).

Example `~/.ckeletin-go.yaml`:

```yaml
app:
  log_level: "info"
  ping:
    output_message: "Pong"
    output_color: "green"
    ui: false
```

- **Global Configurations**: Settings under `app`, like `log_level`, are global and affect the entire application.
- **Command-Specific Configurations**: Settings under `app.ping` are specific to the `ping` command.

### Environment Variables

You can set configuration options using environment variables. The environment variable names are derived from the configuration keys by replacing dots (`.`) with underscores (`_`) and converting to uppercase.

**Examples:**

- To override `app.log_level`:

  ```bash
  export APP_LOG_LEVEL="debug"
  ```

- To override `app.ping.output_message`:

  ```bash
  export APP_PING_OUTPUT_MESSAGE="Hello, World!"
  ```

- To override `app.ping.ui`:

  ```bash
  export APP_PING_UI=true
  ```

### Command-Line Flags

Override configurations at runtime using flags:

```bash
./ckeletin-go ping --message "Hi there!" --color yellow --ui
```

- Flags take precedence over environment variables and configuration files.

---

## Commands

### `ping` Command

The `ping` command is a sample command that demonstrates how to implement a command using Cobra, Viper, Zerolog, and Bubble Tea. It manages its own configurations and defaults, promoting modularity.

#### Usage

```bash
./ckeletin-go ping [flags]
```

#### Flags

- `-m`, `--message`: Custom output message (overrides configuration).
- `-c`, `--color`: Output color (overrides configuration).
- `--ui`: Enable the Bubble Tea UI.

#### Examples

- Default behavior:

  ```bash
  ./ckeletin-go ping
  ```

- Custom message and color:

  ```bash
  ./ckeletin-go ping --message "Hello!" --color cyan
  ```

- Enable UI:

  ```bash
  ./ckeletin-go ping --ui
  ```

- Using environment variables:

  ```bash
  export APP_PING_OUTPUT_MESSAGE="Env Message"
  ./ckeletin-go ping
  ```

---

## Development Workflow

### Taskfile Tasks

**ckeletin-go** uses a `Taskfile.yml` to define and automate development tasks. Here are some of the most common tasks:

- **Setup Development Tools**

  ```bash
  task setup
  ```

- **Format Code**

  ```bash
  task format
  ```

- **Run Linters**

  ```bash
  task lint
  ```

- **Check for Vulnerabilities**

  ```bash
  task vuln
  ```

- **Run Tests with Coverage**

  ```bash
  task test
  ```

- **Run All Quality Checks**

  ```bash
  task check
  ```

- **Build the Application**

  ```bash
  task build
  ```

- **Run the Application**

  ```bash
  task run
  ```

- **Clean Build Artifacts**

  ```bash
  task clean
  ```

### Pre-Commit Hooks with Lefthook

To maintain code quality and consistency, Lefthook is set up to run tasks before each commit.

- **Install Lefthook**

  Lefthook is installed during `task setup`. The Git hooks are automatically configured.

- **Hooks Executed**

  - `task format`: Ensures code is formatted.
  - `task lint`: Runs linters to catch issues.
  - `task test`: Runs tests to ensure code works as expected.

### Continuous Integration

The project includes a CI pipeline configured using GitHub Actions. The CI pipeline runs the `task check` command to perform all quality checks, ensuring that code merged into the main branch meets the project's standards.

- **CI Pipeline Tasks**

  - Checkout repository.
  - Set up Go environment.
  - Install Task.
  - Install dependencies (`task setup`).
  - Run checks (`task check`).

---

## Customization

### Changing the Program Name

To change the program name:

1. Update the `BINARY_NAME` variable in `Taskfile.yml`.
2. Modify the `ProgramName` constant in the codebase.
3. Use Go's `ldflags` in `Taskfile.yml` to inject the program name at build time.
4. The name change will propagate throughout the application.

### Adding New Commands

1. **Create a New Command File**

   - Add a new file in the `cmd/` directory (e.g., `cmd/yourcommand.go`).

2. **Implement the Command Logic**

   - Use `cobra.Command` to define the command, its flags, and behavior.
   - Manage command-specific configurations and defaults within the command file.

3. **Register the Command**

   - In `cmd/root.go`, add your new command:

     ```go
     rootCmd.AddCommand(NewYourCommand())
     ```

### Modifying Configurations

- **Add New Configuration Fields**

  - Define new default values within the command file or `initConfig` in `root.go` for global settings.

- **Use Viper to Bind New Configuration Options**

  - Bind new flags and environment variables using Viper within the command file.

- **Update Configuration Files and Environment Variables**

  - Reflect new configuration options in your config files and document the corresponding environment variables.

### Customizing the UI

- **Modify the Bubble Tea Model**

  - Update the UI logic in the `internal/ui/` package or within your command if the UI is command-specific.

- **Customize Colors and Styles**

  - Use configuration options to make UI elements customizable by the user.

- **Enhance Interactivity**

  - Add new interactive components to the UI as needed.

---

## Contributing

Contributions are welcome! Please follow these steps:

1. **Fork the Repository**

   - Click the "Fork" button at the top right of the repository page.

2. **Create a New Branch**

   - For a new feature: `git checkout -b feature/your-feature-name`
   - For a bug fix: `git checkout -b bugfix/your-bugfix-name`

3. **Make Your Changes**

   - Ensure all tasks pass by running `task check`.

4. **Commit Your Changes**

   - Use clear and descriptive commit messages.

5. **Submit a Pull Request**

   - Push your branch to your forked repository.
   - Open a pull request against the main repository's `main` branch.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Additional Notes

- **Testing the Application**

  After making changes, always rebuild and test the application:

  ```bash
  task build
  ```

  Run the application:

  ```bash
  ./ckeletin-go ping --message 'Hello, World!' --color green --ui
  ```

- **Exiting the Program**

  - Press `q`, `Esc`, or `CTRL-C` to exit the program when using the `--ui` flag.

- **Shell Special Characters**

  - Be cautious when using special characters in command-line arguments.
  - Use single quotes or escape characters as needed.

- **Updating Dependencies**

  - Run `go mod tidy` to ensure dependencies are up-to-date.
  - Update tools and dependencies regularly for compatibility.

---

## Note

Remember to keep your development environment and tools updated to the latest stable versions to avoid compatibility issues. Regularly run your test suite and linters to maintain code quality.
