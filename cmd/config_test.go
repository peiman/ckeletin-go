// cmd/config_test.go

package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunConfigValidate(t *testing.T) {
	tests := []struct {
		name              string
		configContent     string
		configPerms       os.FileMode
		setValidateFile   bool
		setCfgFile        bool
		wantErr           bool
		wantOutputContain string
	}{
		{
			name: "Valid config file",
			configContent: `app:
  log_level: info
  ping:
    output_message: "Test"
`,
			configPerms:       0600,
			setValidateFile:   true,
			wantErr:           false,
			wantOutputContain: "Configuration is valid",
		},
		{
			name: "Invalid YAML syntax",
			configContent: `app:
  invalid: [unclosed
`,
			configPerms:       0600,
			setValidateFile:   true,
			wantErr:           true,
			wantOutputContain: "Configuration is invalid",
		},
		{
			name: "Config with warnings (unknown keys)",
			configContent: `app:
  log_level: info
  unknown_key: value
`,
			configPerms:       0600,
			setValidateFile:   true,
			wantErr:           true, // Warnings also return error (exit code 1)
			wantOutputContain: "valid (with warnings)",
		},
		{
			name: "Use global --config flag when --file not set",
			configContent: `app:
  log_level: debug
`,
			configPerms:       0600,
			setValidateFile:   false,
			setCfgFile:        true,
			wantErr:           false,
			wantOutputContain: "Configuration is valid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global state
			viper.Reset()
			validateConfigFile = ""
			origCfgFile := cfgFile
			defer func() { cfgFile = origCfgFile }()

			// Create temp config file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")
			err := os.WriteFile(configFile, []byte(tt.configContent), tt.configPerms)
			require.NoError(t, err, "Failed to create test config")

			// Set up command
			cmd := &cobra.Command{}
			var output bytes.Buffer
			cmd.SetOut(&output)

			// Set config file paths based on test case
			if tt.setValidateFile {
				validateConfigFile = configFile
			} else if tt.setCfgFile {
				cfgFile = configFile
			} else {
				// For default path test, we'd need to set HOME and create config
				// This is complex, so we skip this case in unit tests
				// (it's tested in integration tests)
				t.Skip("Default path testing requires complex setup, tested in integration")
			}

			// Execute
			err = runConfigValidate(cmd, []string{})

			// Verify
			if tt.wantErr {
				assert.Error(t, err, "runConfigValidate should return error")
			} else {
				assert.NoError(t, err, "runConfigValidate should not return error")
			}

			if tt.wantOutputContain != "" {
				assert.Contains(t, output.String(), tt.wantOutputContain,
					"Output should contain expected string")
			}
		})
	}
}

func TestRunConfigValidate_NonexistentFile(t *testing.T) {
	// Reset global state
	validateConfigFile = "/nonexistent/config.yaml"
	defer func() { validateConfigFile = "" }()

	cmd := &cobra.Command{}
	var output bytes.Buffer
	cmd.SetOut(&output)

	err := runConfigValidate(cmd, []string{})

	require.Error(t, err, "Should return error for nonexistent file")
	assert.Contains(t, err.Error(), "validation failed",
		"Error should mention validation failed")
}

// TestConfigCommandRegistered tests that the config command is properly registered
func TestConfigCommandRegistered(t *testing.T) {
	// SETUP & EXECUTION PHASE
	// RootCmd should have config command as a child
	var foundConfig bool
	for _, c := range RootCmd.Commands() {
		if c.Name() == "config" {
			foundConfig = true
			break
		}
	}

	// ASSERTION PHASE
	assert.True(t, foundConfig, "config command should be registered in RootCmd")
}

// TestConfigCommandMetadata tests the config command's metadata and structure
func TestConfigCommandMetadata(t *testing.T) {
	// SETUP PHASE
	var configCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "config" {
			configCmd = c
			break
		}
	}

	require.NotNil(t, configCmd, "config command should be found")

	// ASSERTION PHASE - test parent command metadata
	tests := []struct {
		name     string
		got      string
		contains string
	}{
		{
			name:     "Use field",
			got:      configCmd.Use,
			contains: "config",
		},
		{
			name:     "Short description",
			got:      configCmd.Short,
			contains: "Configuration",
		},
		{
			name:     "Long description",
			got:      configCmd.Long,
			contains: "configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, strings.ToLower(tt.got), strings.ToLower(tt.contains),
				"%s should contain %q", tt.name, tt.contains)
		})
	}
}

// TestConfigValidateCommandRegistered tests that validate subcommand is registered
func TestConfigValidateCommandRegistered(t *testing.T) {
	// SETUP PHASE
	var configCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "config" {
			configCmd = c
			break
		}
	}

	require.NotNil(t, configCmd, "config command should be found")

	// EXECUTION & ASSERTION PHASE
	var foundValidate bool
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			foundValidate = true
			break
		}
	}

	assert.True(t, foundValidate, "validate subcommand should be registered under config command")
}

// TestConfigValidateCommandMetadata tests the validate subcommand's metadata
func TestConfigValidateCommandMetadata(t *testing.T) {
	// SETUP PHASE
	var configCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "config" {
			configCmd = c
			break
		}
	}

	require.NotNil(t, configCmd, "config command should be found")

	var validateCmd *cobra.Command
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			validateCmd = c
			break
		}
	}

	require.NotNil(t, validateCmd, "validate subcommand should be found")

	// ASSERTION PHASE
	tests := []struct {
		name     string
		got      string
		contains string
	}{
		{
			name:     "Use field",
			got:      validateCmd.Use,
			contains: "validate",
		},
		{
			name:     "Short description",
			got:      validateCmd.Short,
			contains: "Validate",
		},
		{
			name:     "Long description mentions validation",
			got:      validateCmd.Long,
			contains: "validate",
		},
		{
			name:     "Long description mentions security",
			got:      validateCmd.Long,
			contains: "security",
		},
		{
			name:     "Example provided",
			got:      validateCmd.Example,
			contains: "config validate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, strings.ToLower(tt.got), strings.ToLower(tt.contains),
				"%s should contain %q", tt.name, tt.contains)
		})
	}

	// Verify RunE is set
	assert.NotNil(t, validateCmd.RunE, "validateCmd.RunE should be set")
}

// TestConfigValidateCommandFlags tests that the validate command has correct flags
func TestConfigValidateCommandFlags(t *testing.T) {
	// SETUP PHASE
	var configCmd *cobra.Command
	for _, c := range RootCmd.Commands() {
		if c.Name() == "config" {
			configCmd = c
			break
		}
	}

	require.NotNil(t, configCmd, "config command should be found")

	var validateCmd *cobra.Command
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			validateCmd = c
			break
		}
	}

	require.NotNil(t, validateCmd, "validate subcommand should be found")

	// EXECUTION & ASSERTION PHASE
	// Check that --file flag exists
	fileFlag := validateCmd.Flags().Lookup("file")
	require.NotNil(t, fileFlag, "--file flag should exist")

	// Check flag shorthand
	assert.Equal(t, "f", fileFlag.Shorthand, "Flag shorthand should be 'f'")

	// Check flag usage text
	hasConfigFile := strings.Contains(fileFlag.Usage, "Config file") ||
		strings.Contains(fileFlag.Usage, "config file")
	assert.True(t, hasConfigFile, "Flag usage should mention config file")

	// Check default value is empty
	assert.Empty(t, fileFlag.DefValue, "Default value should be empty")
}
