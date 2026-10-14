# Framework Changelog

All notable changes to the ckeletin framework layer (`.ckeletin/`) are documented here.

This changelog follows [Keep a Changelog](https://keepachangelog.com/) format.
Only framework changes are tracked here — project-level changes belong in the root `CHANGELOG.md`.

## [Unreleased]

## [0.2.0] - 2026-10-14

This version rolls up every framework change since 0.1.0. `task ckeletin:update` checks out `.ckeletin/` only, so this file is the record a derived CLI receives. The root `CHANGELOG.md` stays in the reference repository.

### Added

- **`task generate:command` writes a working command and lives in this Taskfile.** The generated command maps one string option through a `config.Key*` constant, refuses an empty value, prints it with `ui.RenderSuccess` (so `--output json` carries the message), and registers `internal/<name>` in `.go-arch-lint.yml`. The project Taskfile forwards `ckeletin:generate:command` and passes `name` and `CLI_ARGS`. A project task that still runs the generator itself is warned at update time, with the forwarding to paste in. `name=<name>` is the documented form. `ping` stays the color and UI demo in the reference CLI.
- **Source license check reads `.license-ignore`.** One module path per line, each passed to go-licenses as `--ignore`. A module that is not listed still fails. The reference repository does not ship that file.
- **`task generate:config:schema`** writes a JSON Schema from the config registry.
- **Conformance tooling subject-gates a derived project.** `task conform` in a project whose module is not the reference implementation announces that and exits 0, so an inherited conformance job does not block that project's release. Met claims must resolve to a real check or test. The published report is deterministic.
- **One test-environment file for every `test:*` task.** When `.ckeletin.test-env.sh` exists (or `CKELETIN_TEST_ENV_SETUP` points elsewhere), each test task sources it in the same shell as `go test`.
- **`task check` reports two kinds of drift.** A non-fatal nudge when `.ckeletin/` is behind upstream, and release-infrastructure drift (GoReleaser deprecations, stale pins) for files the update does not replace.
- **Framework changelog gate.** `task check` fails when a change under `.ckeletin/` since the current version is not accompanied by an edit to this file. Edits under `.ckeletin/docs/` are exempt.

### Changed

- **Command `run*` functions over 35 lines fail `task check`.** 30 lines or fewer passes, 31–35 warns. A `ckeletin:allow-custom-command` marker without a justification fails.
- **`.go-version` is the minimum Go toolchain.** A newer local Go passes. An older Go fails. CI installs the pin.
- **govulncheck exit 3 is always a finding.** A network failure is a different exit status whose text names the connectivity problem. The pre-commit scan no longer treats the word `timeout` inside a findings trace as a network skip.
- **The SBOM scan rebuilds a stale CycloneDX file**, and rebuilds the binary first when that is older than `go.mod`, `go.sum`, or `.go-version`.
- **`task setup` installs pinned tools.** `task setup:all` installs the security and audit tools as well.
- **JSON output types live in `.ckeletin/pkg/output`.** The post-update migration removes a stale `internal/ui/json.go` and tells you which imports to change.
- **Tree-walking checks skip nested checkouts** (`.claude/worktrees/` and any nested `go.mod`). The update adds a `.claude` exclude to `.go-arch-lint.yml` when the file has an exclude list and no such entry.

### Fixed

- **JSON mode keeps the audit log.** Only the console writer is silenced. A log file that was requested and cannot be opened fails the command.
- **`config validate` runs per-option validation**, including log level, color, output format, and numeric limits. Log file paths that traverse or resolve through a symlink are rejected.
- **Generators and coverage checks fail closed.** An interrupted constants write cannot leave a truncated `keys_generated.go`, and a coverage or documentation write failure is not reported as success.
- **Import rewriting uses `go/ast`.** Update and init no longer rewrite the upstream module path inside comments and string constants with `sed`.

## [0.1.0] - 2026-04-01

### Added
- Framework versioning via `.ckeletin/VERSION`
- `task ckeletin:version` command to display framework version
- `task ckeletin:update:dry-run` for safe update preview
- Post-update build verification in `task ckeletin:update`
- Framework CHANGELOG (this file)
- AI agent configuration stack (`AGENTS.md`, `CLAUDE.md`, `.claude/rules/`, `.claude/hooks.json`)
- 14 Architecture Decision Records (ADR-000 through ADR-014)
- 37 validation and build scripts
- Comprehensive task system with tiered quality checks

### Infrastructure
- Centralized configuration registry with auto-generated constants
- Structured logging with Zerolog (dual console + file output)
- Bubble Tea UI framework integration
- Test utilities (`testutil` package)
- License compliance checking (dual-tool: go-licenses + lichen)
- Security scanning (semgrep, gitleaks, govulncheck)
