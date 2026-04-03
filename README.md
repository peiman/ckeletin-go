<div align="center">

![ckeletin-go](logo/ckeletin-go-banner.png)

**AI-first Go development framework. Ship production CLIs that AI agents build correctly.**

<!-- Row 1: Build Quality & Security -->
[![Build Status](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml/badge.svg)](https://github.com/peiman/ckeletin-go/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/codecov/c/github/peiman/ckeletin-go)](https://codecov.io/gh/peiman/ckeletin-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/peiman/ckeletin-go)](https://goreportcard.com/report/github.com/peiman/ckeletin-go)
[![CodeQL](https://github.com/peiman/ckeletin-go/actions/workflows/github-code-scanning/codeql/badge.svg)](https://github.com/peiman/ckeletin-go/security/code-scanning)

<!-- Row 2: Project Metadata -->
[![Version](https://img.shields.io/github/v/release/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/releases)
[![License](https://img.shields.io/github/license/peiman/ckeletin-go)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/peiman/ckeletin-go.svg)](https://pkg.go.dev/github.com/peiman/ckeletin-go)
[![Go Version](https://img.shields.io/github/go-mod/go-version/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/blob/main/go.mod)

<!-- Row 3: Community & Activity -->
[![GitHub stars](https://img.shields.io/github/stars/peiman/ckeletin-go?style=social)](https://github.com/peiman/ckeletin-go/stargazers)
[![Last Commit](https://img.shields.io/github/last-commit/peiman/ckeletin-go)](https://github.com/peiman/ckeletin-go/commits/main)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

</div>

---

## TL;DR

ckeletin-go is an AI-first Go CLI framework — every architectural rule is machine-checkable, so AI coding agents produce correct, well-structured code from day one. You get production-ready infrastructure with an updatable framework layer, so you can focus on your feature.

- **AI agents build correctly here** — `AGENTS.md`, `CLAUDE.md`, hooks, and automated enforcement mean agents follow your architecture, not fight it
- **Updatable framework** — `.ckeletin/` updates independently via `task ckeletin:update`. Your code is never touched. AI agent infrastructure improves automatically
- **Read the code in 5 minutes** — Ultra-thin commands (~20 lines each). No framework magic to decode
- **Ship with confidence** — ≥85% test coverage, automated architecture validation, GPL/AGPL blocking. Every rule is machine-checkable
- **One command setup** — `task init name=myapp module=...` updates 40+ files. Start coding in 2 minutes

```bash
git clone https://github.com/peiman/ckeletin-go.git && cd ckeletin-go
task setup && task init name=myapp module=github.com/you/myapp
task build && ./myapp ping
```

Or [open in GitHub Codespaces](https://github.com/codespaces/new?hide_repo_select=true&repo=peiman/ckeletin-go) for a pre-configured environment with all tools installed.

---

## Who Is This For?

**You use AI coding agents and need them to produce correct code.**
The layered AI configuration — `AGENTS.md`, `CLAUDE.md`, hooks, enforcement — means agents work within your architecture, not around it. This is what "agent-ready" actually looks like.

**You want to make your own codebase agent-ready.**
Study the pattern: `AGENTS.md` → behavioral rules → automated hooks → machine-checkable enforcement. It works in any project.

**Your boss needs a CLI tool by next sprint. You've never built one.**
Clone, `task init`, and you have production-ready infrastructure in 2 minutes. The ADRs teach you the patterns as you build.

**You're a senior dev who's tired of rebuilding the same scaffolding.**
The updatable framework means you set up once and receive improvements over time. The enforced patterns mean your code stays clean even as the team grows.

---

## Agent-Ready Architecture

Most scaffolds produce code that AI agents can write *in* but not write *well in*. Agents guess at conventions, misconfigure flags, and drift from intended patterns. ckeletin-go solves this with **enforcement by automation** — every architectural rule is machine-checkable, so violations are caught whether the code comes from a human or an AI.

### The AI Configuration Stack

```
AGENTS.md          → Universal project guide (any AI assistant)
CLAUDE.md          → Claude Code-specific behavioral rules
.claude/rules/     → Granular rules loaded automatically
.claude/hooks.json → Auto-installs tools, validates commits
task check         → Single gateway that catches all violations
```

**`AGENTS.md`** gives any AI agent complete project context: architecture, commands, conventions, testing thresholds, and decision trees. It's structured as a specification, not prose — designed for machine consumption.

**`CLAUDE.md`** adds Claude Code-specific rules: mandatory task commands, code placement decision trees, priority cascade (Security → License → Correctness → Coverage → Style).

**Hooks and enforcement** close the loop. SessionStart hooks auto-install tools. Pre-commit hooks validate changes. `task check` runs the same quality gates regardless of who wrote the code.

### Why This Matters

- **Determinism**: `task test` always runs the right flags. Agents don't guess `go test -race -coverprofile=... -count=1 ./...`
- **Architectural memory**: ADRs explain *why* patterns exist, preventing agents from optimizing away guardrails they don't understand
- **Automated enforcement**: 14 ADRs, each with machine-checkable validation. No honor system
- **Framework evolution**: `task ckeletin:update` improves the AI configuration alongside everything else

### Using With AI Agents

**Claude Code**: Reads `CLAUDE.md` and `.claude/rules/` automatically. Hooks fire on session start. No configuration needed.

**Cursor / Copilot / Codex**: Point your agent at `AGENTS.md` for full project context. The task-based workflow and automated enforcement work with any tool.

**The pattern is reusable.** The `AGENTS.md` → rules → hooks → enforcement approach works in any codebase. ckeletin-go is a reference implementation.

---

## What You Get

ckeletin-go is both a **scaffold** (fork, customize, ship) and a **framework** (updatable infrastructure that keeps working for you):

```
myapp/
├── .ckeletin/              # FRAMEWORK — updated via `task ckeletin:update`
│   ├── Taskfile.yml        # Quality checks, build tasks, validation
│   ├── pkg/                # Config registry, logger, testutil packages
│   ├── scripts/            # Enforcement scripts (architecture, patterns, security)
│   └── docs/adr/           # Framework ADRs (000-099)
│
├── cmd/                    # YOUR commands (ultra-thin, ≤30 lines)
├── internal/               # YOUR business logic
├── pkg/                    # YOUR public reusable packages
├── docs/adr/               # YOUR ADRs (100+)
├── Taskfile.yml            # YOUR task aliases + custom tasks
└── .golangci.yml           # YOUR tool configs (customize freely)
```

**The scaffold** gets you started: clone, `task init`, customize `cmd/` and `internal/`, ship.

**The framework** keeps working: enforced architecture, validated patterns, type-safe config, structured logging — all updated independently of your code via `task ckeletin:update`.

**AI agents improve with the framework.** When `.ckeletin/` updates, the AI configuration stack — validation scripts, enforcement rules, task definitions — evolves with it. Your agent gets more effective over time without you changing anything.

---

## Quick Start

1. **Clone and set up tools:**
   ```bash
   git clone https://github.com/peiman/ckeletin-go.git
   cd ckeletin-go
   task setup
   ```

2. **Initialize with your project details:**
   ```bash
   task init name=myapp module=github.com/you/myapp
   ```
   This updates module path, imports (40+ files), binary name, and config — automatically.

3. **Build and run:**
   ```bash
   task build
   ./myapp ping
   ```

**Alternative:** [Open in GitHub Codespaces](https://github.com/codespaces/new?hide_repo_select=true&repo=peiman/ckeletin-go) — all tools pre-installed, ready to go.

---

## Architecture

ckeletin-go follows a principled architecture with automated enforcement:

- **Layered architecture** — 4-layer pattern (Entry → Command → Business Logic → Infrastructure) with validation ([ADR-009](.ckeletin/docs/adr/009-layered-architecture-pattern.md))
- **Ultra-thin commands** — ~20-30 lines, delegate to business logic ([ADR-001](.ckeletin/docs/adr/001-ultra-thin-command-pattern.md))
- **Centralized configuration** — Type-safe registry with auto-generated constants ([ADR-002](.ckeletin/docs/adr/002-centralized-configuration-registry.md))
- **Dependency injection** — Over mocking, for testability ([ADR-003](.ckeletin/docs/adr/003-dependency-injection-over-mocking.md))
- **Dual-tool license compliance** — Source + binary analysis ([ADR-011](.ckeletin/docs/adr/011-license-compliance.md))
- **Dev-only commands** — Via build tags ([ADR-012](.ckeletin/docs/adr/012-dev-commands-build-tags.md))
- **Enforcement by automation** — Every ADR has machine-checkable validation, catching violations from humans and AI agents alike ([ADR-014](.ckeletin/docs/adr/014-adr-enforcement-policy.md))

All architectural decisions are documented in **[Architecture Decision Records](docs/adr/)**.

---

## Key Features

- **Modular Command Structure**: Add, remove, or update commands without breaking the rest
- **Layered Architecture**: Enforced 4-layer pattern with automated validation to prevent drift
- **Structured Logging**: Zerolog dual output (console + file) for debugging and production
- **Bubble Tea UI**: Optional interactive terminal UIs
- **Single-Source Configuration**: Defaults in config files, overrides via env vars and flags
- **Enterprise License Compliance**: Dual-tool checking with automatic GPL/AGPL blocking ([ADR-011](.ckeletin/docs/adr/011-license-compliance.md))
- **Task Automation**: One Taskfile for all build, test, and lint commands
- **High Test Coverage**: ≥85% enforced by CI. Hundreds of real tests
- **Beautiful Check Output**: `pkg/checkmate` — thread-safe, TTY-aware terminal output library

---

## Getting Started

### Prerequisites

- **Go**: Version specified in `.go-version` (currently 1.26.1)
- **Task**: Install from [taskfile.dev](https://taskfile.dev/#/installation)
- **Git**: For version control

### Build from Source

```bash
git clone https://github.com/peiman/ckeletin-go.git
cd ckeletin-go
task setup && task build
```

### Using the Scaffold

```bash
task init name=myapp module=github.com/myuser/myapp
```

This single command updates module path, imports (40+ files), binary name, and config. Then:

```bash
task check    # Run quality checks
task build    # Build your binary
./myapp --version
```

---

## Configuration

Viper-based configuration with a centralized registry ([ADR-002](.ckeletin/docs/adr/002-centralized-configuration-registry.md)). Precedence (highest to lowest): command-line flags → environment variables → config file → defaults.

```yaml
# ~/.config/myapp/config.yaml
app:
  log_level: "debug"
  ping:
    output_message: "Hello World!"
```

```bash
export MYAPP_APP_LOG_LEVEL="debug"           # Environment override
./myapp ping --message "Hi there!" --color yellow  # Flag override
```

See [Configuration Reference](docs/configuration.md) for all options, auto-generated docs, and config templates.

---

## Commands

Built-in commands: `ping` (demo), `config validate`, `check` (quality gates), `dev` (dev-only tools), and `doctor` (environment health).

```bash
./myapp ping --message "Hello!" --color cyan
./myapp config validate
task doctor
```

Add new commands with `task generate:command name=mycommand`. See [Command Reference](docs/commands.md) for details.

---

## Development Workflow

All commands defined in `Taskfile.yml` — used identically in local dev, pre-commit hooks, and CI:

```bash
task check     # Run all quality checks (mandatory before commits)
task test      # Run tests with coverage
task build     # Build the binary
task format    # Format all Go code
task doctor    # Check environment health
```

See [Development Workflow](docs/development-workflow.md) for the full reference including tools, license compliance, CI, and releases.

---

## Framework Updates

Your project code and the framework layer are independent:

```bash
task ckeletin:update:dry-run             # Preview changes (safe)
task ckeletin:update:check-compatibility # Test build compatibility (safe)
task ckeletin:update                     # Apply update (creates a commit)
```

| Framework (`.ckeletin/`) | Project (yours) |
|--------------------------|-----------------|
| `Taskfile.yml` — Quality checks, build tasks | `Taskfile.yml` — Your aliases + custom tasks |
| `pkg/config/` — Configuration registry | `cmd/*.go` — Your commands |
| `pkg/logger/` — Zerolog dual-output logging | `internal/` — Your business logic |
| `scripts/` — Validation and check scripts | `pkg/` — Your public packages |
| `docs/adr/000-099` — Framework decisions | `docs/adr/100+` — Your decisions |

See [Framework Update Guide](docs/framework-updates.md) for the safe update workflow, breaking change handling, and how AI agents can manage updates.

---

## Contributing

1. Fork & create a feature branch
2. Make changes, run `task check`
3. Commit with [Conventional Commits](https://www.conventionalcommits.org/) format
4. Open a PR against `main`

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

---

## License

MIT License. See [LICENSE](LICENSE).
