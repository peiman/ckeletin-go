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

// govulncheck exits 3 when it finds vulnerabilities. A findings trace can
// contain the substring "timeout" (http2bufferedWriterTimeoutWriter). The
// pre-commit and dependency checks must treat exit 3 as findings, and only
// treat a real network failure (a non-3 exit) as a network problem.

func writeExec(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o755))
}

func vulnCheckEnv(t *testing.T, bin string) []string {
	t.Helper()
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "SKIP_VULN_CHECK=") || strings.HasPrefix(kv, "PATH=") || strings.HasPrefix(kv, "TMPDIR=") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env,
		"PATH="+bin+string(os.PathListSeparator)+os.Getenv("PATH"),
		"TMPDIR="+t.TempDir(),
	)
	return env
}

func runVulnScript(t *testing.T, script, bin string) (string, error) {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	cmd := exec.Command("bash", filepath.Join(root, ".ckeletin", "scripts", script))
	cmd.Env = vulnCheckEnv(t, bin)
	cmd.Dir = t.TempDir()
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func TestVulnFast_FindingsTraceContainingTimeoutFails(t *testing.T) {
	bin := t.TempDir()
	writeExec(t, bin, "govulncheck", `#!/bin/sh
echo "Vulnerability #1: GO-2026-6603"
echo "net/http.http2bufferedWriterTimeoutWriter.Write"
exit 3
`)

	out, err := runVulnScript(t, "check-vuln-fast.sh", bin)
	require.Error(t, err, "findings must fail the pre-commit scan\n%s", out)
	assert.NotContains(t, out, "network unavailable")
	assert.Contains(t, out, "Security vulnerabilities detected")
}

func TestVulnFast_NetworkErrorIsASkip(t *testing.T) {
	bin := t.TempDir()
	writeExec(t, bin, "govulncheck", `#!/bin/sh
echo "govulncheck: dial tcp: lookup vuln.go.dev: no such host" >&2
exit 1
`)

	out, err := runVulnScript(t, "check-vuln-fast.sh", bin)
	require.NoError(t, err, "a network failure stays a skip on the fast path\n%s", out)
	assert.Contains(t, out, "network unavailable")
}

func TestCheckDeps_FindingsAreNotLabeledNetwork(t *testing.T) {
	bin := t.TempDir()
	writeExec(t, bin, "go", "#!/bin/sh\nexit 0\n")
	writeExec(t, bin, "go-mod-outdated", "#!/bin/sh\nexit 0\n")
	writeExec(t, bin, "govulncheck", `#!/bin/sh
echo "Vulnerability #1: GO-2026-6603"
echo "net/http.http2bufferedWriterTimeoutWriter.Write"
exit 3
`)

	out, err := runVulnScript(t, "check-deps.sh", bin)
	require.Error(t, err, "findings must fail the dependency check\n%s", out)
	assert.NotContains(t, out, "network error")
	assert.Contains(t, out, "Security vulnerabilities found")
}

func TestCheckDeps_NetworkErrorKeepsItsLabel(t *testing.T) {
	bin := t.TempDir()
	writeExec(t, bin, "go", "#!/bin/sh\nexit 0\n")
	writeExec(t, bin, "go-mod-outdated", "#!/bin/sh\nexit 0\n")
	writeExec(t, bin, "govulncheck", `#!/bin/sh
echo "govulncheck: dial tcp: lookup vuln.go.dev: no such host" >&2
exit 1
`)

	out, err := runVulnScript(t, "check-deps.sh", bin)
	require.Error(t, err, "a network failure still fails the full dependency check\n%s", out)
	assert.Contains(t, out, "network error")
	assert.NotContains(t, out, "Security vulnerabilities found")
}
