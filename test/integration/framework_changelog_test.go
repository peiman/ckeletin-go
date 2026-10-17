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

func frameworkChangelogScript(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	p := filepath.Join(root, ".ckeletin", "scripts", "check-framework-changelog.sh")
	require.FileExists(t, p)
	return p
}

func runFrameworkChangelog(t *testing.T, dir string) (string, int) {
	t.Helper()
	cmd := exec.Command("bash", frameworkChangelogScript(t))
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv()
	out, err := cmd.CombinedOutput()
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	require.NoError(t, err, "changelog check errored\n%s", out)
	return string(out), 0
}

func writeFrameworkTree(t *testing.T, dir, version, changelog string, files map[string]string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".ckeletin"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "VERSION"), []byte(version+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "CHANGELOG.md"), []byte(changelog), 0o644))
	for rel, body := range files {
		full := filepath.Join(dir, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(full), 0o755))
		require.NoError(t, os.WriteFile(full, []byte(body), 0o644))
	}
}

func gitCommitAll(t *testing.T, dir, message string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		require.NoError(t, gitCmd(dir, "init").Run())
		require.NoError(t, gitCmd(dir, "config", "user.email", "t@example.com").Run())
		require.NoError(t, gitCmd(dir, "config", "user.name", "Test").Run())
	}
	require.NoError(t, gitCmd(dir, "add", "-A").Run())
	require.NoError(t, gitCmd(dir, "commit", "-m", message).Run())
}

const frameworkChangelog020 = `# Framework Changelog

## [Unreleased]

## [0.2.0] - 2026-10-14

### Added

- A recorded framework change.
`

func TestFrameworkChangelog_RequiresVersionHeading(t *testing.T) {
	dir := t.TempDir()
	writeFrameworkTree(t, dir, "0.2.0", "# Framework Changelog\n\n## [Unreleased]\n\n## [0.1.0] - 2026-04-01\n", nil)
	gitCommitAll(t, dir, "init")

	out, code := runFrameworkChangelog(t, dir)
	assert.Equal(t, 1, code, out)
	assert.Contains(t, out, "0.2.0")
}

func TestFrameworkChangelog_RequiresChangelogWhenFrameworkCodeChanges(t *testing.T) {
	dir := t.TempDir()
	writeFrameworkTree(t, dir, "0.2.0", frameworkChangelog020, map[string]string{
		".ckeletin/scripts/example.sh": "echo ok\n",
	})
	gitCommitAll(t, dir, "framework 0.2.0")

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "scripts", "example.sh"), []byte("echo changed\n"), 0o644))

	out, code := runFrameworkChangelog(t, dir)
	assert.Equal(t, 1, code, out)
	assert.Contains(t, out, ".ckeletin/scripts/example.sh")
	assert.Contains(t, out, ".ckeletin/CHANGELOG.md")
}

func TestFrameworkChangelog_AllowsCodeChangeRecordedInChangelog(t *testing.T) {
	dir := t.TempDir()
	writeFrameworkTree(t, dir, "0.2.0", frameworkChangelog020, map[string]string{
		".ckeletin/scripts/example.sh": "echo ok\n",
	})
	gitCommitAll(t, dir, "framework 0.2.0")

	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "scripts", "example.sh"), []byte("echo changed\n"), 0o644))
	updated := frameworkChangelog020 + "\n## [Unreleased]\n\n### Fixed\n\n- example.sh prints changed.\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "CHANGELOG.md"), []byte(updated), 0o644))

	out, code := runFrameworkChangelog(t, dir)
	assert.Equal(t, 0, code, out)
}

func TestFrameworkChangelog_AllowsDocsOnlyEdits(t *testing.T) {
	dir := t.TempDir()
	writeFrameworkTree(t, dir, "0.2.0", frameworkChangelog020, map[string]string{
		".ckeletin/docs/adr/000-example.md": "# ADR\n",
	})
	gitCommitAll(t, dir, "framework 0.2.0")

	require.NoError(t, os.WriteFile(
		filepath.Join(dir, ".ckeletin", "docs", "adr", "000-example.md"),
		[]byte("# ADR\n\nA comment-only edit.\n"),
		0o644,
	))

	out, code := runFrameworkChangelog(t, dir)
	assert.Equal(t, 0, code, out)
}

