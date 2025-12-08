# ADR-006: Structured Logging with Zerolog

## Status
Accepted (Updated: 2025-10-29 - Added dual logging, log rotation, sampling, and runtime adjustment)

## Context

Logging is essential for debugging and monitoring. Requirements:
- Structured logging for machine parsing
- Different log levels for different audiences
- Good performance
- Easy to use
- Flexible output (console, JSON, file)
- Developer-friendly console output
- Machine-readable detailed file logs

## Decision

Use **Zerolog** for structured logging with:
- Centralized initialization in `internal/logger`
- **Dual logging** support: console (user-friendly) + file (detailed JSON)
- Console-friendly output in development
- Context-rich log messages
- Log sanitization for security
- Level-based filtering per output

### Implementation

#### Basic Setup (Console Only)

```go
// internal/logger/logger.go
func Init(out io.Writer) error {
    // Creates console writer with appropriate filtering
    // Falls back to legacy single-output mode if file logging disabled
    return nil
}
defer logger.Cleanup()

// Usage throughout codebase
log.Info().Str("config_file", path).Msg("Loading configuration")
log.Error().Err(err).Str("key", key).Msg("Invalid config value")
log.Debug().Str("detail", value).Msg("Detailed debug info")
```

#### Dual Logging (Console + File)

```go
// Configuration via flags or config file
--log-console-level info      // INFO+ to console
--log-file-enabled            // Enable file logging
--log-file-path ./logs/app.log
--log-file-level debug        // DEBUG+ to file
```

**Result:**
- **Console:** User-friendly, colored, INFO+ messages
- **File:** Machine-parseable JSON, DEBUG+ messages with full context

#### FilteredWriter Pattern

```go
// internal/logger/filtered_writer.go
type FilteredWriter struct {
    Writer   io.Writer
    MinLevel zerolog.Level
}

func (w FilteredWriter) WriteLevel(level zerolog.Level, p []byte) (int, error) {
    if level >= w.MinLevel {
        return w.Writer.Write(p)
    }
    return len(p), nil // Filtered
}
```

This allows different log levels to different outputs:
- Console: INFO, WARN, ERROR (clean, readable)
- File: DEBUG, INFO, WARN, ERROR (detailed, structured)

### Log Sanitization

```go
// internal/logger/sanitize.go
SanitizeLogString(s)  // Truncates long strings, removes control chars
SanitizePath(p)       // Sanitizes file paths, hides usernames
SanitizeError(err)    // Sanitizes error messages
```

### Configuration Options

```yaml
# config.yaml
app:
  log_level: info              # Legacy (backward compatible)
  log:
    console_level: info        # Console log level
    file_enabled: true         # Enable file logging
    file_path: ./logs/app.log  # Log file path
    file_level: debug          # File log level
    color_enabled: auto        # Console colors (auto/true/false)
```

## Consequences

### Positive
- **Dual logging**: Clean console + detailed file logs
- **Performance**: 12% overhead for dual output, 0 allocations
- **Structured**: Machine-parseable JSON in files
- **Developer UX**: Human-friendly console with colors
- **Debugging**: DEBUG logs available in file without console noise
- **Audit trail**: Permanent record of all operations
- **Backward compatible**: Existing code works unchanged
- **Security**: Automatic value sanitization, secure file permissions (0600)

### Negative
- Disk space: Log files can grow (mitigated by rotation in future)
- Complexity: More configuration options
- I/O overhead: ~12% performance impact with dual output

### Mitigations
- Log rotation can be added via lumberjack integration
- Sanitization helpers prevent data leaks
- Centralized logger initialization
- Clear examples in codebase
- File logging is opt-in (disabled by default)
- Cleanup function ensures files are closed properly

## Performance

Benchmark results (see `internal/logger/filtered_writer_bench_test.go`):
- Single logger: 196 ns/op
- Dual logger: 220 ns/op (+12% overhead)
- FilteredWriter: 2-3 ns/op per write
- **Zero allocations** for all operations

## Related Decisions
- Log sanitization prevents injection attacks
- MaxLogLength prevents memory exhaustion
- Flexible output for testing (io.Writer)
- FilteredWriter enables per-output level control
- Secure file permissions (0600) prevent information disclosure

## References
- `internal/logger/logger.go` - Initialization and dual logging setup
- `internal/logger/filtered_writer.go` - Level-based filtering
- `internal/logger/sanitize.go` - Security helpers
- `internal/logger/logger_bench_test.go` - Performance tests
- `internal/logger/filtered_writer_bench_test.go` - Dual logging benchmarks
- `internal/logger/dual_logger_prototype.go` - Prototype implementation

