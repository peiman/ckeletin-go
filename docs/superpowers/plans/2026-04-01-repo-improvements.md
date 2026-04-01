# Repository Improvements Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Address all 28 recommendations from the repo improvement analysis — fixing bugs, closing security gaps, improving the framework update mechanism, increasing test coverage, and building visibility.

**Architecture:** Changes span CI config, Taskfile definitions, Go source, shell scripts, and documentation. Organized into 4 phases: Immediate (bug fixes), Short-Term (framework & testing), Medium-Term (security, DX, marketing), Long-Term (AST rewriting, advanced testing, extensions).

**Tech Stack:** Go, Bash, GitHub Actions YAML, Taskfile, GoReleaser, cosign, grype, go/ast, asciinema.

**Spec:** `docs/superpowers/specs/2026-04-01-repo-improvement-analysis.md`

---

## Phase 1: Immediate Fixes (Recommendations 1-4)

### Task 1: Fix Go Version Mismatch

**Files:**
- Modify: `go.mod:3`

- [ ] **Step 1: Check current versions**

```bash
cat .go-version
head -3 go.mod
```

Expected: `.go-version` says `1.26.1`, `go.mod` says `go 1.24.2`.

- [ ] **Step 2: Update go.mod to match .go-version**

Use the Edit tool to change the `go` directive in `go.mod` to match `.go-version`. The `.go-version` file is the source of truth (it's what CI and local tooling use).

```bash
go mod tidy
```

- [ ] **Step 3: Verify alignment**

```bash
cat .go-version
head -3 go.mod
go version
```

All three should reference 1.26.x.

- [ ] **Step 4: Run tests**

```bash
task test
```

- [ ] **Step 5: Commit**

```bash
git add go.mod go.sum
git commit -m "fix: align go.mod version with .go-version (1.26.1)

- go.mod said 1.24.2 while .go-version said 1.26.1
- .go-version is the source of truth for CI and local tooling"
```

---

### Task 2: Update GitHub Topics and Description

**Files:**
- None (GitHub API calls only)

- [ ] **Step 1: Update repo topics**

```bash
gh repo edit --remove-topic boilerplate --remove-topic skaffold --remove-topic skeleton
gh repo edit --add-topic ai-agent --add-topic claude-code --add-topic scaffold --add-topic framework --add-topic cobra --add-topic viper --add-topic cli-template
```

Keep existing: `cli`, `go`, `golang`.

- [ ] **Step 2: Update repo description**

```bash
gh repo edit --description "A production-ready Go CLI scaffold powered by an updatable framework layer — built for humans and AI agents alike"
```

- [ ] **Step 3: Verify**

```bash
gh repo view --json description,repositoryTopics --jq '{description, topics: .repositoryTopics}'
```

---

### Task 3: Enable Semgrep on PRs

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Read the CI workflow to find the semgrep skip condition**

Search for the PR-skipping condition in ci.yml. It will be something like:
```yaml
if: github.event_name != 'pull_request'
```
in or near the SAST/semgrep step.

- [ ] **Step 2: Remove or modify the condition**

Remove the `if:` condition that skips semgrep on PRs, or change it so semgrep runs on all events. Read the surrounding context carefully — the condition may apply to a broader block. Only remove the semgrep-specific skip, not other conditional logic.

- [ ] **Step 3: Verify the workflow syntax**

```bash
# If actionlint is available:
actionlint .github/workflows/ci.yml 2>/dev/null || echo "actionlint not installed, skip"
```

- [ ] **Step 4: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: enable semgrep SAST scanning on pull requests

- Previously skipped on PRs, meaning custom SAST rules
  (including ADR enforcement) weren't checked until main
- Adds ~7s to PR checks"
```

---

## Phase 2: Short-Term — Framework & Testing (Recommendations 5-11)

### Task 4: Add Framework Versioning

**Files:**
- Create: `.ckeletin/VERSION`
- Modify: `.ckeletin/Taskfile.yml` (add `version` task)
- Modify: `Taskfile.yml` (add alias)

- [ ] **Step 1: Create VERSION file**

Create `.ckeletin/VERSION` with the initial version:

```
0.1.0
```

Use semver. This is the first versioned release of the framework layer.

- [ ] **Step 2: Add `ckeletin:version` task to framework Taskfile**

Read `.ckeletin/Taskfile.yml` and add a new task in the appropriate section (near the top, with other info tasks like `doctor`):

```yaml
  version:
    desc: Show framework version
    cmds:
      - |
        echo "ckeletin framework v$(cat .ckeletin/VERSION)"
```

- [ ] **Step 3: Add alias in root Taskfile**

Read `Taskfile.yml` and add an alias:

```yaml
  ckeletin-version:
    desc: Show ckeletin framework version
    cmds: [task: ckeletin:version]
```

- [ ] **Step 4: Test the task**

```bash
task ckeletin-version
```

Expected: `ckeletin framework v0.1.0`

- [ ] **Step 5: Commit**

```bash
git add .ckeletin/VERSION .ckeletin/Taskfile.yml Taskfile.yml
git commit -m "feat: add framework versioning (.ckeletin/VERSION)

- Initial framework version: 0.1.0
- Add task ckeletin:version to display framework version
- Enables tracking of framework updates and changelog"
```

---

### Task 5: Add Dry-Run Mode to Framework Update

**Files:**
- Modify: `.ckeletin/Taskfile.yml` (the `update` task)

- [ ] **Step 1: Read the current update task**

Read `.ckeletin/Taskfile.yml` and find the `update` task definition. Understand the full flow.

- [ ] **Step 2: Add a `update:dry-run` task**

Add a new task that runs the same steps as `update` but stops before committing and shows a diff instead:

```yaml
  update:dry-run:
    desc: Preview framework update without applying
    cmds:
      - |
        set -e
        CURRENT_MODULE=$(head -1 go.mod | awk '{print $2}')
        UPSTREAM_MODULE="github.com/peiman""/ckeletin-go"
        if [ "$CURRENT_MODULE" = "$UPSTREAM_MODULE" ]; then
          echo "❌ Cannot update the upstream repository itself"
          exit 1
        fi

        # Set up remote if needed
        if ! git remote get-url ckeletin-upstream >/dev/null 2>&1; then
          git remote add ckeletin-upstream "https://github.com/peiman""/ckeletin-go.git"
        fi
        git fetch ckeletin-upstream

        # Show what would change
        echo "📋 Preview of framework update:"
        echo ""
        git diff HEAD ckeletin-upstream/main -- .ckeletin/ | head -200
        echo ""

        DIFF_STAT=$(git diff --stat HEAD ckeletin-upstream/main -- .ckeletin/)
        if [ -z "$DIFF_STAT" ]; then
          echo "✅ Framework is already up-to-date"
        else
          echo "$DIFF_STAT"
          echo ""
          echo "Run 'task ckeletin:update' to apply these changes."
        fi
```

- [ ] **Step 3: Add alias in root Taskfile**

```yaml
  ckeletin-update-dry-run:
    desc: Preview framework update without applying
    cmds: [task: ckeletin:update:dry-run]
```

- [ ] **Step 4: Test**

```bash
task ckeletin-update-dry-run
```

Expected: Shows "already up-to-date" (since we're on the upstream repo itself, it should fail with the safety check — test on the logic, not the output).

- [ ] **Step 5: Commit**

```bash
git add .ckeletin/Taskfile.yml Taskfile.yml
git commit -m "feat: add dry-run mode to framework update

- task ckeletin:update:dry-run previews changes without applying
- Shows diff and stat of what would change
- Enables safe preview before committing updates"
```

---

### Task 6: Improve Command Generator

**Files:**
- Modify: `.ckeletin/scripts/scaffold/` or the generate:command task in Taskfile
- May need to create: `.ckeletin/scripts/generate-command.go` or modify existing generator

- [ ] **Step 1: Read the current generate:command task**

Find and read the `generate:command` task definition in `Taskfile.yml` and/or `.ckeletin/Taskfile.yml`. Understand what it currently generates.

- [ ] **Step 2: Understand the full pattern by reading existing commands**

Read these files to understand the pattern that should be generated:
- `cmd/ping.go` (command wrapper)
- `internal/ping/ping.go` (business logic with Executor pattern)
- `internal/ping/ping_test.go` (test with table-driven tests)
- `internal/config/commands/ping_config.go` (config metadata)

- [ ] **Step 3: Write a Go-based generator**

Create `.ckeletin/scripts/generate-command/main.go` that:
1. Takes a command name as argument
2. Generates `cmd/<name>.go` (ultra-thin wrapper using helpers)
3. Generates `internal/<name>/<name>.go` (Executor pattern with Config struct)
4. Generates `internal/<name>/<name>_test.go` (table-driven test skeleton)
5. Generates `internal/config/commands/<name>_config.go` (config provider registration)
6. Runs `task generate:config:key-constants` to update generated constants

Use `text/template` for the templates. Follow the exact patterns from the ping command.

The command template (`cmd/<name>.go`):
```go
package cmd

import (
	"{{.Module}}/internal/config"
	"{{.Module}}/internal/config/commands"
	"{{.Module}}/internal/{{.Name}}"
	"github.com/spf13/cobra"
)

var {{.Name}}Cmd = MustNewCommand(commands.{{.NameTitle}}Metadata, run{{.NameTitle}})

func init() {
	MustAddToRoot({{.Name}}Cmd)
}

func run{{.NameTitle}}(cmd *cobra.Command, args []string) error {
	cfg := {{.Name}}.Config{
		// TODO: Add config fields
	}
	return {{.Name}}.NewExecutor(cfg, cmd.OutOrStdout()).Execute()
}
```

The business logic template (`internal/<name>/<name>.go`):
```go
package {{.Name}}

import (
	"io"

	"github.com/rs/zerolog/log"
)

// Config holds configuration for the {{.Name}} command.
type Config struct {
	// TODO: Add config fields
}

// Executor handles the execution of the {{.Name}} command.
type Executor struct {
	cfg    Config
	writer io.Writer
}

// NewExecutor creates a new {{.Name}} executor.
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	return &Executor{cfg: cfg, writer: writer}
}

