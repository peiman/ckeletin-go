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
)

// TestScaffoldInit tests the complete scaffold initialization workflow
func TestScaffoldInit(t *testing.T) {
	// Skip if task is not installed
	if _, err := exec.LookPath("task"); err != nil {
		t.Skip("task command not found, skipping scaffold init test")
	}

	// Create temp directory for test
	tmpDir := t.TempDir()

	// Copy entire project to temp directory
	t.Logf("Copying project to temp directory: %s", tmpDir)
	if err := copyProjectFiles(tmpDir); err != nil {
		t.Fatalf("Failed to copy project files: %v", err)
	}

	// Initialize git repo (needed for Taskfile VERSION variable)
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if output, err := initCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to init git repo: %v\nOutput: %s", err, string(output))
	}

	// Configure git user (required for commits in CI)
	configEmailCmd := exec.Command("git", "config", "user.email", "test@ckeletin-go.example")
	configEmailCmd.Dir = tmpDir
	if output, err := configEmailCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set git user.email: %v\nOutput: %s", err, string(output))
	}

	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tmpDir
	if output, err := configNameCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to set git user.name: %v\nOutput: %s", err, string(output))
	}

	// Add and commit files (needed for git describe)
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to add files: %v\nOutput: %s", err, string(output))
	}

	commitCmd := exec.Command("git", "commit", "-m", "Initial commit")
	commitCmd.Dir = tmpDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit: %v\nOutput: %s", err, string(output))
	}

	// Run: task init name=testapp module=github.com/test/testapp
	testName := "testapp"
	testModule := "github.com/test/testapp"
	t.Logf("Running: task init name=%s module=%s", testName, testModule)

	cmd := exec.Command("task", "init", "name="+testName, "module="+testModule)
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task init failed: %v\nOutput: %s", err, string(output))
	}
	t.Logf("task init output:\n%s", string(output))

	// Verify: go.mod contains new module path
	t.Run("go.mod updated", func(t *testing.T) {
		goModPath := filepath.Join(tmpDir, "go.mod")
		content, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("Failed to read go.mod: %v", err)
		}

		if !strings.Contains(string(content), "module "+testModule) {
			t.Errorf("go.mod does not contain new module path\nContent:\n%s", string(content))
		}
	})

	// Verify: Taskfile.yml contains new binary name
	t.Run("Taskfile.yml updated", func(t *testing.T) {
		taskfilePath := filepath.Join(tmpDir, "Taskfile.yml")
		content, err := os.ReadFile(taskfilePath)
		if err != nil {
			t.Fatalf("Failed to read Taskfile.yml: %v", err)
		}

		if !strings.Contains(string(content), "BINARY_NAME: "+testName) {
			t.Errorf("Taskfile.yml does not contain new binary name\nContent:\n%s", string(content))
		}
	})

	// Verify: .goreleaser.yml contains new project name
	t.Run(".goreleaser.yml updated", func(t *testing.T) {
		goreleaserPath := filepath.Join(tmpDir, ".goreleaser.yml")
		content, err := os.ReadFile(goreleaserPath)
		if err != nil {
			t.Fatalf("Failed to read .goreleaser.yml: %v", err)
		}

		if !strings.Contains(string(content), "project_name: "+testName) {
			t.Errorf(".goreleaser.yml does not contain new project name\nContent:\n%s", string(content))
		}
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

		if err != nil {
			t.Fatalf("Failed to walk directory: %v", err)
		}

		if len(filesWithOldRefs) > 0 {
			t.Errorf("Found old module references in %d files:\n%s",
				len(filesWithOldRefs), strings.Join(filesWithOldRefs, "\n"))
		}
	})

	// Skip quality checks in integration test - they're validated in the main CI build job
	// Integration test focuses on verifying the scaffold init process works correctly
	// Quality checks require tools (golangci-lint, goimports, bash scripts) not available in test env

	// Run: task build (produces binary)
	t.Run("task build succeeds", func(t *testing.T) {
		t.Logf("Running: task build")
		cmd := exec.Command("task", "build")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("task build failed: %v\nOutput: %s", err, string(output))
		}

		// Verify binary exists
		binaryPath := filepath.Join(tmpDir, testName)
		if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
			t.Errorf("Binary %s does not exist after build", testName)
		}
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
		if err != nil {
			t.Fatalf("Binary execution failed: %v\nOutput: %s", err, string(output))
		}

		// Verify output contains binary name
		if !strings.Contains(string(output), testName) {
			t.Errorf("Binary output does not contain expected name %q\nOutput: %s",
				testName, string(output))
		}
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
