# pkg/ Cleanup on Scaffold Init — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** When `task init name=myapp` scaffolds a new project, clean out `pkg/` (ckeletin-go's libraries) and preserve references to `github.com/peiman/ckeletin-go/pkg/checkmate` as an external dependency so the check command still works.

**Architecture:** The scaffold-init script currently does a blind find-replace of the old module path in all Go files. We modify it to skip replacements inside lines that reference `pkg/` packages, then delete the local `pkg/` directory. After `go mod tidy` runs (already part of `task init`), Go resolves checkmate as an external dependency from the published ckeletin-go module.

**Tech Stack:** Go, testify/assert, testify/require, integration tests with binary execution

**Key constraint — `//go:build ignore`:** The scaffold-init script uses `//go:build ignore` and runs via `go run <file>`. Files with this tag cannot be tested with `go test`. We restructure the script into a proper package directory (`.ckeletin/scripts/scaffold/`) so the code is testable normally.

**Key constraint — module resolution:** After scaffold init, the derived project imports `github.com/peiman/ckeletin-go/pkg/checkmate` as an external dependency. For `go mod tidy` to resolve this, a published version of ckeletin-go must include `pkg/checkmate`. In integration tests (before publishing), we add a `replace` directive pointing to the local project root.

---

## File Map

| File | Action | Responsibility |
|------|--------|---------------|
| `.ckeletin/scripts/scaffold/main.go` | Create | Scaffold-init script (moved from scaffold-init.go) |
| `.ckeletin/scripts/scaffold/helpers.go` | Create | `ReplaceModulePreservingPkg` + `RemovePkgDirectory` functions |
| `.ckeletin/scripts/scaffold/helpers_test.go` | Create | Unit tests for helpers |
| `.ckeletin/scripts/scaffold-init.go` | Delete | Replaced by scaffold/ package |
| `Taskfile.yml` | Modify | Update `init` task to use `go run ./.ckeletin/scripts/scaffold/` |
| `test/integration/scaffold_init_test.go` | Modify | Update e2e test + add pkg/ cleanup assertions + replace directive |

---

### Task 1: Restructure scaffold-init.go into a testable package

**Files:**
- Create: `.ckeletin/scripts/scaffold/main.go`
- Create: `.ckeletin/scripts/scaffold/helpers.go`
- Delete: `.ckeletin/scripts/scaffold-init.go`
- Modify: `Taskfile.yml`

The existing `scaffold-init.go` uses `//go:build ignore` which makes it untestable. We move it to a proper package directory so `go test` works.

- [ ] **Step 1: Create the scaffold package directory**

```bash
mkdir -p .ckeletin/scripts/scaffold
```

- [ ] **Step 2: Move scaffold-init.go to scaffold/main.go, removing the build constraint**

Copy `.ckeletin/scripts/scaffold-init.go` to `.ckeletin/scripts/scaffold/main.go`. Remove the `//go:build ignore` line. Keep everything else identical.

The file should start with:

```go
package main

import (
    // ... same imports as before
)

func main() {
    // ... same main() as before
}
```

- [ ] **Step 3: Delete the old file**

```bash
rm .ckeletin/scripts/scaffold-init.go
```

- [ ] **Step 4: Update Taskfile.yml init task**

In `Taskfile.yml`, find the `init` task. Change:

```yaml
- go run .ckeletin/scripts/scaffold-init.go "{{.OLD_MODULE}}" "{{.MODULE}}" "{{.OLD_NAME}}" "{{.NAME}}"
```

to:

```yaml
- go run ./.ckeletin/scripts/scaffold/ "{{.OLD_MODULE}}" "{{.MODULE}}" "{{.OLD_NAME}}" "{{.NAME}}"
```

- [ ] **Step 5: Verify the script still works by building it**

```bash
go build ./.ckeletin/scripts/scaffold/
```

Expected: Compiles without errors.

- [ ] **Step 6: Update the scaffold init test's skipFiles reference**

In `test/integration/scaffold_init_test.go`, the test may reference the old `scaffold-init.go` path. Search for `scaffold-init.go` and update to `scaffold/main.go` if needed. Also check `.ckeletin/scripts/scaffold-init.go` in the skipFiles map inside the scaffold code itself (now at `scaffold/main.go`).

- [ ] **Step 7: Commit**

