# Contributing to ckeletin-go

Thank you for your interest in contributing to ckeletin-go! This document provides guidelines and steps for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Adding a New Command](#adding-a-new-command)
- [Adding a Configuration Option](#adding-a-configuration-option)
- [Testing Requirements](#testing-requirements)
- [Code Quality Standards](#code-quality-standards)
- [Submitting Changes](#submitting-changes)
- [Code Review Process](#code-review-process)

## Code of Conduct

Be respectful, professional, and constructive in all interactions. This project aims to foster an inclusive and welcoming environment for all contributors.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git
- A GitHub account

### Initial Setup

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/ckeletin-go.git
   cd ckeletin-go
   ```

2. **Install development tools:**
   ```bash
   task setup
   ```
   This installs:
   - Task runner
   - goimports (code formatting)
   - golangci-lint (linting)
   - gotestsum (test runner)
   - govulncheck (vulnerability scanning)
   - lefthook (git hooks)

3. **Verify your setup:**
   ```bash
   task check
   ```
   This should pass all quality checks.

4. **Read the architecture documentation:**
   - `README.md` - Project overview
   - `CLAUDE.md` - Development guidelines
   - `docs/adr/*.md` - Architecture Decision Records (ADRs)

## Development Workflow

### Before You Start Coding

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Understand the codebase structure:**
   ```
   ckeletin-go/
   ├── cmd/                    # Ultra-thin CLI commands (~20-30 lines)
   ├── internal/config/        # Centralized configuration registry
   ├── internal/logger/        # Logging infrastructure
   ├── internal/<feature>/     # Feature business logic
   └── docs/adr/              # Architecture decisions
   ```

3. **Review relevant ADRs:**
   - [ADR-001](docs/adr/001-ultra-thin-command-pattern.md) - Command structure
   - [ADR-002](docs/adr/002-centralized-configuration-registry.md) - Configuration
   - [ADR-003](docs/adr/003-dependency-injection-over-mocking.md) - Testing approach

### During Development

1. **Format your code frequently:**
   ```bash
   task format
   ```

2. **Run tests as you go:**
   ```bash
   task test
   ```

3. **Check for linting issues:**
   ```bash
   task lint
   ```

### Before Committing

**This is mandatory - all checks must pass:**

```bash
task check
```

This runs:
- ✅ Format verification
- ✅ Linting (go vet + golangci-lint)
- ✅ Pattern enforcement (ultra-thin commands, no scattered SetDefaults)
- ✅ Dependency verification
- ✅ Vulnerability scanning
- ✅ Tests with coverage requirements

Pre-commit hooks will also run automatically via Lefthook.

## Adding a New Command

Follow the **ultra-thin command pattern** ([ADR-001](docs/adr/001-ultra-thin-command-pattern.md)):

### Step 1: Scaffold the Command

```bash
task generate:command name=mycommand
```

This creates:
- `cmd/mycommand.go` - Ultra-thin CLI wrapper
- `internal/config/commands/mycommand_config.go` - Configuration metadata

### Step 2: Define Configuration Options

Edit `internal/config/commands/mycommand_config.go`:

```go
package commands

import "github.com/peiman/ckeletin-go/internal/config"

// MycommandMetadata defines the command metadata
var MycommandMetadata = config.CommandMetadata{
	Use:   "mycommand",
	Short: "Brief description of your command",
	Long:  "Detailed description of what your command does",
	ConfigPrefix: "app.mycommand",
	FlagOverrides: map[string]string{
		"app.mycommand.option1": "opt1",
		"app.mycommand.option2": "opt2",
	},
}

func init() {
	config.RegisterOptionsProvider(MycommandOptions)
}

// MycommandOptions returns configuration options for mycommand
func MycommandOptions() []config.ConfigOption {
	return []config.ConfigOption{
		{
			Key:          "app.mycommand.option1",
			DefaultValue: "default_value",
			Description:  "Description of option1",
			Type:         "string",
			Example:      "example_value",
		},
		// Add more options as needed
	}
}
```

### Step 3: Generate Type-Safe Constants

```bash
task generate:config:key-constants
```

This updates `internal/config/keys_generated.go` with new constants like:
- `KeyAppMycommandOption1`
- `KeyAppMycommandOption2`

### Step 4: Implement Business Logic

Create `internal/mycommand/mycommand.go`:

```go
package mycommand

import (
	"io"
	"github.com/rs/zerolog/log"
)

// Config holds configuration for the mycommand command
type Config struct {
	Option1 string
	Option2 string
}

// Executor handles the execution of the mycommand command
type Executor struct {
	cfg    Config
	writer io.Writer
}

// NewExecutor creates a new mycommand executor
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	return &Executor{
		cfg:    cfg,
		writer: writer,
	}
}

// Execute runs the mycommand logic
func (e *Executor) Execute() error {
	log.Debug().
		Str("option1", e.cfg.Option1).
		Str("option2", e.cfg.Option2).
		Msg("Starting mycommand execution")

	// Your business logic here

	log.Info().Msg("Mycommand completed successfully")
	return nil
}
```

### Step 5: Wire the Command

Edit `cmd/mycommand.go` to wire everything together (keep it ultra-thin, ~20-30 lines):

```go
package cmd

import (
	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/peiman/ckeletin-go/internal/config/commands"
	"github.com/peiman/ckeletin-go/internal/mycommand"
	"github.com/spf13/cobra"
)

var mycommandCmd = MustNewCommand(commands.MycommandMetadata, runMycommand)

func init() {
	MustAddToRoot(mycommandCmd)
}

func runMycommand(cmd *cobra.Command, args []string) error {
	cfg := mycommand.Config{
		Option1: getConfigValueWithFlags[string](cmd, "opt1", config.KeyAppMycommandOption1),
		Option2: getConfigValueWithFlags[string](cmd, "opt2", config.KeyAppMycommandOption2),
	}
	return mycommand.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
```

### Step 6: Add Tests

Create `internal/mycommand/mycommand_test.go`:

```go
package mycommand

import (
	"bytes"
	"testing"
)

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name       string
		cfg        Config
		wantOutput string
		wantErr    bool
	}{
		{
			name: "Successful execution",
			cfg: Config{
				Option1: "value1",
				Option2: "value2",
			},
			wantErr: false,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outBuf := &bytes.Buffer{}
			executor := NewExecutor(tt.cfg, outBuf)

			err := executor.Execute()

			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
```

### Step 7: Validate and Test

```bash
# Validate ultra-thin pattern
task validate:commands

# Run tests
task test

# Run all checks
task check
```

### Step 8: Update Documentation

- Add command usage to `README.md`
- Add entry to `CHANGELOG.md` under `[Unreleased]`

## Adding a Configuration Option

To add a configuration option to an **existing** command:

### Step 1: Add to Config Options

Edit `internal/config/commands/<command>_config.go`:

```go
func CommandOptions() []config.ConfigOption {
	return []config.ConfigOption{
		// Existing options...
		{
			Key:          "app.command.new_option",
			DefaultValue: "default",
			Description:  "Description of the new option",
			Type:         "string",
			Example:      "example",
			Validation:   nil, // Optional: add validation function
		},
	}
}
```

### Step 2: Regenerate Constants

```bash
task generate:config:key-constants
```

### Step 3: Update Command Config Struct

Edit `internal/<command>/<command>.go`:

```go
type Config struct {
	// Existing fields...
	NewOption string
}
```

### Step 4: Wire in Command

Edit `cmd/<command>.go`:

```go
cfg := command.Config{
	// Existing fields...
	NewOption: getConfigValueWithFlags[string](cmd, "new-option", config.KeyAppCommandNewOption),
}
```

Update `FlagOverrides` in metadata if needed.

### Step 5: Update Tests

Add test cases covering the new option.

## Testing Requirements

### Coverage Requirements

| Package Type | Minimum Coverage | Target Coverage |
|-------------|------------------|-----------------|
| Overall | 80% | 85%+ |
| `cmd/*` | 80% | 90%+ |
| `internal/config` | 80% | 90%+ |
| `internal/logger` | 80% | 90%+ |
| Other packages | 70% | 80%+ |

### Testing Principles

Follow [ADR-003](docs/adr/003-dependency-injection-over-mocking.md):

1. **Use dependency injection** over mocking frameworks
2. **Inject concrete implementations** via constructors
3. **Use table-driven tests** for multiple scenarios
4. **Follow AAA pattern**: Arrange (Setup) → Act (Execute) → Assert

### Example Test Structure

```go
func TestFeature(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid input",
			input:    "test",
			expected: "test_processed",
			wantErr:  false,
		},
		{
			name:    "invalid input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			// (arrange test data, dependencies)

			// EXECUTE
			got, err := ProcessFeature(tt.input)

			// ASSERT
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if got != tt.expected {
					t.Errorf("got %v, want %v", got, tt.expected)
				}
			}
		})
	}
}
```

### Running Tests

```bash
# Run all tests with coverage
task test

# Run tests with race detection
task test:race

# View detailed coverage
task test:coverage:text

# Generate HTML coverage report
task test:coverage:html

# Run integration tests
task test:integration
```

## Code Quality Standards

### Formatting

- **Always use `goimports`** (via `task format`)
- Follow standard Go formatting conventions
- Keep line length reasonable (~120 characters)

### Linting

All code must pass:
- `go vet ./...`
- `golangci-lint run`

Common issues to avoid:
- Unused variables or imports
- Error handling without logging
- Missing godoc comments on exported functions
- Ineffective assignments

### Logging

Use **structured logging with Zerolog**:

```go
import "github.com/rs/zerolog/log"

// Good: Structured logging
log.Info().
    Str("user", username).
    Int("attempts", count).
    Msg("User login successful")

// Bad: Unstructured logging
log.Info().Msg("User " + username + " login successful")

// Bad: Using fmt.Println
fmt.Println("User login")  // ❌ Never use this
```

### Error Handling

```go
// Good: Wrapped errors with context
if err != nil {
    log.Error().Err(err).Str("file", path).Msg("Failed to read file")
    return fmt.Errorf("failed to read file %s: %w", path, err)
}

// Bad: Generic errors
if err != nil {
    return err
}
```

### Configuration

**Never hardcode config keys:**

```go
// Good: Type-safe constants
message := viper.GetString(config.KeyAppPingOutputMessage)

// Bad: Hardcoded strings
message := viper.GetString("app.ping.output_message")  // ❌
```

**Never call `viper.SetDefault()` directly:**

```go
// Bad: Scattered SetDefault calls
viper.SetDefault("app.my.value", "default")  // ❌ Fails check-defaults

// Good: Use the registry
// Add to internal/config/commands/<command>_config.go
```

## Submitting Changes

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <concise summary>

- <bullet point detail 1>
- <bullet point detail 2>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding or updating tests
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `perf`: Performance improvement
- `build`: Changes to build system or dependencies
- `ci`: CI configuration changes
- `chore`: Other changes that don't modify src or test files

**Examples:**

```
feat: add JSON output format to docs command

- Implemented JSON marshaling for config documentation
- Added --format flag with json/yaml/markdown options
- Updated tests to cover JSON format
- Added example to README
```

```
fix: correct color rendering in non-TTY environments

- Fixed color detection logic in logger
- Added fallback for when TERM is not set
- Improved test coverage for color scenarios
```

### Pull Request Process

1. **Ensure all checks pass:**
   ```bash
   task check
   ```

2. **Update CHANGELOG.md:**
   Add your changes under the `[Unreleased]` section:
   ```markdown
   ## [Unreleased]

   ### Added
   - New feature description

   ### Fixed
   - Bug fix description
   ```

3. **Create a pull request:**
   - Use a clear, descriptive title
   - Reference any related issues
   - Describe what changed and why
   - Include screenshots for UI changes

4. **Address review feedback:**
   - Respond to all comments
   - Make requested changes
   - Re-run `task check` after changes

## Code Review Process

### What Reviewers Look For

1. **Architectural Compliance:**
   - ✅ Commands are ultra-thin (~20-30 lines)
   - ✅ No direct `viper.SetDefault()` calls
   - ✅ Configuration uses generated constants
   - ✅ Business logic is in `internal/` packages

2. **Code Quality:**
   - ✅ All tests pass with adequate coverage
   - ✅ Code is formatted (`task format`)
   - ✅ No linting issues (`task lint`)
   - ✅ Proper error handling and logging

3. **Testing:**
   - ✅ New features have tests
   - ✅ Bug fixes have regression tests
   - ✅ Coverage meets requirements

4. **Documentation:**
   - ✅ Public functions have godoc comments
   - ✅ README updated if needed
   - ✅ CHANGELOG.md updated
   - ✅ Complex logic has explanatory comments

### Review Timeline

- Initial review: Within 2-3 business days
- Follow-up reviews: Within 1-2 business days
- Merge: After approval and all checks pass

## Questions or Need Help?

- **Issues:** Open an issue on GitHub for bugs or feature requests
- **Discussions:** Use GitHub Discussions for questions
- **Documentation:** Check `docs/adr/` for architectural guidance

Thank you for contributing to ckeletin-go!
