// internal/config/commands/docs_config.go
//
// Docs command configuration: metadata + options
//
// This file is the single source of truth for the docs command configuration.
// It combines command metadata (Use, Short, Long, flags) with configuration options.

package commands

import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"

// DocsConfigMetadata defines all metadata for the docs config subcommand
var DocsConfigMetadata = config.CommandMetadata{
	Use:   "config",
	Short: "Generate configuration documentation",
	Long: `Generate documentation about all configuration options.

This command generates detailed documentation about all available configuration
options, including their default values, types, and environment variable names.

The documentation can be output in various formats using the --format flag.`,
	ConfigPrefix: "app.docs",
	FlagOverrides: map[string]string{
		"app.docs.output_format": "format",
		"app.docs.output_file":   "output",
	},
	Examples: []string{
		"docs config",
		"docs config --format yaml",
		"docs config --output docs/config.md",
	},
	SeeAlso: []string{"ping"},
}

// DocsOptions returns configuration options for the docs command
func DocsOptions() []config.ConfigOption {
	return []config.ConfigOption{
		{
			Key:          "app.docs.output_format",
			DefaultValue: "markdown",
			Description:  "Output format for documentation (markdown, yaml)",
			Type:         "string",
			Required:     false,
			Example:      "yaml",
		},
		{
			Key:          "app.docs.output_file",
			DefaultValue: "",
			Description:  "Output file for documentation (defaults to stdout)",
			Type:         "string",
			Required:     false,
			Example:      "/path/to/output.md",
		},
	}
}

// Self-register docs options provider at init time
func init() {
	config.RegisterOptionsProvider(DocsOptions)
}
