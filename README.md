# ckeletin-go

A professional CLI application skeleton for Go projects, providing a solid foundation for building command-line tools with industry-standard libraries and best practices.

## Overview

ckeletin-go is a starter template for building professional-grade command-line applications in Go. It offers a well-structured foundation that follows Go best practices and leverages proven industry-standard libraries.

### Who is this for?
- Developers starting new CLI projects who want a solid foundation
- Teams looking to standardize their CLI application structure
- Anyone wanting to learn about Go CLI application best practices

### Key Features
- Modern CLI structure using Cobra
- Configuration management with Viper
- Structured logging with Zerolog
- Robust error handling
- Comprehensive test coverage
- GitHub Actions CI pipeline
- Standard Go project layout

## Getting Started

### Prerequisites
- Go 1.21 or later
- Make (optional, for using Makefile commands)

### Installation
```bash
# Clone the repository
git clone https://github.com/peiman/ckeletin-go.git
cd ckeletin-go

# Install dependencies
go mod download

# Build the project
make build
```

### Quick Start for Your Project
1. Create your project:
```bash
git clone https://github.com/peiman/ckeletin-go.git my-cli-app
cd my-cli-app
rm -rf .git
git init
```

2. Update the module name:
```bash
go mod edit -module github.com/yourusername/my-cli-app
```

3. Build and run:
```bash
make build
./ckeletin-go
```

## Core Components

### Command Line Interface
Built with [Cobra](https://github.com/spf13/cobra), providing:
- Intuitive command and subcommand structure
- Automatic help generation
- Flag handling
- Shell completions
- Command aliasing

### Configuration Management
Using [Viper](https://github.com/spf13/viper) for:
- JSON configuration file support
- Environment variable binding
- Default configuration values
- Dynamic configuration reloading
- Configuration file watching

### Logging
Implemented with [Zerolog](https://github.com/rs/zerolog), offering:
- Structured JSON logging
- Level-based logging (debug, info, warn, error)
- High-performance logging
- Contextual logging with fields
- Log rotation support

### Error Handling
Custom error handling package in `internal/errors`:
- Structured error types with error codes
- Error wrapping and context preservation
- Consistent error handling patterns
- Error logging integration

Example usage:
```go
if err := operation(); err != nil {
    return errors.NewAppError(errors.ErrOperationFailed, "operation description", err)
}
```

## Project Structure
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

## Development

### Available Make Commands
```bash
make build          # Build the binary
make test          # Run tests
make test-race     # Run tests with race detector
make lint          # Run linter
make format        # Format code
make vuln          # Check for vulnerabilities
make check         # Run all quality checks
make clean         # Clean build artifacts
make run           # Run the application
```

### Testing
Following Go's standard practices:
- Tests are co-located with the code they test (`_test.go` files)
- Unit tests and integration tests in the same package
- Test helpers in `test_helpers_test.go` files
- Race condition detection with `make test-race`
- Coverage reports in coverage.txt

### Adding New Commands
1. Create a new command file in `cmd/`:
```go
package cmd

import "github.com/spf13/cobra"

var newCmd = &cobra.Command{
    Use:   "new",
    Short: "Brief description",
    Run: func(cmd *cobra.Command, args []string) {
        // Command implementation
    },
}

func init() {
    rootCmd.AddCommand(newCmd)
}
```

2. Add corresponding tests in `cmd/new_test.go`

## Documentation

### Core Libraries
- [Cobra Documentation](https://pkg.go.dev/github.com/spf13/cobra)
  - [User Guide](https://cobra.dev/)
- [Viper Documentation](https://pkg.go.dev/github.com/spf13/viper)
- [Zerolog Documentation](https://pkg.go.dev/github.com/rs/zerolog)

### Configuration
Default configuration file: `ckeletin-go.json`
```bash
# Use custom config file
./ckeletin-go --config /path/to/config.json

# Set log level
./ckeletin-go --log-level debug
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines
- Write tests for new features
- Update documentation
- Follow Go best practices
- Ensure all checks pass (`make check`)
- Use meaningful commit messages

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.