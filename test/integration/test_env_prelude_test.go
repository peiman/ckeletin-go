package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// teProjectRoot returns the absolute project root.
func teProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	return root
}

func teFrameworkTaskfile(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(teProjectRoot(t), ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err, "failed to read framework Taskfile")
	return string(b)
}

// extractPrelude returns the actual _TEST_ENV_PRELUDE var value from the
// framework Taskfile, so the mechanism test exercises the real string (no
// duplication / drift).
func extractPrelude(t *testing.T) string {
	t.Helper()
	for line := range strings.SplitSeq(teFrameworkTaskfile(t), "\n") {
		s := strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(s, "_TEST_ENV_PRELUDE:"); ok {
			v := strings.TrimSpace(after)
			v = strings.TrimPrefix(v, "'")
			v = strings.TrimSuffix(v, "'")
			return v
		}
	}
	t.Fatal("_TEST_ENV_PRELUDE not found in framework Taskfile")
	return ""
}

// TestTestEnvPrelude_SourcesFileWhenPresent asserts the real prelude string
// sources .ckeletin.test-env.sh (in the same shell) when it exists — verified
// in POSIX sh, which is what Task uses.
func TestTestEnvPrelude_SourcesFileWhenPresent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test-env prelude integration test in short mode")
	}
	prelude := extractPrelude(t)
	dir := t.TempDir()
	marker := filepath.Join(dir, "sourced.marker")
	envFile := "export CKELETIN_TEST_ENV_PROBE=active\ntouch " + marker + "\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin.test-env.sh"), []byte(envFile), 0o644))

	cmd := exec.Command("sh", "-c", prelude+` printf 'PROBE=%s\n' "$CKELETIN_TEST_ENV_PROBE"`)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "prelude run failed.\nOutput:\n%s", out)
	assert.FileExists(t, marker, "the test-env file must be sourced (marker created)")
	assert.Contains(t, string(out), "PROBE=active", "sourced exports must be visible to the test shell")
}

// TestTestEnvPrelude_NoOpWhenAbsent asserts the prelude is a clean no-op (exit 0,
// no error) when no test-env file is present — backward compatibility.
func TestTestEnvPrelude_NoOpWhenAbsent(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test-env prelude integration test in short mode")
	}
	prelude := extractPrelude(t)
	dir := t.TempDir() // no .ckeletin.test-env.sh here

	cmd := exec.Command("sh", "-c", prelude+` echo ran`)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "prelude must not fail when the file is absent.\nOutput:\n%s", out)
	assert.Contains(t, string(out), "ran")
}

// TestTestEnvPrelude_AllTestTasksWired is the anti-rot guard (issue #95): every
// test:* task that invokes `go test`/`gotestsum` directly MUST include the
// prelude, so a newly-added sibling cannot silently lose the consumer's test-env
// isolation. Tasks that delegate (task: test / deps: [test]) are exempt.
func TestTestEnvPrelude_AllTestTasksWired(t *testing.T) {
	content := teFrameworkTaskfile(t)
	lines := strings.Split(content, "\n")
	taskHeader := regexp.MustCompile(`^  ([a-zA-Z][a-zA-Z0-9:_-]*):\s*$`)

	var curName string
	var body []string
	var offenders []string

	flush := func() {
		if curName == "" {
			return
		}
		joined := strings.Join(body, "\n")
		runsGoTest := strings.Contains(joined, "gotestsum ") || strings.Contains(joined, "go test ")
		inScope := strings.HasPrefix(curName, "test") || strings.HasPrefix(curName, "bench")
		if inScope && runsGoTest && !strings.Contains(joined, "_TEST_ENV_PRELUDE") {
			offenders = append(offenders, curName)
		}
	}

	for _, line := range lines {
		if m := taskHeader.FindStringSubmatch(line); m != nil {
			flush()
			curName = m[1]
			body = nil
			continue
		}
		body = append(body, line)
	}
	flush()

	assert.Empty(t, offenders,
		"these test:* tasks run go test but are missing {{._TEST_ENV_PRELUDE}} "+
			"(consumers would lose test-env isolation on them — issue #95): %v", offenders)
}
