# ckeletin-go — Project Guide for AI Agents

> **This file is a reference implementation.** The pattern — structured project
> guide, behavioral rules, automated hooks, machine-checkable enforcement — works
> in any codebase. See the README for how the pieces fit together.

## About This Project

**ckeletin-go** is a production-ready Go CLI scaffold powered by an updatable framework layer — built for humans and AI agents alike.

The `.ckeletin/` directory contains the **framework** — config registry, logging, validation scripts, task definitions, and ADRs (000-099). Your code lives in `cmd/`, `internal/`, `pkg/`. Framework updates via `task ckeletin:update` without touching your code.

Every architectural rule in this project is machine-checkable. `task check` is the single gateway — run it before every commit. If it passes, the code is correct regardless of who wrote it.

Key characteristics:
- Ultra-thin command pattern (commands ≤30 lines, logic in `internal/`)
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Framework-level `--output json` flag for machine-readable output
- Bubble Tea for interactive UIs
- Test-driven development (TDD) — tests first, always
- Dependency injection over mocking
- 85% minimum test coverage, enforced by CI
- Public reusable packages in `pkg/` (e.g., `checkmate` for beautiful CLI output)

**Platform:** macOS and Linux (primary). Windows is supported for core functionality; interactive features (TUI, colored output) may have limitations.

## Commands

Use `task` commands for all standard workflows. The `task` runner wraps Go tooling with correct flags, coverage settings, and checks.

| Scenario | Command |
|----------|---------|
| Build | `task build` |
| Run all tests | `task test` |
| Format code | `task format` |
| Lint code | `task lint` |
| Before commits | `task check` |
| Trivial changes only | `task check:fast` |
| Debug one test | `go test -v -run TestName ./path/...` |
| Quick compile check | `go build ./...` |
| Run benchmarks | `task bench` |
| Integration tests | `task test:integration` |
| Vulnerability check | `task check:vuln` |
| Regenerate config constants | `task generate:config:key-constants` |
| Generate config JSON Schema | `task generate:config:schema` |

**Daily workflow:** `task format` → `task test` → `task lint` → `task check`

**`task check:fast`** skips race detection and integration tests. Use only for docs, comments, or typo fixes. Use full `task check` for any code logic changes.

**What `task check` runs (in order):**
```
Code Quality        → format, lint
Architecture        → validate:defaults, commands, constants, task-naming,
                      architecture, layering, package-organization,
                      config-consumption, output, security, dev-build-tags
Security Scanning   → check:secrets, check:sast
Dependencies        → check:deps, check:license, check:sbom:vulns
Tests               → test:full (unit + integration + race detection)
```

**If `task check` fails:** Fix the issue, don't work around it.
- Format issues → `task format`
- Lint issues → Read output and fix code
- Test failures → Debug and fix tests
- Coverage drops → Add more tests

## Code Organization

```
ckeletin-go/
├── .ckeletin/             # Framework layer (upstream template)
│   ├── docs/adr/          # Framework ADRs (000-014)
│   ├── pkg/config/        # Config registry, constants, validation
│   │   ├── registry.go    # Config option definitions
│   │   └── keys_generated.go  # Auto-generated constants
│   ├── pkg/logger/        # Logging infrastructure (Zerolog)
│   ├── scripts/           # Build, validation, and utility scripts
│   └── Taskfile.yml       # Framework task definitions
├── cmd/                   # Commands (ultra-thin, ≤30 lines each)
│   ├── root.go            # Root command setup
│   └── *.go               # Feature commands
├── internal/              # Private application code
│   ├── check/             # Check command (executor, timing, checks)
│   ├── dev/               # Dev command logic
│   ├── ping/              # Ping command logic
│   └── */                 # Other internal packages
├── pkg/                   # Public reusable libraries (importable by others)
│   └── checkmate/         # Beautiful terminal output for check results
├── test/integration/      # Integration tests
├── docs/adr/              # Project-specific ADRs
├── Taskfile.yml           # Project tasks (includes .ckeletin/Taskfile.yml)
├── AGENTS.md              # This file (universal project guide)
└── CLAUDE.md              # Claude Code-specific behavioral rules
```

**Key principles:**
1. **Ultra-thin commands**: `cmd/*.go` files are wiring only (≤30 lines) — read config, create structs, call `internal/`. Loops, conditionals, or string manipulation → move to `internal/`.
2. **Business logic in `internal/`**: Private implementation packages.
3. **Framework code in `.ckeletin/`**: Config registry, logger, scripts, validators.
4. **Public libraries in `pkg/`**: Importable by external consumers.

