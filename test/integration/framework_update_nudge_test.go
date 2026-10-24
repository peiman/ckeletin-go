package integration

import (
	"encoding/json"
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
	cmd := gitCmd(dir, args...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed.\nOutput:\n%s", args, string(out))
}

// runNudge runs the nudge script in dir with a controlled environment. The
// behaviour-affecting vars (CI, opt-out, TTL) are stripped from the inherited
// environment so the test is deterministic even when it runs inside CI; only
// the provided overrides are applied. Extra args select --check or --json.
func runNudge(t *testing.T, dir string, overrides map[string]string, args ...string) (string, int) {
	t.Helper()
	cmdArgs := append([]string{nudgeScriptPath(t)}, args...)
	cmd := exec.Command("bash", cmdArgs...)
	cmd.Dir = dir

	drop := map[string]bool{"CI": true, "CKELETIN_SKIP_UPDATE_CHECK": true, "CKELETIN_UPDATE_CHECK_TTL": true}
	for k := range overrides {
		drop[k] = true
	}
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

func writeVersion(t *testing.T, dir, version string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".ckeletin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "VERSION"), []byte(version+"\n"), 0o644))
}

// setupDownstream builds an "upstream" repo with a .ckeletin/ tree and a
// "downstream" clone whose origin is renamed to ckeletin-upstream, mirroring a
// real scaffolded project. Both sides start at framework version 0.1.0.
func setupDownstream(t *testing.T) (upstream, downstream string) {
	t.Helper()
	upstream = t.TempDir()
	downstream = t.TempDir()

	nudgeRunGit(t, upstream, "init", "-b", "main")
	nudgeRunGit(t, upstream, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, upstream, "config", "user.name", "Test User")
	require.NoError(t, os.MkdirAll(filepath.Join(upstream, ".ckeletin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, ".ckeletin", "f"), []byte("a\n"), 0o644))
	writeVersion(t, upstream, "0.1.0")
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "go.mod"), []byte("module github.com/test/up\n"), 0o644))
	nudgeRunGit(t, upstream, "add", "-A")
	nudgeRunGit(t, upstream, "commit", "-m", "c1")

	clone := gitCmd(filepath.Dir(downstream), "clone", "--quiet", upstream, downstream)
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

// advanceUpstreamVersion commits a new framework version on upstream.
func advanceUpstreamVersion(t *testing.T, upstream, version string) {
	t.Helper()
	writeVersion(t, upstream, version)
	nudgeRunGit(t, upstream, "add", "-A")
	nudgeRunGit(t, upstream, "commit", "-m", "version "+version)
}

// advanceUpstreamFile commits a framework-tree edit that does not change VERSION.
func advanceUpstreamFile(t *testing.T, upstream string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(upstream, ".ckeletin", "f"), []byte("a\nb\n"), 0o644))
	nudgeRunGit(t, upstream, "commit", "-am", "file only")
}

func TestUpdateNudge_WhenBehind(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code, "nudge must never fail the build")
	assert.Contains(t, out, nudgeMarker, "should nudge when upstream VERSION is newer.\nOutput:\n%s", out)
	assert.Contains(t, out, "task ckeletin:update", "should point at the update command")
	assert.NotContains(t, out, "commit(s) behind", "the notice compares versions, not commits.\nOutput:\n%s", out)
}

func TestUpdateNudge_IgnoresNonVersionCommit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamFile(t, upstream)

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "a framework commit that does not bump VERSION is not an update.\nOutput:\n%s", out)
}

func TestUpdateNudge_LocalAhead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	writeVersion(t, downstream, "0.3.0")
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local ahead")
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "a local version ahead of upstream is not an update.\nOutput:\n%s", out)
}

func TestUpdateNudge_NumericPatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	writeVersion(t, downstream, "0.2.9")
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local 0.2.9")
	advanceUpstreamVersion(t, upstream, "0.2.10")

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.Contains(t, out, nudgeMarker, "0.2.10 is newer than 0.2.9.\nOutput:\n%s", out)
}

func TestUpdateNudge_PrereleaseIsOlder(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	writeVersion(t, downstream, "1.0.0-rc.1")
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local rc")
	advanceUpstreamVersion(t, upstream, "1.0.0")

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.Contains(t, out, nudgeMarker, "1.0.0 is newer than 1.0.0-rc.1.\nOutput:\n%s", out)
}

func TestUpdateNudge_SkipsInCI(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, map[string]string{"CI": "true"})
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent in CI.\nOutput:\n%s", out)
}

