# Claude Code Session Setup

Tools auto-install via SessionStart hook (`.claude/hooks.json` → `.ckeletin/scripts/install_tools.sh`).
Installs: task, goimports, golangci-lint, gotestsum, govulncheck.

If tools fail to install: `bash .ckeletin/scripts/install_tools.sh`

**After upgrading Go versions**, rebuild dev tools: `task setup`
- Dev tools are compiled Go binaries — may be incompatible with newer Go
- Symptom: `go-licenses` failing with "package does not have module info"
- Detection: `task doctor`

**First session verification:**
```
task --list    # Should show all tasks
go build ./... # Should compile cleanly
task test      # Should pass with ≥85% coverage
```

If any fail, run `task setup` to rebuild tools, then retry.

This project uses `includeCoAuthoredBy: false` — commits do not include Claude Code attribution.
