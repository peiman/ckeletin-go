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

// .go-version is the minimum toolchain CI installs. A newer local Go still
// has those fixes, so task check must accept it. An older Go must fail.

func runGoVersionCheck(t *testing.T, pin, reported string, args ...string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".go-version"), []byte(pin+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/fixture\n\ngo "+pin+"\n"), 0o644))

	bin := t.TempDir()
	writeExec(t, bin, "go", "#!/bin/sh\necho \"go version go"+reported+" darwin/arm64\"\n")

	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	cmd := exec.Command("bash", append([]string{filepath.Join(root, ".ckeletin", "scripts", "check-go-version.sh")}, args...)...)
	cmd.Dir = dir
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "PATH=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env, "PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestGoVersion_NewerThanPinPasses(t *testing.T) {
	out, err := runGoVersionCheck(t, "1.26.9", "1.27.2")
	require.NoError(t, err, "a newer toolchain meets the pin\n%s", out)
	assert.Contains(t, out, "Go version OK")
	assert.Contains(t, out, "1.27.2")
}

func TestGoVersion_ExactPinPasses(t *testing.T) {
	out, err := runGoVersionCheck(t, "1.26.9", "1.26.9")
	require.NoError(t, err, "the pinned toolchain passes\n%s", out)
	assert.Contains(t, out, "Go version OK")
}

func TestGoVersion_OlderThanPinFails(t *testing.T) {
	out, err := runGoVersionCheck(t, "1.26.9", "1.26.4")
	require.Error(t, err, "an older toolchain must fail\n%s", out)
	assert.NotContains(t, out, "Go version OK")
	assert.NotContains(t, out, "Update .go-version to 1.26.4")
	assert.Contains(t, out, "1.26.9")
}

func TestGoVersion_MinorFlagAllowsAnotherPatchOfTheSameMinor(t *testing.T) {
	out, err := runGoVersionCheck(t, "1.26.9", "1.26.4", "--minor")
	require.NoError(t, err, "--minor matches the minor and ignores the patch\n%s", out)
	assert.Contains(t, out, "Go version OK")
}
