// internal/config/core_options.go
//
// Core application configuration options
//
// This file contains application-wide configuration options that apply across
// all commands and are not specific to any particular command.
// These are fundamental settings like logging level that affect the entire application.

package config

// CoreOptions returns core application configuration options
// These settings affect the overall behavior of the application
func CoreOptions() []ConfigOption {
	return []ConfigOption{
		{
			Key:          "app.log_level",
			DefaultValue: "info",
			Description:  "Logging level for the application (trace, debug, info, warn, error, fatal, panic)",
			Type:         "string",
			Required:     false,
			Example:      "debug",
		},
		// Add other application-wide settings here
	}
}
