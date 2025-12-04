# Configuration Files

Configuration files for development tools live in the **project root**, not in this directory.

## Why Configs Stay in Project Root

1. **User Customization**: Each project has different requirements for linting, security scanning, etc.
2. **Tool Compatibility**: Most tools expect configs in project root and don't support "extends" patterns
3. **Clear Ownership**: Configs in project root = user-owned, updatable independently from framework

## Standard Configuration Files

| File | Tool | Purpose |
|------|------|---------|
| `.golangci.yml` | golangci-lint | Code quality linting |
| `.lefthook.yml` | lefthook | Git hooks (pre-commit, pre-push) |
| `.go-arch-lint.yml` | go-arch-lint | Architecture validation |
| `.lichen.yaml` | lichen | License compliance (binary) |
| `.gitleaks.toml` | gitleaks | Secret detection |
| `.semgrep.yml` | semgrep | Static analysis security |
| `.goreleaser.yml` | goreleaser | Release automation |

## Customizing Configs

When you clone the template, you get sensible defaults. Customize as needed:

```yaml
# .golangci.yml - Add/remove linters
linters:
  enable:
    - your-preferred-linter
```

## Framework Updates Don't Touch Configs

When you run `task ckeletin:update`, your configuration files in the project root are preserved. Only `.ckeletin/` contents are updated.
