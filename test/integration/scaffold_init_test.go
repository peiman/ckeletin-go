// test/integration/scaffold_init_test.go
//
// Integration tests for scaffold initialization (task init)
//
// These tests verify the complete scaffold initialization workflow:
// - Copying entire project to temp directory
// - Running task init with custom name/module
// - Validating all files are updated correctly
// - Running task check to ensure quality standards
// - Building and executing the customized binary

package integration

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upstreamModule is the canonical upstream module path.
// This constant is used to detect if we're in a derived project.
// NOTE: This value should NOT be replaced by scaffold-init because it's
// a const declaration, not a string in an import or go.mod.
const upstreamModule = "github.com/peiman/ckeletin-go"

// TestScaffoldInit tests the complete scaffold initialization workflow
func TestScaffoldInit(t *testing.T) {
	// Skip in derived projects - this test only makes sense in the upstream repo
	// because it tests the scaffold initialization process itself
	currentModule := getCurrentModule(t)
	if currentModule != upstreamModule {
		t.Skipf("Scaffold init test only runs in upstream repo (current: %s, upstream: %s)", currentModule, upstreamModule)
	}

	// Check if task is available, use fallback if not
	_, taskErr := exec.LookPath("task")
	useTaskFallback := taskErr != nil
	if useTaskFallback {
		t.Log("task command not found, using direct script execution as fallback")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Copy entire project to temp directory
	t.Logf("Copying project to temp directory: %s", tmpDir)
	err := copyProjectFiles(tmpDir)
	require.NoError(t, err, "failed to copy project files")

	// Initialize git repo (needed for Taskfile VERSION variable)
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	output, err := initCmd.CombinedOutput()
	require.NoError(t, err, "failed to init git repo\nOutput: %s", string(output))

	// Configure git user (required for commits in CI)
	configEmailCmd := exec.Command("git", "config", "user.email", "test@ckeletin-go.example")
	configEmailCmd.Dir = tmpDir
	output, err = configEmailCmd.CombinedOutput()
	require.NoError(t, err, "failed to set git user.email\nOutput: %s", string(output))

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tmpDir
	output, err = configNameCmd.CombinedOutput()
	require.NoError(t, err, "failed to set git user.name\nOutput: %s", string(output))

	// Add and commit files (needed for git describe)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	output, err = addCmd.CombinedOutput()
	require.NoError(t, err, "failed to add files\nOutput: %s", string(output))

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tmpDir
	output, err = commitCmd.CombinedOutput()
	require.NoError(t, err, "failed to commit\nOutput: %s", string(output))

	// Run: task init name=testapp module=github.com/test/testapp
	testName := "testapp"
	testModule := "github.com/test/testapp"
	oldModule := upstreamModule // Use constant to avoid replacement by scaffold-init
	oldName := "ckeletin-go"

	// Get project root for replace directive (needed so go mod tidy can resolve
	// external pkg/ imports from local source before a release is published)
	projectRoot := getProjectRoot(t)

	if useTaskFallback {
		// Fallback: run scaffold script directly
		t.Logf("Running: go run ./.ckeletin/scripts/scaffold/ %s %s %s %s", oldModule, testModule, oldName, testName)
		cmd := exec.Command("go", "run", "./.ckeletin/scripts/scaffold/", oldModule, testModule, oldName, testName)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "scaffold init failed\nOutput: %s", string(output))
		t.Logf("scaffold init output:\n%s", string(output))

		// Add replace directive so go mod tidy can resolve ckeletin-go/pkg/checkmate
		// from local source (before a published release includes pkg/)
		addReplaceDirective(t, tmpDir, oldModule, projectRoot)

		// Run go mod tidy
		t.Logf("Running: go mod tidy")
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = tmpDir
		tidyOutput, err := tidyCmd.CombinedOutput()
		require.NoError(t, err, "go mod tidy failed\nOutput: %s", string(tidyOutput))
	} else {
		// Add replace directive BEFORE task init (task init runs go mod tidy internally)
		// We need to add it to go.mod before the module path gets changed by scaffold-init
		// Actually, we add it after scaffold-init but before go mod tidy.
		// Since task init runs both, we use the fallback path for this test instead.
		// For now, inject the directive into go.mod pre-emptively — scaffold-init
		// only replaces the module line and import statements, not replace directives.
		addReplaceDirective(t, tmpDir, oldModule, projectRoot)

		// Use task command
		t.Logf("Running: task init name=%s module=%s", testName, testModule)
		cmd := exec.Command("task", "init", "name="+testName, "module="+testModule)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "task init failed\nOutput: %s", string(output))
		t.Logf("task init output:\n%s", string(output))
	}

	// Verify: go.mod contains new module path
	t.Run("go.mod updated", func(t *testing.T) {
		goModPath := filepath.Join(tmpDir, "go.mod")
		content, err := os.ReadFile(goModPath)
		require.NoError(t, err, "failed to read go.mod")

		assert.Contains(t, string(content), "module "+testModule,
			"go.mod should contain new module path")
	})

	// Verify: Taskfile.yml contains new binary name
	t.Run("Taskfile.yml updated", func(t *testing.T) {
		taskfilePath := filepath.Join(tmpDir, "Taskfile.yml")
		content, err := os.ReadFile(taskfilePath)
		require.NoError(t, err, "failed to read Taskfile.yml")

		assert.Contains(t, string(content), "BINARY_NAME: "+testName,
			"Taskfile.yml should contain new binary name")
	})

	// Verify: .goreleaser.yml contains new project name
	t.Run(".goreleaser.yml updated", func(t *testing.T) {
		goreleaserPath := filepath.Join(tmpDir, ".goreleaser.yml")
		content, err := os.ReadFile(goreleaserPath)
		require.NoError(t, err, "failed to read .goreleaser.yml")

		assert.Contains(t, string(content), "project_name: "+testName,
			".goreleaser.yml should contain new project name")
	})

	// Verify: No old module references remain in Go import statements (except pkg/ imports).
	// The AST-based import rewriter only modifies actual import paths — comments and
	// string constants are intentionally left unchanged (they don't affect compilation).
	t.Run("no old module references except pkg/ imports", func(t *testing.T) {
		pkgPrefix := upstreamModule + "/pkg/"
		var staleRefs []string

		// importLineRe matches Go import lines: optional name, then quoted path
		importLineRe := regexp.MustCompile(`^\s*(\w+\s+)?"[^"]*"`)

		err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip vendor, .git, and non-Go files
			if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
				return nil
			}

			// Skip test files that intentionally keep upstream module references
			if strings.HasSuffix(path, "scaffold_init_test.go") {
				return nil
			}

			// Skip rewrite-imports test fixtures — they contain upstream module
			// references as test data (string constants, not actual imports)
			if strings.Contains(path, "rewrite-imports") {
				return nil
			}

			// Skip scaffold helpers test fixtures for the same reason
			if strings.Contains(path, filepath.Join("scripts", "scaffold")) && strings.HasSuffix(path, "_test.go") {
				return nil
			}

			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				// Only check lines that are import statements, not comments or strings.
				// The AST-based rewriter correctly targets only import paths.
				inImportBlock := false
				lines := strings.Split(string(content), "\n")
				for lineNum, line := range lines {
					trimmed := strings.TrimSpace(line)

					// Track import block boundaries
					if strings.HasPrefix(trimmed, "import (") {
						inImportBlock = true
						continue
					}
					if inImportBlock && trimmed == ")" {
						inImportBlock = false
						continue
					}

					// Check single-line imports and lines within import blocks
					isImportLine := false
					if strings.HasPrefix(trimmed, "import \"") || strings.HasPrefix(trimmed, "import\t\"") {
						isImportLine = true
					}
					if inImportBlock && importLineRe.MatchString(trimmed) {
						isImportLine = true
					}

					if isImportLine && strings.Contains(line, upstreamModule) {
						// pkg/ imports are allowed — they reference ckeletin-go as external dep
						if strings.Contains(line, pkgPrefix) {
							continue
						}
						relPath, _ := filepath.Rel(tmpDir, path)
						staleRefs = append(staleRefs, fmt.Sprintf("%s:%d: %s", relPath, lineNum+1, strings.TrimSpace(line)))
					}
				}
			}

			return nil
		})

		require.NoError(t, err, "failed to walk directory")

		assert.Empty(t, staleRefs,
			"found stale module references in import statements (pkg/ imports are allowed):\n%s",
			strings.Join(staleRefs, "\n"))
	})

	// Verify: pkg/ directory was removed by scaffold init
	t.Run("pkg/ directory removed", func(t *testing.T) {
		pkgDir := filepath.Join(tmpDir, "pkg")
		_, err := os.Stat(pkgDir)
		assert.True(t, os.IsNotExist(err), "pkg/ should be removed after scaffold init")
	})

	// Verify: checkmate imports reference the original ckeletin-go module (external dep)
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
			assert.NotContains(t, string(content), testModule+"/pkg/checkmate",
				"%s should NOT import checkmate from derived module", filepath.Base(f))
		}
	})

	// Verify: go.mod references the original ckeletin-go module (for external pkg/ deps)
	t.Run("go.mod has ckeletin-go dependency", func(t *testing.T) {
		goModContent, err := os.ReadFile(filepath.Join(tmpDir, "go.mod"))
		require.NoError(t, err)
		assert.Contains(t, string(goModContent), oldModule,
			"go.mod should reference the original ckeletin-go module for external pkg/ imports")
	})

	// Skip quality checks in integration test - they're validated in the main CI build job
	// Integration test focuses on verifying the scaffold init process works correctly
	// Quality checks require tools (golangci-lint, goimports, bash scripts) not available in test env

	// Run: task build (produces binary) or go build directly
	t.Run("build succeeds", func(t *testing.T) {
		if useTaskFallback {
			// Fallback: run go build directly
			binaryName := testName
			if runtime.GOOS == "windows" {
				binaryName += ".exe"
			}
			t.Logf("Running: go build -o %s", binaryName)
			cmd := exec.Command("go", "build", "-o", binaryName)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "go build failed\nOutput: %s", string(output))
		} else {
			t.Logf("Running: task build")
			cmd := exec.Command("task", "build")
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "task build failed\nOutput: %s", string(output))
		}

		// Verify binary exists (with .exe on Windows)
		binaryName := testName
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}
		binaryPath := filepath.Join(tmpDir, binaryName)
		_, err := os.Stat(binaryPath)
		assert.False(t, os.IsNotExist(err), "binary %s should exist after build", binaryName)
	})

	// Run: ./testapp --version (binary works)
	t.Run("binary executes", func(t *testing.T) {
		binaryName := testName
		// On Windows, executables have .exe extension
		if runtime.GOOS == "windows" {
			binaryName += ".exe"
		}
		binaryPath := filepath.Join(tmpDir, binaryName)
		t.Logf("Running: %s --version", binaryPath)

		cmd := exec.Command(binaryPath, "--version")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "binary execution failed\nOutput: %s", string(output))

		// Verify output contains binary name
		assert.Contains(t, string(output), testName,
			"binary output should contain expected name")
	})

	// Run: task check (the real user workflow — everything should pass)
	// This catches issues like #72 (binary name in check scripts), #73 (arch-lint
	// stale references), #74 (validator too strict). Skipped when task is not
	// available or on Windows (shell checks require bash).
	t.Run("task check passes", func(t *testing.T) {
		if useTaskFallback {
			t.Skip("task command not available")
		}
		if runtime.GOOS == "windows" {
			t.Skip("task check requires bash (shell-based validators)")
		}

		t.Log("Running: task check")
		cmd := exec.Command("task", "check")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "CI=true", "SKIP_SECRET_SCAN=1")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err,
			"task check should pass after task init\nOutput:\n%s", string(output))
	})
}

