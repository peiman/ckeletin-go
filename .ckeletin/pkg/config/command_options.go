// internal/config/options.go
//
// ConfigOption type definition and related methods
//
// This file defines the core ConfigOption type used throughout the configuration system.
// It doesn't contain any actual configuration values - those are in the specific
// options files (core_options.go, ping_options.go, docs_options.go, etc.)
//
// The Registry() function in registry.go aggregates all options from the various files.

package config

import (
	"fmt"
	"strings"
)

// ConfigOption represents a single configuration option with metadata
type ConfigOption struct {
	// Key is the Viper configuration key (e.g., "app.log_level")
	Key string

	// DefaultValue is the default value for this option
	DefaultValue interface{}

	// Description is a human-readable description of the option
	Description string

	// Type is the data type of the option (string, int, bool, etc.)
	Type string

	// Required indicates whether this option is required
	Required bool

	// Example provides an example value for documentation
	Example string

	// EnvVar is the corresponding environment variable name (computed automatically)
	EnvVar string
}

// EnvVarName returns the full environment variable name for this option,
// based on the EnvPrefix and the option's key
func (o ConfigOption) EnvVarName(prefix string) string {
	key := strings.ReplaceAll(o.Key, ".", "_")
	return fmt.Sprintf("%s_%s", prefix, strings.ToUpper(key))
}

// DefaultValueString returns a string representation of the default value
func (o ConfigOption) DefaultValueString() string {
	if o.DefaultValue == nil {
		return "nil"
	}
	return fmt.Sprintf("%v", o.DefaultValue)
}

// ExampleValueString returns a string representation of the example value,
// or the default value if no example is provided
func (o ConfigOption) ExampleValueString() string {
	if o.Example != "" {
		return o.Example
	}
	return o.DefaultValueString()
}
