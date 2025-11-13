# ADR-000: Task-Based Single Source of Truth for CI/Local Alignment

## Status
Accepted

## Context

A common problem in software projects is **environment drift** between CI and local development:

- Developers run different commands locally than CI runs
- CI uses different flags, tools, or versions than developers
- When CI fails, developers struggle to reproduce the failure locally
- Quality checks are duplicated across multiple places (CI YAML, Makefile, scripts, docs)
- Adding a new check requires updating multiple files
- "Works on my machine" syndrome is common

This leads to:
- Wasted time debugging CI-specific failures
- Inconsistent code quality (CI catches what local checks miss)
- Maintenance burden (keeping CI and local commands in sync)
- Poor developer experience (can't trust local checks)
- Documentation drift (README says one thing, CI does another)

### Traditional Anti-Patterns

**Pattern 1: Duplicate Logic**
```yaml
# .github/workflows/ci.yml
- run: gofmt -l -w .
- run: golangci-lint run --config .golangci.yml
- run: go test -v -coverprofile=coverage.txt ./...

# Makefile or local scripts
format:
	gofmt -l -w .   # Duplicated!
lint:
	golangci-lint run  # Missing --config flag! Drift!
test:
	go test -v ./...  # Missing coverage! Drift!
```

**Pattern 2: CI-Specific Scripts**
```yaml
# .github/workflows/ci.yml
- run: ./scripts/ci-test.sh

# Developer can't easily run this (expects CI env vars, paths, etc.)
```

**Pattern 3: No Local Enforcement**
```yaml
# CI runs comprehensive checks
- run: go vet ./...
- run: golangci-lint run
- run: gosec ./...

# Developer runs minimal checks
$ go test ./...  # Missing vet, lint, security checks!
```

## Decision

We adopt a **Task-based Single Source of Truth (SSOT)** pattern where:

1. **Taskfile.yml is the canonical source** for all development commands
2. **CI runs exactly the same Task commands** as developers run locally
3. **All quality checks are composed into `task check`** - a single command
4. **Pre-commit hooks use Task commands** for consistency
5. **Documentation references Task commands** as the standard interface

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Taskfile.yml (SSOT)                     │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  check:                                              │   │
│  │    - task: check:format                              │   │
│  │    - task: lint                                      │   │
│  │    - task: validate:defaults   # Pattern enforcement│   │
│  │    - task: validate:commands   # Pattern enforcement│   │
│  │    - task: validate:constants  # Pattern enforcement│   │
│  │    - task: check:deps                                │   │
│  │    - task: test                                      │   │
│  └─────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
           ▲                    ▲                    ▲
           │                    │                    │
    ┌──────┴──────┐      ┌─────┴─────┐      ┌──────┴──────┐
    │  Developer  │      │  Lefthook │      │   CI/CD     │
    │             │      │ Pre-commit│      │  (GitHub    │
    │ task check  │      │           │      │   Actions)  │
    │             │      │task format│      │             │
    │             │      │task lint  │      │ task check  │
    │             │      │task test  │      │             │
    └─────────────┘      └───────────┘      └─────────────┘
```

### Implementation

**Taskfile.yml - The SSOT:**
```yaml
check:
  desc: Run all quality checks
  cmds:
    - task: check:format          # Formatting validation
    - task: lint                  # go vet + golangci-lint
    - task: validate:defaults     # ADR-002 enforcement
    - task: validate:commands     # ADR-001 enforcement
    - task: validate:constants    # ADR-005 enforcement
    - task: check:deps            # Verification + vulnerabilities
    - task: test                  # Tests with coverage
```

**GitHub Actions CI - Uses SSOT:**
```yaml
- name: Install Task
  run: curl -sL https://taskfile.dev/install.sh | sh -s -- -b "$INSTALL_DIR"

- name: Install Project Dependencies
  run: task setup

- name: Run Quality Checks
  run: task check  # ← Single command runs everything
```

**Lefthook Pre-commit - Uses SSOT:**
```yaml
pre-commit:
  parallel: true
  commands:
    format:
      run: task format:staged -- {staged_files}
    lint:
      run: task lint
    validate-constants:
      run: task validate:constants
    verify-deps:
      run: task check:deps:verify
    test:
      run: task test
```

**Developer Workflow - Uses SSOT:**
```bash
# Before commit
$ task check

# CI runs the same command
# Result: Zero drift, reproducible failures
```

## Consequences

### Positive

- **Zero Environment Drift**: Impossible for CI and local to diverge
- **Reproducible CI Failures**: `task check` runs exactly what CI runs
- **Single Point of Maintenance**: Update Taskfile once, applies everywhere
- **Self-Documenting**: `task --list` shows all available commands
- **Developer Experience**: Simple, consistent interface across the project
- **Pattern Enforcement**: Architectural patterns validated in CI automatically
- **Onboarding**: New developers learn one tool (`task`) not multiple
- **Confidence**: Developers trust that local checks match CI
- **Composability**: Tasks can be composed (check → check:deps → check:vuln)
- **Granular Control**: Run specific checks (task check:vuln) or all (task check)

### Negative

- **Tool Dependency**: Requires Task installation (mitigated by `task setup`)
- **Learning Curve**: Developers must learn Taskfile syntax
- **Indirection**: One more layer between developer and underlying tools
- **CI Bootstrap**: CI must install Task before running checks

### Mitigations

- **Easy Installation**: `task setup` installs all required tools
- **Clear Documentation**: CLAUDE.md, README.md reference Task commands
- **CI Template**: GitHub Actions workflow pre-configured with Task installation
- **Task Binary Cache**: CI caches Task binary for fast installation
- **Version Pinning**: Task version pinned in CI for reproducibility
- **Fallback**: Individual tasks can still be run directly if needed

## Pattern Enforcement

This ADR enables **automated enforcement** of other architectural patterns:

```yaml
check:
  cmds:
    - task: validate:defaults     # Enforces ADR-002 (No scattered SetDefaults)
    - task: validate:commands     # Enforces ADR-001 (Ultra-thin commands)
    - task: validate:constants    # Enforces ADR-005 (Auto-generated constants)
```

Unlike most projects that document patterns but rely on manual code review, **this project validates architectural patterns in CI automatically**.

## Task Composition

Tasks are composed in layers for flexibility:

```
task check (everything)
  ├─ task check:format
  ├─ task lint
  ├─ task validate:defaults (custom validation)
  ├─ task validate:commands (custom validation)
  ├─ task validate:constants (custom validation)
  ├─ task check:deps (composed task)
  │   ├─ task check:deps:verify
  │   ├─ task check:deps:outdated
  │   └─ task check:vuln
  └─ task test
```

Developers can:
- Run everything: `task check`
- Run a category: `task check:deps`
- Run individual check: `task check:vuln`

CI always runs: `task check` (complete validation)

## Version Pinning

To ensure reproducibility, Task version is pinned in CI:

```yaml
env:
  TASK_VERSION: '3.39.0'

- name: Install Task
  run: |
    curl -sL https://taskfile.dev/install.sh | sh -s -- -b "$INSTALL_DIR" v${{ env.TASK_VERSION }}
```

This guarantees that:
- CI behavior is reproducible over time
- Task updates are intentional (update TASK_VERSION explicitly)
- Old commits can be checked out and CI will behave identically

## Compliance Validation

The pattern is validated through multiple mechanisms:

1. **Pre-commit hooks** run subset of checks before commits land
2. **CI runs `task check`** on every push and pull request
3. **Release process requires** `task check` to pass before tagging
4. **Documentation mandates** `task check` before commits

Example from CLAUDE.md:
```markdown
### Before Committing (MANDATORY)
```bash
task check  # Run ALL checks - this is non-negotiable
```
```

## Task Naming Convention

### Pattern: action:target

All tasks follow a simple, consistent pattern:

```
action:target[:subvariant]
```

Where:
- **action** is what you're doing (check, validate, test, generate, build, clean, format, bench)
- **target** is what you're doing it to (a resource, variant, or modifier)

### Automated Enforcement

Task naming is **validated automatically** via `task validate:task-naming`:

```bash
$ task validate:task-naming
🔍 Validating ADR-000: Task naming convention...
✅ All tasks follow ADR-000 naming convention
```

If you create an invalid task name:

```bash
❌ Task 'fuzz': Standalone task not in allowed list
   💡 Found similar tasks:
      • test:fuzz
      • test:fuzz:config
   💡 Suggested fix: Use pattern 'test:fuzz'
```

**Where it runs:**
- ✅ Local: `task check` (before every commit)
- ✅ Pre-commit: Via lefthook hooks
- ✅ CI: Part of quality gate pipeline

**How it works:**
- Parses all task names from Taskfile.yml
- Validates action is from approved list
- Provides smart suggestions using pattern matching
- Detects typos (e.g., `tset:race` → suggests `test:race`)

**Examples:**

```yaml
# Action applied to different targets
check:format                  # Check format
check:vuln                    # Check vulnerabilities
check:deps                    # Check dependencies (orchestrator)
check:deps:verify             # Check deps, verify subvariant
check:deps:outdated           # Check deps, outdated subvariant
validate:commands             # Validate commands
validate:constants            # Validate constants
generate:config:key-constants # Generate config key constants
generate:config:template      # Generate config YAML template
generate:docs                 # Generate docs (orchestrator)
generate:docs:config          # Generate configuration documentation
test:race                     # Test with race detection
test:integration              # Integration test
test:coverage:patch           # Test coverage, patch subvariant
bench:cmd                     # Benchmark cmd package
build:release                 # Build release artifacts
clean:local                   # Clean local artifacts
clean:release                 # Clean release artifacts
format:staged                 # Format staged files

# Standalone actions (no target needed)
format                  # Format everything
test                    # Test everything
build                   # Build
clean                   # Clean everything (orchestrator)
check                   # Check everything (orchestrator)
lint                    # Lint
run                     # Run
install                 # Install
setup                   # Setup
tidy                    # Tidy
```

**Benefits:**

- **Simple**: One pattern to learn - `action:target`
- **Discoverable**: `task check:<TAB>` shows all checks, `task test:<TAB>` shows all test variants
- **Consistent**: Always read as "action on target"
- **Scalable**: Easy to add new tasks following the same pattern

### Why This Pattern Matters

**Scripts are implementation details. Task is the interface.**

```yaml
# .lefthook.yml - uses Task commands, not scripts
format:
  run: task format:staged -- {staged_files}
validate-constants:
  run: task validate:constants
verify-deps:
  run: task check:deps:verify
```

**Benefits:**

- Rename or refactor scripts → only update Taskfile.yml
- Lefthook, CI, and local commands remain unchanged
- Consistent "always use Task" rule with zero exceptions
- Task is the SSOT interface for ALL environments (local, Lefthook, CI)

### Complex Multi-Action Example: License Compliance

When a domain requires multiple verification methods AND multiple output artifacts, the pattern scales elegantly with orchestrators:

```yaml
# CHECK - Multiple verification methods
check:license              # Orchestrator: runs all verification methods
check:license:source       # Source-based check (go-licenses, fast)
check:license:binary       # Binary-based check (lichen, accurate)

# GENERATE - Multiple artifact types
generate:license           # Orchestrator: creates all artifacts
generate:license:report    # Generate CSV report
generate:license:files     # Save license files to third_party/
generate:attribution       # Generate NOTICE file
```

**Why this works:**

1. **Orchestrators provide simple interface**: `task check:license` runs everything
2. **Subvariants allow granular control**: `task check:license:source` for fast dev feedback
3. **Clear action:target:variant pattern**: Each task name states exactly what it does
4. **Consistent with existing patterns**: Same structure as `check:deps`, `generate:docs`
5. **Scales to complexity**: Can add more variants without breaking pattern

**Usage patterns:**

```bash
# Development: fast feedback
task check:license:source        # ~2-5 seconds

# Release: accurate verification
task check:license:binary        # ~10-15 seconds

# CI: comprehensive check
task check:license               # Runs both via orchestrator

# Artifacts: generate all
task generate:license            # All reports + NOTICE
```

**Pattern clarity:**

- ✅ `check:license` = orchestrator (runs subvariants)
- ✅ `check:license:source` = specific check method
- ✅ `generate:license:report` = generate one artifact type
- ❌ `check:license:report` = WRONG (mixing check + generate actions)

See [ADR-011](011-license-compliance.md) for full implementation details.

## Related ADRs

- [ADR-001](001-ultra-thin-command-pattern.md) - Ultra-thin commands enforced via `task validate:commands`
- [ADR-002](002-centralized-configuration-registry.md) - Config registry enforced via `task validate:defaults`
- [ADR-005](005-auto-generated-config-constants.md) - Config constants enforced via `task validate:constants`
- [ADR-008](008-release-automation-with-goreleaser.md) - Release process uses `task check` as quality gate
- [ADR-011](011-license-compliance.md) - License compliance enforced via `task check:license` (dual-tool orchestrator)

## References

- `Taskfile.yml` - Single source of truth for all commands
- `.github/workflows/ci.yml` - CI implementation using Task
- `.lefthook.yml` - Pre-commit hooks using Task
- `CLAUDE.md` - Developer guidelines referencing Task commands
- `README.md` - User documentation showing Task workflow
- [Task Documentation](https://taskfile.dev/) - Task runner reference