// copyProjectFiles recursively copies all project files to the destination
func copyProjectFiles(dstRoot string) error {
	// Get current working directory (project root)
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}

	// Go up two levels from test/integration to project root
	projectRoot = filepath.Join(projectRoot, "..", "..")
	projectRoot, err = filepath.Abs(projectRoot)
	if err != nil {
		return err
	}

	// Walk the project directory and copy all files
	return filepath.Walk(projectRoot, func(srcPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path from project root
		relPath, err := filepath.Rel(projectRoot, srcPath)
		if err != nil {
			return err
		}

		// Skip certain directories
		skipDirs := []string{".git", "vendor", "dist", ".task"}
		for _, skip := range skipDirs {
			if strings.HasPrefix(relPath, skip) || strings.Contains(relPath, string(filepath.Separator)+skip+string(filepath.Separator)) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Skip test binaries
		if strings.HasSuffix(relPath, "-test") || strings.HasSuffix(relPath, "-test.exe") {
			return nil
		}

		// Skip coverage and test output files
		skipFiles := []string{"coverage.txt", "test-output.json", "coverage.html", "bench-results.txt"}
		for _, skip := range skipFiles {
			if filepath.Base(srcPath) == skip {
				return nil
			}
		}

		dstPath := filepath.Join(dstRoot, relPath)

		// Create directories
		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// Copy files
		srcFile, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := dstFile.ReadFrom(srcFile); err != nil {
			return err
		}

		// Preserve file permissions
		return os.Chmod(dstPath, info.Mode())
	})
}