// Execute runs the {{.Name}} logic.
func (e *Executor) Execute() error {
	log.Debug().Msg("Starting {{.Name}} execution")

	// TODO: Implement business logic

	return nil
}
```

The test template (`internal/<name>/<name>_test.go`):
```go
package {{.Name}}

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecutor_Execute(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "successful execution",
			cfg:     Config{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			executor := NewExecutor(tt.cfg, buf)

			err := executor.Execute()

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			_ = assert.ObjectsAreEqual // ensure testify imported
		})
	}
}
```

The config template (`internal/config/commands/<name>_config.go`):
```go
package commands

import "{{.Module}}/.ckeletin/pkg/config"

// {{.NameTitle}}Metadata defines the command metadata.
var {{.NameTitle}}Metadata = config.CommandMetadata{
	Use:          "{{.Name}}",
	Short:        "TODO: Brief description of {{.Name}} command",
	Long:         "TODO: Detailed description of what {{.Name}} does",
	ConfigPrefix: "app.{{.Name}}",
	FlagOverrides: map[string]string{
		// TODO: Add flag overrides
	},
}

func init() {
	config.RegisterOptionsProvider({{.NameTitle}}Options)
}

// {{.NameTitle}}Options returns configuration options for {{.Name}}.
func {{.NameTitle}}Options() []config.ConfigOption {
	return []config.ConfigOption{
		// TODO: Add config options
	}
}
```

- [ ] **Step 4: Update the Taskfile task**

Update the `generate:command` task in `Taskfile.yml` to call the new generator:

```yaml
  generate:command:
    desc: Generate a new command with full pattern (cmd + internal + tests + config)
    cmds:
      - go run ./.ckeletin/scripts/generate-command/ {{.CLI_ARGS}}
      - task: ckeletin:generate:config:key-constants
      - task: ckeletin:format
    requires:
      vars: [name]
