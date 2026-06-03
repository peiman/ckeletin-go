package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// chProjectRoot returns the absolute project root.
func chProjectRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	require.NoError(t, err, "failed to resolve project root")
	return root
}

// chIsUpstream reports whether this is the ckeletin-go framework repo itself.
// .claude/ and .gitignore are project-owned (not synced by ckeletin:update), so
// this guard only asserts the framework's own state, not a downstream's. The
// module path is split so scaffold-init does not rewrite it in a downstream copy.
func chIsUpstream(t *testing.T) bool {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(chProjectRoot(t), "go.mod"))
	require.NoError(t, err, "failed to read go.mod")
	return strings.Contains(string(b), "module github.com/peiman"+"/ckeletin-go")
}

// TestClaudeHooks_LiveInSettingsJSON guards issue #96: the scaffolded Claude Code
// hooks must live in .claude/settings.json (which Claude Code actually reads),
// NOT in the non-functional .claude/hooks.json (which it never reads, so the
// hooks silently never fired). Prevents regression in the framework and in every
// scaffolded project.
func TestClaudeHooks_LiveInSettingsJSON(t *testing.T) {
	if !chIsUpstream(t) {
		t.Skip("guards the framework's own .claude config; downstream projects own theirs")
	}
	root := chProjectRoot(t)

	// hooks.json must be gone — it is not a location Claude Code reads hooks from.
	_, err := os.Stat(filepath.Join(root, ".claude", "hooks.json"))
	assert.Truef(t, os.IsNotExist(err),
		".claude/hooks.json must not exist — hooks belong in .claude/settings.json (issue #96)")

	// settings.json must exist, be valid JSON, and declare both hooks.
	b, err := os.ReadFile(filepath.Join(root, ".claude", "settings.json"))
	require.NoError(t, err, ".claude/settings.json must exist")

	var cfg struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	require.NoError(t, json.Unmarshal(b, &cfg), ".claude/settings.json must be valid JSON")
	assert.Contains(t, cfg.Hooks, "SessionStart", "settings.json must declare the SessionStart hook")
	assert.Contains(t, cfg.Hooks, "PreToolUse", "settings.json must declare the PreToolUse hook")

	// .gitignore must whitelist settings.json (so it is committed/team-shared) and
	// must no longer whitelist the removed hooks.json.
	gi, err := os.ReadFile(filepath.Join(root, ".gitignore"))
	require.NoError(t, err, ".gitignore must exist")
	assert.Contains(t, string(gi), "!.claude/settings.json",
		".gitignore must whitelist .claude/settings.json so it is committed")
	assert.NotContains(t, string(gi), "!.claude/hooks.json",
		".gitignore must not whitelist the removed hooks.json")
}
