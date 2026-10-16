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

// frameworkNames builds the upstream identity from parts. A contiguous literal
// in this file would be rewritten when a derived project copies the fixture.
func frameworkNames() (module, name, envPrefix string) {
	name = "ckeletin" + "-go"
	module = "github.com/peiman/" + name
	envPrefix = "CKELETIN" + "_GO"
	return module, name, envPrefix
}

func localDiffRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	return root
}

func localDiffEnv() []string {
	var env []string
	for _, kv := range hermeticGitEnv() {
		if strings.HasPrefix(kv, "GOWORK=") {
			continue
		}
		env = append(env, kv)
	}
	return env
}

func copyRewriters(t *testing.T, ckeletinDir string) {
	t.Helper()
	src := filepath.Join(localDiffRoot(t), ".ckeletin", "scripts")
	dst := filepath.Join(ckeletinDir, "scripts", "rewrite-imports")
	require.NoError(t, os.MkdirAll(dst, 0o755))
	for _, name := range []string{"main.go", "main_test.go"} {
		body, err := os.ReadFile(filepath.Join(src, "rewrite-imports", name))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dst, name), body, 0o644))
	}
	body, err := os.ReadFile(filepath.Join(src, "rewrite-derived-names.sh"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(
		filepath.Join(ckeletinDir, "scripts", "rewrite-derived-names.sh"), body, 0o755))
}

func applyDerivedRewrites(t *testing.T, dir, oldModule, newModule, oldName, newName string) {
	t.Helper()
	cmd := exec.Command("go", "run", "./.ckeletin/scripts/rewrite-imports/",
		"-old", oldModule, "-new", newModule, "-dir", ".ckeletin")
	cmd.Dir = dir
	cmd.Env = localDiffEnv()
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "rewrite-imports failed.\n%s", out)

	cmd = exec.Command("bash", "./.ckeletin/scripts/rewrite-derived-names.sh",
		oldModule, newModule, oldName, newName, ".ckeletin")
	cmd.Dir = dir
	cmd.Env = localDiffEnv()
	out, err = cmd.CombinedOutput()
	require.NoError(t, err, "rewrite-derived-names failed.\n%s", out)
}

// setupDerived builds an upstream framework tree and a derived clone whose
// .ckeletin/ differs only by the module path and binary name the update rewrites.
func setupDerived(t *testing.T) (upstream, downstream string) {
	t.Helper()
	module, name, envPrefix := frameworkNames()
	derivedModule := "github.com/test/down"
	derivedName := "down"

	upstream = t.TempDir()
	downstream = t.TempDir()

	nudgeRunGit(t, upstream, "init", "-b", "main")
	nudgeRunGit(t, upstream, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, upstream, "config", "user.name", "Test User")

	ckeletin := filepath.Join(upstream, ".ckeletin")
	writeVersion(t, upstream, "0.2.0")
	require.NoError(t, os.WriteFile(filepath.Join(ckeletin, "note.md"), []byte(
		"module "+module+"\nbinary "+name+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(ckeletin, "sample.go"), []byte(
		"package sample\n\n"+
			"import \""+module+"/.ckeletin/pkg/config\"\n\n"+
			"// binary "+name+"\n"+
			"const Name = \""+name+"\"\n"+
			"const Env = \""+envPrefix+"\"\n"), 0o644))
	copyRewriters(t, ckeletin)
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "go.mod"), []byte(
		"module "+module+"\n\ngo 1.26.9\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, "Taskfile.yml"), []byte(
		"version: '3'\n\nvars:\n  BINARY_NAME: "+name+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(upstream, ".gitignore"), []byte(
		".ckeletin/cache/\n.DS_Store\n"), 0o644))
	nudgeRunGit(t, upstream, "add", "-A")
	nudgeRunGit(t, upstream, "commit", "-m", "framework 0.2.0")

	clone := gitCmd(filepath.Dir(downstream), "clone", "--quiet", upstream, downstream)
	out, err := clone.CombinedOutput()
	require.NoError(t, err, "git clone failed.\n%s", out)
	nudgeRunGit(t, downstream, "config", "user.email", "test@ckeletin-go.example")
	nudgeRunGit(t, downstream, "config", "user.name", "Test User")
	nudgeRunGit(t, downstream, "remote", "rename", "origin", "ckeletin-upstream")

	require.NoError(t, os.WriteFile(filepath.Join(downstream, "go.mod"), []byte(
		"module "+derivedModule+"\n\ngo 1.26.9\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(downstream, "Taskfile.yml"), []byte(
		"version: '3'\n\nvars:\n  BINARY_NAME: "+derivedName+"\n"), 0o644))
	applyDerivedRewrites(t, downstream, module, derivedModule, name, derivedName)
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "derived names")
	return upstream, downstream
}

