// internal/config/registry.go
//
// THIS FILE IS THE SINGLE SOURCE OF TRUTH FOR APPLICATION CONFIGURATION
//
// IMPORTANT: All default values and configuration options MUST be defined here.
// Never use viper.SetDefault() directly in command files or other code.
//
// To add a new configuration option:
// 1. Add it to the Registry() function below with all metadata
// 2. Use viper.Get*() functions to access it in your code
// 3. The default will be automatically applied

package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
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

// Registry returns a list of all known configuration options
func Registry() []ConfigOption {
	return []ConfigOption{
		{
			Key:          "app.log_level",
			DefaultValue: "info",
			Description:  "Logging level for the application (trace, debug, info, warn, error, fatal, panic)",
			Type:         "string",
			Required:     false,
			Example:      "debug",
		},
		{
			Key:          "app.ping.output_message",
			DefaultValue: "Pong",
			Description:  "Default message to display for the ping command",
			Type:         "string",
			Required:     false,
			Example:      "Hello World!",
		},
		{
			Key:          "app.ping.output_color",
			DefaultValue: "white",
			Description:  "Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)",
			Type:         "string",
			Required:     false,
			Example:      "green",
		},
		{
			Key:          "app.ping.ui",
			DefaultValue: false,
			Description:  "Enable interactive UI for the ping command",
			Type:         "bool",
			Required:     false,
			Example:      "true",
		},
	}
}

// SetDefaults applies all default values from the registry to Viper
func SetDefaults() {
	for _, opt := range Registry() {
		viper.SetDefault(opt.Key, opt.DefaultValue)
	}
}