func TestUpdateNudge_OptOut(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, map[string]string{"CKELETIN_SKIP_UPDATE_CHECK": "1"})
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must honour the opt-out.\nOutput:\n%s", out)
}

func TestUpdateNudge_UpToDate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	_, downstream := setupDownstream(t)

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent when versions match.\nOutput:\n%s", out)
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
	writeVersion(t, dir, "0.1.0")
	nudgeRunGit(t, dir, "add", "-A")
	nudgeRunGit(t, dir, "commit", "-m", "init")

	out, code := runNudge(t, dir, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "must be silent without an upstream remote.\nOutput:\n%s", out)
}

func TestUpdateNudge_Offline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	dir := t.TempDir()
	nudgeRunGit(t, dir, "init", "-b", "main")
	nudgeRunGit(t, dir, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, dir, "config", "user.name", "Test User")
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/test/down\n"), 0o644))
	writeVersion(t, dir, "0.1.0")
	nudgeRunGit(t, dir, "add", "-A")
	nudgeRunGit(t, dir, "commit", "-m", "init")
	nudgeRunGit(t, dir, "remote", "add", "ckeletin-upstream", filepath.Join(t.TempDir(), "missing.git"))

	out, code := runNudge(t, dir, nil)
	assert.Equal(t, 0, code, "offline notice must not fail the build.\nOutput:\n%s", out)
	assert.NotContains(t, out, nudgeMarker)
}

func TestUpdateNudge_JSON(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, nil, "--json")
	assert.Equal(t, 0, code, "json check must not fail.\nOutput:\n%s", out)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &body), "json check output:\n%s", out)
	assert.Equal(t, "0.1.0", body["current"])
	assert.Equal(t, "0.2.0", body["latest"])
	assert.Equal(t, true, body["update_available"])
}

func TestUpdateNudge_JSONLocalAhead(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	writeVersion(t, downstream, "0.3.0")
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local ahead")
	advanceUpstreamVersion(t, upstream, "0.2.0")

	out, code := runNudge(t, downstream, map[string]string{"CI": "true", "CKELETIN_SKIP_UPDATE_CHECK": "1"}, "--json")
	assert.Equal(t, 0, code)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &body), "json check output:\n%s", out)
	assert.Equal(t, false, body["update_available"], "explicit json check still runs in CI and when the notice is opted out")
	assert.Equal(t, "0.3.0", body["current"])
	assert.Equal(t, "0.2.0", body["latest"])
}

func TestUpdateNudge_UpstreamExempt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "9.9.9")
	// Split the binary name. task init rewrites "ckeletin-go" in Go files, and
	// this test has to keep naming the real upstream module after that.
	upstreamModule := "github.com/peiman/" + "ckeletin" + "-go"
	require.NoError(t, os.WriteFile(filepath.Join(downstream, "go.mod"), []byte("module "+upstreamModule+"\n"), 0o644))

	out, code := runNudge(t, downstream, nil)
	assert.Equal(t, 0, code)
	assert.NotContains(t, out, nudgeMarker, "the upstream repository is exempt.\nOutput:\n%s", out)

	jsonOut, jsonCode := runNudge(t, downstream, nil, "--json")
	assert.Equal(t, 0, jsonCode)
	var body map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(jsonOut)), &body), "json:\n%s", jsonOut)
	assert.Equal(t, false, body["applicable"])
}

func TestUpdateNudge_IgnoresWindowsTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping update-nudge integration test in short mode")
	}
	upstream, downstream := setupDownstream(t)
	advanceUpstreamVersion(t, upstream, "0.2.0")

	// Windows timeout.exe is not GNU timeout. It accepts "timeout 5" and does
	// not run the following command, so a fetch wrapped in it never happens.
	bin := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(bin, "timeout"), []byte("#!/bin/sh\nexit 1\n"), 0o755))

	out, code := runNudge(t, downstream, map[string]string{"PATH": bin + string(os.PathListSeparator) + os.Getenv("PATH")})
	assert.Equal(t, 0, code, "nudge must never fail the build.\nOutput:\n%s", out)
	assert.Contains(t, out, nudgeMarker, "a non-GNU timeout on PATH must not replace git fetch.\nOutput:\n%s", out)
}

func TestUpdateNudge_WiredIntoCheck(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	taskfile, err := os.ReadFile(filepath.Join(root, ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)
	text := string(taskfile)
	assert.Contains(t, text, "check-framework-updates.sh")
	assert.Contains(t, text, "CKELETIN_UPDATE_RESULT")
}
