package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hfProjectRoot returns the absolute project root (two levels up).
func hfProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	return root
}

// hfScriptPath returns the absolute path to validate-hook-forwardings.sh.
func hfScriptPath(t *testing.T) string {
	t.Helper()
	p := filepath.Join(hfProjectRoot(t), ".ckeletin", "scripts", "validate-hook-forwardings.sh")
	require.FileExists(t, p, "validate-hook-forwardings script must exist")
	return p
}

// runHookForwardings runs the validator in dir. Extra "KEY=value" env entries
// (e.g. CKELETIN_EXPECTED_FORWARDINGS) may be supplied to keep tests hermetic.
func runHookForwardings(t *testing.T, dir string, env ...string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", hfScriptPath(t))
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "validator errored.\nOutput:\n%s", string(out))
	return string(out), 0
}

// TestHookForwardings_PassesOnFramework guards that ckeletin-go's own base hooks
// reference only namespaced or forwarded tasks (issue #100). Exit 0.
func TestHookForwardings_PassesOnFramework(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hook-forwardings integration test in short mode")
	}
	out, code := runHookForwardings(t, hfProjectRoot(t))
	assert.Equal(t, 0, code, "framework base hooks must be consistent.\nOutput:\n%s", out)
	assert.Contains(t, out, "namespaced or forwarded")
}

// TestHookForwardings_DetectsUnforwardedBareTask asserts a base hook that runs a
// bare task with no forwarding (the exact #100 breakage: `task test:scaffold`)
// is flagged. The script reads the real expected-forwardings.txt, so a temp
// base.yml referencing a task absent from it must fail.
func TestHookForwardings_DetectsUnforwardedBareTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hook-forwardings integration test in short mode")
	}
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".ckeletin", "configs"), 0o755))
	base := "pre-push:\n  commands:\n    scaffold:\n      run: task test:scaffold\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "configs", "lefthook.base.yml"), []byte(base), 0o644))

	// Hermetic forwarding list that deliberately omits test:scaffold.
	expected := filepath.Join(dir, "expected.txt")
	require.NoError(t, os.WriteFile(expected, []byte("# forwardings\nlint\ntest:coverage\n"), 0o644))

	out, code := runHookForwardings(t, dir, "CKELETIN_EXPECTED_FORWARDINGS="+expected)
	assert.Equal(t, 1, code, "an unforwarded bare task reference must fail.\nOutput:\n%s", out)
	assert.Contains(t, out, "test:scaffold", "should name the offending task")
	assert.Contains(t, out, "no forwarding", "should explain the failure")
}

// TestHookForwardings_AllowsNamespacedTask asserts a base hook that runs a
// namespaced `ckeletin:*` task is accepted (it resolves in any consumer).
func TestHookForwardings_AllowsNamespacedTask(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping hook-forwardings integration test in short mode")
	}
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".ckeletin", "configs"), 0o755))
	base := "pre-push:\n  commands:\n    scaffold:\n      run: task ckeletin:test:scaffold\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "configs", "lefthook.base.yml"), []byte(base), 0o644))

	out, code := runHookForwardings(t, dir)
	assert.Equal(t, 0, code, "a namespaced task reference must pass.\nOutput:\n%s", out)
}
