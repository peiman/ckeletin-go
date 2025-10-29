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
		// Legacy log level option (maintained for backward compatibility)
		// If app.log.console_level is not set, this value is used for console output
		{
			Key:          "app.log_level",
			DefaultValue: "info",
			Description:  "Logging level for the application (trace, debug, info, warn, error, fatal, panic). Used as console level if app.log.console_level is not set.",
			Type:         "string",
			Required:     false,
			Example:      "debug",
		},

		// Dual logging configuration options
		{
			Key:          "app.log.console_level",
			DefaultValue: "",
			Description:  "Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses app.log_level.",
			Type:         "string",
			Required:     false,
			Example:      "info",
		},
		{
			Key:          "app.log.file_enabled",
			DefaultValue: false,
			Description:  "Enable file logging to capture detailed logs",
			Type:         "bool",
			Required:     false,
			Example:      "true",
		},
		{
			Key:          "app.log.file_path",
			DefaultValue: "./logs/ckeletin-go.log",
			Description:  "Path to the log file (created with secure 0600 permissions)",
			Type:         "string",
			Required:     false,
			Example:      "/var/log/ckeletin-go/app.log",
		},
		{
			Key:          "app.log.file_level",
			DefaultValue: "debug",
			Description:  "File log level (trace, debug, info, warn, error, fatal, panic)",
			Type:         "string",
			Required:     false,
			Example:      "debug",
		},
		{
			Key:          "app.log.color_enabled",
			DefaultValue: "auto",
			Description:  "Enable colored console output (auto, true, false). Auto detects TTY.",
			Type:         "string",
			Required:     false,
			Example:      "true",
		},
		// Add other application-wide settings here
	}
}
