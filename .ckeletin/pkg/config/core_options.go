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
		// Log rotation options (lumberjack)
		{
			Key:          "app.log.file_max_size",
			DefaultValue: 100,
			Description:  "Maximum size in megabytes before log file is rotated",
			Type:         "int",
			Required:     false,
			Example:      "100",
		},
		{
			Key:          "app.log.file_max_backups",
			DefaultValue: 3,
			Description:  "Maximum number of old log files to retain",
			Type:         "int",
			Required:     false,
			Example:      "3",
		},
		{
			Key:          "app.log.file_max_age",
			DefaultValue: 28,
			Description:  "Maximum number of days to retain old log files",
			Type:         "int",
			Required:     false,
			Example:      "28",
		},
		{
			Key:          "app.log.file_compress",
			DefaultValue: false,
			Description:  "Compress rotated log files with gzip",
			Type:         "bool",
			Required:     false,
			Example:      "true",
		},
		// Log sampling options
		{
			Key:          "app.log.sampling_enabled",
			DefaultValue: false,
			Description:  "Enable log sampling for high-volume scenarios",
			Type:         "bool",
			Required:     false,
			Example:      "true",
		},
		{
			Key:          "app.log.sampling_initial",
			DefaultValue: 100,
			Description:  "Number of messages to log per second before sampling",
			Type:         "int",
			Required:     false,
			Example:      "100",
		},
		{
			Key:          "app.log.sampling_thereafter",
			DefaultValue: 100,
			Description:  "Number of messages to log thereafter per second",
			Type:         "int",
			Required:     false,
			Example:      "100",
		},
		// Add other application-wide settings here
	}
}
