# Claude Code Guidelines for ckeletin-go

This document provides guidelines specifically for Claude Code when working on the ckeletin-go project.

## About This Project

**ckeletin-go** is a Go-based CLI skeleton/template generator with:
- Ultra-thin command pattern (commands are ~20-30 lines)
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Bubble Tea for interactive UIs
- Comprehensive testing with high coverage standards
- Dependency injection over mocking

## Getting Started

### Automatic Setup
When you start a new session, development tools are automatically installed:
- Task runner (task command)
- Code formatters (goimports, gofmt)
- Linters (golangci-lint)
- Test runners (gotestsum)
- Security scanners (govulncheck)

**Important:** Tools install automatically via SessionStart hook. You'll see a success message when ready.

### First Steps
1. Read `Taskfile.yml` to understand available commands
2. Review `README.md` for project overview
3. Check `docs/adr/*.md` for architectural decisions
4. Understand the codebase structure before making changes

## Task Command Usage (CRITICAL)

**ALWAYS use `task` commands - never run go/script commands directly.**

### Essential Task Commands

| Command | Purpose | When to Use |
|---------|---------|-------------|
| `task check` | Run ALL quality checks | **Before every commit** |
| `task format` | Format all Go code | When code needs formatting |
| `task format:check` | Check formatting (no changes) | In CI or verification |
| `task lint` | Run golangci-lint | Fix code quality issues |
| `task test` | Run tests with coverage | During development |
| `task test:integration` | Run integration tests | Full system testing |
| `task bench` | Run benchmarks | Performance validation |
| `task vuln` | Check for vulnerabilities | Security audit |
| `task deps:check` | Check dependency updates | Maintenance |
| `task generate:constants` | Regenerate config constants | After registry changes |

### Development Workflow

**1. Before You Start Coding**
```bash
# Understand available commands
task --list

# Read the Taskfile to understand what each task does
cat Taskfile.yml
```

**2. During Development**
```bash
# Format your code frequently
task format

# Run tests for the package you're working on
task test

# Check for linting issues
task lint
```

**3. Before Committing (MANDATORY)**
```bash
# Run ALL checks - this is non-negotiable
task check

# Fix any failures before committing
# Common fixes:
# - Format issues: task format
# - Lint issues: Read golangci-lint output and fix
# - Test failures: Debug and fix the tests
# - Coverage drops: Add more tests
```

### Why Task Commands Matter

1. **Consistency**: Everyone runs the same checks
2. **Completeness**: `task check` runs everything needed
3. **Efficiency**: Tasks are optimized and cached
4. **Documentation**: Taskfile.yml is self-documenting
5. **CI Alignment**: Local tasks match CI pipeline

### If a Task Fails

**Don't work around failures - fix them!**

