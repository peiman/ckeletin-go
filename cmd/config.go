// cmd/config.go
//
// ckeletin:allow-custom-command

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/internal/config/validator"
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

//nolint:errcheck,revive // CLI output function - fmt.Fprintf errors to stdout are acceptable
func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Determine which config file to validate
	configPath := validateConfigFile
	if configPath == "" {
		// Use the global --config flag if set
		if cfgFile != "" {
			configPath = cfgFile
		} else {
			// Use default config path
			configPaths := ConfigPaths()
			configPath = configPaths.DefaultPath
		}
	}

	// Run validation
	result, err := validator.Validate(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Print results
	fmt.Fprintf(cmd.OutOrStdout(), "Validating: %s\n\n", result.ConfigFile)

	// Print errors
	if len(result.Errors) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "❌ Errors (%d):\n", len(result.Errors))
		for i, err := range result.Errors {
			fmt.Fprintf(cmd.OutOrStdout(), "  %d. %v\n", i+1, err)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// Print warnings
	if len(result.Warnings) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "⚠️  Warnings (%d):\n", len(result.Warnings))
		for i, warning := range result.Warnings {
			fmt.Fprintf(cmd.OutOrStdout(), "  %d. %s\n", i+1, warning)
		}
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}

	// Print summary and return appropriate status
	if result.Valid && len(result.Warnings) == 0 {
		// Exit code 0: Valid with no warnings
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "✅ Configuration is valid!")
		return nil
	} else if result.Valid && len(result.Warnings) > 0 {
		// Exit code 1: Valid but with warnings (return error to get exit code 1)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "✅ Configuration is valid (with warnings)")
		cmd.SilenceUsage = true
		return fmt.Errorf("validation completed with warnings")
	} else {
		// Exit code 1: Invalid (has errors)
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), "❌ Configuration is invalid")
		cmd.SilenceUsage = true
		return fmt.Errorf("validation failed")
	}
}
