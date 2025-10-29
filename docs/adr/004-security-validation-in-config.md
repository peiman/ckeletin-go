# ADR-004: Security Validation in Configuration

## Status
Accepted

## Context

Configuration files can be attack vectors:
- World-writable configs allow unauthorized modification
- Large config files can cause DoS
- Excessively long string values can exhaust memory
- Large arrays can cause performance issues

## Decision

Implement multi-layered security validation:

### 1. File Security
```go
// internal/config/security.go
ValidateConfigFilePermissions(path) // Prevents world-writable
ValidateConfigFileSize(path, 1MB)    // Prevents DoS
```

### 2. Value Limits
```go
// internal/config/limits.go
const (
    MaxStringValueLength = 10 * 1024  // 10 KB
    MaxSliceLength       = 1000       // 1000 elements
    MaxConfigFileSize    = 1 * 1024 * 1024 // 1 MB
)
```

### 3. Validation on Load
```go
// cmd/root.go
if err := config.ValidateConfigFileSecurity(path, config.MaxConfigFileSize); err != nil {
    return err
}
```

## Consequences

### Positive
- Prevents common security issues
- DoS attack prevention
- Clear error messages with remediation
- Defense in depth

### Negative
- Adds overhead to config loading
- May reject legitimate large configs

### Mitigations
- Configurable limits
- Clear error messages
- Documentation of limits

## References
- `internal/config/security.go` - Permission checks
- `internal/config/limits.go` - Value size limits
- `test/integration/error_scenarios_test.go` - Security tests
