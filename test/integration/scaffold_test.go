// test/integration/scaffold_test.go

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestScaffoldInitAndUpdate tests the full user journey:
// 1. Clone the repo
// 2. Run task init
// 3. Verify build works
// 4. Simulate framework update
// 5. Verify build still works
func TestScaffoldInitAndUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get the repo root (two levels up from test/integration)
	repoRoot, err := filepath.Abs("../..")
	require.NoError(t, err)

	// Create temp directories
	tempDir := t.TempDir()
	upstreamDir := filepath.Join(tempDir, "upstream")
	projectDir := filepath.Join(tempDir, "myapp")

	// Step 1: Create bare upstream repo
	t.Log("Creating bare upstream repo...")
	require.NoError(t, os.MkdirAll(upstreamDir, 0755))
	runCmd(t, upstreamDir, "git", "init", "--bare")

	// Push current state to upstream
	t.Log("Pushing current state to upstream...")
	// Remove remote if it exists, then add fresh
	_ = exec.Command("git", "-C", repoRoot, "remote", "remove", "test-scaffold-upstream").Run()
	runCmd(t, repoRoot, "git", "remote", "add", "test-scaffold-upstream", upstreamDir)
	defer func() {
		_ = exec.Command("git", "-C", repoRoot, "remote", "remove", "test-scaffold-upstream").Run()
	}()

	// Get current branch
	branch := strings.TrimSpace(runCmdOutput(t, repoRoot, "git", "rev-parse", "--abbrev-ref", "HEAD"))
	runCmd(t, repoRoot, "git", "push", "test-scaffold-upstream", branch+":main")

	// Step 2: Clone as user project
	t.Log("Cloning as user project...")
	runCmd(t, tempDir, "git", "clone", upstreamDir, "myapp")

	// Step 3: Run task init
	t.Log("Running task init...")
	runCmd(t, projectDir, "task", "init", "name=myapp", "module=github.com/testuser/myapp")

	// Step 4: Verify build works after init
	t.Log("Verifying build after init...")
	runCmd(t, projectDir, "go", "build", "./...")

	// Commit init changes
	runCmd(t, projectDir, "git", "add", "-A")
	runCmd(t, projectDir, "git", "commit", "-m", "init: myapp")

	// Step 5: Verify imports in .ckeletin/ are updated
	t.Log("Verifying module paths in .ckeletin/...")
	loggerContent, err := os.ReadFile(filepath.Join(projectDir, ".ckeletin/pkg/logger/logger.go"))
	require.NoError(t, err)
	assert.Contains(t, string(loggerContent), "github.com/testuser/myapp/.ckeletin/pkg/config",
		"Module path should be updated in .ckeletin/ after init")
	assert.NotContains(t, string(loggerContent), "github.com/peiman/ckeletin-go",
		"Old module path should not exist in .ckeletin/ after init")

	// Step 6: Simulate framework update
	// We simulate this by resetting .ckeletin/ to upstream (with old module paths)
	// and then running the module path replacement
	t.Log("Simulating framework update...")
	runCmd(t, projectDir, "git", "remote", "add", "ckeletin-upstream", upstreamDir)
	runCmd(t, projectDir, "git", "fetch", "ckeletin-upstream", "main")

	// Checkout upstream .ckeletin/ (this has OLD module paths)
	runCmd(t, projectDir, "git", "checkout", "ckeletin-upstream/main", "--", ".ckeletin/")

	// Verify upstream has old module paths
	loggerBeforeReplace, err := os.ReadFile(filepath.Join(projectDir, ".ckeletin/pkg/logger/logger.go"))
	require.NoError(t, err)
	assert.Contains(t, string(loggerBeforeReplace), "github.com/peiman/ckeletin-go",
		"After checkout from upstream, should have old module path")

	// Step 7: Update module paths (this is the critical step we're testing)
	t.Log("Running module path replacement...")
	findCmd := exec.Command("find", ".ckeletin", "-name", "*.go", "-exec",
		"sed", "-i", "", "s|github.com/peiman/ckeletin-go|github.com/testuser/myapp|g", "{}", ";")
	findCmd.Dir = projectDir
	require.NoError(t, findCmd.Run(), "Failed to update module paths")

	// Step 8: Verify build still works after update
	t.Log("Verifying build after update...")
	runCmd(t, projectDir, "go", "build", "./...")

	// Step 9: Verify module paths are correct after update
	t.Log("Verifying module paths after update...")
	loggerContentAfter, err := os.ReadFile(filepath.Join(projectDir, ".ckeletin/pkg/logger/logger.go"))
	require.NoError(t, err)
	assert.Contains(t, string(loggerContentAfter), "github.com/testuser/myapp/.ckeletin/pkg/config",
		"Module path should be updated in .ckeletin/ after update")
	assert.NotContains(t, string(loggerContentAfter), "github.com/peiman/ckeletin-go",
		"Old module path should not exist in .ckeletin/ after update")

	t.Log("✅ Scaffold init and update test passed!")
}

// runCmd runs a command and fails the test if it errors
func runCmd(t *testing.T, dir string, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	require.NoError(t, cmd.Run(), "Command failed: %s %v", name, args)
}

// runCmdOutput runs a command and returns its output
func runCmdOutput(t *testing.T, dir string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.Output()
	require.NoError(t, err, "Command failed: %s %v", name, args)
	return string(output)
}