**30-line guidance:** Target ≤30. 31-35 acceptable if refactoring reduces clarity. Beyond 35 requires refactoring. Example:
```go
// cmd/ping.go — wiring only, no business logic
func runPing(cmd *cobra.Command, args []string) error {
    cfg := ping.Config{
        Message: getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage),
        Color:   getConfigValueWithFlags[string](cmd, "color", config.KeyAppPingOutputColor),
    }
    return ping.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
```

## Architecture Decision Records (ADRs)

Read `.ckeletin/docs/adr/*.md` before making architectural changes.

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
| ADR-014 | Enforcement Policy | Every ADR must have automated enforcement |

**Quick lookup — "I'm working on..."**

| Task | Read |
|------|------|
| Adding a command | ADR-001, ADR-009 |
| Adding config option | ADR-002, ADR-005 |
| Writing tests | ADR-003 |
| Adding logging | ADR-006 |
| Adding dependency | ADR-011 |
| Creating UI | ADR-007 |
| Adding/modifying an ADR | ADR-014 |

Every ADR must have an `## Enforcement` section ([ADR-014](.ckeletin/docs/adr/014-adr-enforcement-policy.md)).

## Conventions

### Configuration Management

1. **Define** in `.ckeletin/pkg/config/registry.go`
2. **Generate** constants: `task generate:config:key-constants` → creates `keys_generated.go`
3. **Use** type-safe retrieval: `viper.GetBool(config.KeyAppFeatureEnabled)`

Rules:
- Never hardcode config keys as strings — use `config.Key*` constants
- Always run `task generate:config:key-constants` after registry changes
- Add validation functions for complex config values

### Logging

Zerolog structured logging with dual output:
- **Console**: INFO+ level, colored, human-friendly
- **File**: DEBUG+ level, JSON format

Log level rules:
- Can return this error? → `log.Debug()` + `return err`
- User input error? → Formatted output only (no log)
- Important normal flow event? → `log.Info()`
- Recoverable issue? → `log.Warn()`
- Unrecoverable system failure? → `log.Error()`

Use `log.Error()` only for unrecoverable failures where no error can be returned. Semgrep rule `ckeletin-log-error-and-return` enforces this. See [ADR-006](.ckeletin/docs/adr/006-structured-logging-with-zerolog.md).

### JSON Output Mode (`--output json`)

Every command supports `--output json` for machine-readable output. When active:
- Stdout emits exactly one JSON envelope
- Stderr is silenced (zerolog disabled)
- Audit log file continues unchanged

**JSON envelope structure:**
```json
{
  "status": "success",
  "command": "ping",
  "data": { "message": "Hello", "color": "white", "timestamp": "now" },
  "error": null
}
```

On error:
```json
{
  "status": "error",
  "command": "ping",
  "data": null,
  "error": { "message": "invalid color value", "code": "CONFIG_VALIDATION" }
}
```

**Config:** `app.output_format` (default: `"text"`, valid: `"text"` or `"json"`)
**Env var:** `CKELETINGO_APP_OUTPUT_FORMAT=json`
**Constant:** `config.KeyAppOutputFormat`

**How it works for command authors:**
- Commands that call `ui.RenderSuccess(out, message, data)` get JSON for free — the `data` argument becomes the envelope's `.data` field
- For custom JSON shapes, implement `output.JSONResponder` on your data type:
  ```go
  type MyResult struct { /* fields */ }
  func (r MyResult) JSONResponse() interface{} { return r }
  ```
- The `check` command uses `JSONResponder` to emit a flat list of results instead of its internal state

**Types (in `.ckeletin/pkg/output/json.go`):**
- `output.JSONEnvelope` — the standard envelope wrapper
- `output.JSONError` — structured error with message and optional code
- `output.JSONResponder` — interface for custom JSON data shapes
- `output.IsJSONMode()` — check if JSON mode is active
- `output.RenderJSON(out, envelope)` — marshal envelope to writer

### Testing

- **TDD is mandatory** — Write failing tests FIRST, then implement to make them pass. Test + implementation are committed together as one atomic unit. Never commit tests without the code that makes them pass, or code without its tests
- All tests must use `testify/assert` or `testify/require`
- Use table-driven tests for multiple scenarios
- Unit tests: `*_test.go` in same package
- Integration tests: `test/integration/`
- Dependency injection over mocking ([ADR-003](.ckeletin/docs/adr/003-testing-strategy.md))

### Golden File Testing

Golden files are reference snapshots of CLI output. Never blindly update them.

```bash
task test:golden         # Run golden tests
task test:golden:update  # Update (then review with git diff!)
```

After updating: `git diff test/integration/testdata/` — review every change. See [docs/testing.md](docs/testing.md).

### Checkmate Library (pkg/checkmate/)

Beautiful terminal output for CLI check results. Thread-safe, auto-detects TTY (colors in terminal, plain in CI), customizable themes.

