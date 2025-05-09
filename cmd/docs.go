// cmd/docs.go

package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// Define output format types
const (
	FormatMarkdown = "markdown"
	FormatYAML     = "yaml"
)

var (
	// Flag for output format
	docsOutputFormat string
	// Flag for output file
	docsOutputFile string
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

// Variable to mock file opening for testing
var openOutputFile = func(path string) (io.WriteCloser, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
}

func init() {
	// Add the config subcommand to the docs command
	docsCmd.AddCommand(configCmd)

	// Add the docs command to the root command
	RootCmd.AddCommand(docsCmd)

	// Add flags to config command
	configCmd.Flags().StringVarP(&docsOutputFormat, "format", "f", FormatMarkdown,
		"Output format (markdown, yaml)")
	configCmd.Flags().StringVarP(&docsOutputFile, "output", "o", "",
		"Output file (defaults to stdout)")
}

func runDocsConfig(cmd *cobra.Command, args []string) error {
	var writer io.Writer = cmd.OutOrStdout()
	var file io.WriteCloser
	var closeErr error
	
	// If output file is specified, create it
	if docsOutputFile != "" {
		var err error
		file, err = openOutputFile(docsOutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer func() {
			// Capture the close error so we can return it
			closeErr = file.Close()
			if closeErr != nil {
				log.Error().Err(closeErr).Str("file", docsOutputFile).Msg("Failed to close output file")
			}
		}()
		writer = file
		log.Info().Str("file", docsOutputFile).Msg("Writing documentation to file")
	}

	var err error
	switch strings.ToLower(docsOutputFormat) {
	case FormatMarkdown:
		err = generateMarkdownDocs(writer)
	case FormatYAML:
		err = generateYAMLConfig(writer)
	default:
		err = fmt.Errorf("unsupported format: %s", docsOutputFormat)
	}

	// If there was no error from the operation but there was a close error, return the close error
	if err == nil && closeErr != nil {
		return fmt.Errorf("failed to close output file: %w", closeErr)
	}

	return err
}

// generateMarkdownDocs generates Markdown documentation for all configuration options
func generateMarkdownDocs(w io.Writer) error {
	// Get environment variable prefix
	envPrefix := EnvPrefix()

	// Get configurations paths
	configPaths := ConfigPaths()

	// Write header
	fmt.Fprintf(w, "# %s Configuration\n\n", binaryName)
	fmt.Fprintf(w, "This document describes all available configuration options for %s.\n\n", binaryName)

	// Configuration sources
	fmt.Fprintf(w, "## Configuration Sources\n\n")
	fmt.Fprintf(w, "Configuration can be provided in multiple ways, in order of precedence:\n\n")
	fmt.Fprintf(w, "1. Command-line flags\n")
	fmt.Fprintf(w, "2. Environment variables (with prefix `%s_`)\n", envPrefix)
	fmt.Fprintf(w, "3. Configuration file (%s)\n", configPaths.DefaultPath)
	fmt.Fprintf(w, "4. Default values\n\n")

	// Configuration options
	fmt.Fprintf(w, "## Configuration Options\n\n")

	// Table header
	fmt.Fprintf(w, "| Key | Type | Default | Environment Variable | Description |\n")
	fmt.Fprintf(w, "|-----|------|---------|---------------------|-------------|\n")

	// Table rows for each option
	registry := config.Registry()
	for _, opt := range registry {
		defaultVal := fmt.Sprintf("`%v`", opt.DefaultValueString())
		required := ""
		if opt.Required {
			required = " (Required)"
		}

		fmt.Fprintf(w, "| `%s` | %s | %s | `%s` | %s%s |\n",
			opt.Key,
			opt.Type,
			defaultVal,
			opt.EnvVarName(envPrefix),
			opt.Description,
			required,
		)
	}

	// Example configuration file
	fmt.Fprintf(w, "\n## Example Configuration\n\n")
	fmt.Fprintf(w, "### YAML Configuration File (%s)\n\n", configPaths.DefaultFullName)
	fmt.Fprintf(w, "```yaml\n")

	// Group options by top-level key for a nicer YAML structure
	groups := make(map[string][]config.ConfigOption)
	for _, opt := range registry {
		parts := strings.SplitN(opt.Key, ".", 2)
		if len(parts) > 1 {
			topLevel := parts[0]
			groups[topLevel] = append(groups[topLevel], opt)
		} else {
			groups[""] = append(groups[""], opt)
		}
	}

	// Generate YAML example
	for topLevel, options := range groups {
		if topLevel != "" {
			fmt.Fprintf(w, "%s:\n", topLevel)
		}

		for _, opt := range options {
			parts := strings.SplitN(opt.Key, ".", 2)
			key := opt.Key
			if len(parts) > 1 {
				key = parts[1]
			}

			fmt.Fprintf(w, "  # %s\n", opt.Description)
			fmt.Fprintf(w, "  %s: %s\n\n", key, opt.ExampleValueString())
		}
	}

	fmt.Fprintf(w, "```\n\n")

	// Environment variables
	fmt.Fprintf(w, "### Environment Variables\n\n")
	fmt.Fprintf(w, "```bash\n")
	for _, opt := range registry {
		fmt.Fprintf(w, "# %s\n", opt.Description)
		fmt.Fprintf(w, "export %s=%s\n\n", opt.EnvVarName(envPrefix), opt.ExampleValueString())
	}
	fmt.Fprintf(w, "```\n")

	return nil
}

// generateYAMLConfig generates a YAML configuration template
func generateYAMLConfig(w io.Writer) error {
	registry := config.Registry()

	// Group options by top-level key for a nicer YAML structure
	groups := make(map[string][]config.ConfigOption)
	for _, opt := range registry {
		parts := strings.SplitN(opt.Key, ".", 2)
		if len(parts) > 1 {
			topLevel := parts[0]
			groups[topLevel] = append(groups[topLevel], opt)
		} else {
			groups[""] = append(groups[""], opt)
		}
	}

	// Generate YAML
	for topLevel, options := range groups {
		if topLevel != "" {
			fmt.Fprintf(w, "%s:\n", topLevel)
		}

		for _, opt := range options {
			parts := strings.SplitN(opt.Key, ".", 2)
			key := opt.Key
			if len(parts) > 1 {
				key = parts[1]
			}

			fmt.Fprintf(w, "  # %s\n", opt.Description)
			fmt.Fprintf(w, "  %s: %s\n\n", key, opt.ExampleValueString())
		}
	}

	return nil
}
