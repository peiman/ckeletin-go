# Development Workflow

All commands are defined in `Taskfile.yml` — used identically in local dev, pre-commit hooks, and CI. See [ADR-000](../.ckeletin/docs/adr/000-task-based-single-source-of-truth.md).

## Essential Commands

```bash
task doctor    # Check your development environment
task check     # Run all quality checks (mandatory before commits)
task format    # Format all Go code
task test      # Run tests with coverage
task build     # Build the binary
```

Run `task --list` for the complete reference.

## Development Tools

Pinned tool versions for reproducible builds:

```bash
task doctor                    # Check environment and tool versions
task check:tools:installed     # Fast existence check
task check:tools:version       # Strict version verification (CI)
task check:tools:updates       # Discover available updates
```

## License Compliance

```bash
task check:license:source      # Fast check during development (~2-5s)
task check:license:binary      # Accurate check before releases (~10-15s)
task generate:license          # Generate all license artifacts
```

See [ADR-011](../.ckeletin/docs/adr/011-license-compliance.md) and [licenses.md](licenses.md) for details.

## Pre-Commit Hooks

`task setup` installs Lefthook hooks that run format, lint, and test on commit.

## Continuous Integration

GitHub Actions runs `task check` on each commit or PR. The CI pipeline includes:

- Multi-OS testing (Ubuntu, macOS, Windows)
- Code quality (golangci-lint, vet, gofmt)
- Architecture validation (go-arch-lint)
- Security scanning (semgrep, gitleaks, govulncheck)
- Test coverage enforcement (≥85%)
- License compliance (dual-tool)

## Creating Releases

Uses [GoReleaser](https://goreleaser.com/) for automated multi-platform releases:

```bash
task check                     # Ensure quality
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0         # CI builds binaries, creates GitHub release
```

**Supported platforms:** Linux (amd64, arm64), macOS (Intel, Apple Silicon), Windows (amd64).

See [ADR-008](../.ckeletin/docs/adr/008-release-automation-with-goreleaser.md) for details.
