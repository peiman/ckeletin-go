# ckeletin Framework

This directory contains the ckeletin framework - reusable infrastructure for building Go CLI applications.

## What's in This Directory

```
.ckeletin/
├── Taskfile.yml           # Framework tasks (quality checks, builds, etc.)
├── pkg/                   # Framework Go packages
│   ├── config/            # Configuration registry and validation
│   ├── logger/            # Zerolog dual-output logging
│   ├── ui/                # Bubble Tea UI utilities
│   └── testutil/          # Test helpers
├── scripts/               # Build, validation, and check scripts
├── configs/               # Documentation for config file strategy
└── docs/adr/              # Framework ADRs (000-099)
```

## Framework vs Project Code

| Location | Owner | Updated By |
|----------|-------|------------|
| `.ckeletin/` | Framework | `task ckeletin:update` |
| Everything else | You | Your changes |

## Updating the Framework

When ckeletin releases updates, pull them in:

```bash
# Simple way
task ckeletin:update

# Manual way (if you need more control)
git subtree pull --prefix=.ckeletin https://github.com/peiman/ckeletin-go.git main --squash
```

Your project files (cmd/, internal/, configs in root) are never touched by framework updates.

## Using Framework Packages

Import from `.ckeletin/pkg/`:

```go
import (
    "github.com/youruser/yourproject/.ckeletin/pkg/config"
    "github.com/youruser/yourproject/.ckeletin/pkg/logger"
    "github.com/youruser/yourproject/.ckeletin/pkg/ui"
)
```

## Using Framework Tasks

Framework tasks are namespaced with `ckeletin:`:

```bash
task ckeletin:check      # Run all quality checks
task ckeletin:test       # Run tests
task ckeletin:build      # Build binary
task ckeletin:lint       # Run linters
```

Convenience aliases are defined in your root Taskfile.yml:

```bash
task check               # Same as task ckeletin:check
task test                # Same as task ckeletin:test
```

## Don't Modify This Directory

Changes to `.ckeletin/` will be overwritten on framework updates. Instead:

- **Customize configs**: Edit files in project root (`.golangci.yml`, etc.)
- **Add commands**: Create in `cmd/` (not `.ckeletin/cmd/`)
- **Add business logic**: Create in `internal/` (not `.ckeletin/pkg/`)
- **Project ADRs**: Add to `docs/adr/` starting at 100+
