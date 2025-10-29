# ADR-006: Structured Logging with Zerolog

## Status
Accepted (Updated: 2025-10-29 - Added dual logging support)

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
