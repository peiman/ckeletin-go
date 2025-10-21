// cmd/docs.go

package cmd

import (
	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// docsCmd represents the docs command
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation",
	Long:  `Generate documentation about the application, including configuration options.`,
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Generate configuration documentation",
	Long: `Generate documentation about all configuration options.

This command generates detailed documentation about all available configuration
options, including their default values, types, and environment variable names.

The documentation can be output in various formats using the --format flag.`,
	RunE: runDocsConfig,
}

func init() {
	// Add the config subcommand to the docs command
	docsCmd.AddCommand(configCmd)

	// Add the docs command to the root command
	RootCmd.AddCommand(docsCmd)

	// Auto-register flags from config registry for app.docs.* keys.
	RegisterFlagsForPrefixWithOverrides(configCmd, "app.docs.", map[string]string{
		"app.docs.output_format": "format",
		"app.docs.output_file":   "output",
	})

	// Setup command configuration inheritance
	setupCommandConfig(configCmd)
}

func runDocsConfig(cmd *cobra.Command, args []string) error {
	// Get configuration values from Viper by key (flags already bound)
	outputFormat := getKeyValue[string]("app.docs.output_format")
	outputFile := getKeyValue[string]("app.docs.output_file")

	log.Debug().
		Str("format", outputFormat).
		Str("output_file", outputFile).
		Msg("Documentation configuration loaded")

	// Create application info for the documentation generator
	appInfo := docs.AppInfo{
		BinaryName: binaryName,
		EnvPrefix:  EnvPrefix(),
	}

	// Set config paths
	paths := ConfigPaths()
	appInfo.ConfigPaths.DefaultPath = paths.DefaultPath
	appInfo.ConfigPaths.DefaultFullName = paths.DefaultFullName

	// Create document generator configuration
	cfg := docs.NewConfig(
		cmd.OutOrStdout(),
		docs.WithOutputFormat(outputFormat),
		docs.WithOutputFile(outputFile),
	)

	// Create generator and generate documentation
	generator := docs.NewGenerator(cfg)
	generator.SetAppInfo(appInfo)
	return generator.Generate()
}

// Options for the docs command live in internal/config/docs_options.go
