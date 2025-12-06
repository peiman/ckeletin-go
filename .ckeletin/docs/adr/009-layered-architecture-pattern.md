# ADR-009: Layered Architecture Pattern

## Status
Accepted

## Context

### The Problem: Architectural Degradation

As CLI applications grow, they often suffer from architectural degradation:
- Business logic mixed with CLI framework code (Cobra)
- Direct dependencies between unrelated features
- Difficult to test components in isolation
- Unclear where new code should go
- Circular dependencies and tight coupling
- Hard to understand system boundaries

This leads to:
- Fragile code that breaks when changing unrelated features
- Slow development as codebase becomes harder to navigate
- Testing difficulties requiring complex mocking
- Onboarding friction for new developers

### Why Layering?

Layering provides:
- **Clear dependency flow**: Outer layers depend on inner layers, never reverse
- **Separation of concerns**: Each layer has a specific responsibility
- **Testability**: Layers can be tested independently
- **Framework independence**: Business logic doesn't know about CLI frameworks
- **Predictability**: Developers know where code belongs

### Alternatives Considered

**1. Monolithic/Flat Structure**
- All code in one package or loosely organized packages
- **Pros**: Simple initially, no structure overhead
- **Cons**: Scales poorly, becomes unmaintainable, no boundaries
- **Why not**: ckeletin-go is a template/scaffold - architecture matters

**2. Hexagonal Architecture (Ports & Adapters)**
- Core domain surrounded by adapters, ports define interfaces
- **Pros**: Strong isolation, clear boundaries, highly testable
- **Cons**: More abstraction layers, overhead for simple CLI
- **Why not**: Too complex for CLI application needs

**3. Clean Architecture (Uncle Bob)**
- Multiple concentric layers (entities, use cases, adapters, frameworks)
- **Pros**: Complete separation, highly flexible
- **Cons**: 4+ layers, significant boilerplate, overkill for CLI
- **Why not**: CLI apps don't need entity/use case separation

**4. MVC (Model-View-Controller)**
- Classic pattern separating data, presentation, logic
- **Pros**: Well-understood, proven pattern
- **Cons**: Designed for UI apps with persistent state, not CLI
- **Why not**: Poor fit for stateless command execution model

## Decision

We adopt a **4-layer architecture** optimized for CLI applications:

```
┌──────────────────────────────────────────────────┐
│              1. Entry Layer                      │
│                 (main.go)                        │
│  - Bootstrap application                         │
│  - Execute root command                          │
└─────────────────┬────────────────────────────────┘
                  │ depends on
                  ▼
┌──────────────────────────────────────────────────┐
│              2. Command Layer                    │
│                  (cmd/)                          │
│  - Ultra-thin wrappers (~20-30 lines)            │
│  - CLI framework integration (Cobra)             │
│  - Flag/argument parsing                         │
│  - Delegation to business logic                  │
└─────────────────┬────────────────────────────────┘
                  │ depends on
                  ▼
        ┌─────────┴─────────┐
        ▼                   ▼
┌──────────────────┐  ┌──────────────────────────┐
│  3. Business     │  │  3. Infrastructure       │
│     Logic        │  │     Services             │
│  (internal/*)    │  │  (internal/config,       │
│                  │  │   internal/logger,       │
│  - ping/         │  │   internal/ui)           │
│  - docs/         │  │                          │
│  - Feature logic │  │  - Configuration (Viper) │
│  - No CLI deps   │  │  - Logging (Zerolog)     │
│  - No framework  │  │  - UI (Bubble Tea)       │
└─────────┬────────┘  └──────────┬───────────────┘
          │                      │
          └──────────┬───────────┘
                     │ depends on
                     ▼
          ┌──────────────────────┐
          │  4. External Systems │
          │  (Network, FS, etc.) │
          └──────────────────────┘
```

### Layer Responsibilities

**1. Entry Layer (main.go)**
- Minimal application bootstrap
- Root command execution
- **Imports**: cmd/
- **Imported by**: Nothing (entry point)

