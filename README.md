# ckeletin-go

A professional CLI application skeleton for Go projects, providing a solid foundation for building command-line tools with industry-standard libraries and best practices.

## Overview

ckeletin-go is a starter template for building professional-grade command-line applications in Go. It provides a well-structured foundation with:

- Modern CLI framework using Cobra
- Robust configuration management with Viper
- Structured logging with Zerolog
- Comprehensive test coverage
- GitHub Actions CI pipeline
- Proper error handling
- Professional project layout following Go standards

This skeleton is ideal for developers who want to:
- Start a new CLI project without reinventing the wheel
- Follow Go best practices from day one
- Have a solid testing foundation
- Use industry-standard libraries

## Features

- **Command Line Interface**: Built with [Cobra](https://github.com/spf13/cobra)
  - Subcommand support
  - Automatic help generation
  - Flag handling
  - Modern CLI patterns

- **Configuration Management**: Using [Viper](https://github.com/spf13/viper)
  - JSON config file support
  - Environment variable binding
  - Default configuration handling
  - Dynamic configuration reloading

- **Structured Logging**: Implemented with [Zerolog](https://github.com/rs/zerolog)
  - JSON-structured logging
  - Level-based logging
  - High performance
  - Contextual logging

- **Professional Structure**
  - Standard Go project layout
  - Clean separation of concerns
  - Internal and public packages
  - Proper error handling

- **Development Tools**
  - Comprehensive Makefile
  - golangci-lint configuration
  - GitHub Actions CI
  - Comprehensive test suite

## Installation

### Prerequisites
- Go 1.21 or later
- Make (optional, for using Makefile commands)

### Building from Source
```bash
# Clone the repository
git clone https://github.com/peiman/ckeletin-go.git
cd ckeletin-go

# Install development tools
make setup

# Build the project
make build
```

## Development Setup

### Installing Dependencies

The project uses Go modules for dependency management:

```bash
# Install main dependencies
go mod download

# Install development tools
make setup
```

Development tools are managed through `tools.go` and are automatically installed when running `make setup`. These include:

- `gotestsum`: Enhanced test runner with better output formatting
- `golangci-lint`: Comprehensive linting tool
- `goimports`: Import management and formatting
- `govulncheck`: Vulnerability checking
- `richgo`: Enriched test output

### Available Make Commands

```bash
make help           # Show available commands
make setup         # Install development tools
make build         # Build the binary
make test          # Run tests with coverage
make test-race     # Run tests with race detector
make test-pretty   # Run tests with formatted output
make test-watch    # Run tests in watch mode
make lint          # Run linters
make format        # Format code
make vuln          # Check for vulnerabilities
make check         # Run all quality checks
```

### Development Workflow

1. Fork and clone the repository
2. Install dependencies and tools:
   ```bash
   make setup
   ```

3. Create a new branch:
   ```bash
   git checkout -b feature/your-feature
   ```

4. Make your changes, following these guidelines:
   - Write tests first (TDD approach)
   - Keep code coverage high
   - Follow Go best practices
   - Use the provided tooling

5. Verify your changes:
   ```bash
   make check      # Runs all quality checks
   ```

6. Commit and push your changes:
   ```bash
   git commit -m "feat: add new feature"
   git push origin feature/your-feature
   ```

### Testing

Tests follow Go's standard testing patterns:

- Tests are in `_test.go` files alongside the code they test
- Integration tests use `TestMain` for setup
- Table-driven tests for comprehensive coverage
- Race condition checking with `make test-race`

Running tests:
```bash
make test          # Regular tests with coverage
make test-race     # Tests with race detection
make test-pretty   # Tests with formatted output
make test-watch    # Tests in watch mode
```

### Project Structure
```
ckeletin-go/
├── cmd/                    - CLI commands
│   ├── root.go            - Main command
│   ├── root_test.go       - Command tests
│   └── version.go         - Version command
├── internal/              - Private application code
│   ├── errors/            - Error handling
│   └── infrastructure/    - Core infrastructure
├── pkg/                   - Public packages
├── main.go                - Application entry
└── main_test.go           - Main package tests
```

## Usage

### Basic Commands
```bash
# Run the application
./ckeletin-go

# Show version
./ckeletin-go version

# Show help
./ckeletin-go --help
```

### Configuration
The application looks for `ckeletin-go.json` in the current directory. You can specify a different config file:
```bash
./ckeletin-go --config /path/to/config.json
```

### Logging
Control log level using the --log-level flag:
```bash
./ckeletin-go --log-level debug
```

## Error Handling

The project implements a robust error handling strategy:

- Custom error types in `internal/errors`
- Consistent error wrapping pattern
- Structured error logging
- Error code standardization

Example:
```go
if err := operation(); err != nil {
    return errors.NewAppError(errors.ErrOperationFailed, "operation description", err)
}
```

## Contributing

1. Install development tools:
   ```bash
   make setup
   ```

2. Create your feature branch
3. Make your changes
4. Run quality checks:
   ```bash
   make check
   ```
5. Commit your changes
6. Push and create a Pull Request

Please ensure:
- Tests are added for new features
- Documentation is updated
- All checks pass (`make check`)
- Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
