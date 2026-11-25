package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func outputPatternScript(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs(filepath.Join("..", "..", ".ckeletin", "scripts", "validate-output-patterns.sh"))
	require.NoError(t, err)
	return abs
}

func TestOutputAllowlist_Adversarial(t *testing.T) {
	script := outputPatternScript(t)
	offender := "package bad\n\nimport \"fmt\"\n\nfunc speak() { fmt.Println(\"no\") }\n"

	t.Run("missing allowlist", func(t *testing.T) {
		dir := t.TempDir()
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code)
		assert.Contains(t, out, "Output allowlist is missing")
		assert.NotContains(t, out, "malformed")
		assertNoTraceback(t, out)
	})

	t.Run("malformed allowlist", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"), []byte("internal/bad/bad.go\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code)
		assert.Contains(t, out, "Output allowlist is malformed")
		assert.Contains(t, out, ".ckeletin-output-allowlist:1")
		assert.NotContains(t, out, "Output allowlist is missing")
		assertNoTraceback(t, out)
	})

	t.Run("new print fails", func(t *testing.T) {
		dir := fixtureWithPrint(t, offender)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"), []byte("# none\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "fmt.Println")
		assert.Contains(t, out, "progress.Reporter")
		assertNoTraceback(t, out)
	})

	t.Run("listed path passes", func(t *testing.T) {
		dir := fixtureWithPrint(t, offender)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/bad/bad.go count=1 full-screen TUI writes the frame\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.Equal(t, 0, code, out)
		assert.Contains(t, out, "compliant")
	})

	t.Run("app ui package is scanned", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "ui"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "ui", "x.go"), []byte(offender), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"), []byte("# none\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "internal/ui/x.go")
		assert.Contains(t, out, "fmt.Println")
	})

	t.Run("extra line in allowlisted file fails", func(t *testing.T) {
		two := "package bad\n\nimport \"fmt\"\n\nfunc speak() {\n\tfmt.Println(\"no\")\n\tfmt.Println(\"again\")\n}\n"
		dir := fixtureWithPrint(t, two)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/bad/bad.go count=1 full-screen TUI writes the frame\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "expected 1 direct output lines, found 2")
		assert.Contains(t, out, "internal/bad/bad.go")
	})

	t.Run("pinned count passes", func(t *testing.T) {
		dir := fixtureWithPrint(t, offender)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/bad/bad.go count=1 full-screen TUI writes the frame\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.Equal(t, 0, code, out)
	})

	t.Run("unpinned entry fails", func(t *testing.T) {
		dir := fixtureWithPrint(t, offender)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/bad/bad.go full-screen TUI writes the frame\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, ".ckeletin-output-allowlist:1")
		assert.Contains(t, out, "count= or lines=")
	})

	t.Run("unpinned quiet file fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "quiet"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "quiet", "x.go"), []byte("package quiet\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/quiet/x.go a reason\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, ".ckeletin-output-allowlist:1")
		assert.Contains(t, out, "count= or lines=")
	})

	t.Run("stale count on a quiet file fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "quiet"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "quiet", "x.go"), []byte("package quiet\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/quiet/x.go count=1 was a print\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "expected 1 direct output lines, found 0")
	})

	t.Run("stale lines on a quiet file fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "quiet"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "quiet", "x.go"), []byte("package quiet\n"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/quiet/x.go lines=1 was a print\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "expected output lines 1, found none")
	})

	t.Run("deleted file fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/gone.go count=0 removed\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "internal/gone.go")
		assert.Contains(t, out, "does not exist")
	})

	t.Run("directory entry fails", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "ui"), 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/ui count=1 a whole package\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, ".ckeletin-output-allowlist:1")
		assert.Contains(t, out, "not a file; pin each file")
	})

	t.Run("pinned lines fail when the line moves", func(t *testing.T) {
		dir := fixtureWithPrint(t, offender)
		require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin-output-allowlist"),
			[]byte("internal/bad/bad.go lines=1 full-screen TUI writes the frame\n"), 0o644))
		out, code := runAllowlist(t, script, dir)
		assert.NotEqual(t, 0, code, out)
		assert.Contains(t, out, "expected output lines 1")
	})
}

func fixtureWithPrint(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "internal", "bad"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "internal", "bad", "bad.go"), []byte(source), 0o644))
	return dir
}

func runAllowlist(t *testing.T, script, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	code := 0
	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		require.True(t, ok, "script error: %v\n%s", err, out)
		code = exitErr.ExitCode()
	}
	return string(out), code
}

func assertNoTraceback(t *testing.T, out string) {
	t.Helper()
	assert.NotContains(t, out, "Traceback")
	assert.NotContains(t, out, "panic:")
	for _, line := range strings.Split(out, "\n") {
		assert.NotRegexp(t, `^bash: line [0-9]+:`, line)
	}
}