// getCurrentModule reads the module path from go.mod
func getCurrentModule(t *testing.T) string {
	t.Helper()

	// Get project root (two levels up from test/integration)
	projectRoot, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get project root")

	goModPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err, "failed to read go.mod")

	// Parse first line: "module github.com/..."
	lines := strings.Split(string(content), "\n")
	if len(lines) == 0 {
		t.Fatal("go.mod is empty")
	}

	firstLine := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(firstLine, "module ") {
		t.Fatalf("unexpected go.mod format: %s", firstLine)
	}

	return strings.TrimPrefix(firstLine, "module ")
}

// TestFrameworkUpdate tests the framework update workflow.
// It simulates a downstream project receiving a .ckeletin/ update from upstream
// and verifies that the project still builds, has no stale module references,
// and that all task aliases resolve to existing framework tasks.
func TestFrameworkUpdate(t *testing.T) {
	// Skip in derived projects - this test only makes sense in the upstream repo
	currentModule := getCurrentModule(t)
	if currentModule != upstreamModule {
		t.Skipf("Framework update test only runs in upstream repo (current: %s, upstream: %s)", currentModule, upstreamModule)
	}

	_, taskErr := exec.LookPath("task")
	useTaskFallback := taskErr != nil
	if useTaskFallback {
		t.Log("task command not found, using direct script execution as fallback")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Step 1: Copy project to temp dir and set up git
	t.Log("Copying project to temp directory")
	err := copyProjectFiles(tmpDir)
	require.NoError(t, err, "failed to copy project files")

	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@ckeletin-go.example")
	runGit(t, tmpDir, "config", "user.name", "Test User")
	runGit(t, tmpDir, "add", ".")
	runGit(t, tmpDir, "commit", "-m", "Initial commit")

	// Step 2: Run scaffold init to simulate a downstream project
	testName := "updatetest"
	testModule := "github.com/test/updatetest"
	oldModule := upstreamModule // Use constant to avoid replacement by scaffold-init
	oldName := "ckeletin-go"

	// Get project root for replace directive
	projectRoot := getProjectRoot(t)

	if useTaskFallback {
		t.Log("Running scaffold script directly")
		cmd := exec.Command("go", "run", "./.ckeletin/scripts/scaffold/", oldModule, testModule, oldName, testName)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "scaffold init failed\nOutput: %s", string(output))

		// Add replace directive for external pkg/ imports
		addReplaceDirective(t, tmpDir, oldModule, projectRoot)

		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = tmpDir
		tidyOutput, err := tidyCmd.CombinedOutput()
		require.NoError(t, err, "go mod tidy failed\nOutput: %s", string(tidyOutput))
	} else {
		// Add replace directive before task init (which runs go mod tidy internally)
		addReplaceDirective(t, tmpDir, oldModule, projectRoot)

		t.Logf("Running: task init name=%s module=%s", testName, testModule)
		cmd := exec.Command("task", "init", "name="+testName, "module="+testModule)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "task init failed\nOutput: %s", string(output))
	}

	if useTaskFallback {
		// Direct script doesn't reset git history, so commit the initialized state
		runGit(t, tmpDir, "add", ".")
		runGit(t, tmpDir, "commit", "-m", "Initialize as downstream project")
	} else {
		// task init resets git history (rm -rf .git && git init && commit).
		// Re-set local git config since the original .git was destroyed.
		runGit(t, tmpDir, "config", "user.email", "test@ckeletin-go.example")
		runGit(t, tmpDir, "config", "user.name", "Test User")
	}

	// Step 3: Simulate framework update by re-copying .ckeletin/ from source
	srcCkeletin := filepath.Join(projectRoot, ".ckeletin")
	dstCkeletin := filepath.Join(tmpDir, ".ckeletin")

	// Remove existing .ckeletin/ and re-copy from source (simulates git checkout upstream -- .ckeletin/)
	err = os.RemoveAll(dstCkeletin)
	require.NoError(t, err, "failed to remove .ckeletin")

	err = copyDir(srcCkeletin, dstCkeletin)
	require.NoError(t, err, "failed to copy .ckeletin from source")

	// Use AST-based import rewriter (matches real update workflow)
	t.Log("Rewriting imports with AST rewriter")
	rewriteCmd := exec.Command("go", "run", "./.ckeletin/scripts/rewrite-imports/",
		"-old", upstreamModule, "-new", testModule, "-dir", ".ckeletin")
	rewriteCmd.Dir = tmpDir
	rewriteOutput, err := rewriteCmd.CombinedOutput()
	require.NoError(t, err, "AST import rewrite failed\nOutput: %s", string(rewriteOutput))

	// Replace binary name in .ckeletin/ Go string literals (matches real update workflow).
	// The AST rewriter only handles import paths; string literals like
	// "./logs/ckeletin-go.log" need separate replacement.
	t.Log("Replacing binary name in .ckeletin/ Go string literals")
	replaceNameInCkeletinGoFiles(t, dstCkeletin, oldName, testName)

	// Also replace non-import references (e.g., string constants in non-Go files)
	// that the AST rewriter doesn't touch
	err = filepath.Walk(dstCkeletin, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil || info.IsDir() {
			return walkErr
		}
		// Only process non-Go files (AST rewriter handles .go files)
		if strings.HasSuffix(path, ".go") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		updated := strings.ReplaceAll(string(content), upstreamModule, testModule)
		if updated != string(content) {
			return os.WriteFile(path, []byte(updated), info.Mode())
		}
		return nil
	})
	require.NoError(t, err, "failed to replace module paths in non-Go files")

	// Run go mod tidy after update
	tidyCmd := exec.Command("go", "mod", "tidy")
	tidyCmd.Dir = tmpDir
	tidyOutput, err := tidyCmd.CombinedOutput()
	require.NoError(t, err, "go mod tidy after update failed\nOutput: %s", string(tidyOutput))

	// Format code (matches real update workflow step)
	if !useTaskFallback {
		t.Log("Formatting code after update")
		fmtCmd := exec.Command("task", "format")
		fmtCmd.Dir = tmpDir
		fmtOutput, err := fmtCmd.CombinedOutput()
		if err != nil {
			t.Logf("task format output: %s", string(fmtOutput))
		}
	}

	// Regenerate config constants (matches real update workflow step)
	if !useTaskFallback {
		t.Log("Regenerating config constants after update")
		genCmd := exec.Command("task", "ckeletin:generate:config:key-constants")
		genCmd.Dir = tmpDir
		genOutput, err := genCmd.CombinedOutput()
		if err != nil {
			t.Logf("constant generation output: %s", string(genOutput))
		}
	}

	// Subtests
	t.Run("build succeeds after update", func(t *testing.T) {
		cmd := exec.Command("go", "build", "./...")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err, "go build failed after framework update\nOutput: %s", string(output))
	})

	t.Run("no stale module references in imports", func(t *testing.T) {
		// Only check import statements — the AST rewriter intentionally leaves
		// string constants, comments, and test fixtures unchanged.
		importLineRe := regexp.MustCompile(`^\s*(\w+\s+)?"[^"]*"`)
		var staleRefs []string

		err := filepath.Walk(filepath.Join(tmpDir, ".ckeletin"), func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
				return walkErr
			}
			// Skip test files — they contain upstream module as test fixture data
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			inImportBlock := false
			lines := strings.Split(string(content), "\n")
			for lineNum, line := range lines {
				trimmed := strings.TrimSpace(line)

				if strings.HasPrefix(trimmed, "import (") {
					inImportBlock = true
					continue
				}
				if inImportBlock && trimmed == ")" {
					inImportBlock = false
					continue
				}

				isImportLine := false
				if strings.HasPrefix(trimmed, "import \"") || strings.HasPrefix(trimmed, "import\t\"") {
					isImportLine = true
				}
				if inImportBlock && importLineRe.MatchString(trimmed) {
					isImportLine = true
				}

				if isImportLine && strings.Contains(line, upstreamModule) {
					relPath, _ := filepath.Rel(tmpDir, path)
					staleRefs = append(staleRefs, fmt.Sprintf("%s:%d: %s", relPath, lineNum+1, trimmed))
				}
			}
			return nil
		})
		require.NoError(t, err, "failed to walk .ckeletin")
		assert.Empty(t, staleRefs,
			"found stale upstream module references in import statements after update:\n%s",
			strings.Join(staleRefs, "\n"))
	})

	t.Run("task aliases resolve", func(t *testing.T) {
		projectTaskfile := filepath.Join(tmpDir, "Taskfile.yml")
		frameworkTaskfile := filepath.Join(tmpDir, ".ckeletin", "Taskfile.yml")

		aliasTargets := parseTaskAliasTargets(t, projectTaskfile)
		frameworkTasks := parseFrameworkTaskNames(t, frameworkTaskfile)

		require.NotEmpty(t, aliasTargets, "no alias targets found in project Taskfile")
		require.NotEmpty(t, frameworkTasks, "no tasks found in framework Taskfile")

		// Build a set of framework task names for fast lookup
		taskSet := make(map[string]bool)
		for _, name := range frameworkTasks {
			taskSet[name] = true
		}

		var unresolved []string
		for _, target := range aliasTargets {
			// target is like "ckeletin:check" — strip the "ckeletin:" prefix
			// to get the task name in .ckeletin/Taskfile.yml
			parts := strings.SplitN(target, ":", 2)
			if len(parts) != 2 {
				continue
			}
			taskName := parts[1]
			if !taskSet[taskName] {
				unresolved = append(unresolved, target+" (task '"+taskName+"' not found in framework)")
			}
		}

		assert.Empty(t, unresolved,
			"alias targets in Taskfile.yml point to missing framework tasks:\n%s",
			strings.Join(unresolved, "\n"))
	})

	// Verify that task check passes after a framework update — catches validator,
	// lint, and architecture check breakage that go build alone wouldn't find.
	t.Run("task check passes after update", func(t *testing.T) {
		if useTaskFallback {
			t.Skip("task command not available")
		}
		if runtime.GOOS == "windows" {
			t.Skip("task check requires bash (shell-based validators)")
		}

		t.Log("Running: task check")
		cmd := exec.Command("task", "check")
		cmd.Dir = tmpDir
		cmd.Env = append(os.Environ(), "CI=true", "SKIP_SECRET_SCAN=1")
		output, err := cmd.CombinedOutput()

		assert.NoError(t, err,
			"task check should pass after framework update\nOutput:\n%s", string(output))
	})
}

