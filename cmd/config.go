// cmd/config.go
//
// ckeletin:allow-custom-command — parent + subcommand wiring with a local
// --file flag, not config-registry-driven, so the NewCommand/metadata pattern
// does not apply (cmd/catalog.go cites this same exemption).

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/validator"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/internal/ui"
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
	result, err := validator.Validate(resolveValidateConfigPath())
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	exitErr := validator.ExitCodeForResult(result)

	// JSON mode: emit exactly one envelope (no human text). Like the `check`
	// command, return output.ErrRendered on failure so main.go signals a
	// non-zero exit without emitting a second envelope.
	if output.IsJSONMode() {
		if rerr := ui.RenderValidationJSON(cmd.OutOrStdout(), result, exitErr); rerr != nil {
			return rerr
		}
		if exitErr != nil {
			cmd.SilenceUsage = true
			return output.ErrRendered
		}
		return nil
	}

	// Text mode: human-readable formatting.
	validator.FormatResult(result, cmd.OutOrStdout())
	if exitErr != nil {
		cmd.SilenceUsage = true
		return exitErr
	}
	return nil
}

// resolveValidateConfigPath picks the file `config validate` targets: the
// --file flag, the config file viper already loaded, the global --config flag,
// or the default user config location — in that order.
func resolveValidateConfigPath() string {
	if validateConfigFile != "" {
		return validateConfigFile
	}
	if configFileUsed != "" {
		return configFileUsed
	}
	if cfgFile != "" {
		return cfgFile
	}
	if defaultDir := defaultUserConfigDir(ConfigPaths()); defaultDir != "" {
		return filepath.Join(defaultDir, "config.yaml")
	}
	return ""
}