func TestFrameworkChangelog_UncommittedVersionBumpDoesNotReplayHistory(t *testing.T) {
	dir := t.TempDir()
	writeFrameworkTree(t, dir, "0.1.0", "# Framework Changelog\n\n## [0.1.0] - 2026-04-01\n", map[string]string{
		".ckeletin/scripts/example.sh": "echo ok\n",
	})
	gitCommitAll(t, dir, "framework 0.1.0")

	writeFrameworkTree(t, dir, "0.2.0", frameworkChangelog020, nil)

	out, code := runFrameworkChangelog(t, dir)
	assert.Equal(t, 0, code, out)
}

func TestFrameworkChangelog_PassesOnThisRepo(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	out, code := runFrameworkChangelog(t, root)
	assert.Equal(t, 0, code, out)
}

func TestGenerateCommand_ForwardsIntoTheFrameworkTask(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)

	project, err := os.ReadFile(filepath.Join(root, "Taskfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(project), "task: ckeletin:generate:command")
	assert.Contains(t, string(project), "CLI_ARGS: '{{.CLI_ARGS}}'")
	assert.NotContains(t, string(project), "scripts/generate-command/")

	framework, err := os.ReadFile(filepath.Join(root, ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)
	body := string(framework)
	start := strings.Index(body, "  generate:command:\n")
	require.NotEqual(t, -1, start, "framework Taskfile must define generate:command")
	// The next framework task is indented at the same level. Slice on that so the
	// assertions cannot pass by finding the generator path in some other task.
	next := strings.Index(body[start+1:], "\n  generate:")
	require.NotEqual(t, -1, next)
	task := body[start : start+1+next]
	assert.Contains(t, task, "generate-command/")
	assert.Contains(t, task, ".name | default .CLI_ARGS")
}

func TestGenerateCommandMigration_WarnsWhenTheProjectTaskDoesNotForward(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	script := filepath.Join(root, ".ckeletin", "scripts", "migrate-post-update.sh")

	dir := t.TempDir()
	old := "" +
		"version: '3'\n" +
		"tasks:\n" +
		"  generate:command:\n" +
		"    desc: Generate a new command\n" +
		"    cmds:\n" +
		"      - go run ./.ckeletin/scripts/generate-command/ {{.CLI_ARGS}}\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(old), 0o644))

	cmd := exec.Command("bash", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	text := string(out)
	assert.Contains(t, text, "does not forward to ckeletin:generate:command")
	assert.Contains(t, text, "name: '{{.name}}'")
	assert.Contains(t, text, "CLI_ARGS: '{{.CLI_ARGS}}'")
	assert.Equal(t, 1, strings.Count(text, "does not forward to ckeletin:generate:command"))
}

func TestRewriteDerivedNames_MatchesInit(t *testing.T) {
	// Build the upstream name from parts. task init rewrites that name, and its
	// env-var form, in this file when the update fixture copies it into a
	// derived project. A contiguous literal would make the fixture lie.
	upstreamName := "ckeletin" + "-go"
	upstreamModule := "github.com/peiman/" + upstreamName
	upstreamEnv := "CKELETIN" + "_GO"
	derivedName := "updatetest"
	derivedModule := "github.com/test/" + derivedName
	derivedEnv := "UPDATETEST"

	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	script := filepath.Join(root, ".ckeletin", "scripts", "rewrite-derived-names.sh")
	require.FileExists(t, script)

	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, ".ckeletin", "scripts"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "Taskfile.yml"), []byte(""+
		"vars:\n"+
		"  _BINARY_NAME_DEFAULT: "+upstreamName+"\n"+
		"  _MODULE_PATH_DEFAULT: "+upstreamModule+"\n"+
		"  # see https://github.com/peiman/"+upstreamName+"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "scripts", "note.sh"), []byte(
		"echo \""+upstreamModule+"\"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "sample_test.go"), []byte(""+
		"package sample\n"+
		"\n"+
		"import \""+upstreamModule+"/pkg/checkmate\"\n"+
		"\n"+
		"// comment before rewrite: "+upstreamName+" and ckeletin-rust.\n"+
		"const logPath = \"./logs/"+upstreamName+".log\"\n"+
		"const envPrefix = \""+upstreamEnv+"\"\n"), 0o644))
	// The marker keeps upstream module literals. An unmarked Go file is renamed.
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "kept_test.go"), []byte(""+
		"// ckeletin:keep-upstream-names — fixture literals must survive the derived-name rewrite.\n"+
		"package kept\n"+
		"\n"+
		"const oldModule = \""+upstreamModule+"\"\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, ".ckeletin", "plain_test.go"), []byte(""+
		"package plain\n"+
		"\n"+
		"const oldModule = \""+upstreamModule+"\"\n"), 0o644))
	fixtureSrc := filepath.Join(root, ".ckeletin", "scripts", "rewrite-imports", "main_test.go")
	fixtureBody, readErr := os.ReadFile(fixtureSrc)
	require.NoError(t, readErr)
	fixtureDst := filepath.Join(dir, ".ckeletin", "scripts", "rewrite-imports", "main_test.go")
	require.NoError(t, os.MkdirAll(filepath.Dir(fixtureDst), 0o755))
	require.NoError(t, os.WriteFile(fixtureDst, fixtureBody, 0o644))

	cmd := exec.Command("bash", script,
		upstreamModule, derivedModule, upstreamName, derivedName, ".ckeletin")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	taskfile, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)
	assert.Contains(t, string(taskfile), "_BINARY_NAME_DEFAULT: "+derivedName)
	assert.Contains(t, string(taskfile), "_MODULE_PATH_DEFAULT: "+derivedModule)
	assert.Contains(t, string(taskfile), "https://github.com/test/"+derivedName)
	assert.NotContains(t, string(taskfile), upstreamName)

	// Shell files are not rewritten by task init, so an update must leave them alone.
	scriptBody, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "scripts", "note.sh"))
	require.NoError(t, err)
	assert.Contains(t, string(scriptBody), upstreamModule)

	goBody, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "sample_test.go"))
	require.NoError(t, err)
	assert.Contains(t, string(goBody), "import \""+upstreamModule+"/pkg/checkmate\"")
	assert.Contains(t, string(goBody), derivedName+" and ckeletin-rust")
	assert.Contains(t, string(goBody), "\"./logs/"+derivedName+".log\"")
	assert.Contains(t, string(goBody), "const envPrefix = \""+derivedEnv+"\"")

	kept, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "kept_test.go"))
	require.NoError(t, err)
	assert.Contains(t, string(kept), upstreamModule)
	assert.NotContains(t, string(kept), derivedName)

	plain, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "plain_test.go"))
	require.NoError(t, err)
	assert.NotContains(t, string(plain), upstreamModule)
	assert.Contains(t, string(plain), derivedName)

	rewrittenFixture, err := os.ReadFile(fixtureDst)
	require.NoError(t, err)
	assert.Contains(t, string(rewrittenFixture), upstreamModule, "marked fixture must keep the upstream module literal")
	assert.NotContains(t, string(rewrittenFixture), derivedName)
}

