// internal/config/registry.go
//
// Centralized configuration registry.
//
// IMPORTANT:
// - All default values and configuration options are defined close to their commands
//   (e.g., internal/config/<command>_options.go).
// - Those files self-register providers here via RegisterOptionsProvider in init().
// - Never use viper.SetDefault() directly in command files or elsewhere.
//
// Thread-Safety Notes:
//
// This registry follows a safe initialization-time-only pattern:
//  1. RegisterOptionsProvider() is called only during package init() (single-threaded)
//  2. Registry() is called during startup in PersistentPreRunE (single-threaded)
//  3. SetDefaults() writes to Viper during startup only (single-threaded)
//  4. After initialization, the registry is read-only (no mutations)
//  5. Commands execute sequentially (Cobra's execution model)
//
// This pattern ensures thread-safety without requiring locks or synchronization.
// The optionsProviders slice is only appended to during init(), never during runtime.
// Viper itself is not thread-safe for writes, but our usage pattern avoids concurrent access.

package config

import (
	"github.com/spf13/viper"
)

// optionsProviders holds registered providers that return config options.
var optionsProviders []func() []ConfigOption

// RegisterOptionsProvider registers a provider for configuration options.
// Call this from an init() function in the corresponding options file.
func RegisterOptionsProvider(provider func() []ConfigOption) {
	optionsProviders = append(optionsProviders, provider)
}

// Registry returns a list of all known configuration options
// This is the single source of truth for all configuration options
func Registry() []ConfigOption {
	// Start with application-wide core options
	// These affect the entire application regardless of command
	options := CoreOptions()

	// Append command-specific options via registered providers
	for _, provider := range optionsProviders {
		options = append(options, provider()...)
	}

	return options
}

// SetDefaults applies all default values from the registry to Viper
func SetDefaults() {
	for _, opt := range Registry() {
		viper.SetDefault(opt.Key, opt.DefaultValue)
	}
}