// TestPostUpdateMigration tests that the post-update migration script
// correctly handles stale configuration and is idempotent.
func TestPostUpdateMigration(t *testing.T) {
	// Skip in derived projects
	currentModule := getCurrentModule(t)
	if currentModule != upstreamModule {
		t.Skipf("Migration test only runs in upstream repo (current: %s)", currentModule)
	}

	projectRoot := getProjectRoot(t)
	migrationScript := filepath.Join(projectRoot, ".ckeletin", "scripts", "migrate-post-update.sh")
	if _, err := os.Stat(migrationScript); os.IsNotExist(err) {
		t.Skip("migrate-post-update.sh not found")
	}

	t.Run("cleans stale pkg references from arch lint config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a .go-arch-lint.yml with stale pkg/** references but no pkg/ directory
		staleConfig := `version: 3
workdir: .
components:
  public:
    in: pkg/**
  internal:
    in: internal/**
commonComponents:
  - public
  - internal
deps:
  public:
    canDependOn: []
  internal:
    canDependOn:
      - public
`
		err := os.WriteFile(filepath.Join(tmpDir, ".go-arch-lint.yml"), []byte(staleConfig), 0644)
		require.NoError(t, err)

		// Ensure no pkg/ directory exists
		pkgDir := filepath.Join(tmpDir, "pkg")
		assert.NoDirExists(t, pkgDir, "pkg/ should not exist for this test")

		// Run migration script
		cmd := exec.Command("bash", migrationScript)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "migration script failed\nOutput: %s", string(output))

		// Verify pkg/** references are removed
		content, err := os.ReadFile(filepath.Join(tmpDir, ".go-arch-lint.yml"))
		require.NoError(t, err)
		contentStr := string(content)

		assert.NotContains(t, contentStr, "pkg/**",
			"pkg/** references should be removed after migration")
		assert.NotContains(t, contentStr, "public:",
			"public component should be removed after migration")
		assert.NotContains(t, contentStr, "- public",
			"public commonComponent entry should be removed after migration")
		// internal should still exist
		assert.Contains(t, contentStr, "internal",
			"internal component should still exist after migration")
	})

	t.Run("is idempotent when no pkg dir and no config", func(t *testing.T) {
		tmpDir := t.TempDir()

		// No .go-arch-lint.yml, no pkg/ — migration should be a no-op
		cmd := exec.Command("bash", migrationScript)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		assert.NoError(t, err,
			"migration should succeed as no-op when config doesn't exist\nOutput: %s", string(output))
	})

	t.Run("is idempotent when pkg dir exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create pkg/ directory and .go-arch-lint.yml with pkg/** references
		err := os.MkdirAll(filepath.Join(tmpDir, "pkg"), 0755)
		require.NoError(t, err)

		configWithPkg := `version: 3
components:
  public:
    in: pkg/**
  internal:
    in: internal/**
`
		err = os.WriteFile(filepath.Join(tmpDir, ".go-arch-lint.yml"), []byte(configWithPkg), 0644)
		require.NoError(t, err)

		// Run migration — should NOT remove pkg/** because pkg/ dir exists
		cmd := exec.Command("bash", migrationScript)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "migration script failed\nOutput: %s", string(output))

		// Verify pkg/** references are preserved
		content, err := os.ReadFile(filepath.Join(tmpDir, ".go-arch-lint.yml"))
		require.NoError(t, err)
		assert.Contains(t, string(content), "pkg/**",
			"pkg/** references should be preserved when pkg/ directory exists")
	})

	t.Run("is idempotent on repeated runs", func(t *testing.T) {
		tmpDir := t.TempDir()

		staleConfig := `version: 3
components:
  public:
    in: pkg/**
  internal:
    in: internal/**
commonComponents:
  - public
deps:
  public:
    canDependOn: []
`
		err := os.WriteFile(filepath.Join(tmpDir, ".go-arch-lint.yml"), []byte(staleConfig), 0644)
		require.NoError(t, err)

		// Run migration twice
		for i := 0; i < 2; i++ {
			cmd := exec.Command("bash", migrationScript)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			assert.NoError(t, err,
				"migration run %d should succeed\nOutput: %s", i+1, string(output))
		}

		// Verify final state is clean
		content, err := os.ReadFile(filepath.Join(tmpDir, ".go-arch-lint.yml"))
		require.NoError(t, err)
		assert.NotContains(t, string(content), "pkg/**",
			"pkg/** should be removed after repeated migration runs")
	})
}

