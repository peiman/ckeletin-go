# Claude Code Guidelines for ckeletin-go

**Read [AGENTS.md](AGENTS.md) first** — it contains all project knowledge (architecture, commands, conventions, testing, licensing). This file contains Claude-specific behavioral rules only.

## Non-Negotiable Rules

1. **TDD: write tests FIRST, commit together** — Always write failing tests before implementation code. Test + implementation go in one atomic commit. Never commit tests without the code that makes them pass, or code without its tests
2. **`task check` before every commit** — Non-negotiable, runs all quality checks
3. **Commands ≤30 lines** — `cmd/*.go` files wire things together; logic goes in `internal/`
4. **Use `config.Key*` constants** — Never hardcode config strings; run `task generate:config:key-constants` after registry changes
5. **Never reduce test coverage** — 85% minimum overall, use `testify/assert`
6. **Check licenses after `go get`** — Run `task check:license:source` immediately
7. **Never `--no-verify`** — Ask user permission first with justification
8. **ALWAYS use `task` commands** — See `.claude/rules/task-commands.md` for the full translation table

**When rules conflict:** Security → License compliance → Correctness → Coverage → Style

## Command Translation (MANDATORY)

**STOP — use the task equivalent, not the raw command:**

| Instead of (NEVER) | Use (ALWAYS) |
|--------------------|--------------|
| `go test ./...` | `task test` |
| `go build` | `task build` |
| `golangci-lint run` | `task lint` |
| `goimports -w .` | `task format` |
| `go vet ./...` | `task lint` |
| `go mod tidy` | `task tidy` |
| Multiple checks manually | `task check` |

**ONLY exception:** `go test -v -run TestName ./path/...` for debugging a specific test.

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

## Claude-Specific Behaviors

- **Use the Edit tool** for file modifications — NEVER use `sed`, `awk`, or shell redirects to edit code
- **NEVER use `--no-verify`** on git commands. Only justified when: pre-commit hook is actually broken (not just failing), emergency security patch with user approval, or user has explicitly approved after reviewing justification. **Never justified:** "I'll fix it later", "The tests are flaky", "It works on my machine".
- **Unused variables**: When lint flags them, investigate intent before deleting. See `.claude/rules/unused-vars.md`.
- **Don't work around failures** — if `task check` fails, fix the root cause. Read the error output. Check `Taskfile.yml` to understand what the task does. If stuck, ask the user.
- **Don't propose changes to code you haven't read** — always read files before suggesting modifications
- **Read ADRs before architectural changes** — check `.ckeletin/docs/adr/*.md`

## Claude-Specific Setup

See `.claude/rules/claude-setup.md` for session initialization details.

Tools auto-install via SessionStart hook. If tools fail: `bash .ckeletin/scripts/install_tools.sh`

After Go upgrade: `task setup` to rebuild tools. Verify with: `task --list && task test`

## Anti-Patterns (Consolidated)

| DON'T | DO |
|-------|-----|
| `go test ./...` for full suite | `task test` |
| `goimports -w .` | `task format` |
| `git commit` without checks | `task check && git commit` |
| Put logic in `cmd/*.go` | Put logic in `internal/*` |
| Use `sed`/`awk` for edits | Use the Edit tool |
| Hardcode `"app.log.level"` | Use `config.KeyAppLogLevel` |
| Forget to regenerate constants | `task generate:config:key-constants` |
| Write implementation before tests | Write failing test FIRST, then implement (TDD) |
| Commit tests and implementation separately | Atomic commits: test + implementation together |
| Squash merge branches/PRs | Normal merge (preserve atomic commit history) |
| Skip tests for "simple" code | Write tests (85% coverage is mandatory) |
| Mock everything | Use dependency injection ([ADR-003]) |
| Add deps without license check | `go get pkg && task check:license:source` |
| `fmt.Println()` for logging | `log.Info()` with structured fields |
| `log.Error()` for returnable errors | `log.Debug()` + `return err` |
| Delete unused vars without checking | Investigate if they represent missing functionality |

## Known Rule Violations (These Have Happened Before)

- Writing implementation code before writing tests (TDD violation)
- Running `go test ./...` instead of `task test`
- Deleting unused variables without investigating if they represent planned functionality
- Using raw `go`/`golangci-lint`/`goimports` commands instead of `task` equivalents
- Using `sed` to edit files instead of the Edit tool
