// cmd/completion_shell_test.go

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func genCompletion(t *testing.T, args ...string) string {
	t.Helper()
	var buf bytes.Buffer
	RootCmd.SetOut(&buf)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs(append([]string{"completion"}, args...))
	defer RootCmd.SetArgs(nil)
	require.NoError(t, RootCmd.Execute())
	return buf.String()
}

// TestCompletion_PerShell pins that each shell emits its OWN script, not bash for
// everything.
func TestCompletion_PerShell(t *testing.T) {
	cases := []struct {
		shell  string
		marker string
	}{
		{"bash", "bash completion"},
		{"zsh", "#compdef"},
		{"fish", "fish completion"},
		{"powershell", "Register-ArgumentCompleter"},
	}
	for _, tc := range cases {
		out := genCompletion(t, tc.shell)
		assert.Contains(t, out, tc.marker,
			"`completion %s` must emit a %s script (marker %q)", tc.shell, tc.shell, tc.marker)
	}
}

func TestCompletion_ZshDiffersFromBash(t *testing.T) {
	assert.NotEqual(t, genCompletion(t, "bash"), genCompletion(t, "zsh"),
		"zsh completion must differ from bash completion")
}

func TestCompletion_DefaultsToBash(t *testing.T) {
	assert.Contains(t, genCompletion(t), "bash completion",
		"completion with no shell arg should default to bash")
}

func TestCompletion_RejectsUnknownShell(t *testing.T) {
	var buf bytes.Buffer
	RootCmd.SetOut(&buf)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"completion", "tcsh"})
	defer RootCmd.SetArgs(nil)
	assert.Error(t, RootCmd.Execute(), "an unsupported shell should be rejected")
}
