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
| `task ckeletin:update` | Apply the update (latest) | Yes (creates a commit) |
| `task ckeletin:update version=v0.2.0` | Apply a specific version | Yes (creates a commit) |
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

The update is a 14-step process:

1. **Validates environment** — Confirms you're not running on the upstream ckeletin-go repo itself (run `task init` first)
2. **Sets up remote** — Adds or reuses the `ckeletin-upstream` git remote pointing to `https://github.com/peiman/ckeletin-go.git`
3. **Fetches upstream** — `git fetch ckeletin-upstream main` (or a specific tag if `version=vX.Y.Z` is set)
4. **Captures old version** — Records current `.ckeletin/VERSION` for changelog diff
5. **Checks out framework** — `git checkout ckeletin-upstream/main -- .ckeletin/` — only the `.ckeletin/` directory, nothing else
6. **Rewrites imports** — Runs the AST-based import rewriter to replace the upstream module path with your project's module path (from `go.mod`)
7. **Replaces binary name** — Updates string literals in `.ckeletin/` Go files (e.g., `"./logs/ckeletin-go.log"` → `"./logs/myapp.log"`) and env var prefixes
8. **Tidies modules** — `go mod tidy` to update `go.mod`/`go.sum` if dependencies changed
9. **Regenerates constants** — `task ckeletin:generate:config:key-constants` to keep config constants in sync
10. **Formats code** — `task ckeletin:format`
11. **Runs migrations** — `migrate-post-update.sh` for structural changes (e.g., removing stale config entries)
12. **Commits** — `git add .ckeletin go.mod go.sum && git commit` with message including the new version
13. **Verifies build and tests** — `go build ./...` to catch compilation errors, then `go test ./.ckeletin/...` to catch test utility breakage
14. **Reports result** — Shows changelog diff (old → new version), success message, or "already up-to-date"

If the build fails at step 13, the commit still exists so you can inspect it. See [Handling Breaking Changes](#handling-breaking-changes) for recovery.

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

Yes. Pass a version tag to target a specific release instead of the latest:

```bash
# Preview a specific version
task ckeletin:update:dry-run version=v0.2.0

# Check compatibility with a specific version
task ckeletin:update:check-compatibility version=v0.2.0

# Update to a specific version
task ckeletin:update version=v0.2.0
```

Without the `version` parameter, all commands default to the latest `main` branch.

### Why does the update commit use `--no-verify`?

The framework update commit bypasses pre-commit hooks (`--no-verify`) because:

- The commit is machine-generated with deterministic content
- The update task runs its own verification (build + tests) after committing
- User pre-commit hooks (commit message format, secret scanning, etc.) may reject the auto-generated commit message or flag framework files the user doesn't control

If you need pre-commit hooks to run on framework updates, you can amend the commit afterward:

```bash
task ckeletin:update
git commit --amend --no-edit  # Re-runs hooks on the same content
```

## CI Integration

You can automate framework update checking in CI. Here's a GitHub Actions example that runs weekly and creates an issue when updates are available:

```yaml
# .github/workflows/framework-update-check.yml
name: Check Framework Updates

on:
  schedule:
    - cron: '0 9 * * 1'  # Every Monday at 9am UTC
  workflow_dispatch:       # Allow manual trigger

jobs:
  check-update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install Task
        uses: arduino/setup-task@v2
        with:
          version: 3.x

      - name: Check for framework updates
        id: check
        run: |
          OUTPUT=$(task ckeletin:update:dry-run 2>&1)
          echo "$OUTPUT"
          if echo "$OUTPUT" | grep -q "already up-to-date"; then
            echo "has_update=false" >> "$GITHUB_OUTPUT"
          else
            echo "has_update=true" >> "$GITHUB_OUTPUT"
            echo "diff<<EOF" >> "$GITHUB_OUTPUT"
            echo "$OUTPUT" >> "$GITHUB_OUTPUT"
            echo "EOF" >> "$GITHUB_OUTPUT"
          fi

      - name: Check compatibility
        if: steps.check.outputs.has_update == 'true'
        id: compat
        run: |
          if task ckeletin:update:check-compatibility 2>&1; then
            echo "compatible=true" >> "$GITHUB_OUTPUT"
          else
            echo "compatible=false" >> "$GITHUB_OUTPUT"
          fi

      - name: Create issue for available update
        if: steps.check.outputs.has_update == 'true'
        uses: actions/github-script@v7
        with:
          script: |
            const compatible = '${{ steps.compat.outputs.compatible }}' === 'true';
            const label = compatible ? '🟢 compatible' : '🔴 incompatible';
            const body = compatible
              ? `A compatible framework update is available.\n\n\`\`\`\n${{ steps.check.outputs.diff }}\n\`\`\`\n\nRun \`task ckeletin:update\` to apply.`
              : `A framework update is available but has compatibility issues. Review the build errors before updating.\n\n\`\`\`\n${{ steps.check.outputs.diff }}\n\`\`\``;
            await github.rest.issues.create({
              owner: context.repo.owner,
              repo: context.repo.repo,
              title: `Framework update available (${label})`,
              body: body,
              labels: ['framework-update']
            });
```

This workflow:
1. Runs `task ckeletin:update:dry-run` to detect available updates
2. If updates exist, runs `task ckeletin:update:check-compatibility` to test build
3. Creates a GitHub issue with the diff and compatibility status
