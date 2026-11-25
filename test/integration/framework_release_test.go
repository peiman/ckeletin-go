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

func TestReleaseFramework_WritesVersionAndTag(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module github.com/peiman/ckeletin-go\n\ngo 1.26.11\n"), 0o644))
	writeFrameworkTree(t, dir, "0.2.18", `# Framework Changelog

## [Unreleased]

### Fixed

- The audit records a rendered error.

## [0.2.18] - 2026-11-23

### Fixed

- An earlier release.
`, nil)
	gitCommitAll(t, dir, "framework 0.2.18")
	name, err := gitCmd(dir, "branch", "--show-current").Output()
	require.NoError(t, err)
	if strings.TrimSpace(string(name)) != "main" {
		require.NoError(t, gitCmd(dir, "branch", "-M", "main").Run())
	}

	cmd := exec.Command("bash", filepath.Join(repoRoot(t), "scripts", "release-framework.sh"), "--bump", "patch")
	cmd.Dir = dir
	cmd.Env = append(hermeticGitEnv(), "CKELETIN_RELEASE_ALLOW_BRANCH=1")
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))

	version, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "VERSION"))
	require.NoError(t, err)
	assert.Equal(t, "0.2.19\n", string(version))

	changelog, err := os.ReadFile(filepath.Join(dir, ".ckeletin", "CHANGELOG.md"))
	require.NoError(t, err)
	text := string(changelog)
	assert.Contains(t, text, "## [0.2.19] - ")
	assert.Contains(t, text, "The audit records a rendered error.")
	assert.Contains(t, text, "rendered error.\n\n## [0.2.18]")
	unreleased := text[strings.Index(text, "## [Unreleased]"):strings.Index(text, "## [0.2.19]")]
	assert.NotContains(t, unreleased, "- ")

	tag, err := gitCmd(dir, "rev-parse", "framework-v0.2.19^{}").Output()
	require.NoError(t, err)
	head, err := gitCmd(dir, "rev-parse", "HEAD").Output()
	require.NoError(t, err)
	assert.Equal(t, string(head), string(tag))
	assert.Contains(t, string(out), "Tagged framework-v0.2.19")

	again := exec.Command("bash", filepath.Join(repoRoot(t), "scripts", "release-framework.sh"))
	again.Dir = dir
	again.Env = append(hermeticGitEnv(), "CKELETIN_RELEASE_ALLOW_BRANCH=1")
	failOut, failErr := again.CombinedOutput()
	require.Error(t, failErr)
	assert.Contains(t, string(failOut), "Nothing to release")
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err)
	return root
}
