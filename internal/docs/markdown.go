// internal/docs/markdown.go

package docs

import (
	"fmt"
	"io"
)

// AppInfo contains information about the application needed for documentation generation
type AppInfo struct {
	BinaryName  string
	EnvPrefix   string
	ConfigPaths struct {
		DefaultPath     string
		DefaultFullName string
	}
}

// GenerateMarkdownDocs generates Markdown documentation for all configuration options
func (g *Generator) GenerateMarkdownDocs(w io.Writer, appInfo AppInfo) error {
	// Write header
	fmt.Fprintf(w, "# %s Configuration\n\n", appInfo.BinaryName)
	fmt.Fprintf(w, "This document describes all available configuration options for %s.\n\n", appInfo.BinaryName)

	// Configuration sources
	fmt.Fprintf(w, "## Configuration Sources\n\n")
	fmt.Fprintf(w, "Configuration can be provided in multiple ways, in order of precedence:\n\n")
	fmt.Fprintf(w, "1. Command-line flags\n")
	fmt.Fprintf(w, "2. Environment variables (with prefix `%s_`)\n", appInfo.EnvPrefix)
	fmt.Fprintf(w, "3. Configuration file (%s)\n", appInfo.ConfigPaths.DefaultPath)
	fmt.Fprintf(w, "4. Default values\n\n")

	// Configuration options
	fmt.Fprintf(w, "## Configuration Options\n\n")

	// Table header
	fmt.Fprintf(w, "| Key | Type | Default | Environment Variable | Description |\n")
	fmt.Fprintf(w, "|-----|------|---------|---------------------|-------------|\n")

	// Table rows for each option
	registry := g.cfg.Registry()
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
			opt.EnvVarName(appInfo.EnvPrefix),
			opt.Description,
			required,
		)
	}

	// Example configuration file
	fmt.Fprintf(w, "\n## Example Configuration\n\n")
	fmt.Fprintf(w, "### YAML Configuration File (%s)\n\n", appInfo.ConfigPaths.DefaultFullName)
	fmt.Fprintf(w, "```yaml\n")

	// Group options by top-level key for a nicer YAML structure
	if err := generateYAMLContentFunc(w, registry); err != nil {
		return fmt.Errorf("failed to generate YAML content: %w", err)
	}

	fmt.Fprintf(w, "```\n\n")

	// Environment variables
	fmt.Fprintf(w, "### Environment Variables\n\n")
	fmt.Fprintf(w, "```bash\n")
	for _, opt := range registry {
		fmt.Fprintf(w, "# %s\n", opt.Description)
		fmt.Fprintf(w, "export %s=%s\n\n", opt.EnvVarName(appInfo.EnvPrefix), opt.ExampleValueString())
	}
	fmt.Fprintf(w, "```\n")

	return nil
}
