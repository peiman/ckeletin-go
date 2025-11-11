# Testing Guide for ckeletin-go

This guide documents testing patterns, conventions, and best practices for the ckeletin-go project.

## Table of Contents

- [Testing Philosophy](#testing-philosophy)
- [Test Organization](#test-organization)
- [Writing Tests](#writing-tests)
- [Test Patterns](#test-patterns)
- [Platform-Specific Testing](#platform-specific-testing)
- [Anti-Patterns to Avoid](#anti-patterns-to-avoid)
- [Running Tests](#running-tests)

## Testing Philosophy

**ckeletin-go follows ADR-003: Dependency Injection Over Mocking**

Key principles:
1. **Dependency Injection**: Use interface-based DI instead of mocking frameworks
2. **Test Real Behavior**: Focus on observable behavior, not implementation details
3. **Simple Test Implementations**: Create simple test implementations of interfaces
4. **High Coverage**: Maintain 80%+ test coverage across all packages
5. **Integration Tests**: Complement unit tests with end-to-end integration tests

See [ADR-003](adr/003-testing-strategy.md) for full rationale.

## Test Organization

### File Structure

```
ckeletin-go/
├── cmd/                      # Command tests
│   ├── ping_test.go         # Unit tests for ping command
│   └── root_test.go         # Root command and config tests
├── internal/
│   ├── config/
│   │   ├── config_test.go   # Config package tests
│   │   └── validator/
│   │       └── validator_test.go
│   ├── logger/
│   │   └── logger_test.go
│   └── testutil/            # Shared test utilities
│       ├── helpers.go       # Test helpers (platform skips, etc.)
│       └── helpers_test.go
└── test/
    └── integration/         # End-to-end integration tests
        ├── integration_test.go
        └── error_scenarios_test.go
```

### Naming Conventions

- **Test files**: `*_test.go` (co-located with source)
- **Benchmark files**: `*_bench_test.go` (for performance tests)
- **Integration tests**: `test/integration/*_test.go`
- **Test functions**: `TestFunctionName` or `TestFeature_Scenario`
- **Benchmark functions**: `BenchmarkOperationName`

### Test File Size Guidelines

- **Unit test files**: Keep under 400 lines
- **Integration test files**: Keep under 600 lines
- **If file exceeds limit**: Split into focused files (e.g., `root_config_test.go`, `root_init_test.go`)

## Writing Tests

### Table-Driven Tests (Recommended Pattern)

Use table-driven tests for functions with multiple input/output scenarios:

```go
func TestStringDefault(t *testing.T) {
    tests := []struct {
        name  string
        input interface{}
        want  string
    }{
        {
            name:  "Nil value",
            input: nil,
            want:  "",
        },
        {
            name:  "String value",
            input: "test",
            want:  "test",
        },
        {
            name:  "Integer value",
            input: 42,
            want:  "42",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := stringDefault(tt.input)
            if got != tt.want {
                t.Errorf("stringDefault() = %q, want %q", got, tt.want)
            }
        })
    }
}
```

**Why this pattern:**
- Clear test case names
- Easy to add new scenarios
- Excellent debuggability (can run individual cases)
- Self-documenting test intent

### SETUP-EXECUTION-ASSERTION Pattern

Structure tests in three clear phases:

```go
func TestFeature(t *testing.T) {
    // SETUP PHASE
    viper.Reset()
    testLogger, cleanup := setupTestLogger(t)
    defer cleanup()

    // EXECUTION PHASE
    result, err := FeatureUnderTest()

    // ASSERTION PHASE
    if err != nil {
        t.Errorf("FeatureUnderTest() unexpected error: %v", err)
    }
    if result != expected {
        t.Errorf("FeatureUnderTest() = %v, want %v", result, expected)
    }
}
```

**Benefits:**
- Clear separation of concerns
- Easy to understand test flow
- Easier to debug failures

### Using Testify for Assertions

**Prefer testify/assert for cleaner, more readable assertions:**

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestFeature(t *testing.T) {
    // Use assert for test logic assertions
    result := doSomething()
    assert.Equal(t, expected, result, "result should match expected value")
    assert.NoError(t, err, "should not return error")
    assert.Contains(t, output, "expected text")

    // Use require for setup/preconditions (stops test on failure)
    cfg, err := loadConfig()
    require.NoError(t, err, "setup: failed to load config")
    require.NotNil(t, cfg, "setup: config should not be nil")
}
```

**assert vs require:**
- `assert.*`: Continues test after failure (reports all failures)
- `require.*`: Stops test immediately on failure (for setup/preconditions)

### Test Helper Functions

Always mark helper functions with `t.Helper()`:

```go
// setupTestEnvironment prepares a test environment with config and logger.
func setupTestEnvironment(t *testing.T) (cleanup func()) {
    t.Helper()  // Mark as helper so stack traces skip this function

    // Save original state
    origLogger := log.Logger
    viper.Reset()

    // Setup test state
    log.Logger = zerolog.New(os.Stdout).Level(zerolog.DebugLevel)

    // Return cleanup function
    return func() {
        log.Logger = origLogger
        viper.Reset()
    }
}

// Usage in tests:
func TestSomething(t *testing.T) {
    cleanup := setupTestEnvironment(t)
    defer cleanup()

    // Test logic...
}
```

### Temporary Files and Directories

**Always use `t.TempDir()` for temporary directories:**

```go
func TestConfigFile(t *testing.T) {
    tmpDir := t.TempDir()  // Automatically cleaned up after test
    configFile := filepath.Join(tmpDir, "config.yaml")

    // Write test file
    err := os.WriteFile(configFile, []byte("test: content"), 0600)
    require.NoError(t, err)

    // Test logic...
    // No need to manually clean up - t.TempDir() handles it
}
```

**Why `t.TempDir()`:**
- Automatic cleanup
- Unique directory per test (parallel-safe)
- No resource leaks

### Test Fixtures

Store test data files in `testdata/` directories:

```
cmd/
├── ping.go
├── ping_test.go
└── testdata/
    ├── valid.yaml
    ├── invalid.yaml
    └── partial.yaml
```

**Access fixtures:**
```go
func TestLoadConfig(t *testing.T) {
    content, err := os.ReadFile("testdata/valid.yaml")
    require.NoError(t, err)
    // Test with content...
}
```

## Test Patterns

### Dependency Injection Pattern

**Production code:**
```go
// Define interface
type UIRunner interface {
    RunUI(message, color string) error
}

// Constructor accepts interface
func NewExecutor(cfg Config, uiRunner UIRunner, writer io.Writer) *Executor {
    return &Executor{
        cfg:      cfg,
        uiRunner: uiRunner,
        writer:   writer,
    }
}
```

**Test code:**
```go
// Simple test implementation
type MockUIRunner struct {
    CalledWithMessage string
    CalledWithColor   string
    ReturnError       error
}

func (m *MockUIRunner) RunUI(message, col string) error {
    m.CalledWithMessage = message
    m.CalledWithColor = col
    return m.ReturnError
}

// Use in tests
func TestExecutor(t *testing.T) {
    mockRunner := &MockUIRunner{}
    executor := NewExecutor(cfg, mockRunner, &bytes.Buffer{})

    err := executor.Execute()

    assert.NoError(t, err)
    assert.Equal(t, "expected message", mockRunner.CalledWithMessage)
}
```

**No mocking frameworks needed!**

### Parallel Test Execution

**Mark tests as parallel when they don't share state:**

```go
func TestFeatures(t *testing.T) {
    tests := []struct {
        name string
        // ...
    }{
        // test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()  // Safe because each subtest has isolated state

            // Test logic...
        })
    }
}
```

**When NOT to use `t.Parallel()`:**
- Tests modify global state (viper config, logger)
- Tests modify package-level variables
- Tests manipulate environment variables
- Tests access shared files without locking

**Make tests parallel-safe:**
- Use `t.TempDir()` for file operations
- Save and restore global state
- Use test-local variables

### Error Testing Pattern

**Test both success and error cases:**

```go
func TestValidateConfig(t *testing.T) {
    tests := []struct {
        name        string
        config      string
        wantErr     bool
        errContains string  // Check error message content
    }{
        {
            name:    "Valid config",
            config:  "app:\n  log_level: info\n",
            wantErr: false,
        },
        {
            name:        "Invalid YAML",
            config:      "app: [unclosed",
            wantErr:     true,
            errContains: "YAML",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateConfig(tt.config)

            if tt.wantErr {
                assert.Error(t, err)
                if tt.errContains != "" {
                    assert.Contains(t, err.Error(), tt.errContains)
                }
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

### Integration Test Pattern

**Integration tests verify end-to-end workflows:**

```go
// test/integration/integration_test.go

var binaryPath string

func TestMain(m *testing.M) {
    // Build test binary
    binaryName := "ckeletin-go-test"
    if runtime.GOOS == "windows" {
        binaryName += ".exe"
    }

    cmd := exec.Command("go", "build", "-o", binaryName, "../../main.go")
    if err := cmd.Run(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to build test binary: %v\n", err)
        os.Exit(1)
    }
    binaryPath = "./" + binaryName

    // Run tests
    code := m.Run()

    // Cleanup
    os.Remove(binaryPath)
    os.Exit(code)
}

func TestPingCommand(t *testing.T) {
    cmd := exec.Command(binaryPath, "ping", "--message", "Hello")
    var stdout bytes.Buffer
    cmd.Stdout = &stdout

    err := cmd.Run()

    exitCode := getExitCode(err)
    assert.Equal(t, 0, exitCode)
    assert.Contains(t, stdout.String(), "Hello")
}

// Helper for exit codes
func getExitCode(err error) int {
    if err == nil {
        return 0
    }
    if exitErr, ok := err.(*exec.ExitError); ok {
        return exitErr.ExitCode()
    }
    return -1
}
```

## Platform-Specific Testing

### Skipping Tests on Windows

**Use the testutil helpers for consistent platform skipping:**

```go
import "github.com/peiman/ckeletin-go/internal/testutil"

func TestFilePermissions(t *testing.T) {
    testutil.SkipOnWindowsWithReason(t, "file permissions require Unix")

    // Unix-specific test logic...
}
```

**Available helpers:**
- `testutil.SkipOnWindows(t)` - Skip on Windows with default message
- `testutil.SkipOnWindowsWithReason(t, reason)` - Skip with custom message
- `testutil.SkipOnNonWindows(t)` - Skip on non-Windows platforms
- `testutil.SkipOnPlatform(t, "darwin")` - Skip on specific platform

### Platform-Specific Test Cases

**For table-driven tests with platform-specific cases:**

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name          string
        input         string
        want          string
        skipOnWindows bool
    }{
        {
            name:  "Standard case",
            input: "test",
            want:  "result",
        },
        {
            name:          "Unix permissions",
            input:         "/path/file",
            want:          "secured",
            skipOnWindows: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if tt.skipOnWindows {
                testutil.SkipOnWindows(t)
            }

            // Test logic...
        })
    }
}
```

## Anti-Patterns to Avoid

### ❌ Don't Use panic() in Tests

```go
// BAD
func TestMain(m *testing.M) {
    if err := setup(); err != nil {
        panic("setup failed: " + err.Error())  // ❌ Don't panic
    }
    m.Run()
}

// GOOD
func TestMain(m *testing.M) {
    if err := setup(); err != nil {
        fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
        os.Exit(1)  // ✅ Clean exit
    }
    code := m.Run()
    cleanup()
    os.Exit(code)
}
```

### ❌ Don't Overuse t.Fatalf()

```go
// BAD - Stops test immediately, hides other issues
func TestMultipleOperations(t *testing.T) {
    err1 := operation1()
    if err1 != nil {
        t.Fatalf("op1 failed: %v", err1)  // ❌ Stops here
    }

    err2 := operation2()  // Never runs if op1 fails
    if err2 != nil {
        t.Fatalf("op2 failed: %v", err2)
    }
}

// GOOD - Reports all failures
func TestMultipleOperations(t *testing.T) {
    err1 := operation1()
    assert.NoError(t, err1, "operation1 should succeed")

    err2 := operation2()
    assert.NoError(t, err2, "operation2 should succeed")
    // Both failures reported!
}

// ACCEPTABLE - t.Fatalf() for setup only
func TestFeature(t *testing.T) {
    cfg, err := loadConfig()
    if err != nil {
        t.Fatalf("Setup failed: %v", err)  // ✅ OK for setup
    }

    // Test logic uses assert
    result := doWork(cfg)
    assert.NotNil(t, result)
}
```

### ❌ Don't Test Implementation Details

```go
// BAD - Tests internal state
func TestConfigLoading(t *testing.T) {
    LoadConfig("file.yaml")

    // ❌ Checking package-level variable (implementation detail)
    if configFileStatus != "loaded" {
        t.Error("configFileStatus should be 'loaded'")
    }
}

// GOOD - Tests observable behavior
func TestConfigLoading(t *testing.T) {
    err := LoadConfig("file.yaml")

    // ✅ Check public API and observable output
    assert.NoError(t, err)
    assert.Equal(t, "info", GetLogLevel())
}
```

### ❌ Don't Use Magic Numbers/Strings

```go
// BAD
func TestTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    // What does 5 seconds represent?
}

// GOOD
func TestTimeout(t *testing.T) {
    const testTimeout = 5 * time.Second  // Clearly named constant
    ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
    defer cancel()
}
```

### ❌ Don't Leave Commented-Out Tests

```go
// BAD - Dead code
/* func TestFeature(t *testing.T) {
    // This test is commented out because...
} */

// GOOD - Delete it or create a GitHub issue
// If the test is needed but currently broken, create an issue:
// TODO: Re-enable when issue #123 is fixed
func TestFeature(t *testing.T) {
    t.Skip("Temporarily disabled: waiting on issue #123")
    // Test code...
}
```

### ❌ Don't Create Brittle Tests

```go
// BAD - Tests break when code is refactored
func TestUserCreation(t *testing.T) {
    user := CreateUser("John", "Doe")

    // ❌ Depends on struct internals
    if user.firstName != "John" {
        t.Error("firstName not set")
    }
}

// GOOD - Tests public API
func TestUserCreation(t *testing.T) {
    user := CreateUser("John", "Doe")

    // ✅ Uses public methods
    assert.Equal(t, "John Doe", user.FullName())
}
```

### ❌ Don't Write Tests Without Context

```go
// BAD - Unclear what failed and why
if got != want {
    t.Errorf("got %v, want %v", got, want)
}

// GOOD - Clear context in error messages
if got != want {
    t.Errorf("ProcessConfig() returned unexpected value: got %v, want %v", got, want)
}

// BETTER - Use testify
assert.Equal(t, want, got, "ProcessConfig should return validated config")
```

## Running Tests

### Local Development

```bash
# Run all tests
task test

# Run tests with coverage
task test  # Coverage is automatic

# Run specific package tests
go test ./cmd/...
go test ./internal/config/...

# Run specific test
go test ./cmd -run TestPingCommand

# Run specific subtest
go test ./cmd -run TestPingCommand/WithCustomMessage

# Run with race detector
go test -race ./...

# Run benchmarks
task bench
```

### Before Committing

**MANDATORY: Run all quality checks:**

```bash
task check  # Runs: format, lint, tests, coverage, validation
```

**This checks:**
- Code formatting (goimports, gofmt)
- Linting (golangci-lint)
- All tests pass
- Coverage thresholds met (80%+)
- ADR compliance validation

### Integration Tests

```bash
# Run integration tests only
go test ./test/integration/...

# Skip integration tests (for quick iteration)
go test -short ./...
```

### Test Coverage

**Minimum coverage requirements:**

| Package Type | Minimum | Target |
|-------------|---------|--------|
| Overall | 80% | 85%+ |
| `cmd/*` | 80% | 90%+ |
| `internal/config` | 80% | 90%+ |
| `internal/logger` | 80% | 90%+ |
| Other packages | 70% | 80%+ |

**View coverage report:**
```bash
task test
# Coverage shown in output

# Generate HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Additional Resources

- [ADR-003: Testing Strategy](adr/003-testing-strategy.md) - Testing philosophy and rationale
- [ADR-001: Command Pattern](adr/001-command-pattern.md) - How to write thin, testable commands
- [Go Testing Package](https://pkg.go.dev/testing) - Official Go testing docs
- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify) - Assertion library docs
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) - Dave Cheney's guide

## Summary Checklist

Before submitting code, ensure your tests:

- [ ] Use table-driven tests for multiple scenarios
- [ ] Follow SETUP-EXECUTION-ASSERTION structure
- [ ] Use testify/assert for clean assertions
- [ ] Mark helper functions with `t.Helper()`
- [ ] Use `t.TempDir()` for temporary files
- [ ] Use `t.Parallel()` when safe
- [ ] Test both success and error cases
- [ ] Have clear, descriptive test names
- [ ] Have informative error messages
- [ ] Don't test implementation details
- [ ] Use `testutil` helpers for platform skips
- [ ] Meet coverage requirements (80%+)
- [ ] Pass `task check` before committing

**Happy testing!** 🧪
