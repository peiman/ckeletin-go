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

## Enforcement

Security validation operates at two levels:

**1. Runtime Validation (Application Load)**
- `cmd/root.go` calls `config.ValidateConfigFileSecurity()` during initialization
- Checks file permissions, size limits before loading
- Application fails fast with clear error messages
- Cannot bypass via environment variables or flags

**2. Static Validation** (CI/Pre-commit)
```bash
task validate:security  # Validate security patterns in codebase
task check             # Includes security validation
```

**What Gets Validated:**

| Check | Description | Location |
|-------|-------------|----------|
| Constants defined | Security limit constants exist | `internal/config/limits.go` |
| Validation called | Security validation invoked during init | `cmd/root.go` |
| Functions exist | Security validation functions present | `internal/config/security.go` |
| Test coverage | Error scenarios tested | `test/integration/error_scenarios_test.go` |

**3. Integration Tests**
- `test/integration/error_scenarios_test.go` verifies:
  - World-writable file detection
  - Oversized file rejection
  - Invalid config value handling
  - Clear error messages with remediation

**4. Integration**
- **Runtime**: Validation during config load (always)
- **Local**: Part of `task check` (before commits)
- **CI**: Runs in quality gate pipeline

## References
- `internal/config/security.go` - Permission checks
- `internal/config/limits.go` - Value size limits
- `test/integration/error_scenarios_test.go` - Security tests
- `scripts/validate-security-patterns.sh` - Static validation script