```go
p := checkmate.New()
p.CategoryHeader("Code Quality")
p.CheckSuccess("lint passed")
p.CheckFailure("format", "2 files need formatting", "Run: task format")
```

## Git Workflow

[Conventional Commits](https://www.conventionalcommits.org/) format:
```
<type>: <concise summary>

- <bullet point details>
```

**Types:** `feat`, `fix`, `docs`, `test`, `refactor`, `style`, `perf`, `build`, `ci`, `chore`

**Branch naming:** `feat/`, `fix/`, `refactor/`, `docs/` prefixes (e.g., `feat/add-user-auth`)

**Atomic commits:** Tests and the implementation they cover go in the same commit. Every commit should be a complete, passing unit. Never split tests from their implementation across separate commits.

**Normal merge, never squash.** This project uses normal merge (merge commits) — not squash merge. Every atomic commit is preserved on main. This is why atomic commits matter: they survive the merge and keep `git bisect`, `git log`, and the TDD narrative intact. Do not squash when merging branches or PRs.

`task check` must pass before every commit.

## Code Quality

### Test Coverage Requirements

| Package Type | Minimum | Target |
|-------------|---------|--------|
| Overall | 85% | 90%+ |
| `cmd/*` | 80% | 90%+ |
| `.ckeletin/pkg/config` | 80% | 90%+ |
| `.ckeletin/pkg/logger` | 80% | 90%+ |
| Other packages | 70% | 80%+ |

Both per-package and overall thresholds must pass. CI runs `.ckeletin/scripts/check-coverage-project.sh`.

**Exclusions:** TUI code (`*_tui.go`, `internal/check/executor.go`, `internal/check/summary.go`) and `/demo/` directories.

During refactoring, temporary drops up to 2% acceptable if restored before PR merges.

### New Command Checklist

```
[ ] Create cmd/<name>.go (≤30 lines, wiring only)
[ ] Create internal/<name>/ package for business logic
[ ] Add config options to .ckeletin/pkg/config/registry.go
[ ] Run: task generate:config:key-constants
[ ] Write failing tests FIRST in internal/<name>/*_test.go (TDD)
[ ] Implement code to make tests pass
[ ] Add integration test in test/integration/ (if needed)
[ ] Update CHANGELOG.md
[ ] Run: task check (must pass)
```

## License Compliance

Run `task check:license:source` before committing new dependencies.

| Allowed | Denied |
|---------|--------|
| MIT, Apache-2.0, BSD-2/3-Clause, ISC, 0BSD, Unlicense | GPL, AGPL, SSPL, LGPL, MPL |

| Task | When | Speed |
|------|------|-------|
| `task check:license:source` | Before committing deps | ~2-5s |
| `task check:license:binary` | Before release | ~10-15s |

Transitive dependencies matter — if a MIT package depends on GPL code, your project is contaminated. Always run checks after `go mod tidy`.

To remove a violating dependency: `go get pkg@none && go mod tidy`

Details: [docs/licenses.md](docs/licenses.md) and [ADR-011](.ckeletin/docs/adr/011-license-compliance.md)

## Documentation

- **CHANGELOG.md**: Every user-facing change, [Keep a Changelog](https://keepachangelog.com/) format, under `[Unreleased]`
- **README.md**: Update for new features and major changes
- **ADRs**: New ADR for significant architectural changes, numbered sequentially

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| `task: command not found` | Task not installed | `bash .ckeletin/scripts/install_tools.sh` |
| `go-licenses: package does not have module info` | Tools built with old Go | `task setup` |
| Coverage below 85% | Missing tests | `go tool cover -html=coverage.out` to find gaps |
| License check fails | Copyleft dep added | `go get pkg@none && go mod tidy`, find MIT alternative |
| `golangci-lint` timeout | Slow machine | `task lint` (has proper timeout) |
| Validate commands fails | cmd file too long | Move logic to `internal/`, keep ≤30 lines |

**Local passes but CI fails:**
1. Go version mismatch — check `.go-version`
2. Stale tools — `task setup`
3. Missing deps — `go mod tidy`
4. Race conditions — `task test:race` locally

**Cascading failures — fix in this order:**
1. License violation → remove/replace dep, `go mod tidy && task check:license:source`
2. Build failure → fix compilation, `go build ./...`
3. Lint/format → `task format`, fix remaining manually
4. Test failures → `task test`, fix tests or code
5. Coverage drop → `go tool cover -html=coverage.out`, add tests

Each step depends on the previous. Don't fix coverage for code that fails lint.

## Key Resources

- **.ckeletin/docs/adr/ARCHITECTURE.md** — System structure
- **.ckeletin/docs/adr/*.md** — Architectural decisions
- **.semgrep.yml** — Custom SAST rules
- **Taskfile.yml** — All commands and implementations
- **CHANGELOG.md** — History of changes
- **README.md** — Project overview and usage
