# ckeletin-go — Project Guide for AI Agents

ckeletin-go is a Go CLI scaffold. `.ckeletin/` is the updatable framework (config, logging, validators, ADRs 000-099). Application code lives in `cmd/`, `internal/`, and `pkg/`. `task ckeletin:update` replaces the framework and does not touch application code.

Every architectural rule is machine-checkable. `task check` is the gate before every commit.

## Commands

Use `task` for standard workflows. It wraps Go tooling with the flags, coverage, and checks this repo expects.

| Scenario | Command |
|----------|---------|
| Build | `task build` |
| Run all tests | `task test` |
| Format | `task format` |
| Lint | `task lint` |
| Before every commit | `task check` |
| Docs, comments, typos only | `task check:fast` |
| Debug one test | `go test -v -run TestName ./path/...` |
| Quick compile | `go build ./...` |
| Tidy modules | `task tidy` |
| Benchmarks | `task bench` |
| Integration tests | `task test:integration` |
| Vulnerability check | `task check:vuln` |
| Regenerate config constants | `task generate:config:key-constants` |
| Generate config JSON Schema | `task generate:config:schema` |

`task check:fast` skips race detection and integration tests. Any logic change uses full `task check`.

`task check` runs, in order: format and lint; architecture validators (defaults, commands, constants, task naming, architecture, layering, package organization, config consumption, output, security, dev build tags); secrets and SAST; dependency, license, and SBOM checks; then `test:full` (unit, integration, race).

When it fails, fix the cause. Format → `task format`. Lint → fix the finding. Tests → fix tests or code. Coverage → add tests. See Troubleshooting for the order when several fail at once.

## Where code goes

| Path | What belongs there |
|------|--------------------|
| `cmd/` | Wiring only. `run*` functions ≤30 lines. |
| `internal/` | Business logic. |
| `pkg/` | Public libraries importable by other modules. |
| `.ckeletin/` | Framework. Do not edit it in a derived project; updates overwrite it. |
| `docs/adr/` | Project ADRs, numbered 100+. |
| `test/integration/` | Integration tests. |

`task validate:commands` measures each `run*` function from the `func` line through its closing brace. ≤30 passes, 31–35 warns, >35 fails. A `// ckeletin:allow-custom-command` marker needs a justification on that line. See [ADR-001](.ckeletin/docs/adr/001-ultra-thin-command-pattern.md).

Loops, conditionals, and string manipulation belong in `internal/`. `cmd/` reads config, builds a struct, and calls it.

## Hard rules

- TDD. Write the failing test first. Test and implementation land in one commit.
- `task check` passes before every commit.
- Never hardcode config keys. Define them in `.ckeletin/pkg/config/registry.go`, run `task generate:config:key-constants`, and read `config.Key*`.
- Coverage does not go down. Overall minimum 85%. Use `testify/assert` or `testify/require`. Per-package floors are in `.ckeletin/scripts/check-coverage-project.sh`.
- After adding a dependency, run `task check:license:source`. Allowed and denied licenses are in [docs/licenses.md](docs/licenses.md) ([ADR-011](.ckeletin/docs/adr/011-license-compliance.md)).
- Conventional Commits (`feat`, `fix`, `docs`, `test`, `refactor`, `style`, `perf`, `build`, `ci`, `chore`). Branch prefixes match: `feat/`, `fix/`, `refactor/`, `docs/`.
- Normal merge. Do not squash. Atomic commits are the history.
- Do not pass `--no-verify` unless the user has approved it after seeing why the hook itself is broken.

Conflict order: security, then license, then correctness, then coverage, then style.

## Logging

Zerolog. Console is INFO and above. The file is DEBUG and above, JSON. See [ADR-006](.ckeletin/docs/adr/006-structured-logging-with-zerolog.md).

- Can return the error? `log.Debug()` and return it.
- User input error? Formatted output only.
- Normal flow worth recording? `log.Info()`.
- Recoverable? `log.Warn()`.
- Unrecoverable, and nothing can be returned? `log.Error()`.

Semgrep rule `ckeletin-log-error-and-return` rejects `log.Error()` followed by a returned error.

## JSON Output Mode

`--output json` writes exactly one JSON envelope to stdout and silences the console logger. The audit log file still receives entries. Config key `config.KeyAppOutputFormat` (`app.output_format`, `text` or `json`). Commands that call `ui.RenderSuccess` get the envelope for free. Custom shapes implement `output.JSONResponder`. Types live in `.ckeletin/pkg/output/json.go`.

## Testing

Table-driven tests. Unit tests sit next to the code. Integration tests live in `test/integration/`. Inject dependencies; do not mock by default ([ADR-003](.ckeletin/docs/adr/003-dependency-injection-over-mocking.md)).

Golden files: `task test:golden`, and `task test:golden:update` only after reviewing `git diff test/integration/testdata/`. See [docs/testing.md](docs/testing.md).

## ADRs

Read the ADR before changing the architecture it records. Each one has an `## Enforcement` section ([ADR-014](.ckeletin/docs/adr/014-adr-enforcement-policy.md)).

| Working on | Read |
|------------|------|
| A command | ADR-001, ADR-009 |
| A config option | ADR-002, ADR-005 |
| Tests | ADR-003 |
| Logging | ADR-006 |
| A dependency | ADR-011 |
| A TUI | ADR-007 |
| An ADR | ADR-014 |

Index: [.ckeletin/docs/adr/](.ckeletin/docs/adr/).

## Documentation

User-facing changes go in `CHANGELOG.md` under `[Unreleased]`. New features also update `README.md`. A new architectural decision gets the next ADR number.

## Troubleshooting

| Error | Cause | Solution |
|-------|-------|----------|
| `task: command not found` | Task is not installed | `bash .ckeletin/scripts/install_tools.sh` |
| `go-licenses: package does not have module info` | Tools were built with an older Go | `task setup` |
| Coverage below 85% | Missing tests | `go tool cover -html=coverage.out` |
| License check fails | A copyleft dependency was added | `go get pkg@none`, then `task tidy`, then an allowed replacement |
| `golangci-lint` timeout | The machine is slow | `task lint` |
| Validate commands fails | A `run*` function is over 35 lines, or an exemption marker has no justification | Move the logic to `internal/` |
| Go version too old | The installed Go is older than `.go-version` | Install that version or newer. A newer toolchain passes |
| SBOM scan fails on an old report | `reports/sbom/` is gitignored and can hold a stale CycloneDX file | `task check` regenerates it when it is older than the module |

Local pass, CI fail:

1. Go version — `.go-version` is the source of truth.
2. Stale tools — `task setup`.
3. Missing modules — `task tidy`.
4. A race — `task test:race`.

Cascading failures, in this order:

1. License — replace the module, `task tidy`, `task check:license:source`.
2. Build — `go build ./...`.
3. Lint and format — `task format`, then fix what remains.
4. Tests — `task test`.
5. Coverage — `go tool cover -html=coverage.out`.

Do not fix coverage for code that does not lint.

## Key resources

- [.ckeletin/docs/adr/ARCHITECTURE.md](.ckeletin/docs/adr/ARCHITECTURE.md)
- [Taskfile.yml](Taskfile.yml) and [.ckeletin/Taskfile.yml](.ckeletin/Taskfile.yml)
- [.semgrep.yml](.semgrep.yml)
- [CHANGELOG.md](CHANGELOG.md)
- [README.md](README.md)