// runGit executes a git command in the given directory and requires it to succeed.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed\nOutput: %s", strings.Join(args, " "), string(output))
}

// getProjectRoot returns the absolute path to the project root (two levels up from test/integration).
func getProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to get project root")
	return root
}

// addReplaceDirective appends a replace directive to go.mod in the given directory.
// This allows go mod tidy to resolve external pkg/ imports from the local project root
// before a published release includes those packages.
func addReplaceDirective(t *testing.T, dir, module, localPath string) {
	t.Helper()
	goModPath := filepath.Join(dir, "go.mod")
	content, err := os.ReadFile(goModPath)
	require.NoError(t, err, "failed to read go.mod for replace directive")

	directive := fmt.Sprintf("\nreplace %s => %s\n", module, localPath)
	err = os.WriteFile(goModPath, append(content, []byte(directive)...), 0600)
	require.NoError(t, err, "failed to write replace directive to go.mod")
	t.Logf("Added replace directive: %s => %s", module, localPath)
}

// copyDir recursively copies a directory tree from src to dst.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			return err
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dstFile.Close()

		if _, err := io.Copy(dstFile, srcFile); err != nil {
			return err
		}

		return os.Chmod(dstPath, info.Mode())
	})
}

// parseTaskAliasTargets extracts all ckeletin: alias targets from a project Taskfile.
// It finds patterns like [task: ckeletin:X] and returns the full target strings.
func parseTaskAliasTargets(t *testing.T, path string) []string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read %s", path)

	re := regexp.MustCompile(`\[task:\s*(ckeletin:\S+)\]`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	var targets []string
	for _, m := range matches {
		targets = append(targets, m[1])
	}
	return targets
}

