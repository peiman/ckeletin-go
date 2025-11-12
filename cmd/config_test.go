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
			if err := os.WriteFile(configFile, []byte(tt.configContent), tt.configPerms); err != nil {
				t.Fatalf("Failed to create test config: %v", err)
			}

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
			err := runConfigValidate(cmd, []string{})

			// Verify
			if (err != nil) != tt.wantErr {
				t.Errorf("runConfigValidate() error = %v, wantErr %v", err, tt.wantErr)
			}

			output_str := output.String()
			if tt.wantOutputContain != "" && !strings.Contains(output_str, tt.wantOutputContain) {
				t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContain, output_str)
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

	if err == nil {
		t.Error("Expected error for nonexistent file")
	}

	if !strings.Contains(err.Error(), "validation failed") {
		t.Errorf("Expected 'validation failed' error, got: %v", err)
	}
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
	if !foundConfig {
		t.Error("config command not found in RootCmd.Commands()")
	}
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

	if configCmd == nil {
		t.Fatal("config command not found")
	}

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
			if !strings.Contains(strings.ToLower(tt.got), strings.ToLower(tt.contains)) {
				t.Errorf("%s doesn't contain %q\nGot: %s", tt.name, tt.contains, tt.got)
			}
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

	if configCmd == nil {
		t.Fatal("config command not found")
	}

	// EXECUTION & ASSERTION PHASE
	var foundValidate bool
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			foundValidate = true
			break
		}
	}

	if !foundValidate {
		t.Error("validate subcommand not found under config command")
	}
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

	if configCmd == nil {
		t.Fatal("config command not found")
	}

	var validateCmd *cobra.Command
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			validateCmd = c
			break
		}
	}

	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}

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
			if !strings.Contains(strings.ToLower(tt.got), strings.ToLower(tt.contains)) {
				t.Errorf("%s doesn't contain %q\nGot: %s", tt.name, tt.contains, tt.got)
			}
		})
	}

	// Verify RunE is set
	if validateCmd.RunE == nil {
		t.Error("validateCmd.RunE should not be nil")
	}
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

	if configCmd == nil {
		t.Fatal("config command not found")
	}

	var validateCmd *cobra.Command
	for _, c := range configCmd.Commands() {
		if c.Name() == "validate" {
			validateCmd = c
			break
		}
	}

	if validateCmd == nil {
		t.Fatal("validate subcommand not found")
	}

	// EXECUTION & ASSERTION PHASE
	// Check that --file flag exists
	fileFlag := validateCmd.Flags().Lookup("file")
	if fileFlag == nil {
		t.Fatal("--file flag not found")
	}

	// Check flag shorthand
	if fileFlag.Shorthand != "f" {
		t.Errorf("Expected shorthand 'f', got '%s'", fileFlag.Shorthand)
	}

	// Check flag usage text
	if !strings.Contains(fileFlag.Usage, "Config file") && !strings.Contains(fileFlag.Usage, "config file") {
		t.Errorf("Flag usage doesn't mention config file: %s", fileFlag.Usage)
	}

	// Check default value is empty
	if fileFlag.DefValue != "" {
		t.Errorf("Expected default value to be empty, got '%s'", fileFlag.DefValue)
	}
}
