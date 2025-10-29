# ADR-006: Structured Logging with Zerolog

## Status
Accepted

## Context

Logging is essential for debugging and monitoring. Requirements:
- Structured logging for machine parsing
- Different log levels
- Good performance
- Easy to use
- Flexible output (console, JSON)

## Decision

Use **Zerolog** for structured logging with:
- Centralized initialization in `internal/logger`
- Console-friendly output in development
- Context-rich log messages
- Log sanitization for security

### Implementation

```go
// internal/logger/logger.go
func Init(out io.Writer) error {
    log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: out}).
        With().Timestamp().Logger()
    return nil
}

// Usage throughout codebase
log.Info().Str("config_file", path).Msg("Loading configuration")
log.Error().Err(err).Str("key", key).Msg("Invalid config value")
```

### Log Sanitization

```go
// internal/logger/sanitize.go
SanitizeLogString(s)  // Truncates long strings
SanitizePath(p)       // Sanitizes file paths
SanitizeError(err)    // Sanitizes error messages
```

## Consequences

### Positive
- Structured, machine-parseable logs
- Excellent performance (fastest Go logger)
- Rich context without printf-style formatting
- Easy to filter and search logs
- Security: automatic value sanitization

### Negative
- Different API than standard library log
- Need to sanitize sensitive data

### Mitigations
- Sanitization helpers prevent data leaks
- Centralized logger initialization
- Clear examples in codebase

## Related Decisions
- Log sanitization prevents injection attacks
- MaxLogLength prevents memory exhaustion
- Flexible output for testing (io.Writer)

## References
- `internal/logger/logger.go` - Initialization
- `internal/logger/sanitize.go` - Security helpers
- `internal/logger/logger_bench_test.go` - Performance tests
