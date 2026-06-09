# Commands

Production commands support the framework-level `--output json` flag for
machine-readable output (one JSON envelope on stdout). Exception:
`completion`, which emits shell scripts for the shell to source.

## `ping` Command

Sample command demonstrating Cobra, Viper, Zerolog, and Bubble Tea integration:

```bash
./myapp ping
./myapp ping --message "Hello!" --color cyan
./myapp ping --ui
```

## `config validate` Command

Validate a configuration file for correctness, security, and completeness:

```bash
./myapp config validate
./myapp config validate --file /path/to/config.yaml
./myapp config validate --output json
```

Checks performed:

- File existence and readability
- File size limits and permissions (security, ADR-004)
- YAML syntax validity
- Per-option semantic validation against the config registry
  (log levels, colors, formats, numeric ranges)
- Configuration value limits
- Unknown configuration keys

Exit code 0 means valid with no warnings; 1 means errors or warnings.

## `docs config` Command

Generate configuration reference documentation from the config registry:

```bash
./myapp docs config                    # Markdown (default)
./myapp docs config --format yaml      # YAML config template
```

## `catalog` Command

Emit the machine-readable command catalog (CKSPEC-AGENT-006), derived from the
live command tree so it cannot drift from the actual command set:

```bash
./myapp catalog
```

## `completion` Command

Generate shell completion scripts:

```bash
./myapp completion bash
./myapp completion zsh    # also: fish, powershell
```

## `--version` Flag

Print version, commit, build date, and tree state (clean/dirty):

```bash
./myapp --version
./myapp --version --output json
```

## `check` Command (Dev Build Only)

Run comprehensive quality checks with beautiful TUI output:

```bash
./myapp check
./myapp check --category quality
./myapp check --fail-fast --verbose
```

Categories: `environment`, `quality`, `architecture`, `security`,
`dependencies`, `tests`.

## `dev` Command Group (Dev Build Only)

```bash
./myapp dev config    # Inspect configuration
./myapp dev doctor    # Check environment health
./myapp dev progress  # Demo progress reporting (spinner, bar, multi-phase)
```

Dev-only commands are included via `task build:dev` and excluded from
production builds. See [ADR-012](../.ckeletin/docs/adr/012-dev-commands-build-tags.md)
for build tag separation. Environment health can also be checked without a dev
build via `task doctor`.

## Adding New Commands

```bash
task generate:command name=mycommand
```

This creates the command file, business-logic skeleton, tests, and config
options following the ultra-thin pattern. See
[CONTRIBUTING.md](../CONTRIBUTING.md) for the full walkthrough.
