// cmd/docs_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TestRunDocsConfig tests the runDocsConfig function
func TestRunDocsConfig(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		outputFile  string
		runErr      bool
		expectedErr string
	}{
		{
			name:        "Markdown format",
			format:      docs.FormatMarkdown,
			outputFile:  "",
			runErr:      false,
			expectedErr: "",
		},
		{
			name:        "YAML format",
			format:      docs.FormatYAML,
			outputFile:  "",
			runErr:      false,
			expectedErr: "",
		},
		{
			name:        "Invalid format",
			format:      "invalid",
			outputFile:  "",
			runErr:      true,
			expectedErr: "unsupported format: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Create test command
			cmd := &cobra.Command{}
			var output bytes.Buffer
			cmd.SetOut(&output)

			// Clear Viper config to avoid side effects
			viper.Reset()

			// Set up viper with test values
			viper.SetDefault("app.docs.output_format", tt.format)
			viper.SetDefault("app.docs.output_file", tt.outputFile)

			// Save original binaryName, EnvPrefix and ConfigPaths
			origBinaryName := binaryName
			defer func() { binaryName = origBinaryName }()
			binaryName = "testapp"

			// EXECUTION PHASE
			err := runDocsConfig(cmd, []string{})

			// ASSERTION PHASE
			if tt.runErr {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if !strings.Contains(err.Error(), tt.expectedErr) {
					t.Errorf("Expected error containing %q, got %q", tt.expectedErr, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify that the output contains expected content for valid formats
			if !tt.runErr {
				// For non-error cases, check that we generated some output
				if output.Len() == 0 {
					t.Errorf("No output was generated")
				}
			}
		})
	}
}

// TestDocsCommands tests the initialization and correct setup of the docs commands
func TestDocsCommands(t *testing.T) {
	// SETUP PHASE
	// Capture the log output
	consoleBuf := &bytes.Buffer{}
	origLogger := log.Logger
	log.Logger = zerolog.New(consoleBuf)
	defer func() {
		log.Logger = origLogger
	}()

	// Reset RootCmd for clean testing
	oldRoot := RootCmd
	RootCmd = &cobra.Command{Use: "test"}
	defer func() {
		RootCmd = oldRoot
	}()

	// Initialize the commands manually similar to init() function
	docsCmd := &cobra.Command{
		Use:   "docs",
		Short: "Generate documentation",
		Long:  `Generate documentation about the application, including configuration options.`,
	}
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Generate configuration documentation",
		Long:  `Generate documentation about all configuration options.`,
		RunE:  runDocsConfig,
	}

	// Set up command structure
	docsCmd.AddCommand(configCmd)
	RootCmd.AddCommand(docsCmd)

	// Add flags to config command
	configCmd.Flags().StringP("format", "f", docs.FormatMarkdown, "Output format (markdown, yaml)")
	configCmd.Flags().StringP("output", "o", "", "Output file (defaults to stdout)")

	// Bind flags to Viper
	if err := viper.BindPFlag("app.docs.output_format", configCmd.Flags().Lookup("format")); err != nil {
		t.Fatalf("Failed to bind format flag: %v", err)
	}
	if err := viper.BindPFlag("app.docs.output_file", configCmd.Flags().Lookup("output")); err != nil {
		t.Fatalf("Failed to bind output flag: %v", err)
	}

	// Set up command configuration inheritance
	setupCommandConfig(configCmd)

	// EXECUTION PHASE
	// Find the docs command
	foundDocsCmd, _, err := RootCmd.Find([]string{"docs"})
	if err != nil {
		t.Fatalf("Expected to find docs command: %v", err)
	}

	// Find the config subcommand
	foundConfigCmd, _, err := RootCmd.Find([]string{"docs", "config"})
	if err != nil {
		t.Fatalf("Expected to find docs config command: %v", err)
	}

	// ASSERTION PHASE
	// Check docs command properties
	if foundDocsCmd.Use != "docs" {
		t.Errorf("Expected docs command Use to be 'docs', got %s", foundDocsCmd.Use)
	}
	if foundDocsCmd.Short == "" {
		t.Errorf("Docs command should have a Short description")
	}

	// Check config command properties
	if foundConfigCmd.Use != "config" {
		t.Errorf("Expected config command Use to be 'config', got %s", foundConfigCmd.Use)
	}
	if foundConfigCmd.Short == "" {
		t.Errorf("Config command should have a Short description")
	}
	if foundConfigCmd.RunE == nil {
		t.Errorf("Config command should have a RunE function")
	}

	// Check that format and output flags are registered
	formatFlag := foundConfigCmd.Flags().Lookup("format")
	if formatFlag == nil {
		t.Errorf("format flag not found in config command")
	} else {
		if formatFlag.DefValue != docs.FormatMarkdown {
			t.Errorf("format flag default value should be %s, got %s", docs.FormatMarkdown, formatFlag.DefValue)
		}
	}

	outputFlag := foundConfigCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Errorf("output flag not found in config command")
	} else {
		if outputFlag.DefValue != "" {
			t.Errorf("output flag default value should be empty, got %s", outputFlag.DefValue)
		}
	}
}