func runLocalDiff(t *testing.T, dir string, args ...string) (string, int) {
	t.Helper()
	script := filepath.Join(localDiffRoot(t), ".ckeletin", "scripts", "framework-local-diff.sh")
	require.FileExists(t, script)
	cmd := exec.Command("bash", append([]string{script}, args...)...)
	cmd.Dir = dir
	cmd.Env = localDiffEnv()
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "local-diff guard errored.\n%s", out)
	return string(out), 0
}

func gitPorcelain(t *testing.T, dir string) string {
	t.Helper()
	cmd := gitCmd(dir, "status", "--porcelain")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	return string(out)
}

func gitHead(t *testing.T, dir string) string {
	t.Helper()
	cmd := gitCmd(dir, "rev-parse", "HEAD")
	out, err := cmd.Output()
	require.NoError(t, err, string(out))
	return strings.TrimSpace(string(out))
}

func localDiffResult(t *testing.T, out string) map[string]any {
	t.Helper()
	const prefix = "CKELETIN_UPDATE_RESULT="
	var raw string
	count := 0
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, prefix) {
			raw = strings.TrimPrefix(line, prefix)
			count++
		}
	}
	require.Equal(t, 1, count, "expected one result line.\n%s", out)
	var got map[string]any
	require.NoError(t, json.Unmarshal([]byte(raw), &got), raw)
	return got
}

func assertStopped(t *testing.T, got map[string]any, reason, from string) {
	t.Helper()
	assert.Equal(t, "stopped", got["status"])
	assert.Equal(t, reason, got["reason"])
	assert.Equal(t, from, got["from"])
	assert.Nil(t, got["to"])
	assert.Equal(t, false, got["committed"])
	assert.Equal(t, false, got["discarded"])
}

func TestFrameworkLocalDiff_UncommittedStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	note := filepath.Join(downstream, ".ckeletin", "note.md")
	require.NoError(t, os.WriteFile(note, []byte("LOCALEDIT\n"), 0o644))
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("untracked\n"), 0o644))

	head := gitHead(t, downstream)
	before := gitPorcelain(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update")
	assert.Equal(t, 1, code, out)
	assertStopped(t, localDiffResult(t, out), "uncommitted", "0.2.0")
	assert.Contains(t, out, "note.md")
	assert.Contains(t, out, "extra.txt")
	assert.Contains(t, out, "LOCALEDIT")
	assert.Equal(t, head, gitHead(t, downstream))
	assert.Equal(t, before, gitPorcelain(t, downstream))
}

func TestFrameworkLocalDiff_UncommittedStopsEvenWhenDiscard(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("untracked\n"), 0o644))

	before := gitPorcelain(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update", "--discard", "1")
	assert.Equal(t, 1, code, out)
	assertStopped(t, localDiffResult(t, out), "uncommitted", "0.2.0")
	assert.NotContains(t, out, "Discarding local framework changes")
	assert.Equal(t, before, gitPorcelain(t, downstream))
	assert.FileExists(t, extra)
}

func TestFrameworkLocalDiff_CommittedExtraFileStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("local\n"), 0o644))
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local framework file")

	head := gitHead(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update")
	assert.Equal(t, 1, code, out)
	assertStopped(t, localDiffResult(t, out), "local_difference", "0.2.0")
	assert.Contains(t, out, "extra.txt")
	assert.NotContains(t, out, "note.md")
	assert.NotContains(t, out, "sample.go")
	assert.Equal(t, head, gitHead(t, downstream), "a stop must not commit or checkout")
	assert.Empty(t, gitPorcelain(t, downstream))
	assert.FileExists(t, extra)
}

func TestFrameworkLocalDiff_DiscardExtraFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("local\n"), 0o644))
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local framework file")

	result := filepath.Join(t.TempDir(), "result")
	head := gitHead(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update", "--discard", "1", "--result", result)
	assert.Equal(t, 0, code, out)
	assert.Contains(t, out, "Discarding local framework changes")
	assert.NotContains(t, out, "CKELETIN_UPDATE_RESULT=")
	body, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Equal(t, "discarded=true\n", string(body))
	assert.Equal(t, head, gitHead(t, downstream), "the guard allows the update; it does not checkout itself")
	assert.FileExists(t, extra)

	// checkout alone keeps files that exist only locally. The update removes
	// the tree first, which is what drops the extra file.
	nudgeRunGit(t, downstream, "rm", "-r", "--quiet", "--", ".ckeletin")
	nudgeRunGit(t, downstream, "checkout", "ckeletin-upstream/main", "--", ".ckeletin/")
	_, statErr := os.Stat(extra)
	assert.True(t, os.IsNotExist(statErr), "discard continues into a replacement that drops the extra file")
}

func TestFrameworkLocalDiff_DryRunShowsExtraFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("local\n"), 0o644))
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local framework file")

	head := gitHead(t, downstream)
	before := gitPorcelain(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "dry-run")
	assert.Equal(t, 0, code, out)
	assert.Contains(t, out, "extra.txt")
	assert.Contains(t, out, "would stop")
	assert.NotContains(t, out, "CKELETIN_UPDATE_RESULT=")
	assert.Equal(t, head, gitHead(t, downstream))
	assert.Equal(t, before, gitPorcelain(t, downstream))
	assert.FileExists(t, extra)
}

