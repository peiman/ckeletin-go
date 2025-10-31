# ADR-001: Ultra-Thin Command Pattern

## Status
Accepted

## Context

When building CLI applications with Cobra, there's a tendency for command files to become bloated with:
- Business logic mixed with CLI framework code
- Direct viper.SetDefault() calls scattered throughout
- Tight coupling between commands and their dependencies
- Difficulty in testing business logic separately from CLI code

This leads to:
- Poor separation of concerns
- Difficult unit testing
- Code duplication
- Hard-to-maintain command files

## Decision

We adopt an **ultra-thin command pattern** where command files in `cmd/` are kept to ~20-30 lines and serve only as:
1. **Thin wrappers** that glue together the CLI framework and business logic
2. **Configuration retrievers** using `getConfigValueWithFlags[T]()`
3. **Dependency injectors** passing interfaces to business logic

All actual logic lives in `internal/` packages:
- `internal/config/commands/` - Command metadata and config options
- `internal/<command>/` - Business logic with executor pattern
- `cmd/<command>.go` - Ultra-thin CLI wrapper (~30 lines)

Example structure:
```go
// cmd/ping.go (~30 lines)
var pingCmd = MustNewCommand(config.PingMetadata, runPing)

func runPing(cmd *cobra.Command, args []string) error {
    // 1. Retrieve config
    message := getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage)

    // 2. Create executor with dependencies
    executor := ping.NewExecutor(cfg, uiRunner, os.Stdout)

    // 3. Execute and return
    return executor.Execute()
}
```

## Consequences

### Positive

- **Separation of Concerns**: CLI code separate from business logic
- **Testability**: Business logic easily tested without Cobra
- **Reusability**: Business logic can be used in other contexts
- **Maintainability**: Small, focused command files
- **Consistency**: Enforced pattern across all commands
- **Readability**: Clear flow from CLI → business logic

### Negative

- **Learning Curve**: Developers must understand the pattern
- **Indirection**: Extra layer between CLI and logic
- **Boilerplate**: Helper functions needed (MustNewCommand, etc.)

### Mitigations

- **Documentation**: Clear examples in `cmd/README.md`
- **Validation Script**: `scripts/validate-command-patterns.sh` enforces pattern
- **Helpers**: `cmd/helpers.go` reduces boilerplate
- **Code Generation**: Future generators can create command skeletons

## Compliance Validation

Command files are validated to ensure they follow the ultra-thin pattern:

```bash
task validate:commands
```

This checks that command files:
- Use the helper functions (MustNewCommand, MustAddToRoot)
- Don't contain direct viper.SetDefault() calls
- Don't exceed reasonable line counts

## Related ADRs

- [ADR-002](002-centralized-configuration-registry.md) - Centralized configuration eliminates scattered SetDefault calls
- [ADR-003](003-dependency-injection-over-mocking.md) - Dependency injection enables testing without mocks

## References

- `cmd/README.md` - Detailed command pattern documentation
- `cmd/ping.go` - Reference implementation (~31 lines)
- `cmd/docs.go` - Subcommand example
- `cmd/helpers.go` - Pattern enforcement helpers
