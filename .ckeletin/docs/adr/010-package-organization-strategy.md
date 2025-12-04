# ADR-010: Package Organization Strategy

## Status
Accepted

## Context

### The Problem: Where Should Code Live?

Go projects have flexibility in how they organize packages, particularly around:
- **`pkg/`** - Traditionally for public APIs meant to be imported by external projects
- **`internal/`** - Private packages that cannot be imported externally
- **`cmd/`** - Command-line applications
- **Root directory** - Where application entry points live

This flexibility creates questions:
- Should we expose a public API via `pkg/`?
- Where do we draw the boundary between public and private?
- How do we communicate project intent through structure?
- What prevents accidental API surface expansion?

### Why This Matters for ckeletin-go

**ckeletin-go is a CLI application skeleton**, not a library. This fundamental identity should be:
1. **Visible** - Structure immediately shows "this is a CLI tool"
2. **Enforced** - Prevents accidental library creation
3. **Documented** - Clear rationale for future maintainers
4. **Validated** - Automated checks prevent drift

Without clear organization:
- Developers might create `pkg/` packages "for reusability"
- Business logic might leak into root directory
- Project identity becomes ambiguous (CLI tool or library?)
- Accidental public API surface creates maintenance burden

### Alternatives Considered

**1. Traditional Go Project Layout (cmd/, pkg/, internal/)**
```
project/
├── cmd/           # CLI applications
├── pkg/           # Public library code
└── internal/      # Private implementation
```
- **Pros**: Well-known pattern, supports both CLI and library
- **Cons**: Suggests dual purpose, maintenance burden for public API
- **Why not**: We're explicitly NOT a library

**2. Library-First Layout (pkg/, cmd/ optional)**
```
project/
├── pkg/           # Primary public API
├── internal/      # Private helpers
└── cmd/           # Optional CLI wrapper
```
- **Pros**: Clear library intent, common for SDK projects
- **Cons**: Wrong signal - we're a CLI tool first
- **Why not**: Inverts our actual priority (CLI is the product)

**3. Flat Structure (everything at root)**
```
project/
├── command1.go
├── command2.go
└── utils.go
```
- **Pros**: Simple, no directory overhead
- **Cons**: Scales poorly, no boundaries, everything public
- **Why not**: Doesn't scale, exposes everything

**4. Monorepo Style (apps/, libs/, packages/)**
```
project/
├── apps/ckeletin/     # CLI application
├── libs/config/       # Shared libraries
└── packages/utils/    # Common utilities
```
- **Pros**: Supports multiple apps, clear separation
- **Cons**: Overkill for single CLI, encourages premature abstraction
- **Why not**: We're one CLI tool, not multiple apps

## Decision

We adopt a **CLI-first package organization** with no public API surface:

```
ckeletin-go/
├── main.go                    # Entry point (only root-level .go file)
│
├── cmd/                       # CLI command implementations
│   ├── root.go                # Root command + global setup
│   ├── ping.go                # Feature commands
│   ├── docs.go
│   └── *.go                   # Additional commands
│
├── internal/                  # ALL implementation (private)
│   ├── ping/                  # Business logic packages
│   ├── docs/
│   ├── config/                # Infrastructure packages
│   ├── logger/
│   └── ui/
│
├── scripts/                   # Build and validation scripts
├── test/integration/          # Integration tests
├── docs/                      # Documentation
└── (NO pkg/ directory)        # Explicitly absent
```

### Key Principles

**1. No `pkg/` Directory**
- ckeletin-go is a **CLI application**, not a reusable library
- No public Go API to maintain
- All implementation is private (in `internal/`)
- Users interact via compiled binary, not Go imports

**2. `internal/` for All Implementation**
- Go's `internal/` visibility rules prevent external imports
- Enforces "CLI application only" identity
- Freedom to refactor without breaking external consumers
- No semantic versioning burden for internal APIs

**3. `cmd/` for CLI Interface**
- Only public interface is the command-line tool itself
- Cobra commands live here (framework isolation)
- Ultra-thin wrappers (~20-30 lines, see [ADR-001](001-ultra-thin-command-pattern.md))
- No business logic in this layer

**4. `main.go` at Root**
- Single entry point at project root
- Only root-level `.go` file allowed
- Keeps root directory clean
- Conventional Go application pattern

**5. Auxiliary Directories Allowed**
- `scripts/` - Build, validation, and utility scripts
- `test/` - Integration and E2E tests
- `docs/` - Documentation (ADRs, guides)
- `testdata/` - Test fixtures
- `.github/` - CI/CD configuration
- These do NOT contain production Go packages

### Enforcement Rules

**✅ Allowed:**
- `main.go` at root (entry point)
- All packages in `cmd/` or `internal/`
- Go files in `scripts/` (build tools, not packages)
- Test files anywhere (`*_test.go`)

**❌ Forbidden:**
- `pkg/` directory with Go code
- `.go` files at root except `main.go` and `main_test.go`
- Public packages outside `cmd/` or `internal/`
- Business logic in root directory
- Any Go package importable by external projects (except via `cmd/`)

### Enforcement

**1. Filesystem Checks**
```bash
task validate:package-organization
```
Validates:
- No `pkg/` directory with Go packages
- No `.go` files at root except `main.go` and `main_test.go`
- All packages in `cmd/`, `internal/`, `scripts/`, or `test/`

