package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// check-sbom-vulns.sh scans the newest CycloneDX file under reports/sbom.
// That directory is gitignored, so a months-old local file keeps failing
// Grype after the toolchain and modules have moved on. A file older than
// the module must be regenerated before the scan. A current file is reused.

const syftStub = `#!/bin/sh
touch "$SBOM_SYFT_MARKER"
for a in "$@"; do
  case "$a" in
    cyclonedx-json=*|cyclonedx-xml=*|spdx-json=*|spdx=*)
      printf '{}\n' > "${a#*=}"
      ;;
  esac
done
exit 0
`

const grypeStub = `#!/bin/sh
printf '%s\n' "$@" > "$SBOM_GRYPE_ARGS"
exit 0
`

const goBuildStub = `#!/bin/sh
if [ "$1" = "build" ]; then
  prev=
  for a in "$@"; do
    if [ "$prev" = "-o" ]; then
      : > "$a"
    fi
    prev=$a
  done
fi
exit 0
`

func runSBOMCheck(t *testing.T, dir string) (string, string, bool, error) {
	t.Helper()
	bin := t.TempDir()
	argsFile := filepath.Join(bin, "grype-args")
	marker := filepath.Join(bin, "syft-called")
	writeExec(t, bin, "syft", syftStub)
	writeExec(t, bin, "grype", grypeStub)
	writeExec(t, bin, "go", goBuildStub)

	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	cmd := exec.Command("bash", filepath.Join(root, ".ckeletin", "scripts", "check-sbom-vulns.sh"))
	cmd.Dir = dir
	var env []string
	for _, kv := range hermeticGitEnv() {
		if strings.HasPrefix(kv, "PATH=") || strings.HasPrefix(kv, "SBOM_GRYPE_ARGS=") || strings.HasPrefix(kv, "SBOM_SYFT_MARKER=") {
			continue
		}
		env = append(env, kv)
	}
	cmd.Env = append(env,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"SBOM_GRYPE_ARGS="+argsFile,
		"SBOM_SYFT_MARKER="+marker,
	)

	out, err := cmd.CombinedOutput()
	args, readErr := os.ReadFile(argsFile)
	require.NoError(t, readErr, "grype was not invoked\n%s", out)
	_, statErr := os.Stat(marker)
	return string(out), string(args), statErr == nil, err
}

func TestSBOMCheck_RegeneratesWhenOlderThanTheModule(t *testing.T) {
	dir := t.TempDir()
	sbomDir := filepath.Join(dir, "reports", "sbom")
	require.NoError(t, os.MkdirAll(sbomDir, 0o755))
	stale := filepath.Join(sbomDir, "stale.cdx.json")
	require.NoError(t, os.WriteFile(stale, []byte("{}\n"), 0o644))
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(stale, past, past))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/fixture\n\ngo 1.26.9\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.sum"), []byte(""), 0o644))

	out, args, syftCalled, err := runSBOMCheck(t, dir)
	require.NoError(t, err, "a regenerated SBOM should scan clean\n%s", out)
	assert.True(t, syftCalled, "an SBOM older than the module must be regenerated")
	assert.NotContains(t, args, "stale.cdx.json")
	assert.Contains(t, args, "app-")
	assert.Contains(t, args, ".cdx.json")
}

func TestSBOMCheck_ReusesAnSBOMNewerThanTheModule(t *testing.T) {
	dir := t.TempDir()
	sbomDir := filepath.Join(dir, "reports", "sbom")
	require.NoError(t, os.MkdirAll(sbomDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/fixture\n\ngo 1.26.9\n"), 0o644))
	past := time.Now().Add(-2 * time.Hour)
	require.NoError(t, os.Chtimes(filepath.Join(dir, "go.mod"), past, past))
	current := filepath.Join(sbomDir, "current.cdx.json")
	require.NoError(t, os.WriteFile(current, []byte("{}\n"), 0o644))

	out, args, syftCalled, err := runSBOMCheck(t, dir)
	require.NoError(t, err, "a current SBOM should be scanned as it is\n%s", out)
	assert.Contains(t, args, "current.cdx.json")
	assert.False(t, syftCalled, "a current SBOM must not be regenerated")
}