```

- [ ] **Step 5: Test the generator**

```bash
task generate:command name=example
```

Verify it creates:
- `cmd/example.go`
- `internal/example/example.go`
- `internal/example/example_test.go`
- `internal/config/commands/example_config.go`

```bash
go build ./...
go test ./internal/example/...
```

- [ ] **Step 6: Clean up test artifacts**

```bash
rm -rf cmd/example.go internal/example/ internal/config/commands/example_config.go
task generate:config:key-constants
```

- [ ] **Step 7: Commit**

```bash
git add .ckeletin/scripts/generate-command/ Taskfile.yml .ckeletin/Taskfile.yml
git commit -m "feat: improve command generator to scaffold full pattern

- Generates cmd/<name>.go (ultra-thin wrapper)
- Generates internal/<name>/<name>.go (Executor pattern)
- Generates internal/<name>/<name>_test.go (table-driven tests)
- Generates internal/config/commands/<name>_config.go (config provider)
- Auto-runs config constant generation and formatting
- Reduces 8-step manual process to 1 command"
```

---

### Task 7: Add Race Detection to CI

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Read CI workflow test section**

Find the test step in ci.yml. Look for where `task test` or `task check` is called.

- [ ] **Step 2: Add race detection flag**

The `task check` already runs tests. Check if race detection is included. If not, ensure the CI test run includes `-race` on Linux (race detector is not available on Windows). This may already be handled by the task system — read `task test:full` definition in `.ckeletin/Taskfile.yml` to check.

If race detection is already in `task check` (via `test:full` which includes `test:race`), verify this and document. If not, add it.

- [ ] **Step 3: Commit if changes needed**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: ensure race detection runs in CI test pipeline

- Race conditions could previously be merged undetected
- Enabled on Linux only (not available on Windows)"
```

---

### Task 8: Improve internal/check/ Test Coverage

**Files:**
- Modify: `internal/check/*_test.go` (multiple test files)
- Possibly modify: `internal/check/executor.go` (extract testable logic)

- [ ] **Step 1: Assess current coverage**

```bash
go test -coverprofile=check_cover.out ./internal/check/...
go tool cover -func=check_cover.out | tail -20
```

Identify the least-covered functions.

- [ ] **Step 2: Identify testable units**

Read `internal/check/executor.go` and identify functions that can be tested without TUI:
- Check registration and lookup
- Category filtering
- Fail-fast logic
- Result aggregation
- Timing persistence (read/write)

- [ ] **Step 3: Write tests for non-TUI logic**

Focus on:
1. Check category filtering (given a list of checks and a category filter, correct subset returned)
2. Fail-fast behavior (first failure stops execution)
3. Result aggregation (counts, timing, status)
4. Timing file read/write (with temp files)
5. Check definition validation (all checks have names, categories, scripts)

Use table-driven tests with testify assertions.

- [ ] **Step 4: Verify coverage improvement**

```bash
go test -coverprofile=check_cover.out ./internal/check/...
go tool cover -func=check_cover.out | grep total
```

Target: 75%+ (up from 48.5%).

- [ ] **Step 5: Clean up and commit**

```bash
rm check_cover.out
git add internal/check/
git commit -m "test: improve internal/check/ coverage from 48.5% to 75%+

- Add tests for check filtering, fail-fast, result aggregation
- Add tests for timing persistence with temp files
- Focus on non-TUI business logic"
```

---

### Task 9: Add Post-Update Build Check

**Files:**
- Modify: `.ckeletin/Taskfile.yml` (the `update` task)

- [ ] **Step 1: Read the update task**

Find the `update` task in `.ckeletin/Taskfile.yml`.

- [ ] **Step 2: Add build verification after the commit step**

After the framework update commits, add a build check:

```bash
echo "🔍 Verifying project builds after update..."
if ! go build ./...; then
  echo "❌ Build failed after framework update!"
  echo "The update may have introduced breaking changes."
  echo "Check the framework CHANGELOG and fix compilation errors."
  echo "To rollback: git revert HEAD"
  exit 1
fi
echo "✅ Build verified successfully"
```

- [ ] **Step 3: Commit**

```bash
git add .ckeletin/Taskfile.yml
git commit -m "feat: add post-update build verification to framework update

- Runs go build ./... after framework update commit
- Catches breaking API changes immediately
- Provides rollback instructions if build fails"
```

---

### Task 10: Add Framework CHANGELOG

**Files:**
- Create: `.ckeletin/CHANGELOG.md`

- [ ] **Step 1: Create the CHANGELOG**

```markdown
# Framework Changelog

All notable changes to the ckeletin framework layer (`.ckeletin/`) are documented here.

This changelog follows [Keep a Changelog](https://keepachangelog.com/) format.
Only framework changes are tracked here — project-level changes belong in the root `CHANGELOG.md`.

## [0.1.0] - 2026-04-01

### Added
- Framework versioning via `.ckeletin/VERSION`
- `task ckeletin:version` command
- `task ckeletin:update:dry-run` for safe update preview
- Post-update build verification
- Framework CHANGELOG (this file)

### Changed
- AGENTS.md reframed as reusable reference implementation
- README restructured with Agent-Ready Architecture section

### Fixed
- Go version mismatch (.go-version vs go.mod)
```

