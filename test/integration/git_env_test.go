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

// hermeticGitEnv is the process environment without git's own variables.
// A hook run from a worktree exports GIT_DIR (and GIT_INDEX_FILE, and the
// rest of GIT_*) pointing at the calling repository. A test git command that
// inherits them writes that repository instead of the temp directory it was
// given, including setting core.bare=true.
func hermeticGitEnv() []string {
	var env []string
	for _, kv := range os.Environ() {
		if strings.HasPrefix(kv, "GIT_") {
			continue
		}
		env = append(env, kv)
	}
	return env
}

// gitCmd runs git in dir and does not inherit GIT_*.
func gitCmd(dir string, args ...string) *exec.Cmd {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = hermeticGitEnv()
	return cmd
}

// gitExplicit runs git against one repository by flag, so an inherited GIT_DIR
// cannot redirect setup onto another checkout.
func gitExplicit(t *testing.T, gitDir, workTree string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"--git-dir", gitDir, "--work-tree", workTree}, args...)...)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %v failed\n%s", args, out)
}

func TestHermeticGitEnv_DropsGitVariables(t *testing.T) {
	t.Setenv("GIT_DIR", "/real/repo/.git")
	t.Setenv("GIT_WORK_TREE", "/real/repo")
	t.Setenv("GIT_INDEX_FILE", "/real/index")
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("CKELETIN_HERMETIC_KEEP", "yes")

	env := hermeticGitEnv()

	for _, kv := range env {
		key, _, _ := strings.Cut(kv, "=")
		assert.Falsef(t, strings.HasPrefix(key, "GIT_"), "kept %s", kv)
	}
	assert.Contains(t, env, "CKELETIN_HERMETIC_KEEP=yes")
}

// TestGitCmd_InheritedWorktreeGitDirStaysInTheTempRepo is the worktree-hook
// failure: GIT_DIR points at the calling worktree, and git init / git config
// in a temp directory rewrite the real repository (core.bare=true, and the
// test identity in user.email).
func TestGitCmd_InheritedWorktreeGitDirStaysInTheTempRepo(t *testing.T) {
	repo := t.TempDir()
	gitDir := filepath.Join(repo, ".git")
	gitExplicit(t, gitDir, repo, "init", "-b", "main")
	gitExplicit(t, gitDir, repo, "config", "user.email", "test@ckeletin-go.example")
	gitExplicit(t, gitDir, repo, "config", "user.name", "Test User")
	gitExplicit(t, gitDir, repo, "commit", "--allow-empty", "-m", "init")

	wt := t.TempDir()
	gitExplicit(t, gitDir, repo, "worktree", "add", wt, "-b", "wtbranch")
	raw, err := os.ReadFile(filepath.Join(wt, ".git"))
	require.NoError(t, err)
	worktreeGitDir := strings.TrimSpace(strings.TrimPrefix(string(raw), "gitdir: "))
	require.NotEmpty(t, worktreeGitDir)
	if !filepath.IsAbs(worktreeGitDir) {
		worktreeGitDir = filepath.Join(wt, worktreeGitDir)
	}

	t.Setenv("GIT_DIR", worktreeGitDir)
	unsetEnv(t, "GIT_WORK_TREE")
	unsetEnv(t, "GIT_INDEX_FILE")

	other := t.TempDir()
	initCmd := gitCmd(other, "init")
	initOut, err := initCmd.CombinedOutput()
	require.NoError(t, err, "git init in the temp dir failed\n%s", initOut)

	bareCmd := exec.Command("git", "--git-dir", gitDir, "config", "--get", "core.bare")
	bareCmd.Env = hermeticGitEnv()
	bare, err := bareCmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "false", strings.TrimSpace(string(bare)), "git init must not mark the calling repository bare")

	configCmd := gitCmd(other, "config", "user.email", "leaked-by-test@ckeletin-go.example")
	configOut, err := configCmd.CombinedOutput()
	require.NoError(t, err, "git config in the temp dir failed\n%s", configOut)

	emailCmd := exec.Command("git", "--git-dir", gitDir, "config", "--get", "user.email")
	emailCmd.Env = hermeticGitEnv()
	email, err := emailCmd.Output()
	require.NoError(t, err)
	assert.Equal(t, "test@ckeletin-go.example", strings.TrimSpace(string(email)), "git config must not write the calling repository")
}

func unsetEnv(t *testing.T, key string) {
	t.Helper()
	val, ok := os.LookupEnv(key)
	require.NoError(t, os.Unsetenv(key))
	t.Cleanup(func() {
		if ok {
			_ = os.Setenv(key, val)
			return
		}
		_ = os.Unsetenv(key)
	})
}
