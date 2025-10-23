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
			wantErr:           true,
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
