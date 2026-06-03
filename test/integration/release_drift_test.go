package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// driftProjectRoot returns the absolute path to the project root (two levels up
// from test/integration). Named distinctly to avoid colliding with the
// scaffold-tagged helpers when both compile under `-tags scaffold`.
func driftProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	return root
}

// driftRunGit runs a git command in dir and fails the test on error.
func driftRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed.\nOutput:\n%s", args, string(out))
}

// driftScriptPath returns the absolute path to check-release-drift.sh.
func driftScriptPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join(driftProjectRoot(t), ".ckeletin", "scripts", "check-release-drift.sh")
	require.FileExists(t, p, "release-drift script must exist")
	return p
}

// runDrift executes the release-drift script in dir and returns its combined
// output and exit code.
func runDrift(t *testing.T, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", driftScriptPath(t))
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "failed to run drift script.\nOutput:\n%s", string(out))
	return string(out), 0
}

// TestReleaseDrift_CleanFrameworkPasses guards against the framework's own
// release infrastructure regressing (e.g. reintroducing a deprecated GoReleaser
// property, a stale token name, or orphan tags). It must exit 0.
func TestReleaseDrift_CleanFrameworkPasses(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release-drift integration test in short mode")
	}
	out, code := runDrift(t, driftProjectRoot(t))
	assert.Equal(t, 0, code, "framework release infrastructure should be clean.\nOutput:\n%s", out)
	assert.Contains(t, out, "No release-infrastructure drift detected")
}

// TestReleaseDrift_DetectsStaleConfig asserts the three deterministic drift
// signals (stale CI token, outdated cosign flags, orphan tag) are each flagged
// and the script exits non-zero. The GoReleaser config check is intentionally
// not asserted here: it requires a git remote (absent in this temp fixture) and
// is covered by TestReleaseDrift_CleanFrameworkPasses.
func TestReleaseDrift_DetectsStaleConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping release-drift integration test in short mode")
	}
	dir := t.TempDir()

	// .goreleaser.yml carrying the OUTDATED cosign flags.
	goreleaser := "version: 2\n" +
		"signs:\n" +
		"  - cmd: cosign\n" +
		"    args:\n" +
		"      - \"--output-signature=${signature}\"\n" +
		"      - \"--output-certificate=${certificate}\"\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".goreleaser.yml"), []byte(goreleaser), 0o644))

	// A workflow with the stale CKELETIN_GITHUB_TOKEN secret name.
	wfDir := filepath.Join(dir, ".github", "workflows")
	require.NoError(t, os.MkdirAll(wfDir, 0o755))
	ci := "jobs:\n  release:\n    env:\n      GITHUB_TOKEN: ${{ secrets.CKELETIN_GITHUB_TOKEN }}\n"
	require.NoError(t, os.WriteFile(filepath.Join(wfDir, "ci.yml"), []byte(ci), 0o644))

	// A git repo with an ORPHAN v* tag (points at a commit unreachable from HEAD).
	driftRunGit(t, dir, "init")
	driftRunGit(t, dir, "config", "user.email", "test@ckeletin-go.example")
	driftRunGit(t, dir, "config", "user.name", "Test User")
	driftRunGit(t, dir, "add", ".")
	driftRunGit(t, dir, "commit", "-m", "init")
	// Annotated tag avoids environments that reject lightweight tags without a message.
	driftRunGit(t, dir, "tag", "-a", "v0.0.1", "-m", "old release")
	// Amend HEAD so the tagged commit is no longer reachable from HEAD.
	driftRunGit(t, dir, "commit", "--amend", "-m", "amended", "--allow-empty")

	out, code := runDrift(t, dir)
	assert.Equal(t, 1, code, "drift should be detected (exit 1).\nOutput:\n%s", out)
	assert.Contains(t, out, "CKELETIN_GITHUB_TOKEN", "should flag the stale CI token name")
	assert.Contains(t, out, "Outdated cosign", "should flag the outdated cosign flags")
	assert.Contains(t, out, "orphan tags", "should flag the orphan tag")
}
