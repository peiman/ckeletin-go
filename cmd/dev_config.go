//go:build dev

// ckeletin:allow-custom-command
// cmd/dev_config.go
//
// Configuration inspector subcommand (dev-only).
// Provides utilities to inspect, validate, and export configuration.

package cmd

import (
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
		flags := dev.ConfigCommandFlags{
			List:     configList,
			Show:     configShow,
			Export:   configExport,
			Validate: configValidate,
			Prefix:   configPrefix,
		}
		handled, err := dev.RunConfigCommand(cmd.OutOrStdout(), flags)
		if err != nil {
			return err
		}
		if !handled {
			return cmd.Help()
		}
		return nil
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
