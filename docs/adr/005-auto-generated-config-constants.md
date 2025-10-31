# ADR-005: Auto-Generated Configuration Constants

## Status
Accepted

## Context

Using string literals for config keys leads to:
- Typos causing runtime errors
- No compile-time validation
- Difficult refactoring
- Hard to discover available keys

Example problem:
```go
// Typo won't be caught until runtime
value := viper.Get("app.ping.output_messge") // Wrong!
```

## Decision

Auto-generate compile-time constants from the configuration registry:

```bash
go run scripts/generate-config-constants.go
```

Generates `internal/config/keys_generated.go`:
```go
const (
    KeyAppLogLevel            = "app.log_level"
    KeyAppPingOutputMessage   = "app.ping.output_message"
    KeyAppPingOutputColor     = "app.ping.output_color"
)
```

Usage:
```go
// Compile-time safe
message := getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage)

// Typo caught at compile time
message := getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessge) // Error!
```

## Consequences

### Positive
- Compile-time type safety
- IDE autocomplete for config keys
- Refactoring-friendly
- Self-documenting
- Catches typos early

### Negative
- Code generation step required
- Must re-run after config changes

### Mitigations
- Pre-commit hook validates constants are up-to-date
- Simple generation script
- Documented in contributing guide

## Workflow

1. Add new config option to registry
2. Run `task generate:config:key-constants` (or `go run scripts/generate-config-constants.go`)
3. Use generated constant in code
4. Pre-commit hook validates consistency automatically

The `task check-constants` command can be run manually to verify constants are current.

## References
- `scripts/generate-config-constants.go` - Generator
- `scripts/check-constants.sh` - Pre-commit validation script
- `internal/config/keys_generated.go` - Generated constants
- `.lefthook.yml` - Pre-commit validation
- `Taskfile.yml` - `task check-constants` command
