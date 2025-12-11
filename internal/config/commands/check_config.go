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
	Long: `Run code quality checks using checkmate.

Includes the following checks:
  - format: Check code formatting (goimports + gofmt)
  - lint: Run linters (go vet + golangci-lint)
  - test: Run tests with race detection
  - deps: Verify dependency integrity
  - vuln: Scan for vulnerabilities

Use --fail-fast to stop on the first failure.`,
	ConfigPrefix: "app.check",
	FlagOverrides: map[string]string{
		"app.check.fail_fast": "fail-fast",
		"app.check.verbose":   "verbose",
	},
	Examples: []string{
		"check",
		"check --fail-fast",
		"check -v",
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
	}
}

// Self-register check options provider at init time
func init() {
	config.RegisterOptionsProvider(CheckOptions)
}