```bash
git add .ckeletin/scripts/scaffold/ Taskfile.yml
git rm .ckeletin/scripts/scaffold-init.go
git add test/integration/scaffold_init_test.go
git commit -m "refactor: move scaffold-init to proper package for testability

- .ckeletin/scripts/scaffold-init.go → .ckeletin/scripts/scaffold/main.go
- Remove //go:build ignore constraint (package dir doesn't need it)
- Update Taskfile.yml init task to use go run ./.ckeletin/scripts/scaffold/
- No functional changes — same behavior, now testable with go test"
```

---

### Task 2: Unit-test and implement the pkg-preserving replacement function

**Files:**
- Create: `.ckeletin/scripts/scaffold/helpers.go`
- Create: `.ckeletin/scripts/scaffold/helpers_test.go`

- [ ] **Step 1: Create helpers_test.go with table-driven tests**

```go
// .ckeletin/scripts/scaffold/helpers_test.go
package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplaceModulePreservingPkg(t *testing.T) {
	const oldModule = "github.com/peiman/ckeletin-go"
	const newModule = "github.com/user/myapp"

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces standard internal import",
			input:    `import "github.com/peiman/ckeletin-go/internal/check"`,
			expected: `import "github.com/user/myapp/internal/check"`,
		},
		{
			name:     "replaces .ckeletin import",
			input:    `import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"`,
			expected: `import "github.com/user/myapp/.ckeletin/pkg/config"`,
		},
		{
			name:     "preserves pkg/checkmate import",
			input:    `	"github.com/peiman/ckeletin-go/pkg/checkmate"`,
			expected: `	"github.com/peiman/ckeletin-go/pkg/checkmate"`,
		},
		{
			name:     "preserves any pkg/ import",
			input:    `	"github.com/peiman/ckeletin-go/pkg/somefuture"`,
			expected: `	"github.com/peiman/ckeletin-go/pkg/somefuture"`,
		},
		{
			name: "handles mixed import block",
			input: `import (
	"fmt"
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
)`,
			expected: `import (
	"fmt"
	"github.com/user/myapp/internal/ping"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/user/myapp/.ckeletin/pkg/config"
)`,
		},
		{
			name:     "replaces module in go.mod",
			input:    `module github.com/peiman/ckeletin-go`,
			expected: `module github.com/user/myapp`,
		},
		{
			name:     "replaces module reference in comments",
			input:    `// See github.com/peiman/ckeletin-go/internal/check for details`,
			expected: `// See github.com/user/myapp/internal/check for details`,
		},
		{
			name:     "preserves pkg/ reference in comments too",
			input:    `// Uses github.com/peiman/ckeletin-go/pkg/checkmate for output`,
			expected: `// Uses github.com/peiman/ckeletin-go/pkg/checkmate for output`,
		},
		{
			name:     "no changes when no module reference",
			input:    `package main`,
			expected: `package main`,
		},
		{
			name:     "handles empty input",
			input:    ``,
			expected: ``,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := replaceModulePreservingPkg(tt.input, oldModule, newModule)
			assert.Equal(t, tt.expected, got)
		})
	}
}
```

- [ ] **Step 2: Run the test to verify it fails**

```bash
go test -v -run TestReplaceModulePreservingPkg ./.ckeletin/scripts/scaffold/
```

Expected: FAIL — `replaceModulePreservingPkg` doesn't exist yet.

- [ ] **Step 3: Create helpers.go with the implementation**

```go
// .ckeletin/scripts/scaffold/helpers.go
package main

import (
	"os"
	"path/filepath"
	"strings"
)

// replaceModulePreservingPkg replaces oldModule with newModule in content,
// but preserves lines that reference oldModule/pkg/ (external library imports).
// This allows derived projects to keep pkg/ packages as external dependencies
// from the original ckeletin-go module.
func replaceModulePreservingPkg(content, oldModule, newModule string) string {
	pkgPrefix := oldModule + "/pkg/"
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pkgPrefix) {
			continue // Preserve pkg/ references
		}
		lines[i] = strings.ReplaceAll(line, oldModule, newModule)
	}
	return strings.Join(lines, "\n")
}