- [ ] **Step 2: Commit**

```bash
git add .ckeletin/CHANGELOG.md
git commit -m "docs: add framework CHANGELOG for tracking breaking changes

- Follows Keep a Changelog format
- Tracks framework-layer changes separately from project changes
- Initial entry covers v0.1.0 framework improvements"
```

---

## Phase 3: Medium-Term — Security, DX, Marketing (Recommendations 12-20)

### Task 11: Create Asciinema Demo

**Files:**
- Create: demo recording (not committed, uploaded to asciinema.org)
- Modify: `README.md` (add demo link/embed)

- [ ] **Step 1: Install asciinema if needed**

```bash
brew install asciinema 2>/dev/null || pip3 install asciinema
```

- [ ] **Step 2: Plan the demo script**

The demo should show:
1. Clone and init (fast, skip if too slow)
2. Open Claude Code
3. Ask Claude to add a "greet" command
4. Watch it create the right files, follow patterns, run task check
5. Show the result

Write a script of what to type and expected output. Practice the flow before recording.

- [ ] **Step 3: Record**

```bash
asciinema rec demo.cast
# Execute the demo flow
# Exit when done
```

- [ ] **Step 4: Upload and get embed URL**

```bash
asciinema upload demo.cast
```

- [ ] **Step 5: Add to README**

Add the asciinema embed near the top of README (after TL;DR, before What You Get):

```markdown
### See It in Action

[![asciicast](https://asciinema.org/a/XXXXX.svg)](https://asciinema.org/a/XXXXX)
```

- [ ] **Step 6: Commit README change**

```bash
git add README.md
git commit -m "docs: add asciinema demo of AI-agent workflow to README"
```

---

### Task 12: Write Blog Post

**Files:**
- Create: `docs/blog/how-to-make-any-codebase-agent-ready.md` (draft, not committed to main repo — publish externally)

- [ ] **Step 1: Write the blog post**

Structure:
1. The problem: AI agents guess at conventions
2. The insight: enforcement by automation
3. The pattern: AGENTS.md → rules → hooks → task check
4. Reference implementation: ckeletin-go
5. How to apply this to your own codebase
6. Results: what changes when agents have guardrails

Target: 1,500-2,000 words. Practical, not marketing.

- [ ] **Step 2: Review and publish**

Publish on dev.to, Hashnode, or personal blog. Share on:
- r/golang
- Hacker News
- Go Slack / Discord communities
- Twitter/X with #golang #AI #ClaudeCode tags

This is a documentation task — the actual post content depends on the author's voice and platform.

---

### Task 13: Add Binary Signing (cosign)

**Files:**
- Modify: `.github/workflows/ci.yml` (release job)
- Modify: `.goreleaser.yml` (signing config)

- [ ] **Step 1: Read the current release workflow**

Read the release job in ci.yml and the GoReleaser config.

- [ ] **Step 2: Add cosign installation step**

In the release job, before GoReleaser runs:

```yaml
- name: Install cosign
  uses: sigstore/cosign-installer@v3
```

- [ ] **Step 3: Add signing to GoReleaser config**

In `.goreleaser.yml`, add:

```yaml
signs:
  - cmd: cosign
    artifacts: checksum
    output: true
    args:
      - sign-blob
      - "--output-certificate=${certificate}"
      - "--output-signature=${signature}"
      - "${artifact}"
      - "--yes"
```

- [ ] **Step 4: Test locally (dry run)**

```bash
goreleaser release --snapshot --clean
```

Verify signing artifacts would be created.

- [ ] **Step 5: Commit**

```bash
git add .github/workflows/ci.yml .goreleaser.yml
git commit -m "ci: add cosign binary signing to releases

- Signs release checksums with keyless cosign
- Users can verify release authenticity
- Strengthens supply chain security story"
```

---

### Task 14: Add Commit Message Validation

**Files:**
- Modify: `.lefthook.yml` or `.ckeletin/configs/lefthook.base.yml`

- [ ] **Step 1: Read current hook configuration**

Read `.lefthook.yml` and `.ckeletin/configs/lefthook.base.yml`.

- [ ] **Step 2: Add commit-msg hook**

Add to the framework's lefthook config:

```yaml
commit-msg:
  commands:
    conventional-commit:
      run: |
        MSG=$(cat "$1" 2>/dev/null || cat .git/COMMIT_EDITMSG)
        if ! echo "$MSG" | grep -qE '^(feat|fix|docs|test|refactor|style|perf|build|ci|chore)(\(.+\))?: .+'; then
          echo "❌ Commit message must follow Conventional Commits format:"
          echo "   <type>: <description>"
          echo "   Types: feat, fix, docs, test, refactor, style, perf, build, ci, chore"
          echo ""
          echo "   Your message: $MSG"
          exit 1
        fi
```

- [ ] **Step 3: Test with a bad message**

```bash
echo "bad message" | task test  # or test manually
```

- [ ] **Step 4: Commit**

```bash
git add .lefthook.yml .ckeletin/configs/lefthook.base.yml
git commit -m "feat: add conventional commit validation to pre-commit hooks

- Enforces conventional commit format (type: description)
- Supports all standard types: feat, fix, docs, test, etc.
- Prevents non-conventional messages from being committed"
```

---

### Task 15: Clarify Windows Support

**Files:**
- Modify: `AGENTS.md`
- Modify: `README.md` (if needed)

- [ ] **Step 1: Assess current Windows state**

