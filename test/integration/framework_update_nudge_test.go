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

const nudgeMarker = "Framework update available"

// nudgeScriptPath returns the absolute path to check-framework-updates.sh.
func nudgeScriptPath(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	p := filepath.Join(root, ".ckeletin", "scripts", "check-framework-updates.sh")
	require.FileExists(t, p, "framework-updates script must exist")
	return p
}

// nudgeRunGit runs a git command in dir and fails the test on error.
func nudgeRunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed.\nOutput:\n%s", args, string(out))
}

// runNudge runs the nudge script in dir with a controlled environment. The
// behaviour-affecting vars (CI, opt-out, TTL) are stripped from the inherited
// environment so the test is deterministic even when it runs inside CI; only
// the provided overrides are applied.
func runNudge(t *testing.T, dir string, overrides map[string]string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", nudgeScriptPath(t))
	cmd.Dir = dir

	drop := map[string]bool{"CI": true, "CKELETIN_SKIP_UPDATE_CHECK": true, "CKELETIN_UPDATE_CHECK_TTL": true}
	var env []string
	for _, kv := range os.Environ() {
		if k, _, ok := strings.Cut(kv, "="); ok && drop[k] {
			continue
		}
		env = append(env, kv)
	}
	// Force a fetch every run (no throttle) so tests are deterministic regardless
	// of any stamp file; a test may still override this.
	env = append(env, "CKELETIN_UPDATE_CHECK_TTL=0")
	for k, v := range overrides {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "nudge script errored.\nOutput:\n%s", string(out))
	return string(out), 0
}

// setupDownstream builds an "upstream" repo with a .ckeletin/ tree and a
// "downstream" clone whose origin is renamed to ckeletin-upstream, mirroring a
// real scaffolded project. Returns both paths.
func setupDownstream(t *testing.T) (upstream, downstream string) {
	t.Helper()
	upstream = t.TempDir()
	downstream = t.TempDir()

	nudgeRunGit(t, upstream, "init", "-b", "main")
	nudgeRunGit(t, upstream, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, upstream, "config", "user.name", "Test User")
	require.NoError(t, os.MkdirAll(filepath.Join(upstream, ".ckeletin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, ".ckeletin", "f"), []byte("a\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "go.mod"), []byte("module github.com/test/up\n"), 0o644))
	nudgeRunGit(t, upstream, "add", "-A")
	nudgeRunGit(t, upstream, "commit", "-m", "c1")

	clone := exec.Command("git", "clone", "--quiet", upstream, downstream)
	out, err := clone.CombinedOutput()
	require.NoError(t, err, "git clone failed.\nOutput:\n%s", string(out))

	nudgeRunGit(t, downstream, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, downstream, "config", "user.name", "Test User")
	nudgeRunGit(t, downstream, "remote", "rename", "origin", "ckeletin-upstream")
	// Give the downstream a distinct module path (it is not the upstream repo).
	require.NoError(t, os.WriteFile(filepath.Join(downstream, "go.mod"), []byte("module github.com/test/down\n"), 0o644))
	nudgeRunGit(t, downstream, "commit", "-am", "downstream init")
	return upstream, downstream
}

// advanceUpstream adds a commit touching .ckeletin/ in the upstream repo.
func advanceUpstream(t *testing.T, upstream string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(upstream, ".ckeletin", "f"), []byte("a\nb\n"), 0o644))
	nudgeRunGit(t, upstream, "commit", "-am", "c2")
}

func TestUpdateNudge_WhenBehind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstream(t, upstream)

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code, "nudge must never fail the build")
	assert.Contains(t, out, nudgeMarker, "should nudge when .ckeletin/ is behind upstream.\nOutput:\n%s", out)
	assert.Contains(t, out, "task ckeletin:update", "should point at the update command")
}

func TestUpdateNudge_SkipsInCI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstream(t, upstream)

	out, code := runNudge(t, downstream, map[string]string{"CI": "true"})
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent in CI.\nOutput:\n%s", out)
}

func TestUpdateNudge_OptOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstream(t, upstream)

	out, code := runNudge(t, downstream, map[string]string{"CKELETIN_SKIP_UPDATE_CHECK": "1"})
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must honour the opt-out.\nOutput:\n%s", out)
}

func TestUpdateNudge_UpToDate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	_, downstream := setupDownstream(t)
	// Upstream not advanced: downstream's .ckeletin/ is current.

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent when up to date.\nOutput:\n%s", out)
}

func TestUpdateNudge_NoUpstreamRemote(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	dir := t.TempDir()
	nudgeRunGit(t, dir, "init", "-b", "main")
	nudgeRunGit(t, dir, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, dir, "config", "user.name", "Test User")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/down\n"), 0o644))
	nudgeRunGit(t, dir, "add", "-A")
	nudgeRunGit(t, dir, "commit", "-m", "init")

	out, code := runNudge(t, dir, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent without an upstream remote.\nOutput:\n%s", out)
}