func TestFrameworkLocalDiff_NameRewritesDoNotStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	upstream, downstream := setupDerived(t)
	// Ignored cache must not look like a local framework edit.
	require.NoError(t, os.MkdirAll(filepath.Join(downstream, ".ckeletin", "cache"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(downstream, ".ckeletin", "cache", "spec.json"), []byte("{}\n"), 0o644))
	advanceUpstreamVersion(t, upstream, "0.3.0")

	result := filepath.Join(t.TempDir(), "result")
	head := gitHead(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update", "--result", result)
	assert.Equal(t, 0, code, "name rewrites are not a local difference.\n%s", out)
	assert.NotContains(t, out, "CKELETIN_UPDATE_RESULT=")
	assert.NotContains(t, out, "would stop")
	body, err := os.ReadFile(result)
	require.NoError(t, err)
	assert.Equal(t, "discarded=false\n", string(body))
	assert.Equal(t, head, gitHead(t, downstream))
	assert.Empty(t, gitPorcelain(t, downstream))

	// Upstream moved. The guard compared the recorded version, not the tip,
	// so the incoming update is not reported as a local difference.
	cmd := gitCmd(downstream, "diff", "--quiet", "HEAD", "ckeletin-upstream/main", "--", ".ckeletin/VERSION")
	diffOut, diffErr := cmd.CombinedOutput()
	require.Error(t, diffErr, "upstream VERSION should differ from the derived tree.\n%s", diffOut)
}

func TestFrameworkLocalDiff_UnknownVersionStops(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	writeVersion(t, downstream, "9.9.9")
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "unknown framework version")

	head := gitHead(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "update")
	assert.Equal(t, 1, code, out)
	assertStopped(t, localDiffResult(t, out), "unknown_version", "9.9.9")
	assert.Equal(t, head, gitHead(t, downstream))
}

func TestFrameworkLocalDiff_CompatibilityDoesNotModify(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping framework local-diff integration test in short mode")
	}
	_, downstream := setupDerived(t)
	extra := filepath.Join(downstream, ".ckeletin", "extra.txt")
	require.NoError(t, os.WriteFile(extra, []byte("local\n"), 0o644))
	nudgeRunGit(t, downstream, "add", "-A")
	nudgeRunGit(t, downstream, "commit", "-m", "local framework file")

	head := gitHead(t, downstream)
	before := gitPorcelain(t, downstream)
	out, code := runLocalDiff(t, downstream, "--mode", "compatibility")
	assert.Equal(t, 1, code, out)
	assertStopped(t, localDiffResult(t, out), "local_difference", "0.2.0")
	assert.Equal(t, head, gitHead(t, downstream))
	assert.Equal(t, before, gitPorcelain(t, downstream))
	assert.FileExists(t, extra)
}

func TestFrameworkLocalDiff_TaskfileCallsGuardBeforeCheckout(t *testing.T) {
	body, err := os.ReadFile(filepath.Join(localDiffRoot(t), ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)
	text := string(body)

	update := taskSection(text, "\n  update:\n", "\n  update:dry-run:\n")
	dry := taskSection(text, "\n  update:dry-run:\n", "\n  update:check-compatibility:\n")
	compat := taskSection(text, "\n  update:check-compatibility:\n", "\n  doctor:\n")

	guard := "framework-local-diff.sh"
	for _, part := range []struct {
		name string
		body string
	}{
		{"update", update},
		{"dry-run", dry},
		{"compatibility", compat},
	} {
		assert.Contains(t, part.body, guard, part.name)
		assert.Less(t, strings.Index(part.body, "UPSTREAM_MODULE"), strings.Index(part.body, guard),
			"%s must refuse the upstream repo before the guard", part.name)
	}
	assert.Less(t, strings.Index(update, guard), strings.Index(update, "git rm -r"),
		"update must run the guard before it deletes the framework tree")
	assert.Less(t, strings.Index(update, "git rm -r"), strings.Index(update, "git checkout"),
		"update must remove the local tree before checkout so extra files do not survive")
	assert.Contains(t, update, "discard_local")
	assert.Contains(t, update, `"discarded":%s`)
	assert.NotContains(t, dry, "git checkout")
	assert.Less(t, strings.Index(dry, guard), strings.Index(dry, "git diff --stat"),
		"dry-run still shows the incoming upstream diff after the guard")
	assert.Less(t, strings.Index(compat, guard), strings.Index(compat, "git checkout"))
	assert.Less(t, strings.Index(compat, guard), strings.Index(compat, "cp -R .ckeletin/"))
}

func taskSection(text, start, end string) string {
	from := strings.Index(text, start)
	if from < 0 {
		return ""
	}
	rest := text[from+len(start):]
	to := strings.Index(rest, end)
	if to < 0 {
		return rest
	}
	return rest[:to]
}
