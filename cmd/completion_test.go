// cmd/completion_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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
	if !foundCompletion {
		t.Error("completion command not found in RootCmd.Commands()")
	}
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

	if completionCmd == nil {
		t.Fatal("completion command not found")
	}

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
			name:     "Long description contains bash",
			got:      completionCmd.Long,
			contains: "Bash:",
		},
		{
			name:     "Long description contains zsh",
			got:      completionCmd.Long,
			contains: "Zsh:",
		},
		{
			name:     "Long description contains fish",
			got:      completionCmd.Long,
			contains: "Fish:",
		},
		{
			name:     "Long description contains powershell",
			got:      completionCmd.Long,
			contains: "PowerShell:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(strings.ToLower(tt.got), strings.ToLower(tt.contains)) {
				t.Errorf("%s doesn't contain %q\nGot: %s", tt.name, tt.contains, tt.got)
			}
		})
	}

	// Test DisableFlagsInUseLine is true
	if !completionCmd.DisableFlagsInUseLine {
		t.Error("DisableFlagsInUseLine should be true")
	}
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

			if completionCmd == nil {
				t.Fatal("completion command not found")
			}

			if completionCmd.RunE == nil {
				t.Fatal("completionCmd.RunE is nil")
			}

			// Capture output
			var stdout bytes.Buffer
			completionCmd.SetOut(&stdout)

			// EXECUTION PHASE
			// Call RunE directly to avoid command hierarchy issues
			err := completionCmd.RunE(completionCmd, tt.args)

			// ASSERTION PHASE
			if (err != nil) != tt.wantErr {
				t.Errorf("Expected error=%v, got error=%v", tt.wantErr, err)
			}

			output := stdout.String()
			if tt.outputNotEmpty && output == "" {
				t.Error("Expected non-empty output, got empty string")
			}

			for _, contains := range tt.outputContains {
				if !strings.Contains(strings.ToLower(output), strings.ToLower(contains)) {
					t.Errorf("Output doesn't contain %q\nOutput preview: %s...", contains, output[:min(100, len(output))])
				}
			}
		})
	}
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

	if completionCmd == nil {
		t.Fatal("completion command not found")
	}

	// Verify RunE is set
	if completionCmd.RunE == nil {
		t.Fatal("completionCmd.RunE is nil")
	}

	// EXECUTION PHASE
	var output bytes.Buffer
	completionCmd.SetOut(&output)

	err := completionCmd.RunE(completionCmd, []string{})

	// ASSERTION PHASE
	if err != nil {
		t.Errorf("RunE should not return error, got: %v", err)
	}

	if output.Len() == 0 {
		t.Error("RunE should generate completion output, got empty")
	}

	// Verify it's bash completion (default)
	outputStr := output.String()
	if !strings.Contains(outputStr, "bash") && !strings.Contains(outputStr, "completion") {
		t.Errorf("Output doesn't look like bash completion script:\n%s", outputStr)
	}
}