func TestUpstreamFixtureFilesKeepTheMarker(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	files := []string{
		".ckeletin/scripts/rewrite-imports/main_test.go",
		".ckeletin/scripts/scaffold/helpers_test.go",
		".ckeletin/scripts/scaffold/update_taskfile_test.go",
	}
	for _, rel := range files {
		body, readErr := os.ReadFile(filepath.Join(root, rel))
		require.NoError(t, readErr)
		assert.Contains(t, string(body), "ckeletin:keep-upstream-names", rel)
	}
}

func TestGenerateCommandMigration_SilentWhenTheProjectTaskForwards(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	script := filepath.Join(root, ".ckeletin", "scripts", "migrate-post-update.sh")

	dir := t.TempDir()
	forwarded := "" +
		"version: '3'\n" +
		"tasks:\n" +
		"  generate:command:\n" +
		"    desc: Generate a new command with full pattern (cmd + internal + tests + config)\n" +
		"    cmds:\n" +
		"      - task: ckeletin:generate:command\n" +
		"        vars:\n" +
		"          name: '{{.name}}'\n" +
		"          CLI_ARGS: '{{.CLI_ARGS}}'\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Taskfile.yml"), []byte(forwarded), 0o644))

	cmd := exec.Command("bash", script)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
	assert.NotContains(t, string(out), "does not forward to ckeletin:generate:command")
}
