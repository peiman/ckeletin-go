# Framework Updates

How the ckeletin framework update mechanism works, and how to use it safely.

## Overview

ckeletin-go separates your project code from the framework layer. The `.ckeletin/` directory contains infrastructure — Taskfile, config registry, logger, validation scripts, ADRs — that updates independently without touching your code in `cmd/`, `internal/`, `pkg/`, or `docs/adr/`.

This means you get framework improvements (better validation, new scripts, updated ADRs, improved AI agent configuration) without merge conflicts in your business logic.

## Quick Reference

| Command | Purpose | Modifies files? |
|---------|---------|-----------------|
| `task ckeletin:update:dry-run` | Preview what would change | No (read-only) |
| `task ckeletin:update:check-compatibility` | Test if the update builds with your code | No (restores original) |
| `task ckeletin:update` | Apply the update | Yes (creates a commit) |
| `task ckeletin:version` | Show current framework version | No |

## The Safe Update Workflow

Always follow this sequence:

### Step 1: Preview

```bash
task ckeletin:update:dry-run
```

Shows a `git diff --stat` of what would change in `.ckeletin/` without modifying any files. If the output says "Framework is already up-to-date", there's nothing to do.

### Step 2: Check Compatibility

```bash
task ckeletin:update:check-compatibility
```

This safely tests whether the update will build with your code:

1. Stashes your current `.ckeletin/` state
2. Applies the update temporarily
3. Rewrites imports (AST-based)
4. Runs `go build ./...`
5. Restores your original state regardless of the outcome

Reports either "Compatible — safe to update" or "Incompatible" with the specific build errors.

### Step 3: Apply the Update

```bash
task ckeletin:update
```

Applies the framework update and creates a single, revertable commit. See the next section for exactly what this does.

## What `task ckeletin:update` Does

The update is a 10-step process:

1. **Validates environment** — Confirms you're not running on the upstream ckeletin-go repo itself (run `task init` first)
2. **Sets up remote** — Adds or reuses the `ckeletin-upstream` git remote pointing to `https://github.com/peiman/ckeletin-go.git`
3. **Fetches upstream** — `git fetch ckeletin-upstream main`
4. **Checks out framework** — `git checkout ckeletin-upstream/main -- .ckeletin/` — only the `.ckeletin/` directory, nothing else
5. **Rewrites imports** — Runs the AST-based import rewriter to replace the upstream module path with your project's module path (from `go.mod`)
6. **Regenerates constants** — `task ckeletin:generate:config:key-constants` to keep config constants in sync
7. **Formats code** — `task ckeletin:format`
8. **Commits** — `git add .ckeletin && git commit` with message `"chore: update ckeletin framework"`
9. **Verifies build** — `go build ./...` to catch any breaking changes
10. **Reports result** — Success with instructions to review (`git diff HEAD~1`), or "already up-to-date" if nothing changed