- CI tests on Windows ✓
- GoReleaser builds Windows binaries ✓
- AGENTS.md says "not officially supported" ✗

- [ ] **Step 2: Decide and document**

Two options:
A) Change to "Windows is supported" (CI tests pass, binaries are built)
B) Change to "Windows has limited support — CI tests pass but interactive features (TUI, color) may not work correctly"

Option B is more honest. Update AGENTS.md:

```markdown
**Platform:** macOS and Linux (primary). Windows is supported for core functionality; interactive features (TUI, colored output) may have limitations.
```

- [ ] **Step 3: Commit**

```bash
git add AGENTS.md
git commit -m "docs: clarify Windows support status

- Core functionality supported (CI tests pass, binaries built)
- Interactive features may have limitations on Windows"
```

---

### Task 16: Add Framework Health Check

**Files:**
- Modify: `.ckeletin/Taskfile.yml` (add `doctor` task or extend existing)
- Modify: `Taskfile.yml` (add alias)

- [ ] **Step 1: Read existing doctor task**

Check what `task doctor` currently does.

- [ ] **Step 2: Add framework-specific health checks**

Create a `ckeletin:health` task (or extend doctor):

```yaml
  health:
    desc: Check framework health and update status
    cmds:
      - |
        set -e
        echo "🔍 Framework Health Check"
        echo ""

        # Version
        VERSION=$(cat .ckeletin/VERSION 2>/dev/null || echo "unknown")
        echo "📦 Framework version: v$VERSION"

        # Local modifications
        if git diff --quiet -- .ckeletin/; then
          echo "✅ No local modifications to .ckeletin/"
        else
          echo "⚠️  Local modifications detected in .ckeletin/"
          git diff --stat -- .ckeletin/
        fi

        # Update check
        CURRENT_MODULE=$(head -1 go.mod | awk '{print $2}')
        UPSTREAM_MODULE="github.com/peiman""/ckeletin-go"
        if [ "$CURRENT_MODULE" = "$UPSTREAM_MODULE" ]; then
          echo "ℹ️  This is the upstream repository"
        else
          if git remote get-url ckeletin-upstream >/dev/null 2>&1; then
            git fetch ckeletin-upstream --quiet 2>/dev/null || true
            BEHIND=$(git rev-list HEAD..ckeletin-upstream/main -- .ckeletin/ 2>/dev/null | wc -l | tr -d ' ')
            if [ "$BEHIND" = "0" ]; then
              echo "✅ Framework is up-to-date"
            else
              echo "⚠️  Framework is $BEHIND commit(s) behind upstream"
              echo "   Run: task ckeletin:update:dry-run"
            fi
          else
            echo "ℹ️  No upstream remote configured (run task ckeletin:update to set up)"
          fi
        fi

        # Import consistency
        echo ""
        echo "🔗 Import consistency:"
        INCONSISTENT=$(grep -rn "github.com/peiman/ckeletin-go" --include="*.go" cmd/ internal/ 2>/dev/null | grep -v "_test.go" | grep -v ".ckeletin/" || true)
        if [ -z "$INCONSISTENT" ]; then
          echo "✅ All imports are consistent"
        else
          echo "⚠️  Found references to upstream module in project code:"
          echo "$INCONSISTENT"
        fi
```

- [ ] **Step 3: Add alias**

```yaml
  ckeletin-health:
    desc: Check framework health and update status
    cmds: [task: ckeletin:health]
```

- [ ] **Step 4: Test**

```bash
task ckeletin-health
```

- [ ] **Step 5: Commit**

```bash
git add .ckeletin/Taskfile.yml Taskfile.yml
git commit -m "feat: add framework health check (task ckeletin:health)

- Shows framework version
- Detects local modifications to .ckeletin/
- Checks for available updates from upstream
- Validates import consistency"
```

---

### Task 17: Add Config-Time Validation

**Files:**
- Modify: `.ckeletin/pkg/config/registry.go` or validation code
- Modify: `cmd/root.go` (where config is loaded)

- [ ] **Step 1: Read current validation flow**

Read `cmd/root.go` to understand when config validation runs. Read `.ckeletin/pkg/config/` for existing validation infrastructure.

- [ ] **Step 2: Add early validation**

After Viper loads config but before command execution, validate user-facing values:
- Color values are valid
- Log levels are valid
- String lengths are within bounds

This should happen in the `PersistentPreRunE` of root command, after config is loaded.

- [ ] **Step 3: Write tests**

Test that invalid config values produce clear error messages at load time, not during execution.

- [ ] **Step 4: Commit**

```bash
git add .ckeletin/pkg/config/ cmd/root.go
git commit -m "feat: add config-time validation for user-facing values

- Validate colors, log levels, and string lengths at config load
- Fail early with clear error messages
- Previously validation happened during command execution"
```

---

### Task 18: Extract TUI from check/executor.go

**Files:**
- Modify: `internal/check/executor.go`
- Create: `internal/check/runner.go` (extracted business logic)
- Modify: `internal/check/*_test.go`

- [ ] **Step 1: Read executor.go and identify TUI vs business logic**

Read `internal/check/executor.go` (508 lines). Identify:
- Business logic: check orchestration, filtering, ordering, timing
- TUI logic: Bubble Tea model, rendering, progress display

- [ ] **Step 2: Extract business logic into a runner**

Create `internal/check/runner.go` with the orchestration logic:
- `RunChecks(checks []Check, opts RunOptions) ([]Result, error)`
- `FilterByCategory(checks []Check, category string) []Check`
- `OrderChecks(checks []Check) []Check`

Keep TUI in executor as a thin wrapper that calls the runner.

- [ ] **Step 3: Write tests for the runner**

Test the extracted functions directly without TUI involvement.

