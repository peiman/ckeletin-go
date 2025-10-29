# Claude Code Guidelines for ckeletin-go

This document provides guidelines specifically for Claude Code when working on the ckeletin-go project.

## Automatic Setup (SessionStart Hook)

**This project has a SessionStart hook that automatically sets up the development environment!**

When you start a new Claude Code session:
- `scripts/install_tools.sh` runs automatically via `.claude/hooks.json`
- All required development tools are installed (goimports, golangci-lint, gotestsum, etc.)
- Tools are installed in `/root/go/bin` and added to PATH
- The setup is idempotent - it only installs missing tools
- You'll see a success message when tools are ready

### Hook Configuration
Located in `.claude/hooks.json`:
```json
{
  "hooks": {
    "SessionStart": [
      {
        "matcher": "startup",
        "hooks": [
          {
            "type": "command",
            "command": "bash \"$CLAUDE_PROJECT_DIR\"/scripts/install_tools.sh"
          }
        ]
      }
    ]
  }
}
```

### Why This Matters
- No manual `task setup` needed
- Tools are always available in new sessions
- Consistent development environment
- Faster onboarding for new sessions

## Task Command Usage (CRITICAL)

**ALWAYS use task commands instead of direct go/script commands.**

### At the Start of Each Session
1. Tools are automatically installed by the SessionStart hook (no action needed!)
2. Read `Taskfile.yml` to understand available task commands
3. Before making any changes, understand the project structure

### Common Task Commands

**Instead of direct commands, use task:**

| ❌ DON'T USE | ✅ USE INSTEAD |
|-------------|---------------|
| `goimports -w .` | `task format` |
| `gofmt -w .` | `task format` |
| `./scripts/format-go.sh check` | `task format:check` |
| `golangci-lint run` | `task lint` |
| `go test ./...` | `task test` |
| `go test -bench` | `task bench` |
| `govulncheck ./...` | `task vuln` |
| Multiple checks | `task check` |

### Before Committing Changes
1. **ALWAYS** run `task check` before any commit
2. Fix all linter and test failures
3. Ensure formatting is correct (`task format`)
4. Verify dependencies (`task deps:check`)

### Task Command Hierarchy
- `task check` - Runs ALL quality checks (format, lint, test, deps, etc.)
- `task format` - Formats all Go code (goimports + gofmt)
- `task format:check` - Checks formatting without modifying files
- `task test` - Runs tests with coverage
- `task bench` - Runs performance benchmarks
- `task setup` - Installs all development tools

### If a Task Command Fails
1. Read the error message carefully
2. Check `Taskfile.yml` to understand what the task does
3. Fix the underlying issue (don't work around it)
4. Re-run the task command to verify the fix

## Git Workflow

### Committing Changes
1. Run `task check` first
2. Stage changes: `git add <files>`
3. Commit with conventional commit format (see .cursor/rules/before-git-commit.mdc)
4. Push to remote branch

### Commit Message Format
```
<type>: <concise summary>

- <bullet point details>
- <additional details>
```

Types: feat, fix, docs, test, refactor, style, ci, build, deps, perf, chore

### Branch Naming
- System enforces `claude/` prefix and session ID suffix
- Use descriptive names in the middle: `claude/descriptive-name-<session-id>`

## Code Quality Standards

### Test Coverage
- Minimum 80% overall coverage
- Aim for 90%+ on critical packages (cmd, internal/config)
- Add tests for all new features and bug fixes

### Code Organization
- Commands in `cmd/` (ultra-thin, ~20-30 lines)
- Business logic in `internal/`
- Configuration in `internal/config/`
- Follow existing patterns (see docs/adr/)

### Architecture Decision Records (ADRs)
- Read `docs/adr/` to understand architectural decisions
- Follow established patterns:
  - Ultra-thin command pattern (ADR-001)
  - Centralized configuration registry (ADR-002)
  - Dependency injection over mocking (ADR-003)
  - Security validation (ADR-004)
  - Auto-generated config constants (ADR-005)
  - Structured logging with Zerolog (ADR-006)
  - Bubble Tea for UI (ADR-007)

## Project-Specific Conventions

### Configuration Management
- All config options defined in `internal/config/registry.go`
- Use `getConfigValueWithFlags[T]()` for type-safe retrieval
- Auto-generated constants in `internal/config/keys_generated.go`
- Run `task generate:constants` after registry changes

### Testing Standards
- Use table-driven tests
- Clear phases: setup, execution, assertion
- Integration tests in `test/integration/`
- Benchmarks in `*_bench_test.go` files

### Documentation Requirements
- Update CHANGELOG.md for user-facing changes
- Keep a Changelog format (https://keepachangelog.com/)
- Semantic versioning (https://semver.org/)
- Update README.md for new features

## Common Mistakes to Avoid

1. ❌ Running direct go commands instead of task commands
2. ❌ Committing without running `task check`
3. ❌ Adding logic to command files (keep them ultra-thin)
4. ❌ Hardcoding config keys (use generated constants)
5. ❌ Skipping tests or reducing coverage
6. ❌ Not updating CHANGELOG.md for user-facing changes
7. ❌ Using mocking frameworks (use dependency injection)

## Quick Reference

### Starting Work
```bash
# Tools are auto-installed by SessionStart hook!

# Read the Taskfile
cat Taskfile.yml

# Understand the codebase
ls -la
cat README.md
cat docs/adr/*.md
```

### During Development
```bash
# Format code
task format

# Run linters
task lint

# Run tests
task test

# Check everything
task check
```

### Before Pushing
```bash
# Final quality check
task check

# Stage and commit
git add <files>
git commit -m "type: description"

# Push
git push -u origin <branch-name>
```

## Questions?
- Check `Taskfile.yml` for available commands
- Read `.cursor/rules/*.mdc` for detailed conventions
- Review `docs/adr/*.md` for architectural decisions
- Ask the user for clarification when uncertain
