# Claude Code Guidelines for ckeletin-go

## TL;DR - Non-Negotiable Rules

**Memorize these 7 rules before doing anything else:**

1. **`task check` before every commit** - Non-negotiable, runs all quality checks
2. **Commands ≤30 lines** - `cmd/*.go` files wire things together; logic goes in `internal/`
3. **Use `config.Key*` constants** - Never hardcode config strings; run `task generate:config:key-constants` after registry changes
4. **Never reduce test coverage** - 80% minimum overall, use `testify/assert`
5. **Check licenses after `go get`** - Run `task check:license:source` immediately
6. **Never `--no-verify`** - Ask user permission first with justification
7. **Task for workflows, Go for debugging** - `task test` for full suite, `go test -v -run TestName` for debugging

## Quick Decision Trees

```
Where does this code go?
├── CLI command entry point? → cmd/<name>.go (≤30 lines)
├── Business logic? → internal/<name>/
├── Reusable public API? → pkg/
└── Test helpers? → test/ or *_test.go

Which command to run?
├── All tests? → task test
├── Debug one test? → go test -v -run TestName ./path/...
├── Before commit? → task check (MANDATORY)
├── Format code? → task format
└── Quick compile? → go build ./... (OK for iteration)

Which log level?
├── Can return this error? → log.Debug() + return err
├── User input error? → Formatted output only (no log)
├── Important event in normal flow? → log.Info()
├── Recoverable issue needing attention? → log.Warn()
└── Unrecoverable system failure/bug? → log.Error()
```

