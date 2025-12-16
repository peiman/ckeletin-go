// internal/config/commands/check_config.go
//
// Check command configuration: metadata + options
//
// This file defines the configuration for the 'check' command which runs
// quality checks using pkg/checkmate.

package commands

import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"

// CheckMetadata defines all metadata for the check command
var CheckMetadata = config.CommandMetadata{
	Use:   "check",
	Short: "Run quality checks",
	Long: `Run comprehensive code quality checks using checkmate.

Includes 22 checks across 6 categories:

  Development Environment (2):
    go-version, tools

  Code Quality (2):
    format, lint

  Architecture Validation (10):
    defaults, commands, constants, task-naming, architecture,
    layering, package-org, config-consumption, output-patterns,
    security-patterns

  Security Scanning (2):
    secrets, sast

  Dependencies (5):
    deps, vuln, license-source, license-binary, sbom-vulns

  Tests (1):
    test

Use --parallel to run checks within each category concurrently.
Use --fail-fast to stop on the first failure.`,
	ConfigPrefix: "app.check",
	FlagOverrides: map[string]string{
		"app.check.fail_fast": "fail-fast",
		"app.check.verbose":   "verbose",
		"app.check.parallel":  "parallel",
		"app.check.category":  "category",
		"app.check.timing":    "timing",
		"app.check.tui":       "tui",
	},
	Examples: []string{
		"check",
		"check --tui",
		"check --category=security",
		"check -c security,tests",
		"check --fail-fast",
	},
	SeeAlso: []string{"docs"},
}

// CheckOptions returns configuration options for the check command
func CheckOptions() []config.ConfigOption {
	return []config.ConfigOption{
		{
			Key:          "app.check.fail_fast",
			DefaultValue: false,
			Description:  "Stop on first failed check",
			Type:         "bool",
			ShortFlag:    "f",
			Required:     false,
			Example:      "true",
		},
		{
			Key:          "app.check.verbose",
			DefaultValue: false,
			Description:  "Show verbose output including command details",
			Type:         "bool",
			ShortFlag:    "v",
			Required:     false,
			Example:      "true",
		},
		{
			Key:          "app.check.parallel",
			DefaultValue: true,
			Description:  "Run checks within each category in parallel (disable with --parallel=false)",
			Type:         "bool",
			ShortFlag:    "p",
			Required:     false,
			Example:      "false",
		},
		{
			Key:          "app.check.category",
			DefaultValue: "",
			Description:  "Filter to specific categories (comma-separated: environment,quality,architecture,security,dependencies,tests)",
			Type:         "string",
			ShortFlag:    "c",
			Required:     false,
			Example:      "security,tests",
		},
		{
			Key:          "app.check.timing",
			DefaultValue: true,
			Description:  "Show duration for each check in the output",
			Type:         "bool",
			ShortFlag:    "t",
			Required:     false,
			Example:      "false",
		},
		{
			Key:          "app.check.tui",
			DefaultValue: false,
			Description:  "Use interactive TUI with progress bars (requires TTY)",
			Type:         "bool",
			ShortFlag:    "u",
			Required:     false,
			Example:      "true",
		},
	}
}

// Self-register check options provider at init time
func init() {
	config.RegisterOptionsProvider(CheckOptions)
}