// removePkgDirectory removes the pkg/ directory from the project root.
// After scaffold init, pkg/ packages (like checkmate) are consumed as external
// dependencies from the original ckeletin-go module, not local copies.
func removePkgDirectory(projectRoot string) error {
	pkgDir := filepath.Join(projectRoot, "pkg")
	if _, err := os.Stat(pkgDir); os.IsNotExist(err) {
		return nil // Nothing to remove
	}
	return os.RemoveAll(pkgDir)
}
```

- [ ] **Step 4: Run the test to verify it passes**

```bash
go test -v -run TestReplaceModulePreservingPkg ./.ckeletin/scripts/scaffold/
```

Expected: PASS — all 10 test cases pass.

- [ ] **Step 5: Commit**

```bash
git add .ckeletin/scripts/scaffold/helpers.go .ckeletin/scripts/scaffold/helpers_test.go
git commit -m "feat: add pkg-preserving module replacement and cleanup helpers

- replaceModulePreservingPkg skips lines containing oldModule/pkg/
- removePkgDirectory removes pkg/ after scaffold init
- Unit tests cover import blocks, comments, edge cases"
```

---

### Task 3: Unit-test the removePkgDirectory function

**Files:**
- Modify: `.ckeletin/scripts/scaffold/helpers_test.go`

- [ ] **Step 1: Add tests for removePkgDirectory**

Append to `.ckeletin/scripts/scaffold/helpers_test.go`:

```go
func TestRemovePkgDirectory(t *testing.T) {
	t.Run("removes pkg directory and contents", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a fake pkg/ structure with nested dirs
		pkgDir := filepath.Join(tmpDir, "pkg")
		checkmatDir := filepath.Join(pkgDir, "checkmate")
		demoDir := filepath.Join(pkgDir, "checkmate", "demo")
		require.NoError(t, os.MkdirAll(demoDir, 0750))
		require.NoError(t, os.WriteFile(filepath.Join(checkmatDir, "checkmate.go"), []byte("package checkmate"), 0600))
		require.NoError(t, os.WriteFile(filepath.Join(demoDir, "main.go"), []byte("package main"), 0600))

		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)

		_, err = os.Stat(pkgDir)
		assert.True(t, os.IsNotExist(err), "pkg/ directory should be removed")
	})

	t.Run("no error when pkg directory does not exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)
	})

	t.Run("preserves other directories", func(t *testing.T) {
		tmpDir := t.TempDir()

		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "pkg", "checkmate"), 0750))
		require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "internal", "check"), 0750))
		internalFile := filepath.Join(tmpDir, "internal", "check", "check.go")
		require.NoError(t, os.WriteFile(internalFile, []byte("package check"), 0600))

		err := removePkgDirectory(tmpDir)
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(tmpDir, "internal", "check", "check.go"))
		assert.NoError(t, err, "internal/ files should be preserved")

		_, err = os.Stat(filepath.Join(tmpDir, "pkg"))
		assert.True(t, os.IsNotExist(err), "pkg/ should be removed")
	})
}
```

- [ ] **Step 2: Run the tests**

```bash
go test -v -run TestRemovePkgDirectory ./.ckeletin/scripts/scaffold/
```

Expected: PASS — all 3 test cases pass.

- [ ] **Step 3: Commit**

```bash
git add .ckeletin/scripts/scaffold/helpers_test.go
git commit -m "test: add unit tests for removePkgDirectory

- Tests nested directory removal, missing directory, preserves other dirs"
```

---

### Task 4: Wire helpers into scaffold main.go

**Files:**
- Modify: `.ckeletin/scripts/scaffold/main.go`

- [ ] **Step 1: Replace `strings.ReplaceAll` with `replaceModulePreservingPkg` in `updateGoFiles`**

In `.ckeletin/scripts/scaffold/main.go`, find the `updateGoFiles` function. Find the line:

```go
updated := strings.ReplaceAll(string(content), oldModule, newModule)
```

Replace with:

```go
updated := replaceModulePreservingPkg(string(content), oldModule, newModule)
```

- [ ] **Step 2: Add pkg/ cleanup step to `main()`**

In `main()`, after the template file update step and before the final success message, add:

```go
	fmt.Println("  ✓ Cleaning pkg/ directory (libraries available as external dependencies)")
	if err := removePkgDirectory("."); err != nil {
		fmt.Fprintf(os.Stderr, "Error removing pkg/ directory: %v\n", err)
		os.Exit(1)
	}
