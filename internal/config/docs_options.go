// internal/config/docs_options.go
//
// Docs command configuration options
//
// This file contains configuration options specific to the 'docs' command.
// These settings only affect the behavior of the docs command and are not used elsewhere.

package config

// DocsOptions returns configuration options for the docs command
func DocsOptions() []ConfigOption {
	return []ConfigOption{
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
	RegisterOptionsProvider(DocsOptions)
}