1. Read the error message completely
2. Check `Taskfile.yml` to understand what the task does
3. Fix the root cause (don't just make the error go away)
4. Re-run the task to verify
5. If stuck, ask the user for guidance

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
- Branches use `claude/` prefix with session ID suffix
- Example: `claude/add-logging-feature-<session-id>`
- The system enforces this automatically

## Code Quality Standards

### Test Coverage Requirements

| Package Type | Minimum Coverage | Target Coverage |
|-------------|------------------|-----------------|
| Overall | 80% | 85%+ |
| `cmd/*` | 80% | 90%+ |
| `internal/config` | 80% | 90%+ |
| `internal/logger` | 80% | 90%+ |
| Other packages | 70% | 80%+ |

**Rules:**
- **Never** reduce test coverage
- Add tests for all new features
- Add tests for all bug fixes
- Use table-driven tests for multiple scenarios
- Write clear test names that describe what's being tested

### Code Organization

```
ckeletin-go/
├── cmd/                    # Commands (ultra-thin, 20-30 lines each)
│   ├── root.go            # Root command setup
│   └── *.go               # Feature commands
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   │   ├── registry.go    # Config option definitions
│   │   └── keys_generated.go  # Auto-generated constants
│   ├── logger/            # Logging infrastructure
│   └── */                 # Other internal packages
├── test/
│   └── integration/       # Integration tests
├── docs/
│   └── adr/              # Architecture Decision Records
└── scripts/              # Build and utility scripts
```

**Key Principles:**
1. **Ultra-thin commands**: Commands in `cmd/` should be ~20-30 lines
2. **Business logic in `internal/`**: Keep implementation details internal
3. **Follow ADRs**: All architectural decisions are documented in `docs/adr/`

### Architecture Decision Records (ADRs)

**MUST READ** `docs/adr/*.md` before making architectural changes!

| ADR | Topic | Key Principle |
|-----|-------|---------------|
| ADR-000 | Task-Based Workflow (Foundational) | Single source of truth for dev commands |
| ADR-001 | Command Pattern | Commands are ultra-thin (~20-30 lines) |
| ADR-002 | Config Registry | Centralized config with type safety |
| ADR-003 | Testing Strategy | Dependency injection over mocking |
| ADR-004 | Security | Input validation and safe defaults |
| ADR-005 | Config Constants | Auto-generated from registry |
| ADR-006 | Logging | Structured logging with Zerolog |
| ADR-007 | UI Framework | Bubble Tea for interactive UIs |
| ADR-008 | Release Automation | Multi-platform releases with GoReleaser |

**When to Update ADRs:**
- Making architectural changes
- Changing fundamental patterns
- Introducing new core technologies
- Modifying build/deployment processes

## Project-Specific Conventions

### Configuration Management

**How ckeletin-go handles configuration:**

1. **Define in Registry** (`internal/config/registry.go`)
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
   task generate:constants
   # Creates internal/config/keys_generated.go
   ```

3. **Use Type-Safe Retrieval**
   ```go
   import "github.com/yourusername/ckeletin-go/internal/config"

   enabled := viper.GetBool(config.KeyAppFeatureEnabled)
   ```

**Rules:**
- **Never** hardcode config keys as strings
- **Always** use generated constants from `config.Key*`
- **Always** run `task generate:constants` after registry changes
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

See ADR-006 for details.

### Testing Standards

**Structure your tests like this:**

```go
func TestFeature(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test",
            expected: "test_processed",
            wantErr:  false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup

            // Execute
            got, err := ProcessFeature(tt.input)

            // Assert
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

**Testing Locations:**
- Unit tests: `*_test.go` in same package
- Integration tests: `test/integration/`
- Benchmarks: `*_bench_test.go`

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

## Common Mistakes to Avoid

### Critical Errors

| ❌ Don't Do This | ✅ Do This Instead | Why |
|------------------|-------------------|-----|
| `go test ./...` | `task test` | Task runs coverage, gotestsum, and formatting |
| `goimports -w .` | `task format` | Task handles all formatting consistently |
| `git commit` without checks | `task check && git commit` | Must pass all checks before committing |
| Hardcode `"app.log.level"` | Use `config.KeyAppLogLevel` | Type-safe, refactor-friendly |
| Put logic in `cmd/*.go` | Put logic in `internal/*` | Commands must be ultra-thin (~20-30 lines) |
| Skip tests for "simple" code | Write tests anyway | Coverage requirements are mandatory |
| Mock everything | Use dependency injection | Simpler, more maintainable (ADR-003) |
| Forget CHANGELOG.md | Update it with every change | Users need to know what changed |

### Anti-Patterns Specific to ckeletin-go

1. **Fat Commands** - Commands with >30 lines of logic
   - Move logic to `internal/` packages
   - Commands should only wire things together

2. **Config Key Magic Strings** - Using raw strings for config keys
   - Always use generated constants
   - Run `task generate:constants` after registry changes

3. **Unstructured Logging** - Using `fmt.Println()` or basic log
   - Use `log.Info()`, `log.Error()`, etc.
   - Add structured fields with `.Str()`, `.Int()`, etc.

4. **Skipping Integration Tests** - Only running unit tests
   - Run `task test:integration` for full coverage
   - Integration tests catch real-world issues

## Quick Reference

### New Session Checklist
```bash
# 1. Wait for automatic tool installation
# (SessionStart hook installs task, goimports, golangci-lint, etc.)

# 2. Understand the project
task --list                    # See all available tasks
cat Taskfile.yml              # Understand what tasks do
cat README.md                 # Project overview
ls docs/adr/                  # Review architectural decisions

# 3. Check your environment
task check                    # Ensure everything works
```

### Development Cycle
```bash
# Write code
vim internal/mypackage/feature.go

# Format frequently
task format

# Test your changes
task test

# Check for issues
task lint

# Before committing
task check                    # Run ALL checks

# Commit with conventional format
git add .
git commit -m "feat: add new feature

- Implemented XYZ functionality
- Added tests with 90% coverage
- Updated CHANGELOG.md"

# Push
git push -u origin <branch-name>
```

### Common Tasks Reference

| Task | When to Run | What It Does |
|------|-------------|-------------|
| `task check` | Before EVERY commit | Runs all quality checks |
| `task format` | Multiple times during dev | Formats all Go code |
| `task lint` | When check fails | Shows detailed lint issues |
| `task test` | After code changes | Runs tests with coverage |
| `task generate:constants` | After config registry changes | Regenerates config constants |
| `task deps:check` | Weekly/monthly | Checks for dependency updates |
| `task vuln` | Before releases | Scans for vulnerabilities |
| `task release:check` | Before creating releases | Checks if GoReleaser is installed |
| `task release:test` | Before tagging | Tests release build locally |
| `task release:clean` | After testing releases | Cleans GoReleaser artifacts |

## Getting Help

### When Stuck
1. Check `Taskfile.yml` to understand what a task does
2. Review relevant ADR in `docs/adr/` for architectural context
3. Look at similar code in the codebase for patterns
4. Check `.cursor/rules/*.mdc` for detailed conventions
5. Ask the user for clarification when uncertain

### Key Resources
- **Taskfile.yml** - All available commands and their implementations
- **docs/adr/** - Architectural decisions and rationale
- **CHANGELOG.md** - History of changes
- **README.md** - Project overview and usage
- **.cursor/rules/** - Detailed development rules and conventions
