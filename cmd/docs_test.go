// cmd/docs_test.go

package cmd

import (
	"bytes"
	"testing"

	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				require.Error(t, err, "Expected error but got none")
				assert.Contains(t, err.Error(), tt.expectedErr, "Error should contain expected message")
			} else {
				require.NoError(t, err, "Unexpected error")

				// Verify that the output contains expected content for valid formats
				assert.NotZero(t, output.Len(), "Output should not be empty for valid formats")
			}
		})
	}
}

// TestDocsCommands tests the initialization and correct setup of the docs commands
func TestDocsCommands(t *testing.T) {
	// SETUP PHASE
	// Reset Viper state for clean testing
	viper.Reset()

	// Save logger state and restore after test
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Capture the log output
	consoleBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(consoleBuf)

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
	err := viper.BindPFlag("app.docs.output_format", configCmd.Flags().Lookup("format"))
	require.NoError(t, err, "Failed to bind format flag")

	err = viper.BindPFlag("app.docs.output_file", configCmd.Flags().Lookup("output"))
	require.NoError(t, err, "Failed to bind output flag")

	// Set up command configuration inheritance
	setupCommandConfig(configCmd)

	// EXECUTION PHASE
	// Find the docs command
	foundDocsCmd, _, err := RootCmd.Find([]string{"docs"})
	require.NoError(t, err, "Expected to find docs command")

	// Find the config subcommand
	foundConfigCmd, _, err := RootCmd.Find([]string{"docs", "config"})
	require.NoError(t, err, "Expected to find docs config command")

	// ASSERTION PHASE
	// Check docs command properties
	assert.Equal(t, "docs", foundDocsCmd.Use, "Docs command Use should be 'docs'")
	assert.NotEmpty(t, foundDocsCmd.Short, "Docs command should have a Short description")

	// Check config command properties
	assert.Equal(t, "config", foundConfigCmd.Use, "Config command Use should be 'config'")
	assert.NotEmpty(t, foundConfigCmd.Short, "Config command should have a Short description")
	assert.NotNil(t, foundConfigCmd.RunE, "Config command should have a RunE function")

	// Check that format and output flags are registered
	formatFlag := foundConfigCmd.Flags().Lookup("format")
	require.NotNil(t, formatFlag, "format flag should be registered in config command")
	assert.Equal(t, docs.FormatMarkdown, formatFlag.DefValue, "format flag default value should be markdown")

	outputFlag := foundConfigCmd.Flags().Lookup("output")
	require.NotNil(t, outputFlag, "output flag should be registered in config command")
	assert.Equal(t, "", outputFlag.DefValue, "output flag default value should be empty")
}
