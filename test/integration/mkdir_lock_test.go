package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func mkdirLockScript(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	script := filepath.Join(root, ".ckeletin", "scripts", "check-mkdir-lock.sh")
	require.FileExists(t, script)
	return script
}

func runMkdirLock(t *testing.T, dir string, files ...string) (string, int) {
	t.Helper()
	args := append([]string{mkdirLockScript(t)}, files...)
	cmd := exec.Command("bash", args...)
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv()
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "mkdir lock check errored\n%s", out)
	return string(out), 0
}

func TestMkdirLockCheckPassesOnDirectoryCreate(t *testing.T) {
	dir := t.TempDir()
	ok := filepath.Join(dir, "ok.sh")
	body := "#!/bin/bash\nmkdir -p reports\nmkdir -p \"$OUTPUT_DIR\"\n# until mkdir \"$LOCK\" is the bug, not this comment\n"
	require.NoError(t, os.WriteFile(ok, []byte(body), 0o644))

	out, code := runMkdirLock(t, dir, ok)
	require.Equal(t, 0, code, out)
	require.Contains(t, out, "no mkdir lock")
}

func TestMkdirLockCheckFlagsMkdirUsedAsLock(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.sh")
	body := "#!/bin/bash\nuntil mkdir \"$LOCK\"; do sleep 1; done\nif mkdir \"$LOCK\"; then echo held; fi\nmkdir \"$LOCK\" || exit 1\nmkdir -p \"$lockdir\"\n"
	require.NoError(t, os.WriteFile(bad, []byte(body), 0o644))

	out, code := runMkdirLock(t, dir, bad)
	require.NotEqual(t, 0, code, out)
	require.Contains(t, out, "set -C")
	require.Contains(t, out, "lock condition")
	require.Contains(t, out, "mkdir failure is used as a lock")
	require.Contains(t, out, "path named like a lock")
}

func TestMkdirLockCheckPassesOnThisRepo(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	out, code := runMkdirLock(t, root)
	require.Equal(t, 0, code, out)
}