**2. Integrated into Quality Pipeline**
```bash
task check  # Includes package organization validation
```

**3. CI Enforcement**
- Runs on every PR
- Fails if organization rules violated
- Prevents architectural drift

## Consequences

### Positive

**1. Clear Project Identity**
- File structure immediately shows "this is a CLI tool"
- No confusion about whether to import as library
- Onboarding faster (no question about where code goes)

**2. Internal Freedom**
- Can refactor `internal/` without breaking external consumers
- No semantic versioning burden
- No API compatibility concerns
- Rapid iteration without fear

**3. Prevents Scope Creep**
- Absence of `pkg/` prevents "let's make this a library"
- Forces conscious decision if we want to expose APIs
- Maintains focus on CLI excellence

**4. Enforcement Automation**
- `task validate:package-organization` catches violations
- CI prevents architectural drift
- No reliance on code review alone

**5. Go Ecosystem Alignment**
- `internal/` uses Go's visibility rules
- Conventional `cmd/` and `main.go` placement
- Familiar to Go developers

### Negative

**1. Not Reusable as Library**
- Business logic in `internal/` cannot be imported externally
- If users want Go API, must expose via `pkg/`
- Migration would require conscious refactoring

**2. Strict Structure**
- Cannot "just add a package at root"
- Must think about placement (cmd/ vs internal/)
- More structure than flat layout

**3. Potential Over-Engineering**
- For tiny projects, this might be overkill
- Adds directory overhead for single-file utilities

### Mitigations

**1. Documentation**
- This ADR explains WHY we're CLI-only
- [ARCHITECTURE.md](ARCHITECTURE.md) shows HOW packages organize
- Clear guidance for contributors

**2. If Library Needed Later**
- Extract desired packages from `internal/` to `pkg/`
- Requires conscious decision (not accidental)
- Can maintain CLI tool while adding library mode
- Semantic versioning applies only after extraction

**3. Examples**
- Current codebase demonstrates pattern
- Template files guide new code placement
- `scripts/validate-package-organization.sh` gives instant feedback

## Implementation Details

### Current State Validation

The current project **already follows this pattern**:
- ✅ No Go code in `pkg/` (directory exists but empty)
- ✅ All implementation in `internal/` and `cmd/`
- ✅ Only `main.go` and `main_test.go` at root
- ✅ Auxiliary directories (`scripts/`, `test/`, `docs/`) present

This ADR **documents existing practice** and adds enforcement.

### Directory Purposes

```
ckeletin-go/
│
├── main.go                    # Bootstrap, execute root command
├── main_test.go               # Entry point tests
│
├── cmd/                       # Layer 2: CLI Commands (see ADR-009)
│   └── *.go                   # Cobra commands, ultra-thin (ADR-001)
│
├── internal/                  # Layers 3-4: Business + Infrastructure (ADR-009)
│   ├── ping/, docs/           # Business logic packages
│   ├── config/, logger/, ui/  # Infrastructure packages
│   └── */                     # Additional internal packages
│
├── scripts/                   # Build and validation tooling
│   ├── *.sh                   # Bash scripts
│   └── *.go                   # Go build tools (not packages)
│
├── test/integration/          # Integration tests
├── docs/                      # ADRs, guides, documentation
├── testdata/                  # Test fixtures
├── .github/                   # CI/CD workflows
│
└── (no pkg/)                  # Explicitly absent (CLI only)
```

### When to Create `pkg/`

Only create `pkg/` if we make a **conscious decision** to expose a public Go API.

**Questions to ask first:**
1. Do external projects need to import our code?
2. Are we willing to maintain API compatibility?
3. Should we version the library separately from the CLI?
4. Can users accomplish their goals with the CLI alone?

**If yes to all:** Extract relevant packages from `internal/` to `pkg/`, add semantic versioning, document public API.

**If no:** Keep everything in `internal/`, users interact via CLI binary.

### Adding New Packages

**For new features:**
1. Create business logic in `internal/<feature>/`
2. Create command in `cmd/<feature>.go`
3. No need to update validation (automatically covered)

**For new commands:**
1. Follow [ADR-001](001-ultra-thin-command-pattern.md) (ultra-thin pattern)
2. Follow [ADR-009](009-layered-architecture-pattern.md) (layering rules)
3. Run `task validate:package-organization` to verify

## Related ADRs

- [ADR-009](009-layered-architecture-pattern.md) - Defines what goes in cmd/ vs internal/
- [ADR-001](001-ultra-thin-command-pattern.md) - Pattern for cmd/ packages
- [ADR-002](002-centralized-configuration-registry.md) - Config belongs in internal/config
- [ADR-006](006-structured-logging-with-zerolog.md) - Logger belongs in internal/logger
- [ADR-007](007-bubble-tea-for-interactive-ui.md) - UI belongs in internal/ui

## References

- [Go Project Layout](https://github.com/golang-standards/project-layout) - Community conventions
- [Go `internal/` packages](https://go.dev/doc/go1.4#internalpackages) - Visibility rules
- [ARCHITECTURE.md](ARCHITECTURE.md) - Complete system architecture
- `scripts/validate-package-organization.sh` - Enforcement script
