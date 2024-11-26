
# ckeletin-go

**A professional Golang CLI scaffold for building beautiful and robust command-line applications.**

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

This scaffold integrates essential libraries and tools that follow best practices in Go development, including proper logging, configuration management, error handling, testing, and task automation. It also includes a sample `ping` command that demonstrates how to implement commands using the integrated libraries.

---

## Features

- **CLI Framework with Cobra and Viper**
  - Manage commands and configurations seamlessly.
  - Easily add new commands and flags.

- **Structured Logging with Zerolog**
  - Efficient and leveled logging.
  - Configurable log levels via configuration or flags.

- **Beautiful Terminal UI with Bubble Tea**
  - Create interactive and aesthetically pleasing interfaces.
  - Customize UI elements like colors and messages.

- **Task Automation with Taskfile**
  - Define and automate development tasks.
  - Single source of truth for tasks, used in pre-commit hooks and CI pipeline.

- **Robust Error Handling**
  - Follow Go best practices for error propagation and handling.
  - Provide meaningful error messages to users.

- **Testing and Code Quality**
  - Test-driven development with high test coverage.
  - Static code analysis and vulnerability checks.

- **Configurability**
  - Easily change the program name; changes propagate throughout the application.
  - Customize command behaviors, output messages, colors, and log levels.

---

## Getting Started

### Prerequisites

- **Go**: Version **1.20** or higher (recommended **1.23.x**). You can download it from [golang.org](https://golang.org/dl/).

- **Task**: Task is a task runner similar to Make, but written in Go. Install it by following the instructions at [taskfile.dev](https://taskfile.dev/#/installation).

- **Git**: Git is required for version control and to clone the repository.

### Installation

1. **Clone the Repository**

   ```bash
   git clone https://github.com/peiman/ckeletin-go.git
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
   - Replace all occurrences of `github.com/peiman/ckeletin-go` with your module path.

2. **Change the Program Name**

   - Update the `BINARY_NAME` variable in `Taskfile.yml`.
   - Modify the `ProgramName` constant in the configuration or main package.
   - The name change will propagate throughout the application.

3. **Customize the `ping` Command**

   - Modify the `ping` command to suit your needs.
   - Use it as a template to add new commands.

4. **Run the Application**

   ```bash
   task run
   ```

---

## Configuration

**ckeletin-go** uses Viper for configuration management, allowing you to configure the application via configuration files, environment variables, and flags.

### Configuration File

Create a configuration file in one of the supported formats (JSON, YAML, TOML) and place it in one of the default paths (`./config`, `$HOME/.ckeletin-go.yaml`, etc.).

Example `config.yaml`:

```yaml
app:
  output_message: "Pong"
  output_color: "green"
  log_level: "info"
```

### Environment Variables

You can set configuration options using environment variables. For example:

```bash
export APP_OUTPUT_MESSAGE="Hello, World!"
export APP_OUTPUT_COLOR="blue"
export APP_LOG_LEVEL="debug"
```

### Command-Line Flags

Override configurations at runtime using flags:

```bash
./ckeletin-go ping --message "Hi there!" --color yellow --log-level debug
```

---

## Commands

### `ping` Command

The `ping` command is a sample command that demonstrates how to implement a command using Cobra, Viper, Zerolog, and Bubble Tea.

#### Usage

```bash
./ckeletin-go ping [flags]
```

#### Flags

- `-m`, `--message`: Custom output message (overrides configuration).
- `-c`, `--color`: Output color (overrides configuration).
- `-l`, `--log-level`: Set the log level (e.g., debug, info, warn, error).
- `-u`, `--ui`: Enable the Bubble Tea UI.

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

  - Checkout repository
  - Set up Go environment
  - Install Task
  - Install dependencies (`task setup`)
  - Run checks (`task check`)

---

## Customization

### Changing the Program Name

To change the program name:

1. Update the `BINARY_NAME` variable in `Taskfile.yml`.
2. Modify the `ProgramName` constant in the codebase.
3. Use Go's `ldflags` in `Taskfile.yml` to inject the program name at build time.

### Adding New Commands

1. Use Cobra to generate a new command:

   ```bash
   cobra add yourcommand
   ```

2. Implement the command logic in the generated files.
3. Register the command in the root command.

### Modifying Configurations

- Add new configuration fields in the configuration struct.
- Use Viper to bind new configuration options.
- Update the configuration file, environment variables, and flags to include new options.

### Customizing the UI

- Modify the Bubble Tea model in the UI package.
- Customize colors, styles, and interactive elements.
- Use configuration options to make UI customizable by the user.

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Ensure all tasks pass by running `task check`.
4. Commit your changes with clear commit messages.
5. Submit a pull request.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

**Feel free to reach out if you have any questions or need further assistance!**

---

# Additional Notes

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

# Note

Remember to keep your development environment and tools updated to the latest stable versions to avoid compatibility issues. Regularly run your test suite and linters to maintain code quality.
