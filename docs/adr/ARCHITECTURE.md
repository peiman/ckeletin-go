# ckeletin-go System Architecture

> **Quick Reference:** This document shows **WHAT** the system is and how components interact.
> For **WHY** decisions were made, see individual ADRs linked throughout this document.

**Last Updated:** 2025-11-04
**Status:** Living document (updated as architecture evolves)

---

## Table of Contents

1. [Overview](#overview)
2. [Architectural Layers](#architectural-layers)
3. [Component Structure](#component-structure)
4. [Initialization Sequence](#initialization-sequence)
5. [Configuration Flow](#configuration-flow)
6. [Command Execution Lifecycle](#command-execution-lifecycle)
7. [Testing Architecture](#testing-architecture)
8. [Development Workflow Integration](#development-workflow-integration)
9. [How ADRs Work Together](#how-adrs-work-together)
10. [Package Organization](#package-organization)
11. [Key Design Patterns](#key-design-patterns)

---

## Overview

**ckeletin-go** is a production-ready Go CLI application scaffold that demonstrates modern Go development practices. It provides a complete foundation for building command-line tools with:

- **Ultra-thin command layer** (20-30 lines per command)
- **Centralized configuration** with type-safe access
- **Structured logging** with dual output (console + file)
- **Interactive terminal UIs** using Bubble Tea
- **Automated validation** of architectural patterns
- **Cross-platform support** (Linux, macOS, Windows)

The architecture follows a **4-layer pattern** (Entry → Command → Business Logic → Infrastructure) with automated enforcement of dependency rules. See [ADR-009](009-layered-architecture-pattern.md) for complete details.

---

## Architectural Layers

See [ADR-009](009-layered-architecture-pattern.md) for the rationale, alternatives considered, and enforcement mechanisms for the layered architecture pattern.

```
┌──────────────────────────────────────────────────────────────┐
│                     CLI Entry Layer                          │
│                      (main.go)                               │
│  - Application bootstrap                                     │
│  - Root command execution                                    │
└────────────────────────┬─────────────────────────────────────┘
                         │
                         ▼
┌──────────────────────────────────────────────────────────────┐
│                   Command Layer (cmd/)                       │
│  - Ultra-thin command definitions (~20-30 lines) → ADR-001   │
│  - Cobra command setup                                       │
│  - Flag/argument parsing                                     │
│  - Delegation to business logic                              │
└────────────────────────┬─────────────────────────────────────┘
                         │
         ┌───────────────┼───────────────┬────────────┐
         ▼               ▼               ▼            ▼
┌──────────────┐  ┌──────────────┐  ┌──────────┐  ┌────────────┐
│  Business    │  │   Config     │  │  Logger  │  │    UI      │
│   Logic      │  │  Registry    │  │  (Zero)  │  │  (Tea)     │
│ (internal/*) │  │ (ADR-002/005)│  │ (ADR-006)│  │ (ADR-007)  │
│              │  │              │  │          │  │            │
│ - ping/      │  │ - registry   │  │ - setup  │  │ - models   │
│ - docs/      │  │ - keys_gen   │  │ - dual   │  │ - bubbletea│
│ - validators │  │ - loaders    │  │   output │  │ - lipgloss │
└──────────────┘  └──────────────┘  └──────────┘  └────────────┘
         │               │               │              │
         └───────────────┴───────────────┴──────────────┘
                         │
                         ▼
                  External Systems
              (Network, Filesystem, etc.)
```

**Layer Responsibilities:**

1. **CLI Entry (main.go)**
   - Bootstrap application
   - Execute root command
   - **Imports**: `cmd/` only
   - **Imported by**: Nothing (entry point)

2. **Command Layer (cmd/)**
   - Parse CLI input and bind flags
   - Validate arguments
   - Delegate to business logic (see [ADR-001](001-ultra-thin-command-pattern.md))
   - **Imports**: `internal/*`, Cobra framework
   - **Imported by**: Entry layer only
   - **Key Rule**: Only this layer can import Cobra

3. **Business Logic (internal/ping, internal/docs, etc.)**
   - Domain-specific functionality
   - Framework-independent implementations
   - **Imports**: Infrastructure layer, standard library
   - **Imported by**: Command layer
   - **Key Rules**:
     - ❌ No Cobra imports (framework independence)
     - ❌ No `cmd/` imports (prevents cycles)
     - ❌ Business packages isolated from each other

4. **Infrastructure (internal/config, internal/logger, internal/ui)**
   - Cross-cutting concerns
   - Shared services available to all layers
   - **Imports**: External libraries, standard library
   - **Imported by**: Command layer, Business logic layer
   - **Key Rules**:
     - ❌ Cannot import business logic
     - ❌ Cannot import `cmd/`

### Dependency Rules (Enforced by ADR-009)

See [ADR-009](009-layered-architecture-pattern.md) for complete rationale and alternatives considered.

**Allowed Dependencies:**
- ✅ Entry → Command
- ✅ Command → Business Logic
- ✅ Command → Infrastructure
- ✅ Business Logic → Infrastructure

**Forbidden Dependencies:**
- ❌ Business Logic → Command (would couple to CLI)
- ❌ Business Logic → Business Logic (packages must be isolated)
- ❌ Infrastructure → Business Logic (wrong direction)
- ❌ Infrastructure → Command (wrong direction)
- ❌ `internal/*` → Cobra (only `cmd/` uses framework)

**Example Violations Caught by Validation:**

```go
// ❌ VIOLATION: Business logic importing command layer
// internal/ping/executor.go
import "github.com/peiman/ckeletin-go/cmd"
// Error: Component business shouldn't depend on cmd

// ❌ VIOLATION: Business logic importing other business logic
// internal/ping/executor.go
import "github.com/peiman/ckeletin-go/internal/docs"
// Error: Component business shouldn't depend on internal/docs
```

**Enforcement:**

```bash
task validate:layering  # Runs go-arch-lint to check all rules
```

Configuration: `.go-arch-lint.yml` defines components and allowed dependencies.

**Maintenance Note:** When adding new commands (e.g., `internal/init/`), update `.go-arch-lint.yml` to include the new business logic package. See [ADR-009](009-layered-architecture-pattern.md) for details.

### Validation in Action

When you run `task validate:layering`, go-arch-lint checks all dependency rules and reports violations with clear error messages:

**Example 1: Business logic importing command layer**

```bash
$ task validate:layering
🔍 Validating layered architecture (ADR-009)...
✅ go-arch-lint installed successfully
Component business shouldn't depend on github.com/peiman/ckeletin-go/cmd in internal/ping/ping.go:9
❌ Layered architecture validation failed
```

This violation occurs when business logic tries to import from `cmd/`:
```go
// ❌ internal/ping/ping.go:9
import "github.com/peiman/ckeletin-go/cmd"
```

**Example 2: Business logic importing other business logic**

```bash
Component business shouldn't depend on github.com/peiman/ckeletin-go/internal/docs in internal/ping/ping.go:10
```

This violation occurs when business logic packages try to import each other:
```go
// ❌ internal/ping/ping.go:10
import "github.com/peiman/ckeletin-go/internal/docs"
```

**Fix:** Remove the forbidden import and refactor:
- Extract shared functionality to infrastructure layer (`internal/config`, `internal/logger`, etc.)
- Pass data as parameters between business logic packages
- Use dependency injection for shared services

For complete package organization details, see [Package Organization](#package-organization).

---

## Component Structure

### Core Components

```
ckeletin-go/
│
├── main.go                    # Entry point (root command execution)
│
├── cmd/                       # Command Layer → ADR-001
│   ├── root.go                # Root command setup, global flags, config init
│   ├── ping.go                # Ping command (example thin command)
│   ├── version.go             # Version command
│   ├── docs.go                # Docs command (config documentation)
│   └── template_command.go.example  # Command template
│
├── internal/                  # Private application code
│   │
│   ├── config/                # Configuration Management → ADR-002, ADR-005
│   │   ├── registry.go        # Config option definitions (SSOT)
│   │   ├── keys_generated.go  # Auto-generated type-safe constants
│   │   ├── loader.go          # Config loading logic
│   │   ├── validator.go       # Config validation → ADR-004
│   │   └── commands/          # Per-command config structs
│   │
│   ├── logger/                # Logging Infrastructure → ADR-006
│   │   ├── logger.go          # Logger setup and configuration
│   │   ├── console.go         # Console output (colored, human-friendly)
│   │   └── file.go            # File output (JSON, debug level)
│   │
│   ├── ui/                    # Terminal UI Components → ADR-007
│   │   ├── styles.go          # Lipgloss styles
│   │   └── models.go          # Bubble Tea models
│   │
│   ├── ping/                  # Ping Business Logic
│   │   ├── executor.go        # Ping execution logic
│   │   └── executor_test.go   # Tests → ADR-003
│   │
│   └── docs/                  # Documentation Generation
│       └── generator.go       # Config docs generator
│
├── test/                      # Integration Tests
│   └── integration/
│       └── scaffold_init_test.go
│
└── scripts/                   # Build & Validation Scripts → ADR-000
    ├── format-go.sh           # Code formatting
    ├── validate-*.sh          # Pattern enforcement (ADR validation)
    ├── check-*.sh             # Coverage/quality checks
    └── scaffold-init.go       # Scaffold customization
```

### Component Interactions

```
┌──────────┐     uses      ┌────────────┐     generates     ┌─────────────┐
│ cmd/*.go │──────────────>│ registry.go│<──────────────────│ scripts/    │
│          │               │ (ADR-002)  │                   │ generate-   │
│          │               └────────────┘                   │ constants.go│
│          │                     │                          └─────────────┘
│          │                     │ produces
│          │                     ▼
│          │               ┌─────────────────┐
│          │     imports   │ keys_generated  │
│          │<──────────────│ (ADR-005)       │
│          │               └─────────────────┘
│          │
│          │     uses      ┌──────────┐
│          │──────────────>│ logger/  │
│          │               │ (ADR-006)│
│          │               └──────────┘
│          │
│          │   delegates   ┌──────────┐
│          │──────────────>│internal/ │
│          │               │  pkg/    │
└──────────┘               └──────────┘
```

---

## Initialization Sequence

When a user runs `./ckeletin-go <command>`, the following sequence occurs:

```
┌─────────────────────────────────────────────────────────────┐
│ 1. main() Execution                                         │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 2. root.Execute()                                           │
│    - Cobra framework takes control                          │
│    - Parses CLI arguments and flags                         │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 3. root.init() (runs before execution)                      │
│    - Binds flags to Viper                                   │
│    - Registers configuration options → ADR-002              │
│    - Sets up config file paths                              │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 4. PersistentPreRun() (runs before any command)             │
│    - Loads configuration from file/env/flags                │
│    - Validates configuration → ADR-004                      │
│    - Initializes logger → ADR-006                           │
│    - Sets up dual logging (console + file)                  │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 5. Command.Run() (specific command execution)               │
│    - Ultra-thin command code → ADR-001                      │
│    - Retrieves config values using generated constants      │
│    - Delegates to business logic in internal/               │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 6. Business Logic Execution                                 │
│    - Executor pattern (e.g., ping.Executor)                 │
│    - Uses injected dependencies → ADR-003                   │
│    - Logs structured events → ADR-006                       │
│    - Returns result to command                              │
└──────────────┬──────────────────────────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────────────────────────┐
│ 7. Cleanup & Exit                                           │
│    - Flush logs                                             │
│    - Return exit code                                       │
└─────────────────────────────────────────────────────────────┘
```

**Key Points:**

- Configuration is loaded **once** in PersistentPreRun (not per-command)
- Logger is initialized **before** any command runs
- Commands receive **already-validated** configuration
- Business logic is **isolated** from CLI concerns

---

## Configuration Flow

ckeletin-go uses a **centralized configuration registry** as the single source of truth for all configuration options. See [ADR-002](002-centralized-configuration-registry.md) and [ADR-005](005-auto-generated-config-constants.md) for rationale.

### Flow Diagram

```
┌────────────────────────────────────────────────────────────────┐
│ 1. Developer defines config option in registry.go              │
│                                                                │
│    {                                                           │
│        Key:          "app.ping.timeout",                       │
│        DefaultValue: 5 * time.Second,                          │
│        Description:  "Timeout for ping operations",            │
│        Validation:   validateTimeout,                          │
│    }                                                           │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ 2. Run: task generate:config:key-constants                     │
│    → scripts/generate-config-constants.go                      │
│    → Reads registry.go                                         │
│    → Generates internal/config/keys_generated.go               │
│                                                                │
│    const KeyAppPingTimeout = "app.ping.timeout"                │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ 3. Application startup (root.init)                             │
│    → config.InitializeRegistry()                               │
│    → Binds all options to Viper                                │
│    → Sets default values                                       │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ 4. Configuration loading (PersistentPreRun)                    │
│    Priority order (highest to lowest):                         │
│                                                                │
│    1. CLI Flags          --timeout=10s                         │
│    2. Environment Vars   CKELETIN_APP_PING_TIMEOUT=10s         │
│    3. Config File        app.ping.timeout: 10s                 │
│    4. Registry Defaults  5s                                    │
│                                                                │
│    → Runs validation functions → ADR-004                       │
│    → Fails fast if invalid configuration                       │
└───────────────────────────┬────────────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────────────┐
│ 5. Command execution (cmd/ping.go)                             │
│    → Uses type-safe constant:                                  │
│                                                                │
│    timeout := viper.GetDuration(config.KeyAppPingTimeout)      │
│                         ^^^^^^^^                               │
│                    compile-time safe                           │
│                    refactor-friendly                           │
└────────────────────────────────────────────────────────────────┘
```

### Validation Enforcement

The `task validate:constants` script ensures:

- ✅ All constants in `keys_generated.go` exist in `registry.go`
- ✅ All registry keys have corresponding constants
- ✅ No manual string literals for config keys in code

See [ADR-005](005-auto-generated-config-constants.md) for details.

---

## Command Execution Lifecycle

Commands follow the **ultra-thin pattern** (see [ADR-001](001-ultra-thin-command-pattern.md)). Each command is ~20-30 lines and delegates to business logic.

### Execution Flow

```
User runs: ./ckeletin-go ping example.com --count 3
                            │
                            ▼
┌───────────────────────────────────────────────────────────┐
│ Cobra Router (root.go)                                    │
│ - Matches "ping" to pingCmd                               │
│ - Parses flags: count=3                                   │
│ - Binds to Viper                                          │
└───────────────┬───────────────────────────────────────────┘
                │
                ▼
┌───────────────────────────────────────────────────────────┐
│ PersistentPreRun (root.go)                                │
│ - Loads config (already done, reused)                     │
│ - Logger already initialized                              │
└───────────────┬───────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ pingCmd.Run() (cmd/ping.go) ~25 lines                       │
│                                                             │
│   func(cmd *cobra.Command, args []string) {                 │
│       target := args[0]  // "example.com"                   │
│       count := viper.GetInt(config.KeyAppPingCount)         │
│       timeout := viper.GetDuration(config.KeyAppPingTimeout)│
│                                                             │
│       // Create executor with dependencies                  │
│       executor := ping.NewExecutor(                         │
│           target, count, timeout,                           │
│       )                                                     │
│                                                             │
│       // Execute business logic                             │
│       result, err := executor.Execute()                     │
│       if err != nil { handleError(err); return }            │
│                                                             │
│       // Display result                                     │
│       fmt.Println(result)                                   │
│   }                                                         │
│                                                             │
└───────────────┬─────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ Business Logic (internal/ping/executor.go)                  │
│                                                             │
│ type Executor struct {                                      │
│     target  string                                          │
│     count   int                                             │
│     timeout time.Duration                                   │
│ }                                                           │
│                                                             │
│ func (e *Executor) Execute() (Result, error) {              │
│     // Actual ping implementation                           │
│     // - Network calls                                      │
│     // - Structured logging → ADR-006                       │
│     // - Error handling                                     │
│     return result, nil                                      │
│ }                                                           │
│                                                             │
└───────────────┬─────────────────────────────────────────────┘
                │
                ▼
┌─────────────────────────────────────────────────────────────┐
│ Return to Command                                           │
│ - Format output for user                                    │
│ - Exit with appropriate code                                │
└─────────────────────────────────────────────────────────────┘
```

### Why This Pattern?

- ✅ **Commands stay thin** (~20-30 lines)
- ✅ **Business logic is testable** without Cobra dependency (see [ADR-003](003-dependency-injection-over-mocking.md))
- ✅ **Clear separation** of CLI concerns vs business logic
- ✅ **Validation enforced** by `task validate:commands`

---

## Testing Architecture

Testing follows the **dependency injection over mocking** principle. See [ADR-003](003-dependency-injection-over-mocking.md) for rationale.

### Test Structure

```
┌────────────────────────────────────────────────────────────┐
│ Unit Tests (*_test.go in same package)                     │
│                                                            │
│ - Test business logic directly                             │
│ - Use dependency injection (interfaces)                    │
│ - Table-driven tests                                       │
│ - No mocking frameworks (prefer real implementations)      │
│                                                            │
│ Example: internal/ping/executor_test.go                    │
│   - Tests Executor.Execute() directly                      │
│   - Injects test dependencies                              │
│   - Validates behavior without CLI layer                   │
└────────────────────────────────────────────────────────────┘
                           │
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Integration Tests (test/integration/)                      │
│                                                            │
│ - End-to-end workflow validation                           │
│ - Example: scaffold_init_test.go                           │
│   - Copies entire project to temp dir                      │
│   - Runs `task init` with real Task                        │
│   - Validates all files updated correctly                  │
│   - Builds and executes binary                             │
│   - Cross-platform (Linux, macOS, Windows)                 │
└────────────────────────────────────────────────────────────┘
                           │
                           │
                           ▼
┌────────────────────────────────────────────────────────────┐
│ Coverage Enforcement (scripts/)                            │
│                                                            │
│ - check-coverage-project.sh: Project-wide coverage         │
│ - check-coverage-patch.sh: Changed lines coverage          │
│                                                            │
│ Thresholds:                                                │
│   - Overall: 80% minimum, 85% target                       │
│   - cmd/*: 80% minimum, 90% target                         │
│   - internal/config: 80% minimum, 90% target               │
└────────────────────────────────────────────────────────────┘
```

### Testing Workflow

See [ADR-000](000-task-based-single-source-of-truth.md) for task-based workflow.

```bash
# Unit tests
task test

# Integration tests
task test:integration

# Watch mode (development)
task test:watch

# Race detection
task test:race

# Coverage reports
task test:coverage:text
task test:coverage:html

# Full quality check (includes tests)
task check
```

---

## Development Workflow Integration

The entire development workflow is **task-based**. See [ADR-000](000-task-based-single-source-of-truth.md) for the foundational decision.

### Task as Single Source of Truth

```
┌──────────────────────────────────────────────────────────────┐
│                     Taskfile.yml                             │
│              (Single Source of Truth)                        │
│                                                              │
│  - All development commands                                  │
│  - All CI/CD commands                                        │
│  - All validation scripts                                    │
│  - Pattern enforcement                                       │
└──────────────┬───────────────────────────────┬───────────────┘
               │                               │
               ▼                               ▼
┌──────────────────────────┐   ┌──────────────────────────────┐
│   Local Development      │   │      CI/CD Pipeline          │
│                          │   │    (.github/workflows/)      │
│ $ task check             │   │                              │
│ $ task test              │   │  - task check                │
│ $ task format            │   │  - task test                 │
│ $ task build             │   │  - task build                │
└──────────────────────────┘   └──────────────────────────────┘
               │                               │
               └───────────┬───────────────────┘
                           │
                           ▼
                  Same behavior guaranteed
```

### Pattern Enforcement Through Tasks

Each ADR has **validation automation** tied to task commands:

| ADR | Pattern | Enforcement Task | Validation |
|-----|---------|------------------|------------|
| ADR-000 | Task-based workflow | `task check` | All checks use task |
| ADR-001 | Ultra-thin commands | `task validate:commands` | Script checks line count, patterns |
| ADR-002 | Config registry | `task validate:defaults` | No viper.SetDefault() calls |
| ADR-005 | Config constants | `task validate:constants` | Registry ↔ constants sync |
| ADR-006 | Structured logging | `task check` | Linter rules (no fmt.Println) |
| ADR-009 | Layered architecture | `task validate:layering` | go-arch-lint checks dependencies |
| ADR-010 | Package organization | `task validate:package-organization` | Validates CLI-first structure (no pkg/) |

### Development Cycle

```
┌──────────────────────────────────────────────────────────────┐
│ 1. Write Code                                                │
│    - Follow ADR patterns                                     │
│    - Use generated constants (config.Key*)                   │
│    - Keep commands thin (~20-30 lines)                       │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│ 2. Format                                                    │
│    $ task format                                             │
│    - Runs goimports                                          │
│    - Standardizes formatting                                 │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│ 3. Test                                                      │
│    $ task test                                               │
│    - Run tests with coverage                                 │
│    - Ensure >80% coverage                                    │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│ 4. Check (MANDATORY before commit)                           │
│    $ task check                                              │
│    - Format validation                                       │
│    - Linting (golangci-lint)                                 │
│    - Pattern validation (all ADRs)                           │
│    - Dependency checks                                       │
│    - Tests with coverage                                     │
│    - Vulnerability scan                                      │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│ 5. Commit                                                    │
│    $ git commit -m "feat: description"                       │
│    - Lefthook runs task check:format                         │
│    - Prevents commit if validation fails                     │
└──────────────┬───────────────────────────────────────────────┘
               │
               ▼
┌──────────────────────────────────────────────────────────────┐
│ 6. CI/CD                                                     │
│    - GitHub Actions runs task check                          │
│    - Same validation as local                                │
│    - Cross-platform testing                                  │
└──────────────────────────────────────────────────────────────┘
```

---

## How ADRs Work Together

This table shows how the ADRs interact to create the overall architecture:

| ADR | Scope | Interacts With | How They Connect |
|-----|-------|----------------|------------------|
| **[ADR-000](000-task-based-single-source-of-truth.md)** | Development workflow | All ADRs | Provides task-based enforcement for all patterns |
| **[ADR-001](001-ultra-thin-command-pattern.md)** | Command structure | ADR-002, 003, 006, 009 | Commands use config (002), DI (003), logging (006), follow layering (009) |
| **[ADR-002](002-centralized-configuration-registry.md)** | Configuration SSOT | ADR-001, 004, 005, 009 | Registry used by commands (001), validated (004), generates constants (005), infrastructure layer (009) |
| **[ADR-003](003-dependency-injection-over-mocking.md)** | Testing strategy | ADR-001 | Business logic (called by commands) uses DI for testability |
| **[ADR-004](004-security-validation-in-config.md)** | Security | ADR-002 | Adds validation layer to config registry |
| **[ADR-005](005-auto-generated-config-constants.md)** | Type safety | ADR-001, 002 | Generates constants from registry (002) for use in commands (001) |
| **[ADR-006](006-structured-logging-with-zerolog.md)** | Logging | ADR-001, 009 | Commands and business logic use structured logging, logger is infrastructure layer (009) |
| **[ADR-007](007-bubble-tea-for-interactive-ui.md)** | UI framework | ADR-001, 006, 009 | Interactive commands use Bubble Tea, log with structured logging, UI is infrastructure layer (009) |
| **[ADR-008](008-release-automation-with-goreleaser.md)** | Distribution | ADR-000 | Release process uses task commands |
| **[ADR-009](009-layered-architecture-pattern.md)** | Architecture layers | ADR-001, 002, 006, 007, 010 | Enforces 4-layer pattern with automated validation, commands (001) delegate to business logic, infrastructure includes config (002), logging (006), UI (007), package structure (010) |
| **[ADR-010](010-package-organization-strategy.md)** | Package organization | ADR-009 | Defines CLI-first structure (no pkg/, all in internal/), complements layering rules (009) |

### Dependency Graph

```
                    ADR-000 (Task-based workflow)
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
    ADR-009             ADR-001             ADR-002         ADR-008
   (Layering)          (Commands)           (Config)       (Release)
        │                   │                   │
   [enforces]        ┌──────┼────────┐          │
        │            │      │        │          │
        └────────────┼──────┼────────┼──────────┘
                     │      │        │
                     ▼      ▼        ▼
                 ADR-003  ADR-006  ADR-007  ADR-004  ADR-005
                  (DI)    (Log)     (UI)    (Sec)   (Constants)
                                                     │
                                                     ▼
                                                Validation
                                                  Scripts
```

**Key Relationships:**
- **ADR-009** (Layering) enforces the structure that ADR-001 (Commands) and ADR-002 (Config) operate within
- **ADR-001** (Commands) uses ADR-003 (DI), ADR-006 (Logging), ADR-007 (UI) within the layering constraints
- **ADR-002** (Config) uses ADR-004 (Security validation) and generates ADR-005 (Constants)
- **ADR-000** (Tasks) provides enforcement for all patterns via validation scripts

### Cross-Cutting Concerns

**Security (ADR-004):**

- Applied in: Config loading, file operations, user input
- Validated by: `task check` (linter rules)

**Logging (ADR-006):**

- Used by: All business logic, commands, infrastructure
- Configured by: ADR-002 (config registry)

**Testing (ADR-003):**

- Applied to: All business logic (internal/)
- Enforced by: Coverage thresholds in `task check`

---

## Package Organization

This section explains the directory structure for the layers described in [Architectural Layers](#architectural-layers).

### Why internal/ vs pkg/ vs cmd/?

```
cmd/                # Public CLI interface (Cobra commands)
├── *.go            # Ultra-thin (ADR-001), public API of the tool
└── Commands are the ONLY public Go API

internal/           # Private implementation (not importable by other projects)
├── config/         # Configuration management (ADR-002, 005)
├── logger/         # Logging infrastructure (ADR-006)
├── ui/             # Terminal UI components (ADR-007)
└── */              # Business logic (domain-specific)

pkg/                # (Not used - nothing to expose as library)
```

See [ADR-010](010-package-organization-strategy.md) for the rationale behind this organization strategy.

**Design Decision:** ckeletin-go is a **CLI application**, not a library:

- No `pkg/` directory because nothing is intended for external import
- All implementation in `internal/` to prevent accidental API surface
- Only `cmd/` exposes the CLI interface (via Cobra)

### Package Dependency Rules

```
cmd/           →  can import  →  internal/* (all)
internal/pkg1  →  can import  →  internal/pkg2 (with layering rules)
internal/*     →  CANNOT import → cmd/* (prevents cycles)

Example valid imports in cmd/ping.go:
  ✅ "ckeletin-go/internal/ping"
  ✅ "ckeletin-go/internal/config"
  ✅ "ckeletin-go/internal/logger"

Example invalid imports in internal/ping/executor.go:
  ❌ "ckeletin-go/cmd"  (would create cycle)
```

**Enforcement:** Go compiler prevents cycles. Layering rules are automated via go-arch-lint (see [ADR-009](009-layered-architecture-pattern.md)). Run `task validate:layering` to check compliance.

---

## Key Design Patterns

While not formally documented in ADRs, these patterns are used consistently:

### 1. Executor Pattern

**Used in:** Business logic (internal/*/executor.go)

```go
type Executor struct {
    // Dependencies (injected)
    target string
    config Config
}

func NewExecutor(deps...) *Executor {
    return &Executor{...}
}

func (e *Executor) Execute() (Result, error) {
    // Business logic here
}
```

<!-- TODO: ADR - Executor Pattern (likely extends ADR-001, separates business logic from CLI) -->

**Why:** Separates business logic from CLI, enables testing (ADR-003)

### 2. Options Pattern

**Used in:** Command configuration (internal/config/commands/)

```go
type CommandOptions struct {
    Timeout  time.Duration
    Retries  int
    Verbose  bool
}

func NewOptionsFromViper() CommandOptions {
    return CommandOptions{
        Timeout: viper.GetDuration(config.KeyAppTimeout),
        // ...
    }
}
```

<!-- TODO: ADR - Options Pattern (likely extends ADR-002, type-safe config consumption) -->

**Why:** Centralizes command configuration, type-safe access

### 3. Registry Pattern

**Used in:** Configuration (internal/config/registry.go)

See [ADR-002](002-centralized-configuration-registry.md) for details.

### 4. Factory Pattern

**Used in:** Command creation (cmd/*.go)

```go
func MustNewPingCommand() *cobra.Command {
    cmd := &cobra.Command{...}
    // Setup flags, etc.
    return cmd
}
```

<!-- TODO: ADR - Factory Pattern (likely extends ADR-001, consistent command creation) -->

**Why:** Consistent command creation, validation at startup

---

## Summary

**ckeletin-go's architecture** is built on four foundational pillars:

1. **Layered architecture** ([ADR-009](009-layered-architecture-pattern.md)) - Enforced 4-layer pattern with automated validation
2. **Task-based workflow** ([ADR-000](000-task-based-single-source-of-truth.md)) - SSOT for development
3. **Ultra-thin commands** ([ADR-001](001-ultra-thin-command-pattern.md)) - Clear separation of concerns
4. **Centralized configuration** ([ADR-002](002-centralized-configuration-registry.md)) - Type-safe, validated config

These are supported by:

- **Dependency injection** ([ADR-003](003-dependency-injection-over-mocking.md)) for testability
- **Security validation** ([ADR-004](004-security-validation-in-config.md)) for safety
- **Auto-generated constants** ([ADR-005](005-auto-generated-config-constants.md)) for type safety
- **Structured logging** ([ADR-006](006-structured-logging-with-zerolog.md)) for observability
- **Interactive UIs** ([ADR-007](007-bubble-tea-for-interactive-ui.md)) when needed
- **Automated releases** ([ADR-008](008-release-automation-with-goreleaser.md)) for distribution

All patterns are **enforced through automation** via `task validate:*` commands.

---

## References

- **ADRs:** See individual ADR files in this directory for decision rationale
- **Task Commands:** See `Taskfile.yml` for all available development commands
- **Validation Scripts:** See `scripts/validate-*.sh` for pattern enforcement
- **Contributing Guide:** See `CONTRIBUTING.md` for development workflow
- **AI Guidelines:** See `CLAUDE.md` for AI-assisted development guidelines

---

**For questions about WHY these architectural decisions were made, see the individual ADRs linked throughout this document.**
