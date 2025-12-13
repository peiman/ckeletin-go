//go:build dev

// ckeletin:allow-custom-command
// cmd/dev_config.go
//
// Configuration inspector subcommand (dev-only).
// Provides utilities to inspect, validate, and export configuration.

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/internal/dev"
	"github.com/spf13/cobra"
)

var (
	configList     bool
	configShow     bool
	configExport   string
	configValidate bool
	configPrefix   string
)

var devConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Inspect and validate configuration",
	Long: `Inspect the configuration registry, show effective configuration values,
validate current configuration, and export to various formats.

Examples:
  # List all configuration options
  dev config --list

  # Show effective configuration (merged from defaults, file, env)
  dev config --show

  # Export configuration to JSON
  dev config --export json

  # Validate current configuration
  dev config --validate

  # Show config options with specific prefix
  dev config --prefix app.log`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create config inspector
		ci := dev.NewConfigInspector()

		// Handle --list flag
		if configList {
			table := ci.FormatAsTable()
			fmt.Fprintln(cmd.OutOrStdout(), table)
			return nil
		}

		// Handle --show flag
		if configShow {
			table := ci.FormatEffectiveAsTable()
			fmt.Fprintln(cmd.OutOrStdout(), table)
			return nil
		}

		// Handle --export flag
		if configExport != "" {
			switch configExport {
			case "json":
				jsonStr, err := ci.ExportToJSON(true)
				if err != nil {
					return fmt.Errorf("failed to export to JSON: %w", err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), jsonStr)
				return nil
			default:
				return fmt.Errorf("unsupported export format: %s (supported: json)", configExport)
			}
		}

		// Handle --validate flag
		if configValidate {
			errors := ci.ValidateConfig()
			if len(errors) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "❌ Configuration validation failed:")
				for i, err := range errors {
					fmt.Fprintf(cmd.OutOrStdout(), "  %d. %v\n", i+1, err)
				}
				return fmt.Errorf("validation found %d error(s)", len(errors))
			}
			fmt.Fprintln(cmd.OutOrStdout(), "✅ Configuration is valid")
			return nil
		}

		// Handle --prefix flag
		if configPrefix != "" {
			matches := ci.GetConfigByPrefix(configPrefix)
			if len(matches) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "No configuration options found with prefix: %s\n", configPrefix)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Configuration options with prefix '%s':\n\n", configPrefix)
			for _, opt := range matches {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s (%s): %s\n", opt.Key, opt.Type, opt.Description)
			}
			return nil
		}

		// No flags specified - show help
		return cmd.Help()
	},
}

func init() {
	// Add flags
	devConfigCmd.Flags().BoolVarP(&configList, "list", "l", false, "List all configuration options")
	devConfigCmd.Flags().BoolVarP(&configShow, "show", "s", false, "Show effective configuration values")
	devConfigCmd.Flags().StringVarP(&configExport, "export", "e", "", "Export configuration (format: json)")
	devConfigCmd.Flags().BoolVarP(&configValidate, "validate", "v", false, "Validate current configuration")
	devConfigCmd.Flags().StringVarP(&configPrefix, "prefix", "p", "", "Show config options with specific prefix")

	// Add to dev command
	devCmd.AddCommand(devConfigCmd)
}
