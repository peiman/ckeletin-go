# ADR-001: Ultra-Thin Command Pattern

## Status
Accepted

## Context

### Framework Selection: Why Cobra?

ckeletin-go uses [Cobra](https://github.com/spf13/cobra) as the CLI framework.

**Why Cobra?**
- Industry standard for Go CLIs (used by kubectl, hugo, gh, docker cli, and many others)
- Excellent subcommand support with clean, hierarchical command structure
- Automatic help generation and flag parsing with sensible defaults
- Strong community support and battle-tested stability (10k+ GitHub stars, widely adopted)
- Native integration with Viper for configuration binding
- POSIX-compliant flags with support for shorthand, required flags, and persistent flags

**Alternatives Considered:**
- **urfave/cli**: Simpler but less feature-rich, weaker subcommand nesting, smaller ecosystem
- **kingpin**: Less active maintenance, smaller community, less documentation
- **kong**: Tag-based reflection approach offers less explicit control over command structure
- **Direct flag package**: Low-level, would require significant boilerplate for subcommands and help

Cobra's maturity, explicit command structure, and wide adoption align with our goal of building a maintainable, production-ready CLI application. Its integration with Viper enables the configuration patterns described in ADR-002.

### The Problem: Command Bloat

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

## Implementation Patterns

This section documents the specific implementation patterns that realize the ultra-thin command principle in practice.

### Command Metadata Pattern

**Problem**: Creating Cobra commands involves repetitive boilerplate (Use, Short, Long, flags) scattered across command files, leading to inconsistency and duplication.

**Solution**: Declare command metadata separately in `internal/config/commands/<command>_config.go` and use factory functions to construct commands from metadata.

**Structure**:
```go
// internal/config/commands/ping_config.go
var PingMetadata = config.CommandMetadata{
    Use:   "ping",
    Short: "Responds with a pong",
    Long:  `Description...`,
    ConfigPrefix: "app.ping",
    FlagOverrides: map[string]string{
        "app.ping.output_message": "message",
    },
}

// cmd/ping.go
var pingCmd = MustNewCommand(commands.PingMetadata, runPing)
```

**Benefits**:
- **Separation**: Command metadata separate from implementation
- **Consistency**: `MustNewCommand()` applies consistent patterns
- **Discoverability**: All metadata in `internal/config/commands/`
- **Testability**: Metadata can be validated independently
- **Auto-registration**: Flags registered automatically from ConfigPrefix

**Factory Functions** (`cmd/helpers.go`):
- `MustNewCommand(metadata, runE)` - Creates command, panics on error (for init())
- `NewCommand(metadata, runE)` - Returns `(*cobra.Command, error)` (for runtime)
- `MustAddToRoot(cmd)` - Adds to root and sets up config inheritance

### Executor Pattern

**Problem**: Business logic mixed with CLI framework code makes testing difficult and couples domain logic to Cobra.

**Solution**: Implement business logic in separate executor structs with `Execute() error` method, keeping CLI layer as thin delegation.

**Structure**:
```go
// internal/ping/ping.go
type Executor struct {
    cfg      Config
    uiRunner ui.UIRunner
    writer   io.Writer
}

func NewExecutor(cfg Config, uiRunner ui.UIRunner, writer io.Writer) *Executor {
    return &Executor{cfg: cfg, uiRunner: uiRunner, writer: writer}
}

func (e *Executor) Execute() error {
    // All business logic here
    // Framework-agnostic
    return nil
}

// cmd/ping.go
func runPing(cmd *cobra.Command, args []string) error {
    cfg := ping.Config{/* retrieve config */}
    return ping.NewExecutor(cfg, uiRunner, os.Stdout).Execute()
}
```

**Benefits**:
- **Framework Independence**: Business logic has zero Cobra dependencies
- **Testability**: Test `Execute()` without any CLI framework
- **Dependency Injection**: Dependencies explicit in constructor (see [ADR-003](003-dependency-injection-over-mocking.md))
- **Reusability**: Executor can be used in non-CLI contexts
- **Clear Separation**: CLI concerns vs business logic boundaries explicit

**Pattern Enforced By**:
- **ADR-009** (Layered Architecture): Business logic cannot import cmd/
- **ADR-001** (This ADR): Commands must be thin (~20-30 lines)
- Natural consequence: Logic must live elsewhere → Executor pattern emerges

**Validation**: `task validate:commands` checks command files stay thin, which implicitly requires delegation pattern.

## Related ADRs

- [ADR-002](002-centralized-configuration-registry.md) - Centralized configuration eliminates scattered SetDefault calls; command metadata uses ConfigPrefix
- [ADR-003](003-dependency-injection-over-mocking.md) - Dependency injection enables testing without mocks; executors use DI
- [ADR-009](009-layered-architecture-pattern.md) - Enforces business logic isolation from cmd/, making Executor pattern necessary

## References

- `cmd/README.md` - Detailed command pattern documentation
- `cmd/ping.go` - Reference implementation (~31 lines) showing metadata and executor patterns
- `cmd/helpers.go` - Factory functions (MustNewCommand, MustAddToRoot)
- `internal/ping/ping.go` - Executor pattern reference implementation
- `internal/config/commands/ping_config.go` - Command metadata example
- `cmd/docs.go` - Subcommand example
