package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scannerProjectRoot returns the absolute path to the project root (two levels
// up from test/integration). Named distinctly to avoid colliding with helpers
// in other files of this package.
func scannerProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	return root
}

// plantStaleWorktree creates a nested module copy under .claude/worktrees/
// containing Go code that does not compile against the current tree —
// exactly what Claude Code's worktree feature leaves behind after refactors.
// Returns a cleanup-registered path.
func plantStaleWorktree(t *testing.T, root string) string {
	t.Helper()
	dir := filepath.Join(root, ".claude", "worktrees", "scanner-exclusion-test")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "cmd"), 0o755))
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	goMod := "module github.com/peiman/ckeletin-go\n\ngo 1.26.4\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goMod), 0o644))

	// Unformatted (bad indentation) AND uncompilable: trips both a formatter
	// walking the tree and any scanner that tries to build it.
	broken := "package cmd\n\nimport \"github.com/peiman/ckeletin-go/internal/nonexistent\"\n\nfunc stale() {  nonexistent.RemovedAPI( os.Exit(1) ) }\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "cmd", "broken.go"), []byte(broken), 0o644))
	return dir
}

// runFrameworkScript executes a .ckeletin script from the project root and
// returns combined output plus the exit error (nil on success).
func runFrameworkScript(t *testing.T, root, script string, args ...string) (string, error) {
	t.Helper()
	path := filepath.Join(root, ".ckeletin", "scripts", script)
	require.FileExists(t, path, "framework script must exist")
	cmd := exec.Command("bash", append([]string{path}, args...)...)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// TestScannersExcludeClaudeWorktrees pins the framework contract that
// file-walking checks never descend into .claude/worktrees/ (Claude Code's
// nested worktree checkouts). A stale copy there must not fail the gate,
// must not be mutated by format fix mode, and must not silently degrade
// scanner coverage (a scanner erroring on a nested module skips scanning it).
func TestScannersExcludeClaudeWorktrees(t *testing.T) {
	root := scannerProjectRoot(t)
	staleDir := plantStaleWorktree(t, root)
	staleFile := filepath.Join(staleDir, "cmd", "broken.go")

	t.Run("format check ignores stale worktree", func(t *testing.T) {
		out, err := runFrameworkScript(t, root, "format-go.sh", "check")
		assert.NoError(t, err, "format check must pass with a stale worktree present.\nOutput:\n%s", out)
		assert.NotContains(t, out, "scanner-exclusion-test",
			"format check must not even mention files under .claude/worktrees/")
	})

	t.Run("format fix does not mutate foreign worktree files", func(t *testing.T) {
		before, err := os.ReadFile(staleFile)
		require.NoError(t, err)

		out, ferr := runFrameworkScript(t, root, "format-go.sh", "fix")
		require.NoError(t, ferr, "format fix must succeed.\nOutput:\n%s", out)

		after, err := os.ReadFile(staleFile)
		require.NoError(t, err)
		assert.Equal(t, string(before), string(after),
			"format fix must never rewrite files inside another worktree's checkout")
	})

	t.Run("layering validation ignores stale worktree", func(t *testing.T) {
		out, err := runFrameworkScript(t, root, "validate-layering.sh")
		assert.NoError(t, err, "go-arch-lint must pass with a stale worktree present.\nOutput:\n%s", out)
	})

	t.Run("sast scan ignores stale worktree", func(t *testing.T) {
		if _, lookErr := exec.LookPath("semgrep"); lookErr != nil {
			t.Skip("semgrep not installed; covered in CI")
		}
		out, err := runFrameworkScript(t, root, "check-sast.sh")
		assert.NoError(t, err, "SAST scan must pass with a stale worktree present.\nOutput:\n%s", out)
	})
}
