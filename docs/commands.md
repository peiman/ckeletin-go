# Commands

## `ping` Command

Sample command demonstrating Cobra, Viper, Zerolog, and Bubble Tea integration:

```bash
./myapp ping
./myapp ping --message "Hello!" --color cyan
./myapp ping --ui
```

## `doctor` Command

Check your development environment:
```bash
task doctor
```

## `config validate` Command

Validate configuration files:
```bash
./myapp config validate
./myapp config validate --file /path/to/config.yaml
```

## `check` Command (Dev Build Only)

Run comprehensive quality checks with beautiful TUI output:
```bash
./myapp check
./myapp check --category quality
./myapp check --fail-fast --verbose
```

Categories: Development Environment, Code Quality, Architecture Validation, Dependencies, Tests.

## `dev` Command Group (Dev Build Only)

```bash
./myapp dev config    # Inspect configuration
./myapp dev doctor    # Check environment health
./myapp dev progress  # Show development progress
```

See [ADR-012](../.ckeletin/docs/adr/012-dev-commands-build-tags.md) for build tag separation.

## Adding New Commands

```bash
task generate:command name=mycommand
```

This creates the command file, metadata, and config options following the ultra-thin pattern. See [CONTRIBUTING.md](../CONTRIBUTING.md) for the full walkthrough.
