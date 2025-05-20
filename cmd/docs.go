// cmd/docs.go

package cmd

import (
	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	// Add flags to config command
	configCmd.Flags().StringP("format", "f", docs.FormatMarkdown, "Output format (markdown, yaml)")
	configCmd.Flags().StringP("output", "o", "", "Output file (defaults to stdout)")

	// Bind flags to Viper using consistent naming convention
	if err := viper.BindPFlag("app.docs.output_format", configCmd.Flags().Lookup("format")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'format' flag")
	}
	if err := viper.BindPFlag("app.docs.output_file", configCmd.Flags().Lookup("output")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'output' flag")
	}

	// Setup command configuration inheritance
	setupCommandConfig(configCmd)
}

func runDocsConfig(cmd *cobra.Command, args []string) error {
	// Get configuration values from viper/flags
	outputFormat := getConfigValue[string](cmd, "format", "app.docs.output_format")
	outputFile := getConfigValue[string](cmd, "output", "app.docs.output_file")

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
