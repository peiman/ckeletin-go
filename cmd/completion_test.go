// cmd/completion_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompletionCommandRegistered tests that the completion command is properly registered
func TestCompletionCommandRegistered(t *testing.T) {
	// SETUP PHASE
	// RootCmd should have completion command as a child

	// EXECUTION PHASE
	cmd := RootCmd.Commands()
	var foundCompletion bool
	for _, c := range cmd {
		if c.Name() == "completion" {
			foundCompletion = true
			break
		}
	}

	// ASSERTION PHASE
	assert.True(t, foundCompletion, "completion command should be registered in RootCmd")
}

// TestCompletionCommandMetadata tests the completion command's metadata
func TestCompletionCommandMetadata(t *testing.T) {
	// SETUP PHASE
	cmd := RootCmd.Commands()
	var completionCmd *cobra.Command
	for _, c := range cmd {
		if c.Name() == "completion" {
			completionCmd = c
			break
		}
	}

	require.NotNil(t, completionCmd, "completion command should be found")

	// Long is rendered lazily at help time (see cmd/completion.go init),
	// so assert against the rendered help output, not the raw field.
	var helpBuf bytes.Buffer
	completionCmd.SetOut(&helpBuf)
	defer completionCmd.SetOut(nil)
	require.NoError(t, completionCmd.Help(), "help rendering should succeed")
	helpText := helpBuf.String()

	// ASSERTION PHASE
	tests := []struct {
		name     string
		got      string
		contains string
	}{
		{
			name:     "Use field",
			got:      completionCmd.Use,
			contains: "completion",
		},
		{
			name:     "Short description",
			got:      completionCmd.Short,
			contains: "autocompletion",
		},
		{
			name:     "Help output contains bash",
			got:      helpText,
			contains: "Bash:",
		},
		{
			name:     "Help output contains zsh",
			got:      helpText,
			contains: "Zsh:",
		},
		{
			name:     "Help output contains fish",
			got:      helpText,
			contains: "Fish:",
		},
		{
			name:     "Help output contains powershell",
			got:      helpText,
			contains: "PowerShell:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, strings.ToLower(tt.got), strings.ToLower(tt.contains),
				"%s should contain %q", tt.name, tt.contains)
		})
	}

	// Test DisableFlagsInUseLine is true
	assert.True(t, completionCmd.DisableFlagsInUseLine, "DisableFlagsInUseLine should be true")
}

// TestCompletionCommandExecution tests that the completion command generates output via RunE
func TestCompletionCommandExecution(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		outputContains []string
		outputNotEmpty bool
	}{
		{
			name:           "Default bash completion",
			args:           []string{},
			wantErr:        false,
			outputContains: []string{"bash", "completion"},
			outputNotEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Find completion command
			var completionCmd *cobra.Command
			for _, c := range RootCmd.Commands() {
				if c.Name() == "completion" {
					completionCmd = c
					break
				}
			}

			require.NotNil(t, completionCmd, "completion command should be found")
			require.NotNil(t, completionCmd.RunE, "completionCmd.RunE should be set")

			// Capture output
			var stdout bytes.Buffer
			completionCmd.SetOut(&stdout)

			// EXECUTION PHASE
			// Call RunE directly to avoid command hierarchy issues
			err := completionCmd.RunE(completionCmd, tt.args)

			// ASSERTION PHASE
			if tt.wantErr {
				assert.Error(t, err, "Should return error")
			} else {
				assert.NoError(t, err, "Should not return error")
			}

			output := stdout.String()
			if tt.outputNotEmpty {
				assert.NotEmpty(t, output, "Output should not be empty")
			}

			for _, contains := range tt.outputContains {
				assert.Contains(t, strings.ToLower(output), strings.ToLower(contains),
					"Output should contain %q", contains)
			}
		})
	}
}

// TestCompletionHelpRendersBinaryName pins that the rendered help examples
// include the resolved binary name. binaryName is still empty while
// package-level vars initialize (root.go's init() resolves it afterwards, and
// init() runs in file-name order), so the Long text must be rendered lazily
// at help time, not captured at var-declaration time.
func TestCompletionHelpRendersBinaryName(t *testing.T) {
	// SETUP PHASE
	var completionCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "completion" {
			completionCmd = c
			break
		}
	}
	require.NotNil(t, completionCmd, "completion command should be found")

	var buf bytes.Buffer
	RootCmd.SetOut(&buf)
	RootCmd.SetErr(&bytes.Buffer{})
	completionCmd.SetOut(&buf)
	defer completionCmd.SetOut(nil)
	RootCmd.SetArgs([]string{"completion", "--help"})
	defer RootCmd.SetArgs(nil)

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	require.NoError(t, err, "`completion --help` should succeed")
	require.Equal(t, "ckeletin-go", RootCmd.Name(), "dev builds should resolve the binary name fallback")

	help := buf.String()
	examples := []string{
		"source <(ckeletin-go completion bash)",
		"source <(ckeletin-go completion zsh)",
		"ckeletin-go completion fish | source",
		"ckeletin-go completion powershell | Out-String | Invoke-Expression",
	}
	for _, example := range examples {
		assert.Contains(t, help, example,
			"help should render the binary name in the shell example")
	}
	assert.NotContains(t, help, "<( completion",
		"binary name must not render empty in help examples")
}

// TestCompletionCommandConfigInheritance pins that completion is registered
// via MustAddToRoot, which wraps PreRunE through setupCommandConfig like
// every other command.
func TestCompletionCommandConfigInheritance(t *testing.T) {
	// SETUP PHASE
	var completionCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "completion" {
			completionCmd = c
			break
		}
	}
	require.NotNil(t, completionCmd, "completion command should be found")

	// ASSERTION PHASE
	assert.NotNil(t, completionCmd.PreRunE,
		"completion should get the setupCommandConfig PreRunE wrapper via MustAddToRoot")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestCompletionCommandRunE tests the RunE function directly
func TestCompletionCommandRunE(t *testing.T) {
	// SETUP PHASE
	// Find completion command
	var completionCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "completion" {
			completionCmd = c
			break
		}
	}

	require.NotNil(t, completionCmd, "completion command should be found")
	require.NotNil(t, completionCmd.RunE, "completionCmd.RunE should be set")

	// EXECUTION PHASE
	var output bytes.Buffer
	completionCmd.SetOut(&output)

	err := completionCmd.RunE(completionCmd, []string{})

	// ASSERTION PHASE
	assert.NoError(t, err, "RunE should not return error")
	assert.NotEmpty(t, output.String(), "RunE should generate completion output")

	// Verify it's bash completion (default)
	outputStr := output.String()
	hasBash := strings.Contains(outputStr, "bash")
	hasCompletion := strings.Contains(outputStr, "completion")
	assert.True(t, hasBash || hasCompletion, "Output should look like bash completion script")
}
