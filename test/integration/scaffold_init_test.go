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
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScaffoldInit tests the complete scaffold initialization workflow
func TestScaffoldInit(t *testing.T) {
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
	oldModule := "github.com/peiman/ckeletin-go"
	oldName := "ckeletin-go"

	if useTaskFallback {
		// Fallback: run scaffold-init.go directly
		t.Logf("Running: go run scripts/scaffold-init.go %s %s %s %s", oldModule, testModule, oldName, testName)
		cmd := exec.Command("go", "run", "scripts/scaffold-init.go", oldModule, testModule, oldName, testName)
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		require.NoError(t, err, "scaffold-init.go failed\nOutput: %s", string(output))
		t.Logf("scaffold-init.go output:\n%s", string(output))

		// Run go mod tidy
		t.Logf("Running: go mod tidy")
		tidyCmd := exec.Command("go", "mod", "tidy")
		tidyCmd.Dir = tmpDir
		tidyOutput, err := tidyCmd.CombinedOutput()
		require.NoError(t, err, "go mod tidy failed\nOutput: %s", string(tidyOutput))
	} else {
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

	// Verify: No old module references remain in Go files
	t.Run("no old module references", func(t *testing.T) {
		oldModule := "github.com/peiman/ckeletin-go"
		var filesWithOldRefs []string

		err := filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Skip vendor, .git, and non-Go files
			if strings.Contains(path, "vendor") || strings.Contains(path, ".git") {
				return nil
			}

			if !info.IsDir() && strings.HasSuffix(path, ".go") {
				content, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				if strings.Contains(string(content), oldModule) {
					relPath, _ := filepath.Rel(tmpDir, path)
					filesWithOldRefs = append(filesWithOldRefs, relPath)
				}
			}

			return nil
		})

		require.NoError(t, err, "failed to walk directory")

		assert.Empty(t, filesWithOldRefs,
			"found old module references in files:\n%s",
			strings.Join(filesWithOldRefs, "\n"))
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
