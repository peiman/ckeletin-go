// .ckeletin/pkg/config/core_options.go
//
// Core application configuration options
//
// This file contains application-wide configuration options that apply across
// all commands and are not specific to any particular command.
// These are fundamental settings like logging level that affect the entire application.

package config

import (
	"fmt"
	"time"
)

// CoreOptions returns core application configuration options
// These settings affect the overall behavior of the application
func CoreOptions() []ConfigOption {
	return []ConfigOption{
		// Legacy log level option (maintained for backward compatibility)
		// If app.log.console_level is not set, this value is used for console output
		{
			Key:          "app.log_level",
			DefaultValue: "warn",
			Description:  "Logging level for the application (trace, debug, info, warn, error, fatal, panic). Console level when app.log.console_level is empty. It does not change the file level.",
			Type:         "string",
			Required:     false,
			Example:      "debug",
			Validation:   ValidateLogLevel(false),
		},

		// Dual logging configuration options
		{
			Key:          "app.log.console_level",
			DefaultValue: "",
			Description:  "Console log level (trace, debug, info, warn, error, fatal, panic). If empty, uses app.log_level.",
			Type:         "string",
			Required:     false,
			Example:      "info",
			Validation:   ValidateLogLevel(true),
		},
		{
			Key:          "app.log.file_enabled",
			DefaultValue: true,
			Description:  "Enable file logging to capture detailed logs. Set false to turn the audit log off.",
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
			Validation:   ValidateLogFilePath(),
		},
		{
			Key:          "app.log.file_level",
			DefaultValue: "debug",
			Description:  "File log level (trace, debug, info, warn, error, fatal, panic). Independent of app.log_level. Empty or invalid uses debug.",
			Type:         "string",
			Required:     false,
			Example:      "debug",
			Validation:   ValidateLogLevel(false),
		},
		{
			Key:          "app.log.color_enabled",
			DefaultValue: "auto",
			Description:  "Enable colored console output (auto, true, false). Auto detects TTY.",
			Type:         "string",
			Required:     false,
			Example:      "true",
			// Empty is allowed: the logger treats "" as auto-detection
			Validation: ValidateOneOf([]string{"auto", "true", "false"}, true),
		},
		{
			Key:          "app.log.console_time_format",
			DefaultValue: time.RFC3339,
			Description:  "Go time layout for console log lines. Default is RFC3339. Empty omits the console timestamp.",
			Type:         "string",
			Required:     false,
			Example:      "15:04:05",
			NoFlag:       true,
			Validation:   validateConsoleTimeFormat,
		},
		// Log rotation options (lumberjack)
		{
			Key:          "app.log.file_max_size",
			DefaultValue: 100,
			Description:  "Maximum size in megabytes before log file is rotated",
			Type:         "int",
			Required:     false,
			Example:      "100",
			Validation:   ValidatePositiveInt(),
		},
		{
			Key:          "app.log.file_max_backups",
			DefaultValue: 3,
			Description:  "Maximum number of old log files to retain",
			Type:         "int",
			Required:     false,
			Example:      "3",
			// Zero is meaningful: lumberjack retains all old log files
			Validation: ValidateNonNegativeInt(),
		},
		{
			Key:          "app.log.file_max_age",
			DefaultValue: 28,
			Description:  "Maximum number of days to retain old log files",
			Type:         "int",
			Required:     false,
			Example:      "28",
			// Zero is meaningful: lumberjack skips age-based deletion
			Validation: ValidateNonNegativeInt(),
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
			Validation:   ValidatePositiveInt(),
		},
		{
			Key:          "app.log.sampling_thereafter",
			DefaultValue: 100,
			Description:  "Number of messages to log thereafter per second",
			Type:         "int",
			Required:     false,
			Example:      "100",
			Validation:   ValidatePositiveInt(),
		},
		// Output format option (framework-level)
		{
			Key:          "app.output_format",
			DefaultValue: "text",
			Description:  "Output format: text (human-readable) or json (machine-readable)",
			Type:         "string",
			Required:     false,
			Example:      "json",
			Validation: func(value interface{}) error {
				str, ok := value.(string)
				if !ok {
					return fmt.Errorf("output_format must be a string")
				}
				switch str {
				case "text", "json":
					return nil
				default:
					return fmt.Errorf("invalid output format %q: must be \"text\" or \"json\"", str)
				}
			},
		},
		// Add other application-wide settings here
	}
}

// validateConsoleTimeFormat accepts any Go time layout, including empty.
// Empty means the console line has no timestamp. The audit file is unaffected.
func validateConsoleTimeFormat(value interface{}) error {
	if value == nil {
		return nil
	}
	if _, ok := value.(string); !ok {
		return fmt.Errorf("console_time_format must be a string")
	}
	return nil
}