## Examples

### Console Output (INFO+ level)
```
2025-10-29T01:35:41Z INF File logging enabled file_level=debug path=~/logs/app.log
2025-10-29T01:35:41Z INF Application started successfully
2025-10-29T01:35:42Z WRN Configuration file not found, using defaults
```

### File Output (DEBUG+ level, JSON)
```json
{"level":"info","path":"~/logs/app.log","file_level":"debug","time":"2025-10-29T01:35:41Z","message":"File logging enabled"}
{"level":"debug","time":"2025-10-29T01:35:41Z","message":"No config file found, using defaults"}
{"level":"debug","command":"ping","time":"2025-10-29T01:35:41Z","message":"Applying command-specific configuration"}
{"level":"info","time":"2025-10-29T01:35:41Z","message":"Application started successfully"}
```

## Advanced Features

### Log Rotation (Lumberjack)

Automatic log rotation prevents disk exhaustion:

```yaml
app:
  log:
    file_max_size: 100      # MB before rotation
    file_max_backups: 3     # Old files to keep
    file_max_age: 28        # Days to retain
    file_compress: true     # Gzip old logs
```

**Features:**
- Automatic rotation when file exceeds max size
- Keeps specified number of backup files
- Removes old logs after max age
- Optional gzip compression of rotated logs

### Log Sampling

Reduce log volume in high-traffic scenarios:

```yaml
app:
  log:
    sampling_enabled: true
    sampling_initial: 100      # Log first 100/sec
    sampling_thereafter: 10    # Then log 10/sec
```

**Use case:** During traffic spikes, log first 100 messages per second, then sample 1 in 10 thereafter.

### Runtime Level Adjustment

Change log levels without restarting:

```go
// Adjust console verbosity
logger.SetConsoleLevel(zerolog.DebugLevel)

// Adjust file verbosity
logger.SetFileLevel(zerolog.TraceLevel)

// Query current levels
consoleLevel := logger.GetConsoleLevel()
fileLevel := logger.GetFileLevel()
```

**Use case:** Enable debug logging temporarily for troubleshooting, then revert to info level.

## Configuration Reference

### Complete Configuration

```yaml
app:
  log_level: info              # Legacy (backward compatible)
  log:
    # Dual logging
    console_level: info        # Console log level
    file_enabled: true         # Enable file logging
    file_path: ./logs/app.log  # Log file path
    file_level: debug          # File log level
    color_enabled: auto        # Console colors (auto/true/false)

    # Log rotation (lumberjack)
    file_max_size: 100         # MB before rotation
    file_max_backups: 3        # Old files to keep
    file_max_age: 28           # Days to retain
    file_compress: false       # Gzip old logs

    # Log sampling
    sampling_enabled: false    # Enable sampling
    sampling_initial: 100      # First N/sec
    sampling_thereafter: 100   # Sample thereafter
```

### Command-Line Flags

```bash
# Basic flags
--log-level info
--log-console-level info
--log-file-enabled
--log-file-path ./logs/app.log
--log-file-level debug
--log-color auto

# Rotation flags
--log-file-max-size 100
--log-file-max-backups 3
--log-file-max-age 28
--log-file-compress

# Sampling flags
--log-sampling-enabled
--log-sampling-initial 100
--log-sampling-thereafter 10
```

## Enforcement

Structured logging is enforced through linter rules and architectural patterns:

**1. Output Pattern Validation** (ADR-012)
```bash
task validate:output  # Checks business logic uses logger, not fmt.Print
```
- Detects `fmt.Print*` usage in `internal/*` packages
- Business logic must use `log.Info()`, `log.Error()`, etc.

**2. Linter Integration**
- golangci-lint configured with rules discouraging direct printing
- `forbidigo` linter can flag `fmt.Print*` in production code
- Run via `task lint`

**3. Centralized Initialization**
- Logger initialized in `internal/logger/logger.go`
- Global `log` variable configured with correct outputs
- Cleanup function ensures resources released

**4. Code Organization**
- Infrastructure layer pattern makes logger natural choice
- Direct stdout writes blocked by output validation
- Structured logging is path of least resistance

**5. Integration**
- **Local**: `task lint` checks linter rules
- **CI**: Full linting in quality pipeline
- **Output validation**: `task validate:output` catches fmt.Print usage

**Why No Dedicated `task validate:logging`:**
Output pattern validation (ADR-012) already catches direct printing. Adding a separate logging validator would be redundant. The linter + output validation provide sufficient enforcement.
