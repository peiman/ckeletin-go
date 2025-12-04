# Scripts Directory

This directory contains automation scripts for the ckeletin-go project.

## install_tools.sh

**Purpose:** Automatically install all required development tools for the project.

**Usage:**
```bash
bash scripts/install_tools.sh
```

**When it runs:**
- Automatically via `.claude/hooks.json` SessionStart hook
- When starting a new Claude Code session
- Can be run manually if needed

**What it does:**
1. Adds `/root/go/bin` to PATH
2. Checks if each tool is already installed
3. Installs missing tools using `go install`:
   - `goimports` - Import formatting
   - `govulncheck` - Vulnerability checking
   - `gotestsum` - Test running with better output
   - `golangci-lint` - Linting
   - `go-mod-outdated` - Dependency update checking
   - `lefthook` (optional) - Git hooks manager

**Features:**
- **Idempotent:** Safe to run multiple times, only installs what's missing
- **Fast:** Skips already-installed tools
- **Silent downloads:** Suppresses verbose download messages
- **Informative:** Shows clear success/skip messages for each tool

## Other Scripts

### format-go.sh
Format and verify Go code formatting.

**Usage:**
```bash
# Format all Go files
./scripts/format-go.sh fix

# Check formatting without modifying files
./scripts/format-go.sh check
```

### check-defaults.sh
Verify that all configuration defaults are defined in the centralized registry.

**Usage:**
```bash
./scripts/check-defaults.sh
```

### validate-command-patterns.sh
Ensure command files follow the ultra-thin command pattern (ADR-001).

**Usage:**
```bash
./scripts/validate-command-patterns.sh
```

## Adding New Scripts

When adding new scripts to this directory:
1. Make them executable: `chmod +x scripts/your-script.sh`
2. Add a shebang line: `#!/bin/bash`
3. Include usage comments at the top
4. Document them in this README
5. Consider adding them to Taskfile.yml for easier access
