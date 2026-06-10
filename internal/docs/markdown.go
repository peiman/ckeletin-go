// internal/docs/markdown.go

package docs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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

// NewAppInfo assembles the AppInfo the docs generator embeds. defaultConfigDir
// is the user config directory ("" when none applies); the default config path
// is derived from it.
func NewAppInfo(binaryName, envPrefix, defaultConfigDir string) AppInfo {
	info := AppInfo{
		BinaryName: binaryName,
		EnvPrefix:  envPrefix,
	}
	if defaultConfigDir != "" {
		info.ConfigPaths.DefaultPath = filepath.Join(defaultConfigDir, "config.yaml")
	}
	info.ConfigPaths.DefaultFullName = "config.yaml" // local project config
	return info
}

// GenerateMarkdownDocs generates Markdown documentation for all configuration options
func (g *Generator) GenerateMarkdownDocs(w io.Writer, appInfo AppInfo) error {
	ew := &errWriter{w: w}

	// Write header
	ew.printf("# %s Configuration\n\n", appInfo.BinaryName)
	ew.printf("This document describes all available configuration options for %s.\n\n", appInfo.BinaryName)

	// Configuration sources
	ew.printf("## Configuration Sources\n\n")
	ew.printf("Configuration can be provided in multiple ways, in order of precedence:\n\n")
	ew.printf("1. Command-line flags\n")
	ew.printf("2. Environment variables (with prefix `%s_`)\n", appInfo.EnvPrefix)
	ew.printf("3. Configuration file (%s)\n", sanitizeConfigPath(appInfo.ConfigPaths.DefaultPath))
	ew.printf("4. Default values\n\n")

	// Configuration options
	ew.printf("## Configuration Options\n\n")

	// Table header
	ew.printf("| Key | Type | Default | Environment Variable | Description |\n")
	ew.printf("|-----|------|---------|---------------------|-------------|\n")

	// Table rows for each option
	registry := g.cfg.Registry()
	for _, opt := range registry {
		defaultVal := fmt.Sprintf("`%v`", opt.DefaultValueString())
		required := ""
		if opt.Required {
			required = " (Required)"
		}

		ew.printf("| `%s` | %s | %s | `%s` | %s%s |\n",
			opt.Key,
			opt.Type,
			defaultVal,
			opt.EnvVarName(appInfo.EnvPrefix),
			opt.Description,
			required,
		)
	}

	// Example configuration file
	ew.printf("\n## Example Configuration\n\n")
	ew.printf("### YAML Configuration File (%s)\n\n", appInfo.ConfigPaths.DefaultFullName)
	ew.printf("```yaml\n")

	// Group options by top-level key for a nicer YAML structure
	if ew.err == nil {
		if err := generateYAMLContentFunc(w, registry); err != nil {
			return fmt.Errorf("failed to generate YAML content: %w", err)
		}
	}

	ew.printf("```\n\n")

	// Environment variables
	ew.printf("### Environment Variables\n\n")
	ew.printf("```bash\n")
	for _, opt := range registry {
		ew.printf("# %s\n", opt.Description)
		ew.printf("export %s=%s\n\n", opt.EnvVarName(appInfo.EnvPrefix), opt.ExampleValueString())
	}
	ew.printf("```\n")

	if ew.err != nil {
		return fmt.Errorf("failed to write markdown documentation: %w", ew.err)
	}
	return nil
}

// sanitizeConfigPath replaces user-specific home directories with tilde notation
// This ensures generated documentation doesn't contain user-specific paths like /Users/username
func sanitizeConfigPath(path string) string {
	// Get user's home directory
	home := os.Getenv("HOME")
	if home != "" && strings.HasPrefix(path, home) {
		// Replace home directory with ~
		return strings.Replace(path, home, "~", 1)
	}

	// Also handle /Users/ or /home/ patterns even if HOME isn't set
	if strings.Contains(path, "/Users/") || strings.Contains(path, "/home/") {
		// Extract just the filename if it starts with a home-like path
		parts := strings.Split(path, "/")
		if len(parts) > 0 {
			filename := parts[len(parts)-1]
			return "~/" + filename
		}
	}

	return path
}
