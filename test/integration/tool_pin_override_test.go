package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGolangciLintPinDefaultAndOverride(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))

	defaultOut := drySetupPinned(t, root)
	assert.Contains(t, defaultOut, "golangci-lint@v2.15.0")

	overrideOut := drySetupPinned(t, root, "CKELETIN_TOOL_GOLANGCI_LINT_VERSION=v9.9.9")
	assert.Contains(t, overrideOut, "golangci-lint@v9.9.9")
	assert.NotContains(t, overrideOut, "golangci-lint@v2.15.0")
}

func drySetupPinned(t *testing.T, root string, extraArgs ...string) string {
	t.Helper()
	args := append([]string{"--dry", "ckeletin:setup:pinned"}, extraArgs...)
	cmd := exec.Command("task", args...)
	cmd.Dir = root
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "task --dry ckeletin:setup:pinned failed:\n%s", out)
	return string(out)
}
