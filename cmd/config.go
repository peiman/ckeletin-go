// cmd/config.go
//
// ckeletin:allow-custom-command

package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/validator"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
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
			// Default to the selected user config directory for validation target
			configPaths := ConfigPaths()
			if defaultDir := defaultUserConfigDir(configPaths); defaultDir != "" {
				configPath = filepath.Join(defaultDir, "config.yaml")
			}
		}
	}

	// Run validation
	result, err := validator.Validate(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	exitErr := validator.ExitCodeForResult(result)

	// JSON mode: emit exactly one envelope (no human text). Like the `check`
	// command, return nil afterward so main.go does not emit a second envelope —
	// the envelope's status communicates success/failure.
	if output.IsJSONMode() {
		status := "success"
		var jsonErr *output.JSONError
		if exitErr != nil {
			status = "error"
			jsonErr = &output.JSONError{Message: exitErr.Error()}
		}
		errMsgs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			errMsgs[i] = e.Error()
		}
		if rerr := output.RenderJSON(cmd.OutOrStdout(), output.JSONEnvelope{
			Status:  status,
			Command: output.CommandName(),
			Data: map[string]any{
				"valid":       result.Valid,
				"config_file": result.ConfigFile,
				"errors":      errMsgs,
				"warnings":    result.Warnings,
			},
			Error: jsonErr,
		}); rerr != nil {
			return fmt.Errorf("failed to write JSON output: %w", rerr)
		}
		// The single envelope is written; signal a non-zero exit on failure
		// (errors or warnings) without main.go emitting a second envelope.
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
