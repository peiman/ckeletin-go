// internal/config/registry.go
//
// THIS FILE IS THE SINGLE SOURCE OF TRUTH FOR APPLICATION CONFIGURATION
//
// IMPORTANT: All default values and configuration options MUST be defined here.
// Never use viper.SetDefault() directly in command files or other code.
//
// The Registry() function here aggregates configuration options from all command-specific
// configuration files. To add a new command's configuration, create a new file with a
// function that returns []ConfigOption and add it to the Registry() function.

package config

import (
	"github.com/spf13/viper"
)

// Registry returns a list of all known configuration options
// This is the single source of truth for all configuration options
func Registry() []ConfigOption {
	// Start with application-wide core options
	// These affect the entire application regardless of command
	options := CoreOptions()

	// Append command-specific options
	// These only affect their respective commands
	options = append(options, PingOptions()...) // ping command options
	options = append(options, DocsOptions()...) // docs command options

	// Add any other command options here as they are implemented:
	// options = append(options, NewCommandOptions()...)

	return options
}

// SetDefaults applies all default values from the registry to Viper
func SetDefaults() {
	for _, opt := range Registry() {
		viper.SetDefault(opt.Key, opt.DefaultValue)
	}
}
