# ADR-002: Centralized Configuration Registry

## Status
Accepted

## Context

In typical Cobra/Viper applications, configuration defaults are scattered:
- `viper.SetDefault()` calls in init() functions
- Different command files setting overlapping defaults
- No single source of truth for configuration
- Difficult to generate documentation
- Hard to validate all config options

Problems this causes:
- Configuration drift and inconsistencies
- Duplicate default definitions
- Missing or incorrect documentation
- No compile-time safety for config keys
- Difficult to understand all available options

## Decision

We implement a **centralized configuration registry** where:

1. **Single Source of Truth**: All config options defined in one place
2. **Self-Registration**: Config providers register themselves via init()
3. **Type-Safe Keys**: Auto-generated constants for all config keys
4. **Documentation Generation**: Auto-generate docs from registry
5. **Validation**: Validate all config against registry

### Architecture

```
internal/config/
├── registry.go              # Central registry
├── command_options.go       # ConfigOption type definition
├── core_options.go          # App-wide options (logging, etc.)
├── keys_generated.go        # Auto-generated const keys
├── commands/
│   ├── ping_config.go      # Ping command config (self-registers)
│   └── docs_config.go      # Docs command config (self-registers)
└── validator/
    └── validator.go        # Registry-based validation
```

### Usage

```go
// internal/config/commands/ping_config.go
func init() {
    config.RegisterOptionsProvider(PingOptions)
}

func PingOptions() []config.ConfigOption {
    return []config.ConfigOption{
        {
            Key:          "app.ping.output_message",
            DefaultValue: "Pong!",
            Description:  "Message to display",
            Type:         "string",
        },
    }
}

// cmd/root.go (initialization)
config.SetDefaults() // Applies all registered defaults to Viper

// cmd/ping.go (usage)
message := getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage)
```

## Consequences

### Positive

- **Single Source of Truth**: All config in one place
- **Consistency**: Guaranteed consistent defaults
- **Documentation**: Auto-generate accurate docs
- **Type Safety**: Compile-time checks with generated constants
- **Validation**: Registry-based config validation
- **Discoverability**: Easy to find all options
- **No Scattered SetDefault**: Prevents config drift

### Negative

- **Centralization Overhead**: All options must be registered
- **Code Generation Dependency**: Requires running generation script
- **Learning Curve**: Developers must understand registry pattern

### Mitigations

- **Validation Script**: `scripts/check-defaults.sh` prevents unauthorized SetDefault calls
- **Auto-Generation**: `scripts/generate-config-constants.go` generates type-safe keys
- **Documentation**: Clear examples and conventions guide
- **Pre-commit Hooks**: Automatic validation before commit

## Implementation Details

### ConfigOption Structure

```go
type ConfigOption struct {
    Key          string      // "app.ping.output_message"
    DefaultValue interface{} // "Pong!"
    Description  string      // "Message to display"
    Type         string      // "string"
    Required     bool        // false
    Example      string      // "Hello World"
    EnvVar       string      // Computed automatically
}
```

### Registration Pattern

```go
var optionsProviders []func() []ConfigOption

func RegisterOptionsProvider(provider func() []ConfigOption) {
    optionsProviders = append(optionsProviders, provider)
}

func Registry() []ConfigOption {
    options := CoreOptions() // App-wide options
    for _, provider := range optionsProviders {
        options = append(options, provider()...)
    }
    return options
}
```

### Key Generation

```bash
go run scripts/generate-config-constants.go
```

Generates `internal/config/keys_generated.go`:
```go
const (
    KeyAppLogLevel            = "app.log_level"
    KeyAppPingOutputMessage   = "app.ping.output_message"
    // ... all config keys
)
```

## Validation

The registry enables comprehensive validation:

```bash
task check-defaults  # Ensure no unauthorized viper.SetDefault() calls
```

```go
// Validate all config values against limits
errs := config.ValidateAllConfigValues(viper.AllSettings())

// Check for unknown keys
unknownKeys := findUnknownKeys(settings, knownKeys)
```

## Related ADRs

- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-thin commands rely on centralized config
- [ADR-005](005-auto-generated-config-constants.md) - Type-safe constants from registry

## References

- `internal/config/registry.go` - Registry implementation
- `internal/config/commands/` - Self-registering config providers
- `scripts/check-defaults.sh` - Validation script
- `scripts/generate-config-constants.go` - Key generation
- `docs/configuration.md` - Auto-generated documentation