**2. Command Layer (cmd/)**
- Ultra-thin command definitions (~20-30 lines, see [ADR-001](001-ultra-thin-command-pattern.md))
- CLI framework integration (Cobra)
- Flag/argument parsing and binding
- Delegation to business logic
- **Imports**: internal/*, Cobra
- **Imported by**: Entry layer only
- **Key Rule**: Only this layer can import Cobra

**3. Business Logic Layer (internal/ping, internal/docs, etc.)**
- Feature-specific implementations
- Domain logic with no CLI framework dependencies
- Executor pattern for command execution
- **Imports**: Infrastructure layer, standard library
- **Imported by**: Command layer
- **Key Rules**:
  - No Cobra imports
  - No imports from cmd/
  - Business packages isolated from each other (ping cannot import docs)

**4. Infrastructure Layer (internal/config, internal/logger, internal/ui)**
- Cross-cutting concerns
- Shared services available to all layers
- External system integration (Viper, Zerolog, Bubble Tea)
- **Imports**: External libraries, standard library
- **Imported by**: Command layer, Business logic layer
- **Key Rules**:
  - Cannot import business logic
  - Cannot import cmd/
  - Available to all layers above

### Dependency Rules

**Critical Rules (Enforced):**
1. **Acyclic Dependencies**: Outer layers depend on inner, never reverse
2. **CLI Isolation**: Only `cmd/` can import Cobra (framework independence)
3. **Internal Protection**: `internal/*` packages cannot import `cmd/`
4. **Business Isolation**: Business logic packages don't import each other
5. **Infrastructure Separation**: Infrastructure cannot import business logic

**Allowed Dependencies:**
- Entry → Command
- Command → Business Logic
- Command → Infrastructure
- Business Logic → Infrastructure
- Infrastructure → External Libraries

**Forbidden Dependencies:**
- Business Logic → Command (would couple to CLI)
- Business Logic → Entry (architectural violation)
- Infrastructure → Business Logic (wrong direction)
- Infrastructure → Command (wrong direction)
- `internal/*` → Cobra (only cmd/ uses Cobra)

### Enforcement

**1. Go Compiler Baseline**
- Prevents import cycles automatically
- Enforces `internal/` package visibility rules

**2. Automated Validation** (go-arch-lint)
- Layer dependency rules verified in CI
- Cobra isolation enforced
- Business logic isolation checked
- Configuration: `.go-arch-lint.yml`

**3. Task Command**
```bash
task validate:layering  # Checks all layer rules
task check              # Includes layering validation
```

**4. Validation Script**
- `scripts/validate-layering.sh` runs go-arch-lint
- Fails CI if violations detected
- Reports which packages violate which rules

## Consequences

### Positive

**1. Clear Boundaries**
- Developers know where code belongs
- Architecture visible in file structure
- Onboarding faster (structure is obvious)

**2. Framework Independence**
- Business logic has zero Cobra dependencies
- Can reuse logic in non-CLI contexts (e.g., library, server)
- Easy to swap CLI frameworks if needed

**3. Testability**
- Business logic testable without CLI framework
- No Cobra mocks needed in business logic tests
- Infrastructure components tested independently

**4. Maintainability**
- Changes to CLI framework don't affect business logic
- New commands follow clear pattern (ADR-001 + this ADR)
- Refactoring safer with automated validation

**5. Scalability**
- Adding features has clear location (new package in internal/)
- Feature isolation prevents coupling
- Architecture doesn't degrade as project grows

**6. Automated Quality**
- go-arch-lint prevents architectural drift
- CI catches violations before merge
- No reliance on code review alone

### Negative

**1. Learning Curve**
- Developers must understand layering concept
- More structure than flat architecture
- Requires discipline to maintain

**2. Initial Overhead**
- Setting up validation tooling
- Writing/maintaining `.go-arch-lint.yml`
- More thought required for code placement

**3. Potential Over-Engineering**
- Simple features still need proper layering
- Cannot shortcut even for trivial commands
- More files/packages than monolithic approach

### Mitigations

**1. Documentation**
- This ADR explains WHY layering matters
- [ARCHITECTURE.md](ARCHITECTURE.md) shows HOW layers interact
- Code comments reference ADRs

**2. Automation**
- `task validate:layering` prevents violations automatically
- Clear error messages when rules violated
- Fast feedback loop (<2s validation)

**3. Examples**
- `cmd/ping.go` demonstrates command layer pattern
- `internal/ping/` shows business logic layer
- Template files guide new command creation

**4. Maintenance Tools**
- `.go-arch-lint.yml` uses exclusions (zero-maintenance when adding commands)
- If exclusions don't work: Document that adding commands requires YAML update
- Validation runs automatically in CI

## Implementation Details

### File Organization

```
ckeletin-go/
├── main.go                    # Layer 1: Entry
│
├── cmd/                       # Layer 2: Commands
│   ├── root.go
│   ├── ping.go
│   └── docs.go
│
└── internal/
    ├── ping/                  # Layer 3: Business Logic
    ├── docs/                  # Layer 3: Business Logic
    │
    ├── config/                # Layer 4: Infrastructure
    ├── logger/                # Layer 4: Infrastructure
    └── ui/                    # Layer 4: Infrastructure
```

### go-arch-lint Configuration

See `.go-arch-lint.yml` for complete configuration.

Key features:
- Uses `exclude` patterns for zero-maintenance (infrastructure list is stable)
- Automatically includes new business logic packages under `internal/`
- Enforces external dependency rules (Cobra only in cmd/)

### Adding New Commands

When adding a new command (e.g., `task init`):

1. **Create business logic**: `internal/init/executor.go`
   - Automatically included in business layer (no config update needed)
   - Cannot import cmd/ or other business logic
   - Can import infrastructure (config, logger, ui)

2. **Create command wrapper**: `cmd/init.go`
   - Ultra-thin (~20-30 lines, ADR-001)
   - Imports Cobra and business logic
   - Delegates to executor

3. **Validation**: Run `task validate:layering`
   - Ensures new code follows layer rules
   - Catches violations before commit

## Related ADRs

- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-thin commands define Command Layer pattern
- [ADR-002](002-centralized-configuration-registry.md) - Config registry is Infrastructure Layer
- [ADR-003](003-dependency-injection-over-mocking.md) - DI enables layer isolation in tests
- [ADR-006](006-structured-logging-with-zerolog.md) - Logger is Infrastructure Layer
- [ADR-007](007-bubble-tea-for-interactive-ui.md) - UI framework is Infrastructure Layer

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - Complete system architecture overview
- `.go-arch-lint.yml` - Layering rules configuration
- `scripts/validate-layering.sh` - Validation script
- `cmd/ping.go` - Command layer example (31 lines)
- `internal/ping/ping.go` - Business logic layer example
- [Layered Architecture Pattern](https://herbertograca.com/2017/08/03/layered-architecture/) - General pattern overview