// parseFrameworkTaskNames extracts all task names from a .ckeletin/Taskfile.yml.
// It parses top-level keys under the "tasks:" section.
func parseFrameworkTaskNames(t *testing.T, path string) []string {
	t.Helper()
	content, err := os.ReadFile(path)
	require.NoError(t, err, "failed to read %s", path)

	var names []string
	inTasks := false
	// Match lines with exactly 2-space indent followed by a task name and colon
	re := regexp.MustCompile(`^  ([a-zA-Z0-9][a-zA-Z0-9:._-]*):\s*$`)

	for _, line := range strings.Split(string(content), "\n") {
		if strings.TrimSpace(line) == "tasks:" {
			inTasks = true
			continue
		}
		if !inTasks {
			continue
		}
		// Stop if we hit a non-indented, non-comment line (outside tasks section)
		if len(line) > 0 && line[0] != ' ' && line[0] != '#' {
			break
		}
		if m := re.FindStringSubmatch(line); m != nil {
			names = append(names, m[1])
		}
	}
	return names
}

// replaceNameInCkeletinGoFiles replaces the old binary name with the new name
// in Go string literals within .ckeletin/, skipping import lines.
// Also replaces the env var prefix form (e.g., CKELETIN_GO → UPDATETEST).
// This mirrors what task ckeletin:update does for binary name references
// (e.g., "./logs/ckeletin-go.log" → "./logs/myapp.log").
func replaceNameInCkeletinGoFiles(t *testing.T, ckeletinDir, oldName, newName string) {
	t.Helper()

	// Build replacement pairs: binary name + env var prefix
	type replacement struct{ old, new string }
	replacements := []replacement{{oldName, newName}}

	oldEnv := strings.ToUpper(strings.ReplaceAll(oldName, "-", "_"))
	newEnv := strings.ToUpper(strings.ReplaceAll(newName, "-", "_"))
	if oldEnv != newEnv {
		replacements = append(replacements, replacement{oldEnv, newEnv})
	}

	for _, r := range replacements {
		err := filepath.Walk(ckeletinDir, func(path string, info os.FileInfo, walkErr error) error {
			if walkErr != nil || info.IsDir() {
				return walkErr
			}
			if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}

			content, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}

			original := string(content)
			lines := strings.Split(original, "\n")
			inImportBlock := false

			for i, line := range lines {
				trimmed := strings.TrimSpace(line)

				if strings.HasPrefix(trimmed, "import (") {
					inImportBlock = true
					continue
				}
				if inImportBlock {
					if trimmed == ")" {
						inImportBlock = false
					}
					continue
				}
				if strings.HasPrefix(trimmed, "import ") {
					continue
				}

				lines[i] = strings.ReplaceAll(line, r.old, r.new)
			}

			updated := strings.Join(lines, "\n")
			if updated != original {
				return os.WriteFile(path, []byte(updated), info.Mode())
			}
			return nil
		})

		require.NoError(t, err, "failed to replace %q in .ckeletin Go files", r.old)
	}
}
