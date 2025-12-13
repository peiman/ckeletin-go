// cmd/docs.go

package cmd

import (
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/commands"
	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// docsCmd represents the docs command (parent command)
var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Generate documentation",
	Long:  `Generate documentation about the application, including configuration options.`,
}

var docsConfigCmd = MustNewCommand(commands.DocsConfigMetadata, runDocsConfig)

func init() {
	docsCmd.AddCommand(docsConfigCmd)
	setupCommandConfig(docsConfigCmd)
	MustAddToRoot(docsCmd)
}

func runDocsConfig(cmd *cobra.Command, args []string) error {
	// Get configuration values from Viper by key (flags already bound)
	outputFormat := getKeyValue[string](config.KeyAppDocsOutputFormat)
	outputFile := getKeyValue[string](config.KeyAppDocsOutputFile)

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
	if paths.XDGDir != "" {
		appInfo.ConfigPaths.DefaultPath = paths.XDGDir + "/config.yaml"
	}
	appInfo.ConfigPaths.DefaultFullName = "config.yaml" // local project config

	// Create document generator configuration
	cfg := docs.Config{
		Writer:       cmd.OutOrStdout(),
		OutputFormat: outputFormat,
		OutputFile:   outputFile,
		Registry:     config.Registry,
	}

	// Create generator and generate documentation
	generator := docs.NewGenerator(cfg)
	generator.SetAppInfo(appInfo)
	return generator.Generate()
}
