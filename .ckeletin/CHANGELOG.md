# Framework Changelog

All notable changes to the ckeletin framework layer (`.ckeletin/`) are documented here.

This changelog follows [Keep a Changelog](https://keepachangelog.com/) format.
Only framework changes are tracked here — project-level changes belong in the root `CHANGELOG.md`.

## [Unreleased]

## [0.2.4] - 2026-10-17

### Fixed

- **A normal config file no longer logs at Info.** Mode 0644 matches the permissive-permission check, and that note is now Debug. Group-writable still warns. World-writable is still an error.
- **golangci-lint is pinned to v2.15.0.** v2.12.2 cannot read Go 1.27 export data. A project overrides any tool pin with `CKELETIN_TOOL_<NAME>` in its own Taskfile, or on the task command line. Unset, the framework default is used. An update still replaces `.ckeletin/Taskfile.yml`, so the override has to live outside it.
- **A framework update no longer rewrites the rename tools' fixtures.** A Go file marked `ckeletin:keep-upstream-names` keeps the upstream module path. The ordinary test tasks list `.ckeletin/scripts` explicitly, because `./...` skips a directory whose name starts with `.`.
- **Generated config docs use the project default.** `config.EffectiveRegistry` applies `internal/projectdefaults` to a copy of the registry. `docs config` and the config schema publish that value. The registry itself keeps the framework default. The generated precedence is flag, environment, config file, project default, then framework default.
- **An option can omit its command-line flag.** `ConfigOption.NoFlag` keeps the option in the registry, the config file, the environment, and the generated docs. Flag registration skips it. Generated docs say it is config file or environment only. Flags written out by hand, including the reference root flags, are separate from this loop. An update brings the field. A project that already registers flags has to honor `NoFlag` in its own `cmd/flags.go`, because an update does not replace that file.

## [0.2.3] - 2026-10-17

### Added

- **`task ckeletin:update` stops before it replaces a local framework edit.** Uncommitted changes under `.ckeletin/` always stop the command, including when `discard_local=1` is set. A committed difference from the upstream tree at the project's current `.ckeletin/VERSION` stops the command unless `discard_local=1`. Names the update itself rewrites are not a local difference. The dry run shows the difference and changes nothing. The compatibility check refuses to apply a candidate when a local difference exists, and still restores `.ckeletin/` when the tree is clean. A stop prints `CKELETIN_UPDATE_RESULT` with `status` `stopped`, `reason` (`uncommitted`, `local_difference`, or `unknown_version`), `to` null, and `discarded` false.
- **Project defaults live outside the framework tree.** `config.ApplyProjectDefaultsTOML` replaces the framework default for a registry key. An unknown key fails when configuration is loaded. A config file, environment variable, or flag still wins.

### Changed

- **`CKELETIN_UPDATE_RESULT` includes `discarded`.** It is true only when the command discarded a committed local difference. `up_to_date`, `build_failed`, and `updated` keep their existing fields.

## [0.2.2] - 2026-10-16

### Added

- **An adoption ratchet for the 35-line `run*` limit.** A project that updates from 0.1.0 with long commands already in place runs `task validate:commands:baseline` once. It writes `.ckeletin-command-baseline` in the project root, outside `.ckeletin/`, one `<command> <run func> <lines>` per line. A recorded function at its recorded length is a warning, and the line is shown even on a pass. Growing past the record fails. An entry above the current length fails until it is lowered. An entry for a function back within the limit, or gone, fails until it is removed. So the list only shrinks. Unrecorded long functions fail as before.

## [0.2.1] - 2026-10-16

### Added

- **`task check` notices a newer framework version.** The notice compares `.ckeletin/VERSION` with upstream and prints when upstream is a newer semver. It names `task ckeletin:update`. It does not fail the build, does not run in CI, can be turned off with `CKELETIN_SKIP_UPDATE_CHECK=1`, and stays quiet offline or without the `ckeletin-upstream` remote. A commit that does not change the version is not an update. A local version ahead of upstream is not one either. `task ckeletin:check-update` prints the same comparison on demand. `format=json` prints `current`, `latest`, and `update_available`. `task ckeletin:update` prints `CKELETIN_UPDATE_RESULT` with the version before and after.

### Fixed

- **Layering check installs go-arch-lint v1.18.0** instead of `@latest`. That pin does not fix Windows: through v1.19.0 the linter reports false violations there, so the worktree layering fixture skips on Windows. The unwritable-directory audit-log fixture also skips on Windows, where mode `0500` does not remove the owner's write access. The fail-closed case where a path component is a file still runs there.
- **`task ckeletin:update` re-applies the derived project's names.** After checkout it rewrites the module path and binary name in framework markdown and YAML, and the binary name plus env-var prefix in Go files including tests, the same way `task init` did. A same-version update no longer leaves a `.ckeletin/` diff that fails the changelog gate. Shell scripts are unchanged.

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