- [ ] **Step 4: Verify coverage improvement**

```bash
go test -coverprofile=check_cover.out ./internal/check/...
go tool cover -func=check_cover.out | grep total
```

- [ ] **Step 5: Commit**

```bash
git add internal/check/
git commit -m "refactor: extract business logic from check executor

- Separate orchestration logic from TUI rendering
- New runner.go handles check filtering, ordering, execution
- Executor becomes thin TUI wrapper around runner
- Improves testability of core check logic"
```

---

### Task 19: Add SBOM Scanning to Release Workflow

**Files:**
- Modify: `.github/workflows/ci.yml` (release job)

- [ ] **Step 1: Read the release job**

Find where SBOMs are generated in the release workflow.

- [ ] **Step 2: Add grype scanning step**

After SBOM generation, add:

```yaml
- name: Scan SBOM for vulnerabilities
  uses: anchore/scan-action@v6
  with:
    sbom: "dist/*.spdx.json"
    fail-build: true
    severity-cutoff: high
```

- [ ] **Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add SBOM vulnerability scanning to release workflow

- Scans generated SBOMs with grype before publishing
- Fails release if high-severity vulnerabilities found
- Closes gap between SBOM generation and vulnerability detection"
```

---

## Phase 4: Long-Term (Recommendations 21-28)

### Task 20: Replace sed with go/ast for Import Rewriting

**Files:**
- Create: `.ckeletin/scripts/rewrite-imports/main.go`
- Modify: `.ckeletin/Taskfile.yml` (update task to use new tool)
- Modify: `.ckeletin/scripts/scaffold/main.go` (use new import rewriter)

This is the most significant engineering task. The current sed-based approach (`sed -i "s|old|new|g"`) does blind string replacement in `.go` files, which can corrupt comments, string constants, and partial matches.

Go's `go/ast` and `golang.org/x/tools/go/ast/astutil` packages provide proper import path rewriting that understands Go syntax.

- [ ] **Step 1: Read the current sed-based rewriting in update task**

Find all sed-based module replacement in:
- `.ckeletin/Taskfile.yml` (the update task)
- `.ckeletin/scripts/scaffold/main.go` (the init script)

Document every location where module paths are replaced.

- [ ] **Step 2: Write the AST-based import rewriter**

Create `.ckeletin/scripts/rewrite-imports/main.go`:

```go
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	oldModule := flag.String("old", "", "Old module path to replace")
	newModule := flag.String("new", "", "New module path")
	dir := flag.String("dir", ".", "Directory to process")
	preservePkg := flag.Bool("preserve-pkg", false, "Preserve old/pkg/ imports (for scaffold init)")
	flag.Parse()

	if *oldModule == "" || *newModule == "" {
		fmt.Fprintf(os.Stderr, "Usage: rewrite-imports -old <module> -new <module> [-dir <dir>] [-preserve-pkg]\n")
		os.Exit(1)
	}

	err := filepath.Walk(*dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip non-Go files, vendor, .git, dist, .task
		if info.IsDir() {
			base := filepath.Base(path)
			if base == "vendor" || base == ".git" || base == "dist" || base == ".task" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		return rewriteFile(path, *oldModule, *newModule, *preservePkg)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func rewriteFile(path, oldModule, newModule string, preservePkg bool) error {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", path, err)
	}

	changed := false
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)

		// Skip pkg/ imports if preserving
		if preservePkg && strings.HasPrefix(importPath, oldModule+"/pkg/") {
			continue
		}

		if strings.HasPrefix(importPath, oldModule) {
			newPath := strings.Replace(importPath, oldModule, newModule, 1)
			imp.Path.Value = fmt.Sprintf(`"%s"`, newPath)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	// Sort imports after rewriting
	ast.SortImports(fset, node)

	// Write back
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating %s: %w", path, err)
	}
	defer f.Close()

	if err := format.Node(f, fset, node); err != nil {
		return fmt.Errorf("formatting %s: %w", path, err)
	}

	return nil
}
```

- [ ] **Step 3: Write tests for the import rewriter**

Create `.ckeletin/scripts/rewrite-imports/main_test.go`:

```go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRewriteFile(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		oldModule   string
		newModule   string
		preservePkg bool
		expected    string
	}{
		{
			name: "basic import rewrite",
			input: `package cmd

import (
	"github.com/old/module/internal/ping"
	"github.com/old/module/.ckeletin/pkg/config"
)
`,
			oldModule: "github.com/old/module",
			newModule: "github.com/new/module",
			expected: `package cmd

import (
	"github.com/new/module/.ckeletin/pkg/config"
	"github.com/new/module/internal/ping"
)
`,
		},
		{
			name: "preserve pkg imports",
			input: `package cmd

import (
	"github.com/old/module/internal/ping"
	"github.com/old/module/pkg/checkmate"
)
`,
			oldModule:   "github.com/old/module",
			newModule:   "github.com/new/module",
			preservePkg: true,
			expected: `package cmd

import (
	"github.com/new/module/internal/ping"
	"github.com/old/module/pkg/checkmate"
)
`,
		},
		{
			name: "no matching imports unchanged",
			input: `package cmd

import "fmt"
`,
			oldModule: "github.com/old/module",
			newModule: "github.com/new/module",
			expected: `package cmd

import "fmt"
`,
		},
		{
			name: "string constants not affected",
			input: `package cmd

import "fmt"

const ref = "github.com/old/module/docs"
`,
			oldModule: "github.com/old/module",
			newModule: "github.com/new/module",
			expected: `package cmd

import "fmt"

const ref = "github.com/old/module/docs"
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.go")
			require.NoError(t, os.WriteFile(path, []byte(tt.input), 0644))

			err := rewriteFile(path, tt.oldModule, tt.newModule, tt.preservePkg)
			require.NoError(t, err)

			got, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(got))
		})
	}
}
```

- [ ] **Step 4: Run the tests**

```bash
go test ./.ckeletin/scripts/rewrite-imports/...
```

- [ ] **Step 5: Update the framework update task to use the AST rewriter**

In `.ckeletin/Taskfile.yml`, replace the sed command in the update task:

Old:
```bash
find .ckeletin -name "*.go" -exec sed -i.bak "s|$UPSTREAM_MODULE|$CURRENT_MODULE|g" {} \;
find .ckeletin -name "*.bak" -delete
```

New:
```bash
go run ./.ckeletin/scripts/rewrite-imports/ \
  -old "$UPSTREAM_MODULE" \
  -new "$CURRENT_MODULE" \
  -dir .ckeletin
```

- [ ] **Step 6: Update scaffold init to use the AST rewriter**

In `.ckeletin/scripts/scaffold/main.go`, replace the string-based `replaceModulePreservingPkg()` with a call to the AST rewriter (or import its logic directly).

For the scaffold init, use the `-preserve-pkg` flag:

```bash
go run ./.ckeletin/scripts/rewrite-imports/ \
  -old "$OLD_MODULE" \
  -new "$NEW_MODULE" \
  -dir . \
  -preserve-pkg
```

- [ ] **Step 7: Test the full init workflow**

```bash
# In a temp directory, test that scaffold init works correctly
task init name=testapp module=github.com/test/testapp
go build ./...
task test
```

- [ ] **Step 8: Commit**

```bash
git add .ckeletin/scripts/rewrite-imports/ .ckeletin/Taskfile.yml .ckeletin/scripts/scaffold/
git commit -m "feat: replace sed with go/ast for import rewriting

- AST-based rewriter understands Go syntax
- Only rewrites actual import paths, not comments or strings
- Supports -preserve-pkg flag for scaffold init
- Eliminates string replacement bugs in framework update and init
- Uses go/ast, go/parser, go/format from standard library"
```

---

### Task 21: Add Pre-Update Compatibility Checking

**Files:**
- Create: `.ckeletin/scripts/check-compatibility/main.go`
- Modify: `.ckeletin/Taskfile.yml`

- [ ] **Step 1: Design the compatibility check**

The check should:
1. Fetch latest framework from upstream (without applying)
2. Copy the fetched `.ckeletin/` to a temp location
3. Rewrite imports to use the current module
4. Attempt `go build ./...` with the new framework
5. Report success or list compilation errors

- [ ] **Step 2: Implement as a task**

```yaml
  update:check-compatibility:
    desc: Check if latest framework update is compatible with your code
    cmds:
      - |
        set -e
        CURRENT_MODULE=$(head -1 go.mod | awk '{print $2}')
        UPSTREAM_MODULE="github.com/peiman""/ckeletin-go"

        if [ "$CURRENT_MODULE" = "$UPSTREAM_MODULE" ]; then
          echo "ℹ️  This is the upstream repository"
          exit 0
        fi

        # Fetch latest
        if ! git remote get-url ckeletin-upstream >/dev/null 2>&1; then
          git remote add ckeletin-upstream "https://github.com/peiman""/ckeletin-go.git"
        fi
        git fetch ckeletin-upstream

        # Create temp branch for testing
        TEMP_BRANCH="ckeletin-compat-check-$$"
        git checkout -b "$TEMP_BRANCH" HEAD --quiet

        # Apply update
        git checkout ckeletin-upstream/main -- .ckeletin/ 2>/dev/null
        go run ./.ckeletin/scripts/rewrite-imports/ \
          -old "$UPSTREAM_MODULE" \
          -new "$CURRENT_MODULE" \
          -dir .ckeletin

        # Test build
        echo "🔍 Checking compatibility with latest framework..."
        if go build ./... 2>&1; then
          echo "✅ Compatible — safe to update"
        else
          echo "❌ Incompatible — build errors detected"
          echo "Review the errors above before running task ckeletin:update"
        fi

        # Cleanup
        git checkout - --quiet
        git branch -D "$TEMP_BRANCH" --quiet
```

- [ ] **Step 3: Test and commit**

```bash
git add .ckeletin/Taskfile.yml
git commit -m "feat: add pre-update compatibility check

- task ckeletin:update:check-compatibility tests latest framework
- Creates temp branch, applies update, attempts build
- Reports compatibility before committing changes
- Cleans up temp branch automatically"
```

---

### Task 22: Add Weekly Fuzz Testing to CI

**Files:**
- Create: `.github/workflows/fuzz.yml`

- [ ] **Step 1: Create fuzz workflow**

```yaml
name: Fuzz Testing

on:
  schedule:
    - cron: '0 3 * * 0'  # Weekly, Sunday 3am UTC
  workflow_dispatch:       # Manual trigger

permissions: {}

jobs:
  fuzz:
    runs-on: ubuntu-latest
    permissions:
      contents: read
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version-file: .go-version

      - name: Run fuzz tests
        run: |
          task test:fuzz FUZZTIME=60s
        timeout-minutes: 10

      - name: Upload crash artifacts
        if: failure()
        uses: actions/upload-artifact@v4
        with:
          name: fuzz-crashes
          path: '**/testdata/fuzz/**'
```

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/fuzz.yml
git commit -m "ci: add weekly fuzz testing workflow

- Runs every Sunday at 3am UTC
- 60-second fuzz duration per target
- Uploads crash artifacts on failure
- Also supports manual trigger via workflow_dispatch"
```

---

### Task 23: Add SLSA Provenance Generation

**Files:**
- Modify: `.github/workflows/ci.yml`

- [ ] **Step 1: Add SLSA provenance to release workflow**

After GoReleaser creates the release, add provenance generation:

```yaml
- name: Generate SLSA provenance
  uses: slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@v2.1.0
  with:
    base64-subjects: "${{ steps.goreleaser.outputs.artifacts }}"
```

Note: SLSA provenance generation may require a separate reusable workflow. Read the slsa-framework documentation for the current recommended approach.

- [ ] **Step 2: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "ci: add SLSA provenance generation for releases

- Generates SLSA Level 3 provenance attestations
- Enables cryptographic verification of build environment
- Strengthens supply chain security"
```

---

### Task 24: Migrate Remaining Tests to Testify

**Files:**
- Modify: Multiple `*_test.go` files that use `t.Errorf()` instead of testify

- [ ] **Step 1: Find tests using raw t.Errorf/t.Fatalf**

```bash
grep -rn "t\.Errorf\|t\.Fatalf\|t\.Error(" --include="*_test.go" cmd/ internal/ pkg/ .ckeletin/pkg/
```

- [ ] **Step 2: Migrate each to testify**

Replace patterns:
- `t.Errorf("got %v, want %v", got, want)` → `assert.Equal(t, want, got)`
- `t.Fatalf("unexpected error: %v", err)` → `require.NoError(t, err)`
- `if err != nil { t.Fatal(err) }` → `require.NoError(t, err)`
- `if got != want { t.Errorf(...) }` → `assert.Equal(t, want, got)`

- [ ] **Step 3: Run tests**

```bash
task test
```

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "test: migrate remaining tests to testify assertions

- Replace t.Errorf/t.Fatalf with assert/require
- Consistent assertion patterns across codebase"
```

---

### Task 25: Add Signal Handling Tests

**Files:**
- Create: `test/integration/signal_test.go` or appropriate test file

- [ ] **Step 1: Write signal handling tests**

Test that the CLI handles SIGINT gracefully:

```go
//go:build !windows

package integration

import (
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdownOnSIGINT(t *testing.T) {
	cmd := exec.Command("./ckeletin-go", "ping", "--ui")
	require.NoError(t, cmd.Start())

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	// Send SIGINT
	require.NoError(t, cmd.Process.Signal(syscall.SIGINT))

	// Should exit within 2 seconds
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		// Process exited — check it wasn't a crash
		if exitErr, ok := err.(*exec.ExitError); ok {
			assert.NotEqual(t, -1, exitErr.ExitCode(), "Process should not crash on SIGINT")
		}
	case <-time.After(2 * time.Second):
		cmd.Process.Kill()
		t.Fatal("Process did not exit within 2 seconds of SIGINT")
	}
}
```

- [ ] **Step 2: Run the test**

```bash
task build
go test -v -run TestGracefulShutdown ./test/integration/...
```

- [ ] **Step 3: Commit**

```bash
git add test/integration/
git commit -m "test: add signal handling tests for graceful shutdown

- Test SIGINT handling on Linux/macOS
- Verify process exits cleanly within timeout
- Skipped on Windows (different signal model)"
```

---

### Task 26: Add Atomic Writes to Timing Persistence

**Files:**
- Modify: `internal/check/timing.go`

- [ ] **Step 1: Read current timing persistence**

Read `internal/check/timing.go` and find the `save()` function that writes JSON.

- [ ] **Step 2: Implement atomic write**

Replace direct file write with write-to-temp-then-rename pattern:

```go
func (t *TimingStore) save() error {
	data, err := json.MarshalIndent(t.history, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling timing data: %w", err)
	}

	// Atomic write: write to temp file, then rename
	tmpFile := t.path + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("writing temp timing file: %w", err)
	}
	if err := os.Rename(tmpFile, t.path); err != nil {
		os.Remove(tmpFile) // cleanup on rename failure
		return fmt.Errorf("renaming timing file: %w", err)
	}
	return nil
}
```

- [ ] **Step 3: Add test for atomic write**

Test that a concurrent read during write doesn't get corrupted data.

- [ ] **Step 4: Commit**

```bash
git add internal/check/timing.go internal/check/timing_test.go
git commit -m "fix: use atomic writes for timing persistence

- Write to .tmp file then rename (atomic on most filesystems)
- Eliminates data race risk on concurrent check runs
- Clean up temp file on rename failure"
```

---

### Task 27: Plugin/Extension Architecture (Design Only)

This is a design task, not implementation. Only pursue if adoption warrants.

**Files:**
- Create: `docs/adr/100-plugin-architecture.md` (project ADR, not framework)

- [ ] **Step 1: Write the ADR**

Document the decision to NOT implement plugins yet, with criteria for when to revisit:

```markdown
# ADR-100: Plugin/Extension Architecture

## Status
Deferred

## Context
The framework is currently all-or-nothing. Users can't add composable plugins
(database, gRPC, API client) that extend the framework.

## Decision
Defer plugin architecture until:
1. 3+ external users request it
2. Adoption reaches 50+ stars
3. A concrete plugin use case emerges that can't be solved by adding to internal/

## Consequences
- Framework stays simple and maintainable
- Users who need extensions add them directly to internal/
- Revisit when adoption criteria are met
```

- [ ] **Step 2: Commit**

```bash
git add docs/adr/
git commit -m "docs: add ADR-100 deferring plugin architecture

- Document decision criteria for when to revisit
- Keep framework simple until adoption warrants complexity"
```

---

## Verification

After completing all tasks, run:

- [ ] `task check` — all 23 checks pass
- [ ] `git log --oneline -30` — verify clean commit history
- [ ] Review the improvement analysis against completed work
- [ ] Update `.ckeletin/CHANGELOG.md` with all changes
