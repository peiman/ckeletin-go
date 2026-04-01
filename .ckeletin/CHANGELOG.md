# Framework Changelog

All notable changes to the ckeletin framework layer (`.ckeletin/`) are documented here.

This changelog follows [Keep a Changelog](https://keepachangelog.com/) format.
Only framework changes are tracked here — project-level changes belong in the root `CHANGELOG.md`.

## [Unreleased]

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