```

Place this BEFORE the `go mod tidy` message — the cleanup should happen before tidy so tidy sees the external imports without local copies.

- [ ] **Step 3: Run all unit tests**

```bash
go test -v ./.ckeletin/scripts/scaffold/
```

Expected: PASS.

- [ ] **Step 4: Commit**

```bash
git add .ckeletin/scripts/scaffold/main.go
git commit -m "feat: wire pkg-preserving replacement and cleanup into scaffold init

- updateGoFiles now uses replaceModulePreservingPkg
- pkg/ directory is removed before go mod tidy
- checkmate imports preserved as external dependency from ckeletin-go module"
```

---

### Task 5: Update integration test — fix "no old module references" and add replace directive

**Files:**
- Modify: `test/integration/scaffold_init_test.go`

The existing test has two issues after our change:
1. The "no old module references" check will fail because pkg/ imports are intentionally preserved
2. `go mod tidy` (run by `task init`) needs to resolve `github.com/peiman/ckeletin-go` — which may not be available from the proxy at the exact version that includes `pkg/checkmate`. We add a `replace` directive pointing to the local project root BEFORE running init.

- [ ] **Step 1: Add a replace directive injection before running task init**

In `TestScaffoldInit`, find where the test runs `task init` (or `go run scaffold-init.go`). BEFORE that execution, add code to inject a replace directive into go.mod in the temp directory:

```go
// Add replace directive so go mod tidy can resolve ckeletin-go/pkg/checkmate
// from the local project root (needed before a published release includes pkg/)
projectRoot := getProjectRoot(t)
goModPath := filepath.Join(tmpDir, "go.mod")
goModContent, err := os.ReadFile(goModPath)
require.NoError(t, err)
replaceDirective := fmt.Sprintf("\nreplace %s => %s\n", oldModule, projectRoot)
err = os.WriteFile(goModPath, append(goModContent, []byte(replaceDirective)...), 0600)
require.NoError(t, err)
```

Note: After `task init` replaces the module path in go.mod, the replace directive's LEFT side (`github.com/peiman/ckeletin-go`) won't be touched because it's the external module reference, not the project module. The scaffold-init will replace the `module` line but the `replace` directive stays intact because `replace github.com/peiman/ckeletin-go` — the LHS matches `oldModule` but scaffold-init only processes the module declaration and import statements, not arbitrary go.mod directives. **Verify this assumption during implementation** — if scaffold-init does a blind string replace on go.mod too, the replace directive's LHS will also change. In that case, add the replace directive AFTER init instead.

- [ ] **Step 2: Update "no old module references" subtest**

Find the `"no old module references"` subtest. Replace the assertion to allow `/pkg/` imports:

```go
t.Run("no old module references except pkg/ imports", func(t *testing.T) {
    err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.IsDir() {
            name := info.Name()
            if name == ".git" || name == "vendor" || name == "dist" || name == ".task" {
                return filepath.SkipDir
            }
            return nil
        }
        if !strings.HasSuffix(path, ".go") {
            return nil
        }
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        lines := strings.Split(string(content), "\n")
        for lineNum, line := range lines {
            if strings.Contains(line, oldModule) {
                // pkg/ imports are allowed — they reference ckeletin-go as external dep
                if strings.Contains(line, oldModule+"/pkg/") {
                    continue
                }
                relPath, _ := filepath.Rel(tmpDir, path)
                t.Errorf("stale module reference in %s:%d: %s", relPath, lineNum+1, strings.TrimSpace(line))
            }
        }
        return nil
    })
    require.NoError(t, err)
})
```

- [ ] **Step 3: Run the test**

```bash
go test -v -run TestScaffoldInit -count=1 ./test/integration/...
```

Expected: PASS (or skip if not running on upstream module).

- [ ] **Step 4: Commit**

```bash
git add test/integration/scaffold_init_test.go
git commit -m "test: update scaffold init test for pkg/ preservation

- Add replace directive for local module resolution in tests
- Allow pkg/ imports in 'no old module references' check
- These are intentional external dependency references"
```

---

### Task 6: Add pkg/ cleanup e2e assertions

**Files:**
- Modify: `test/integration/scaffold_init_test.go`

- [ ] **Step 1: Add new subtests for pkg/ cleanup behavior**

Add these subtests inside `TestScaffoldInit`, after the existing subtests:

```go
t.Run("pkg/ directory removed", func(t *testing.T) {
    pkgDir := filepath.Join(tmpDir, "pkg")
    _, err := os.Stat(pkgDir)
    assert.True(t, os.IsNotExist(err), "pkg/ should be removed after scaffold init")
})