If the build fails at step 9, the commit still exists so you can inspect it. See [Handling Breaking Changes](#handling-breaking-changes) for recovery.

## What Gets Updated vs. What Doesn't

| Updated (framework-owned) | NOT updated (project-owned) |
|----------------------------|-----------------------------|
| `.ckeletin/Taskfile.yml` | `Taskfile.yml` (your aliases + custom tasks) |
| `.ckeletin/pkg/config/` | `cmd/*.go` (your commands) |
| `.ckeletin/pkg/logger/` | `internal/` (your business logic) |
| `.ckeletin/pkg/testutil/` | `pkg/` (your public packages) |
| `.ckeletin/scripts/` | `docs/adr/` (your ADRs, 100+) |
| `.ckeletin/docs/adr/` (framework ADRs, 000-099) | `.golangci.yml`, `.goreleaser.yml` |
| `.ckeletin/VERSION` | `go.mod`, `go.sum` |
| `.ckeletin/CHANGELOG.md` | `AGENTS.md`, `CLAUDE.md` |

The boundary is strict: `git checkout ckeletin-upstream/main -- .ckeletin/` ensures only files under `.ckeletin/` are touched. Your code, your configs, and your documentation are never modified.

## How Import Rewriting Works

When you fork ckeletin-go and run `task init`, your module path changes (e.g., from `github.com/peiman/ckeletin-go` to `github.com/you/myapp`). Framework code in `.ckeletin/pkg/` imports need to reflect your module path, not the upstream one.

The update uses an AST-based rewriter (`.ckeletin/scripts/rewrite-imports/main.go`) that:

- Parses each `.go` file using Go's `go/ast` package
- Finds import statements matching the upstream module path
- Replaces them with your project's module path
- Only modifies actual import paths — never comments, strings, or partial matches
- Sorts imports after rewriting

This is fundamentally safer than `sed`-based string replacement because it understands Go syntax. A string like `"github.com/peiman/ckeletin-go"` in a comment or test fixture won't be incorrectly rewritten.

## Handling Breaking Changes

### Detection

The update runs `go build ./...` after committing (step 9). If the build fails:

- The commit still exists so you can inspect what changed
- The error message directs you to `.ckeletin/CHANGELOG.md`
- The task exits with a non-zero exit code

### Recovery

```bash
git revert HEAD
```

This cleanly undoes the framework update commit. Your project returns to its previous state.

### Prevention

Always run compatibility checking first:

```bash
task ckeletin:update:check-compatibility
```

This tests the build without committing anything. If it reports incompatibility, review the build errors and `.ckeletin/CHANGELOG.md` before proceeding.

## Framework Versioning

| Item | Location | Purpose |
|------|----------|---------|
| Version file | `.ckeletin/VERSION` | Current framework version (semver) |
| Changelog | `.ckeletin/CHANGELOG.md` | What changed, in Keep a Changelog format |
| Version command | `task ckeletin:version` | Display the current framework version |

The framework version is independent of your project version. It tracks changes to the infrastructure layer only.

## AI Agents and Framework Updates

The update workflow is designed to be AI-agent-friendly:

- **`task ckeletin:update:dry-run`** is safe to run at any time (read-only)
- **`task ckeletin:update:check-compatibility`** is safe (restores original state)
- **`task ckeletin:update`** creates a single revertable commit
- The entire workflow is deterministic and non-interactive — no prompts, no user input required

An AI agent following the safe update workflow (dry-run → check-compatibility → update) can keep the framework current without human intervention. The build verification step catches breaking changes automatically.

When the framework updates, the AI agent configuration stack (`AGENTS.md` patterns, validation scripts, enforcement rules) improves with it — making the agent more effective over time.

## FAQ

### "Cannot update: this appears to be the upstream repo itself"

You're running `task ckeletin:update` in the original ckeletin-go repository, not a fork. Run `task init name=myapp module=github.com/you/myapp` first to initialize your project.

### "Framework is already up-to-date"

No changes between your `.ckeletin/` and upstream. Nothing to do.

### Build fails after update

1. Read the build error output carefully
2. Check `.ckeletin/CHANGELOG.md` for breaking changes in the latest version
3. Fix compilation errors in your code to match the new framework API
4. If the changes are too disruptive, rollback: `git revert HEAD`

### Import rewriting missed something

The AST rewriter only processes `.go` files under `.ckeletin/`. If you have Go files elsewhere that import `.ckeletin/` packages directly (unusual but possible), they won't be automatically rewritten. Fix manually:

```go
// Before (upstream path):
import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"

// After (your project path):
import "github.com/you/myapp/.ckeletin/pkg/config"
```

### Will I get merge conflicts?

No. The update uses `git checkout` (not `git merge`), so it overwrites `.ckeletin/` entirely with the upstream version. There are no merge conflicts in the traditional sense.

If you had local modifications to files inside `.ckeletin/`, they will be overwritten. This is by design — `.ckeletin/` is upstream-owned. Put your customizations in project-owned files (`Taskfile.yml`, `.golangci.yml`, `cmd/`, `internal/`, etc.).

### Can I pin to a specific framework version?

The update always pulls from `ckeletin-upstream/main` (the latest). To stay on a specific version:

1. Don't run `task ckeletin:update`
2. Check `.ckeletin/VERSION` to see what you're currently on
3. When ready to update, use the safe workflow to review changes before applying
