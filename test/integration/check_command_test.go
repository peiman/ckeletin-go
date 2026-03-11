// test/integration/check_command_test.go
//
// Integration tests for the check command (dev build).
// Exercises the full check pipeline: Execute → runCategorySimple →
// individual check functions → save timing → printFinalSummary.

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// devBinaryPath is the path to the dev-tagged binary.
var devBinaryPath string

// projectRoot returns the absolute path to the project root (two levels up from test/integration/).
func projectRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../..")
	require.NoError(t, err)
	return abs
}

// buildDevBinary builds the binary with dev tags for check command tests.
// Returns an absolute path so the binary works regardless of cmd.Dir.
func buildDevBinary(t *testing.T) string {
	t.Helper()

	if devBinaryPath != "" {
		return devBinaryPath
	}

	binaryName := "ckeletin-go-dev-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	root := projectRoot(t)
	outPath := filepath.Join(root, binaryName)

	cmd := exec.Command("go", "build", "-tags", "dev", "-o", outPath, ".")
	cmd.Dir = root
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build dev binary: %s", string(output))

	devBinaryPath = outPath
	t.Cleanup(func() {
		os.Remove(devBinaryPath)
		devBinaryPath = ""
	})

	return devBinaryPath
}

// TestCheckCommand_QualityCategory runs the check command with --category quality.
// This exercises the highest-impact uncovered code paths:
//   - Executor.Execute (orchestration)
//   - runCategorySimple (non-TUI output)
//   - checkFormat (goimports + gofmt)
//   - checkLint (go vet + golangci-lint)
//   - timingHistory.save (persist timing data)
//   - printFinalSummary (summary box output)
func TestCheckCommand_QualityCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping check command integration test in short mode")
	}

	binary := buildDevBinary(t)
	root := projectRoot(t)

	cmd := exec.Command(binary, "check", "--category", "quality")
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Ensure non-TTY mode (simple output, not TUI)
	cmd.Env = append(os.Environ(), "CI=true")

	err := cmd.Run()
	combinedOutput := stdout.String() + stderr.String()

	// Quality checks should pass on a well-formatted codebase
	exitCode := getExitCode(err)
	assert.Equal(t, 0, exitCode,
		"check --category quality should pass\nOutput:\n%s", combinedOutput)

	// Verify output structure
	assert.Contains(t, combinedOutput, "Code Quality",
		"Should contain Code Quality category header")
	assert.Contains(t, combinedOutput, "format",
		"Should show format check")
	assert.Contains(t, combinedOutput, "lint",
		"Should show lint check")
	assert.Contains(t, combinedOutput, "All",
		"Should contain summary")
}

// TestCheckCommand_DepsCategory runs the check command with --category dependencies.
// Exercises checkDeps, checkVuln, and shellCheck for license/sbom checks.
func TestCheckCommand_DepsCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping check command integration test in short mode")
	}
	if runtime.GOOS == "windows" {
		t.Skip("Skipping deps category test on Windows (license tools not available)")
	}

	binary := buildDevBinary(t)
	root := projectRoot(t)

	cmd := exec.Command(binary, "check", "--category", "dependencies")
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = append(os.Environ(), "CI=true")

	err := cmd.Run()
	combinedOutput := stdout.String() + stderr.String()

	exitCode := getExitCode(err)
	assert.Equal(t, 0, exitCode,
		"check --category dependencies should pass\nOutput:\n%s", combinedOutput)

	assert.Contains(t, combinedOutput, "Dependencies",
		"Should contain Dependencies category header")
}

// TestCheckCommand_InvalidCategory tests error handling for invalid category.
func TestCheckCommand_InvalidCategory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping check command integration test in short mode")
	}

	binary := buildDevBinary(t)
	root := projectRoot(t)

	cmd := exec.Command(binary, "check", "--category", "nonexistent")
	cmd.Dir = root
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	combinedOutput := stdout.String() + stderr.String()

	// Should fail with invalid category
	exitCode := getExitCode(err)
	assert.NotEqual(t, 0, exitCode,
		"check with invalid category should fail")
	assert.True(t,
		strings.Contains(combinedOutput, "invalid") ||
			strings.Contains(combinedOutput, "unknown") ||
			strings.Contains(combinedOutput, "nonexistent"),
		"Should report invalid category\nOutput:\n%s", combinedOutput)
}
