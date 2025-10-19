# Configuration Patterns - ckeletin-go Best Practices

This project follows the **ckeletin-go scaffold** configuration patterns to ensure proper flag/config/environment variable precedence handling.

## Quick Reference

**Configuration Precedence**: `explicit flag > config file > environment variable > default`

**Core Pattern - getConfigValue**:
```go
// CORRECT: Proper precedence handling
dbName := getConfigValue[string](cmd, "db-name", "app.db.name")

// WRONG: Manual precedence logic
dbName := viper.GetString("app.db.name")
if cmd.Flags().Changed("db-name") {
    dbName, _ = cmd.Flags().GetString("db-name")
}
```

## The Problem

When using Cobra flags with Viper configuration, a common bug occurs:

**Scenario**: User sets `db.name: "mydb"` in config file, but CLI always uses empty string default.

**Root Cause**: Flag defaults always override config file values because Cobra can't distinguish between:
- User explicitly set `--db-name ""`
- User didn't set flag at all (should use config/env)

**The Bug**:
```go
// ❌ WRONG: This breaks config file precedence
cmd.Flags().StringP("db-name", "d", "", "Database name")
viper.BindPFlag("app.db.name", cmd.Flags().Lookup("db-name"))

// Later...
dbName := viper.GetString("app.db.name")  // Always "" even if config has value!
```

## The Solution: getConfigValue

```go
// Helper function in cmd/root.go
func getConfigValue[T any](cmd *cobra.Command, flagName string, viperKey string) T {
    var value T

    // Get value from viper (config file or env)
    if v := viper.Get(viperKey); v != nil {
        if typedValue, ok := v.(T); ok {
            value = typedValue
        }
    }

    // If flag was EXPLICITLY set, override
    if cmd.Flags().Changed(flagName) {
        // Type-specific flag getter...
        value = getFlagValue[T](cmd, flagName)
    }

    return value
}
```

**Usage in commands**:
```go
func runCommand(cmd *cobra.Command, args []string) error {
    // ✅ CORRECT: Respects precedence
    dbName := getConfigValue[string](cmd, "db-name", "app.db.name")
    port := getConfigValue[int](cmd, "port", "app.db.port")
    enabled := getConfigValue[bool](cmd, "enabled", "app.feature.enabled")
}
```

## Validation

Configuration patterns are validated automatically:

```bash
# Run validation
task validate-config-patterns

# Part of quality checks
task check  # Includes config pattern validation
```

The linter checks for:
1. ✅ No `viper.SetDefault()` in command files (only in `internal/config/registry.go`)
2. ✅ All `NewXConfig` accept `cmd *cobra.Command` parameter
3. ✅ No manual precedence logic (`viper.Get` + `cmd.Flags().Changed`)
4. ✅ runE functions use `getConfigValue`, not direct viper calls

## Key Principles

✅ **DO:**
- Use `getConfigValue[T](cmd, flagName, viperKey)` for all config
- Define all defaults in `internal/config/registry.go`
- Accept `cmd *cobra.Command` in all config constructors
- Bind flags to viper in `init()` functions

❌ **DON'T:**
- Use `viper.SetDefault()` in command files
- Manually check `cmd.Flags().Changed()` in runE functions
- Read flags directly without checking if they were set
- Skip the `cmd` parameter in config constructors

## Example: Correct Pattern

```go
// cmd/example.go

func init() {
    exampleCmd.Flags().StringP("name", "n", "", "Name (required)")
    exampleCmd.Flags().IntP("count", "c", 0, "Count")

    // Bind flags to viper
    viper.BindPFlag("app.example.name", exampleCmd.Flags().Lookup("name"))
    viper.BindPFlag("app.example.count", exampleCmd.Flags().Lookup("count"))

    // NO viper.SetDefault here!
}

func runExample(cmd *cobra.Command, args []string) error {
    // ✅ Use getConfigValue
    name := getConfigValue[string](cmd, "name", "app.example.name")
    count := getConfigValue[int](cmd, "count", "app.example.count")

    // ... use values
}
```

## References

- **ckeletin-go scaffold**: https://github.com/peiman/ckeletin-go
- **Validation script**: scripts/validate-configuration-patterns.sh
- **Configuration registry**: internal/config/registry.go