**Something broke?** → Jump to [Troubleshooting](#troubleshooting) for cascading failures, rollback, and recovery.

**When rules conflict, prioritize:** Security → License compliance → Correctness → Coverage → Style

<details>
<summary>📋 Table of Contents (click to expand)</summary>

- [TL;DR - Non-Negotiable Rules](#tldr---non-negotiable-rules)
- [About This Project](#about-this-project)
- [Getting Started](#getting-started)
- [Task Command Usage](#task-command-usage-critical)
- [Git Workflow](#git-workflow)
- [Code Quality Standards](#code-quality-standards)
- [Code Organization](#code-organization)
- [Architecture Decision Records](#architecture-decision-records-adrs)
- [Project-Specific Conventions](#project-specific-conventions)
- [License Compliance](#license-compliance)
- [Mistakes and Anti-Patterns](#mistakes-and-anti-patterns)
- [Troubleshooting](#troubleshooting)
- [Getting Help](#getting-help)

</details>

---

## About This Project

**ckeletin-go** is a Go-based CLI skeleton/template generator with:
- Ultra-thin command pattern (commands ≤30 lines)
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Bubble Tea for interactive UIs
- Comprehensive testing with high coverage standards
- Dependency injection over mocking

## Getting Started

**Platform:** This project is developed and tested on macOS and Linux. Windows is not officially supported.

### Automatic Setup
When you start a new session, development tools are automatically installed:
- Task runner (task command)
- Code formatters (goimports, gofmt)
- Linters (golangci-lint)
- Test runners (gotestsum)
- Security scanners (govulncheck)

**Important:** Tools install automatically when you start a session. The SessionStart hook (defined in `.claude/hooks.json`) runs `.ckeletin/scripts/install_tools.sh` which installs task, goimports, golangci-lint, and other required tools. You'll see a success message when ready.

If tools fail to install, run manually: `bash .ckeletin/scripts/install_tools.sh`

### After Upgrading Go

When upgrading Go versions (e.g., 1.25.3 → 1.25.4), rebuild dev tools to avoid compatibility issues:

```bash
task setup  # Rebuilds all dev tools with new Go version
```

**Why this matters:**
- Dev tools (go-licenses, golangci-lint, etc.) are compiled Go binaries
- They may be incompatible when compiled with an older Go version
- Common symptom: `go-licenses` failing with "package does not have module info" errors
- Solution: Rebuild tools with current Go version via `task setup`

**Detecting stale tools:**
```bash
task doctor  # Checks if tools were built with older Go version
```

### First Steps
1. Read `Taskfile.yml` to understand available commands
2. Review `README.md` for project overview
3. Check `.ckeletin/docs/adr/*.md` for architectural decisions
4. Understand the codebase structure before making changes

### First 5 Minutes Verification

After tools install, verify your environment:

```bash
task --list          # Should show all available tasks
go build ./...       # Should compile cleanly
task test            # Should pass with ≥80% coverage
```

If any fail, run `task setup` to rebuild tools, then retry.

## Task Command Usage (CRITICAL)

**Use `task` commands for standard workflows. Direct `go` commands are OK for debugging.**

### Quick Reference

| Scenario | Command |
|----------|---------|
| Build (official) | `task build` |
| Run all tests | `task test` |
| Format code | `task format` |
| Lint code | `task lint` |
| **Before commits** | `task check` |
| **Trivial changes** | `task check:fast` |
| Debug specific test | `go test -v -run TestName ./path/...` |
| Quick compile check | `go build ./...` |

### Essential Task Commands

**Daily workflow:** `task format` → `task test` → `task lint` → `task check`

<details>
<summary>📋 Full task list (click to expand)</summary>

| Command | Purpose |
|---------|---------|
| `task check` | Run ALL quality checks |
| `task format` | Format all Go code |
| `task lint` | Run golangci-lint |
| `task test` | Run tests with coverage |
| `task test:integration` | Run integration tests |
| `task bench` | Run benchmarks |
| `task check:vuln` | Check for vulnerabilities |
| `task check:deps` | Check dependency issues |
| `task generate:config:key-constants` | Regenerate config constants |

</details>

### Development Workflow

```
During development: task format → task test → task lint
Before committing:  task check (MANDATORY - runs everything)
Trivial changes:    task check:fast (docs, comments, typos only)
```

**When to use `task check:fast`:** For documentation-only changes, comment updates, or trivial typo fixes where full validation is overkill. Skips race detection and integration tests. Still runs format, lint, and unit tests. Use full `task check` for any code logic changes.

**What `task check` runs (in order):**
```
Code Quality        → format, lint
Architecture        → validate:defaults, commands, constants, task-naming,
                      architecture, layering, package-organization,
                      config-consumption, output, security
Security Scanning   → check:secrets, check:sast
Dependencies        → check:deps, check:license, check:sbom:vulns
Tests               → test:full (unit + integration + race detection)
```

**If `task check` fails:** Fix the issue, don't work around it. Common fixes:
- Format issues → `task format`
- Lint issues → Read output and fix code
- Test failures → Debug and fix tests
- Coverage drops → Add more tests

### When Direct Go Commands Are OK

Direct `go` commands are acceptable for **debugging and iteration**:

```bash
# Debug a specific test with verbose output
go test -v -run TestSpecificFunction ./internal/check/...

# Run tests with race detector for a specific package
go test -race ./pkg/checkmate/...

# Quick compile check while iterating
go build ./...

# Debug with delve
dlv test ./internal/config/...
```

**Always return to task commands for:**
- Final validation before commits → `task check`
- Official builds → `task build`
- Full test suite with coverage → `task test`
- Code formatting → `task format`

**The principle:** Task commands ensure all flags, coverage settings, and checks are applied consistently. Direct commands are fine for exploration, but finish with `task check`.

**IDE note:** Your editor may auto-format differently than `task format`. Always run `task format` before commits to ensure consistency, regardless of editor settings.

### If a Task Fails

**Don't work around failures - fix them!**

1. Read the error message completely
2. Check `Taskfile.yml` to understand what the task does
3. Fix the root cause (don't just make the error go away)
4. Re-run the task to verify
5. If stuck, ask the user for guidance

### Task Naming Convention

All tasks follow a simple `action:target` pattern:

```bash
# Action on target
task check:format          # Check format
task check:vuln            # Check vulnerabilities
task check:deps:verify     # Check deps, verify subvariant
task validate:commands     # Validate commands (enforces ADR-001)
task validate:constants    # Validate constants (enforces ADR-005)
task test:race             # Test with race detection
task test:integration      # Integration test
task generate:config:key-constants    # Generate constants
task build:release         # Build release
task clean:local           # Clean local artifacts

# Standalone actions
task format                # Format everything
task test                  # Test everything
task check                 # Check everything (orchestrator)
```

**Scripts are implementation details. Task is the interface.**

Lefthook, CI, and local all use Task commands. If you rename scripts, only Taskfile.yml changes.

See [ADR-000](.ckeletin/docs/adr/000-task-based-single-source-of-truth.md) for full details.

## Git Workflow

### Commit Process

**1. Run Quality Checks**
```bash
task check  # MANDATORY - must pass before commit
```

**2. Stage Your Changes**
```bash
git add <files>
# Or for all changes:
git add .
```

**3. Commit with Conventional Format**
```bash
git commit -m "type: concise summary

- Bullet point detail 1
- Bullet point detail 2
- Additional context if needed"
```

**4. Push to Branch**
```bash
git push -u origin <branch-name>
```

**⚠️ NEVER use `--no-verify`** - Do not skip pre-commit hooks. If you believe there's a legitimate reason to skip, you MUST:
1. Ask the user for permission first
2. Provide a clear justification for why skipping is necessary
3. Only proceed if explicitly approved

### Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <concise summary>

- <bullet point details>
- <additional details>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation only
- `test`: Adding or updating tests
- `refactor`: Code change that neither fixes a bug nor adds a feature
- `style`: Code style changes (formatting, missing semi colons, etc)
- `perf`: Performance improvement
- `build`: Changes to build system or dependencies
- `ci`: CI configuration changes
- `chore`: Other changes that don't modify src or test files

**Note:** This project uses `includeCoAuthoredBy: false` - commits do not include Claude Code attribution.

### Branch Naming
Use descriptive branch names with conventional prefixes:
- `feat/` - New features (e.g., `feat/add-user-auth`)
- `fix/` - Bug fixes (e.g., `fix/config-validation`)
- `refactor/` - Code refactoring (e.g., `refactor/logger-cleanup`)
- `docs/` - Documentation updates (e.g., `docs/readme-update`)

## Code Quality Standards

### Test Coverage Requirements

| Package Type | Minimum Coverage | Target Coverage |
|-------------|------------------|-----------------|
| Overall | 80% | 85%+ |
| `cmd/*` | 80% | 90%+ |
| `internal/config` | 80% | 90%+ |
| `internal/logger` | 80% | 90%+ |
| Other packages | 70% | 80%+ |

**How coverage is enforced:**
- Each package must meet its category minimum (70-80% depending on type)
- The overall project must meet 80%
- **Both conditions must pass.** A package at 65% fails even if overall is 85%.
- **Enforcement:** CI runs `.ckeletin/scripts/check-coverage-project.sh` which fails the build if thresholds aren't met. This is automated, not honor system.

**Rules:**
- **Maintain coverage thresholds in every PR.** During refactoring, temporary drops up to 2% are acceptable if restored before the PR merges.
- Add tests for all new features
- Add tests for all bug fixes
- Use table-driven tests for multiple scenarios
- Write clear test names that describe what's being tested

### Code Organization

```
ckeletin-go/
├── .claude/               # Claude Code config (hooks.json for auto-setup)
├── cmd/                    # Commands (ultra-thin, ≤30 lines each)
│   ├── root.go            # Root command setup
│   └── *.go               # Feature commands
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   │   ├── registry.go    # Config option definitions
│   │   └── keys_generated.go  # Auto-generated constants
│   ├── logger/            # Logging infrastructure
│   └── */                 # Other internal packages
├── pkg/                   # Public reusable libraries (importable by others)
│   └── checkmate/         # Beautiful terminal output for check results
├── test/
│   └── integration/       # Integration tests
├── docs/
│   └── adr/              # Architecture Decision Records
└── scripts/              # Build and utility scripts
```

**Key Principles:**
1. **Ultra-thin commands**: Commands in `cmd/` should be ≤30 lines
2. **Business logic in `internal/`**: Keep implementation details internal
3. **Follow ADRs**: Framework decisions in `.ckeletin/docs/adr/`, project decisions in `docs/adr/`

**30-line guidance:** Target ≤30 lines. Commands at 31-35 lines are acceptable if refactoring would reduce clarity. Beyond 35 lines requires refactoring to `internal/`. If you must exceed, add a comment explaining why.

**Example: Ultra-thin command (cmd/ping.go)**
```go
// cmd/ping.go - This is GOOD: wiring only, no business logic
package cmd

func runPing(cmd *cobra.Command, args []string) error {
    cfg := ping.Config{
        Message: getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage),
        Color:   getConfigValueWithFlags[string](cmd, "color", config.KeyAppPingOutputColor),
    }
    return ping.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
// Business logic lives in internal/ping/ping.go
```

**What "wiring" means:** Reading config, creating structs, calling into `internal/`. If you're writing loops, conditionals, or string manipulation in `cmd/`, move it to `internal/`.

### New Command Checklist

When adding a new command (e.g., `analyze`):

```
[ ] Create cmd/analyze.go (≤30 lines, wiring only)
[ ] Create internal/analyze/ package for business logic
[ ] Add config options to .ckeletin/pkg/config/registry.go
[ ] Run: task generate:config:key-constants
[ ] Write unit tests in internal/analyze/*_test.go
[ ] Add integration test in test/integration/ (if needed)
[ ] Update CHANGELOG.md
[ ] Run: task check (must pass)
```

### Architecture Decision Records (ADRs)

**MUST READ** `.ckeletin/docs/adr/*.md` before making architectural changes!

| ADR | Topic | Key Principle |
|-----|-------|---------------|
| ADR-000 | Task-Based Workflow | Single source of truth for dev commands |
| ADR-001 | Command Pattern | Commands are ultra-thin (≤30 lines) |
| ADR-002 | Config Registry | Centralized config with type safety |
| ADR-003 | Testing Strategy | Dependency injection over mocking |
| ADR-004 | Security | Input validation and safe defaults |
| ADR-005 | Config Constants | Auto-generated from registry |
| ADR-006 | Logging | Structured logging with Zerolog |
| ADR-007 | UI Framework | Bubble Tea for interactive UIs |
| ADR-008 | Release Automation | Multi-platform releases with GoReleaser |
| ADR-009 | Layered Architecture | 4-layer dependency rules |
| ADR-010 | Package Organization | pkg/ for public, internal/ for private |
| ADR-011 | License Compliance | Dual-tool license checking |
| ADR-012 | Dev Commands | Build tags for dev-only commands |
| ADR-013 | Structured Output | Shadow logging and checkmate patterns |

**Quick ADR lookup - "I'm working on..."**
| Task | Read |
|------|------|
| Adding a command | ADR-001, ADR-009 |
| Adding config option | ADR-002, ADR-005 |
| Writing tests | ADR-003 |
| Adding logging | ADR-006 |
| Adding dependency | ADR-011 |
| Creating UI | ADR-007 |

**When to Update ADRs:**
- Making architectural changes
- Changing fundamental patterns
- Introducing new core technologies
- Modifying build/deployment processes

## Project-Specific Conventions

### Configuration Management

**How ckeletin-go handles configuration:**

1. **Define in Registry** (`.ckeletin/pkg/config/registry.go`)
   ```go
   {
       Key:          "app.feature.enabled",
       DefaultValue: false,
       Description:  "Enable feature XYZ",
       Validation:   nil,
   }
   ```

2. **Generate Constants**
   ```bash
   task generate:config:key-constants
   # Creates .ckeletin/pkg/config/keys_generated.go
   ```

3. **Use Type-Safe Retrieval**
   ```go
   import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"

   enabled := viper.GetBool(config.KeyAppFeatureEnabled)
   ```

**Rules:**
- **Never** hardcode config keys as strings
- **Always** use generated constants from `config.Key*`
- **Always** run `task generate:config:key-constants` after registry changes
- Add validation functions for complex config values

### Logging Standards

**ckeletin-go uses structured logging with Zerolog:**

```go
import "github.com/rs/zerolog/log"

// Structured logging
log.Info().
    Str("component", "auth").
    Int("user_id", userID).
    Msg("User authenticated successfully")

// Error logging
log.Error().
    Err(err).
    Str("operation", "database_query").
    Msg("Failed to fetch user")
```

**Dual logging system:**
- **Console**: INFO+ level, colored, human-friendly
- **File**: DEBUG+ level, JSON format, for debugging

See [ADR-006](.ckeletin/docs/adr/006-structured-logging-with-zerolog.md) for details.

### Checkmate Library (pkg/checkmate/)

**What it does:** Beautiful terminal output for CLI check results with automatic TTY detection.

```go
import "github.com/peiman/ckeletin-go/pkg/checkmate"

p := checkmate.New()
p.CategoryHeader("Code Quality")
p.CheckHeader("Running linter...")
p.CheckSuccess("lint passed")
p.CheckFailure("format", "2 files need formatting", "Run: task format")
p.CheckSummary(checkmate.StatusSuccess, "All Checks Passed")
```

**Features:** Thread-safe, auto-detects TTY (colors in terminal, plain in CI), customizable themes, progress indicators.

**When to use:** Building CLI tools that run multiple checks and need consistent, beautiful output.

### Testing Standards

**Rules:**
- All new tests MUST use `testify/assert` or `testify/require`
- Use table-driven tests for multiple scenarios
- Unit tests: `*_test.go` in same package
- Integration tests: `test/integration/`

<details>
<summary>📋 Test template example (click to expand)</summary>

```go
import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "test", "test_processed", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ProcessFeature(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, got)
            }
        })
    }
}
```

</details>

### Golden File Testing

**Golden files** are reference snapshots of CLI output. **Never blindly update them - always review changes first!**

```bash
task test:golden         # Run golden tests
task test:golden:update  # Update (then review with git diff!)
```

<details>
<summary>📋 Golden file workflow (click to expand)</summary>

1. Make changes to output code
2. Run `task test:golden:update`
3. **CRITICAL:** Review with `git diff test/integration/testdata/`
4. Commit golden files WITH your code changes

**See [docs/testing.md](docs/testing.md) for full documentation.**

</details>

## License Compliance

**Rule:** Run `task check:license:source` before committing new dependencies.

| Allowed | Denied (will contaminate your code) |
|---------|-------------------------------------|
| MIT, Apache-2.0, BSD-2/3-Clause, ISC, 0BSD, Unlicense | GPL, AGPL, SSPL, LGPL, MPL |

| Task | When | Speed |
|------|------|-------|
| `task check:license:source` | Before committing deps | ~2-5s |
| `task check:license:binary` | Before release | ~10-15s |

**Transitive dependencies matter:** Even if you add a MIT-licensed package, if *that package* depends on GPL code, your project is contaminated. The license tools scan the entire dependency tree, not just direct imports. Always run checks after `go mod tidy`.

<details>
<summary>📋 Detailed license procedures (click to expand)</summary>

### When to Check Licenses

```bash
# After adding a dependency
go get github.com/example/package
task check:license:source  # Fast check

# Before releases
task check:license:binary  # Accurate check
```

### Handling Violations

```bash
# Remove violating dependency
go get github.com/example/gpl-lib@none
go mod tidy
task check:license:source
```

**Or find an alternative:** Search [pkg.go.dev](https://pkg.go.dev) for MIT/Apache-2.0 alternatives.

### Customizing Policy

**Via .lichen.yaml:**
```yaml
allow:
  - "MIT"
  - "Apache-2.0"
override:
  - path: "github.com/example/package"
    licenses: ["MIT"]
```

### Generating Artifacts

```bash
task generate:license:report   # CSV report
task generate:license:files    # License files for distribution
task generate:attribution      # NOTICE file
task generate:license          # All artifacts
```

**Full details:** [docs/licenses.md](docs/licenses.md) and [ADR-011](.ckeletin/docs/adr/011-license-compliance.md)

</details>

### Documentation Requirements

**Keep documentation up to date:**

1. **CHANGELOG.md** - For every user-facing change
   - Follow [Keep a Changelog](https://keepachangelog.com/) format
   - Add entries under `[Unreleased]` section
   - Categories: Added, Changed, Deprecated, Removed, Fixed, Security

2. **README.md** - For new features and major changes
   - Update usage examples
   - Update feature list
   - Keep installation instructions current

3. **ADRs** - For architectural decisions
   - Create new ADR for significant architectural changes
   - Follow existing ADR format
   - Number sequentially (ADR-001, ADR-002, etc.)

4. **Code Comments** - For complex logic
   - Use Go doc comments for public APIs
   - Explain "why" not "what"
   - Keep comments up to date with code

## Mistakes and Anti-Patterns

### Commands & Workflow

| ❌ Don't | ✅ Do | Why |
|----------|-------|-----|
| `go test ./...` for full suite | `task test` | Task runs coverage, gotestsum correctly |
| `goimports -w .` | `task format` | Task handles all formatting |
| `git commit` without checks | `task check && git commit` | Must pass checks first |
| Put logic in `cmd/*.go` | Put logic in `internal/*` | Commands ≤30 lines, wiring only |
| Use `sed`/`awk` for edits | Use the Edit tool | sed often corrupts code |

**Note:** `go test -v -run TestName` is fine for debugging. See "When Direct Go Commands Are OK".

### Configuration

| ❌ Don't | ✅ Do | Why |
|----------|-------|-----|
| Hardcode `"app.log.level"` | Use `config.KeyAppLogLevel` | Type-safe, refactor-friendly |
| Forget to regenerate constants | `task generate:config:key-constants` | Keep registry and constants in sync |

### Testing

| ❌ Don't | ✅ Do | Why |
|----------|-------|-----|
| Skip tests for "simple" code | Write tests anyway | 80% coverage is mandatory |
| Mock everything | Use dependency injection | Simpler, more maintainable ([ADR-003](.ckeletin/docs/adr/003-testing-strategy.md)) |
| Only run unit tests | Run `task test:integration` too | Integration tests catch real issues |

### Dependencies & Licensing

| ❌ Don't | ✅ Do | Why |
|----------|-------|-----|
| Add deps without license check | `go get pkg && task check:license:source` | Prevent GPL contamination |
| Forget CHANGELOG.md | Update with every change | Users need to know what changed |

### Logging

**Which log level?** (See decision tree in TL;DR)

| ❌ Don't | ✅ Do |
|----------|-------|
| `fmt.Println()` or basic log | `log.Info()`, `log.Error()` with structured fields |
| `log.Error()` for returnable errors | `log.Debug()` + `return err` |
| `log.Error()` for user input errors | Formatted output only (no log) |

Use `log.Error()` only for unrecoverable system failures. See [ADR-006](.ckeletin/docs/adr/006-structured-logging-with-zerolog.md).

### Code Quality & Linting

| ❌ Don't | ✅ Do | Why |
|----------|-------|-----|
| Delete unused variables without investigation | Investigate if they represent missing functionality first | May be planned code that wasn't wired up yet |
| Assume lint warnings are just cleanup | Check surrounding code, intent, and TODOs for context | Unused vars often signal forgotten implementations |

**When lint finds unused variables:**
1. **Investigate first** — Is this variable meant to be used somewhere that hasn't been implemented yet?
2. **Check context** — Look at surrounding code, function signatures, and commit history for clues
3. **Flag to user** — If it looks like missing functionality, ask before removing
4. **Only then remove** — Delete only after confirming it's truly dead code, not an incomplete implementation

## Troubleshooting

### Common Errors and Solutions

| Error | Cause | Solution |
|-------|-------|----------|
| `task: command not found` | Task not installed | Run `bash .ckeletin/scripts/install_tools.sh` or `go install github.com/go-task/task/v3/cmd/task@latest` |
| `go-licenses: package does not have module info` | Tools built with old Go version | Run `task setup` to rebuild tools |
| Coverage below 80% | Missing tests | Run `go tool cover -html=coverage.out` to see uncovered lines |
| License check fails | Copyleft dependency added | Remove dep with `go get pkg@none && go mod tidy`, find MIT alternative |
| `golangci-lint` timeout | Large codebase or slow machine | Run `task lint` (has proper timeout settings) |
| Validate commands fails | Command file too long | Move logic to `internal/` package, keep cmd file ≤30 lines |

### Local Passes but CI Fails

1. **Go version mismatch**: Check `.go-version` file matches your local Go
2. **Stale tools**: Run `task setup` to rebuild all dev tools
3. **Missing test dependencies**: Run `go mod tidy`
4. **Race conditions**: Run `task test:race` locally to reproduce

### When `task check` Fails Midway

The checks run in this order - find which category failed:
1. **Code Quality** (format, lint) - Run `task format`, then `task lint` to see issues
2. **Architecture** (validate:*) - Check the specific validator output
3. **Security** (secrets, sast) - Review flagged code patterns
4. **Dependencies** (deps, license) - Check for new/changed dependencies
5. **Tests** (test:full) - Run `task test` for detailed output

### Cascading Failures (Fix in This Order)

When one fix causes another failure, follow this triage order.

**Note:** This order is for *fixing* failures, not the execution order of `task check`. License issues block everything downstream, so fix them first even though `task check` runs format/lint earlier.

```
1. LICENSE VIOLATION (fix first - blocks everything)
   └→ Remove/replace the dependency
   └→ Run: go mod tidy && task check:license:source

2. BUILD FAILURE (fix second - can't test what won't compile)
   └→ Fix compilation errors
   └→ Run: go build ./...

3. LINT/FORMAT ERRORS (fix third - quick wins)
   └→ Run: task format
   └→ Fix remaining lint issues manually

4. TEST FAILURES (fix fourth)
   └→ Run: task test to see failures
   └→ Fix tests or code causing failures

5. COVERAGE DROP (fix last - depends on working tests)
   └→ Run: go tool cover -html=coverage.out
   └→ Add tests for uncovered lines
```

**Key principle:** Each step depends on the previous. Don't add coverage tests for code that fails lint, and don't fix lint for code that won't compile.

### Rollback and Recovery

| Situation | Action |
|-----------|--------|
| Commit broke CI | `git revert HEAD` to undo, then fix on new commit |
| `task format` mangled code | `git checkout -- <file>` to restore |
| Bad dependency added | `go get pkg@none && go mod tidy` |
| Need to abandon changes | `git stash` or `git checkout .` |
| Stuck in bad state | `git status`, commit/stash work, `task check` on clean state |

### Emergency Procedures

**When you MUST ship but checks are failing:**

1. **Identify the blocker category** - Is it license, build, lint, test, or coverage?

2. **Assess severity:**
   - **License violation** - STOP. Never ship GPL-contaminated code. Find alternative or remove dep.
   - **Build failure** - STOP. Can't ship what doesn't compile.
   - **Lint/format** - Can proceed with user approval if purely cosmetic.
   - **Test failure** - Assess: is the test wrong or the code? Flaky test can be skipped with justification.
   - **Coverage drop** - Can proceed if drop is <2% and documented in PR.

3. **Document the exception:**
   ```bash
   git commit -m "fix: emergency patch for X

   - Skipping Y check because: [reason]
   - Follow-up ticket: [link]
   - Approved by: [user]"
   ```

4. **Create follow-up immediately** - Don't let tech debt accumulate.

**When `--no-verify` is justified:**
- Pre-commit hook is broken (not just failing - actually broken)
- Emergency security patch where hook adds unacceptable delay
- User has explicitly approved after reviewing the justification

**Never justified:**
- "I'll fix it later"
- "The tests are flaky"
- "It works on my machine"

## Getting Help

### When Stuck
1. Check `Taskfile.yml` to understand what a task does
2. Review relevant ADR in `.ckeletin/docs/adr/` for architectural context
3. Look at similar code in the codebase for patterns
4. Ask the user for clarification when uncertain

### Key Resources
- **.ckeletin/docs/adr/ARCHITECTURE.md** - System structure (WHAT the system is: components, flows, interactions)
- **.ckeletin/docs/adr/*.md** - Architectural decisions (WHY it's this way: rationale, alternatives, consequences)
- **Taskfile.yml** - All available commands and their implementations
- **CHANGELOG.md** - History of changes
- **README.md** - Project overview and usage