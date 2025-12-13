// cmd/config.go
//
// ckeletin:allow-custom-command

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/validator"
	"github.com/spf13/cobra"
)

// configCmd represents the config command (parent command)
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management commands",
	Long:  `Commands for managing and validating application configuration.`,
}

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration file",
	Long: `Validate a configuration file for correctness, security, and completeness.

This command checks:
- File existence and readability
- File size limits (prevents DoS)
- File permissions (security)
- YAML syntax validity
- Configuration value limits
- Unknown configuration keys

Exit codes:
  0 - Configuration is valid (no warnings)
  1 - Configuration has errors or warnings`,
	Example: `  # Validate default config file
  ckeletin-go config validate

  # Validate specific config file
  ckeletin-go config validate --file /path/to/config.yaml`,
	RunE: runConfigValidate,
}

var validateConfigFile string

func init() {
	// Add validate command to config command
	configCmd.AddCommand(configValidateCmd)

	// Add flags
	configValidateCmd.Flags().StringVarP(&validateConfigFile, "file", "f", "",
		"Config file to validate (default: uses --config flag or default location)")

	// Add config command to root
	MustAddToRoot(configCmd)
}

func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Determine which config file to validate
	configPath := validateConfigFile
	if configPath == "" {
		// Use the config file viper already found, or the global --config flag
		if configFileUsed != "" {
			configPath = configFileUsed
		} else if cfgFile != "" {
			configPath = cfgFile
		} else {
			// Default to XDG path for validation target
			configPaths := ConfigPaths()
			if configPaths.XDGDir != "" {
				configPath = configPaths.XDGDir + "/config.yaml"
			}
		}
	}

	// Run validation
	result, err := validator.Validate(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Format and display results
	validator.FormatResult(result, cmd.OutOrStdout())

	// Determine exit code and suppress usage on validation errors
	if exitErr := validator.ExitCodeForResult(result); exitErr != nil {
		cmd.SilenceUsage = true
		return exitErr
	}

	return nil
}
