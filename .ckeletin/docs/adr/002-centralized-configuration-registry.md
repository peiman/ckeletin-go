# ADR-002: Centralized Configuration Registry

## Status
Accepted

## Context

### Configuration Library: Why Viper?

The centralized configuration registry is built on [Viper](https://github.com/spf13/viper), the de facto standard configuration library for Go applications.

**Why Viper?**
- **Multi-source configuration precedence** with sensible defaults (CLI flags > environment variables > config file > default values)
- **Automatic environment variable mapping** (e.g., `app.log_level` → `CKELETIN_APP_LOG_LEVEL`)
- **Multiple file format support** (YAML, JSON, TOML, HCL, etc.) without code changes
- **Type-safe getters** with fallback support (GetString, GetInt, GetDuration, GetBool, etc.)
- **Native integration with Cobra** for seamless flag binding and command configuration
- **Well-maintained and widely adopted** (used by Hugo, Docker CLI, Kubernetes tools, and 25k+ GitHub stars)
- **Live configuration reloading** capability (watchConfig for hot-reload scenarios)

**Alternatives Considered:**
- **envconfig**: Environment variables only, no file or flag support, loses CLI arg precedence
- **koanf**: More features (e.g., dot-notation access, merge strategies) but added complexity for minimal benefit, smaller ecosystem
- **viper alternatives (cleanenv, etc.)**: Less mature, smaller communities, fewer integrations
- **Manual configuration**: High maintenance burden, would reinvent precedence rules and type conversion

Viper's multi-source precedence model and Cobra integration make it the natural choice for a CLI application. The 12-factor app principle of configuration via environment variables is built-in, while still supporting config files for complex setups.

### The Problem: Scattered Configuration

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
task validate:defaults  # Ensure no unauthorized viper.SetDefault() calls
```

```go
// Validate all config values against limits
errs := config.ValidateAllConfigValues(viper.AllSettings())

// Check for unknown keys
unknownKeys := findUnknownKeys(settings, knownKeys)
```

## Implementation Patterns

This section documents how commands consume configuration from the centralized registry in a type-safe manner.

### Type-Safe Config Consumption Pattern

**Problem**: Direct `viper.Get*()` calls scattered throughout command files create fragility - typos in string keys fail at runtime, no type safety, and difficult refactoring.

**Solution**: Commands use helper functions with generated constants and pass configuration as typed structs to business logic.

**Structure**:
```go
// cmd/ping.go - Retrieve config with helper
func runPing(cmd *cobra.Command, args []string) error {
    cfg := ping.Config{
        Message: getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage),
        Color:   getConfigValueWithFlags[string](cmd, "color", config.KeyAppPingOutputColor),
        UI:      getConfigValueWithFlags[bool](cmd, "ui", config.KeyAppPingUi),
    }
    return ping.NewExecutor(cfg, uiRunner, os.Stdout).Execute()
}

// internal/ping/ping.go - Config struct in business logic
type Config struct {
    Message string
    Color   string
    UI      bool
}
```

**Benefits**:
- **Type Safety**: Generic helper `getConfigValueWithFlags[T]()` provides compile-time type checking
- **Generated Constants**: Uses `config.Key*` constants (auto-generated from registry via [ADR-005](005-auto-generated-config-constants.md))
- **Flag Integration**: Automatically checks command flags before falling back to Viper
- **Centralized Retrieval**: Config assembled in one place (command file) before passing to business logic
- **Refactor-Friendly**: Renaming config keys only requires updating registry and regenerating constants
- **Framework Independence**: Business logic receives plain structs, no Viper/Cobra dependencies

**Helper Function** (`cmd/helpers.go`):
```go
func getConfigValueWithFlags[T any](cmd *cobra.Command, flagName, configKey string) T {
    // 1. Check if flag was explicitly set
    if cmd.Flags().Changed(flagName) {
        val, _ := cmd.Flags().GetString(flagName) // or GetBool, etc.
        return convertToType[T](val)
    }
    // 2. Fall back to Viper (env vars, config file, defaults)
    return viper.Get(configKey).(T)
}
```

**Config Struct Pattern**:
- Config structs live in business logic packages (`internal/ping/ping.go`)
- Commands create config structs from registry keys
- Executors receive config as constructor parameters
- No `viper` or `cobra` imports in business logic

**Enforcement**:

```bash
task validate:config-consumption  # Checks type-safe config pattern
```

Validation script checks:
- ✅ No direct `viper.Get*()` calls in command files (except whitelisted: helpers.go, root.go, flags.go)
- ✅ Commands use `getConfigValueWithFlags[T]()` helper for type-safe retrieval
- ✅ Config passed as typed structs to executors

Additionally, ADR-001's `task validate:commands` indirectly enforces this pattern by requiring commands to be thin - scattering config retrieval throughout business logic would violate line count limits.

## Related ADRs

- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-thin commands rely on centralized config; executor pattern receives config structs
- [ADR-005](005-auto-generated-config-constants.md) - Type-safe constants from registry enable type-safe consumption
- [ADR-009](009-layered-architecture-pattern.md) - Business logic isolation means no direct Viper access, config passed as structs

## References

- `internal/config/registry.go` - Registry implementation
- `internal/config/commands/` - Self-registering config providers
- `cmd/helpers.go` - `getConfigValueWithFlags[T]()` helper function
- `cmd/ping.go` - Reference implementation showing config consumption pattern
- `internal/ping/ping.go` - Config struct example in business logic
- `scripts/check-defaults.sh` - Validation script
- `scripts/generate-config-constants.go` - Key generation
- `docs/configuration.md` - Auto-generated documentation