t.Run("checkmate imported as external dependency", func(t *testing.T) {
    checkFiles := []string{
        filepath.Join(tmpDir, "internal", "check", "executor.go"),
        filepath.Join(tmpDir, "internal", "check", "summary.go"),
        filepath.Join(tmpDir, "internal", "ui", "check.go"),
        filepath.Join(tmpDir, "internal", "ui", "check_test.go"),
    }
    for _, f := range checkFiles {
        content, err := os.ReadFile(f)
        if err != nil {
            t.Logf("Skipping %s (not found): %v", filepath.Base(f), err)
            continue
        }
        assert.Contains(t, string(content), oldModule+"/pkg/checkmate",
            "%s should import checkmate from original ckeletin-go module", filepath.Base(f))
        assert.NotContains(t, string(content), newModule+"/pkg/checkmate",
            "%s should NOT import checkmate from derived module", filepath.Base(f))
    }
})

t.Run("go.mod has ckeletin-go dependency", func(t *testing.T) {
    goModContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
    require.NoError(t, err)
    assert.Contains(t, string(goModContent), oldModule,
        "go.mod should reference the original ckeletin-go module")
})
```

- [ ] **Step 2: Also check TestFrameworkUpdate if it needs updates**

The `TestFrameworkUpdate` test also runs scaffold-init and has module reference checks. Inspect whether it needs the same `replace` directive and pkg/ import allowance. If it does, apply the same pattern. If it only checks `.ckeletin/` files (which don't import from `pkg/`), it may be fine as-is.

- [ ] **Step 3: Run the full integration test suite**

```bash
go test -v -run "TestScaffoldInit|TestFrameworkUpdate" -count=1 ./test/integration/...
```

Expected: PASS for all subtests.

- [ ] **Step 4: Commit**

```bash
git add test/integration/scaffold_init_test.go
git commit -m "test: add e2e assertions for pkg/ cleanup on scaffold init

- Verify pkg/ directory is removed
- Verify checkmate imports reference original ckeletin-go module (4 files)
- Verify go.mod includes ckeletin-go as dependency
- Check TestFrameworkUpdate compatibility
- Existing 'build succeeds' test confirms compilation with external dep"
```

---

### Task 7: Final verification

- [ ] **Step 1: Run all unit tests**

```bash
go test -v ./.ckeletin/scripts/scaffold/
```

Expected: PASS — `TestReplaceModulePreservingPkg` (10 cases) and `TestRemovePkgDirectory` (3 cases).

- [ ] **Step 2: Run `task check`**

```bash
task check
```

Expected: All quality checks pass. If coverage drops, investigate and add tests.

- [ ] **Step 3: Run the integration tests**

```bash
go test -v -run "TestScaffoldInit|TestFrameworkUpdate" -count=1 ./test/integration/...
```

Expected: PASS with all subtests including pkg/ cleanup assertions.

- [ ] **Step 4: Verify the script path update didn't break any references**

```bash
grep -rn "scaffold-init.go" . --include="*.go" --include="*.yml" --include="*.yaml" --include="*.md" | grep -v ".git/"
```

Any remaining references to the old `scaffold-init.go` path need updating. Check CLAUDE.md, AGENTS.md, README.md, CONTRIBUTING.md, and test files.

- [ ] **Step 5: Fix any remaining references and commit**

```bash
task format
git add -A
git commit -m "chore: fix remaining scaffold-init.go references and formatting"
```

(Only if there are changes.)

- [ ] **Step 6: Push**

```bash
git push origin main
```

---

## Release Dependency Note

After this change is merged, derived projects need a published version of ckeletin-go that includes `pkg/checkmate`. Until a new release is tagged:
- The integration test works via `replace` directive (local resolution)
- Real users running `task init` on a fresh clone will get a `go mod tidy` error if no published version has `pkg/checkmate`

**Action required after merge:** Tag a new release so `go mod tidy` can resolve `github.com/peiman/ckeletin-go/pkg/checkmate` from the Go module proxy.
