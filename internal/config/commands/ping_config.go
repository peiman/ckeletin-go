// internal/config/commands/ping_config.go
//
// Ping command configuration: metadata + options
//
// This file is the single source of truth for the ping command configuration.
// It combines command metadata (Use, Short, Long, flags) with configuration options.
//
// USAGE PATTERN:
// After defining options here, run `task generate:config:key-constants` to
// generate type-safe constants. Then use them in your business logic:
//
//	import (
//	    "github.com/yourmodule/.ckeletin/pkg/config"
//	    "github.com/spf13/viper"
//	)
//
//	// ✅ Correct: Use generated constant (type-safe, refactor-friendly)
//	message := viper.GetString(config.KeyAppPingOutputMessage)
//
//	// ❌ Wrong: Hardcoded string (typo-prone, no IDE autocomplete)
//	message := viper.GetString("app.ping.output_message")
//
// Generated constants are in: .ckeletin/pkg/config/keys_generated.go

package commands

import "github.com/peiman/ckeletin-go/.ckeletin/pkg/config"

// PingMetadata defines all metadata for the ping command
var PingMetadata = config.CommandMetadata{
	Use:   "ping",
	Short: "Responds with a pong",
	Long: `The ping command demonstrates configuration, logging, and optional Bubble Tea UI.
- Without arguments, prints "Pong".
- Supports overriding its output and an optional interactive UI.`,
	ConfigPrefix: "app.ping",
	FlagOverrides: map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	},
	Examples: []string{
		"ping",
		"ping --message 'Hello World!'",
		"ping --color green",
		"ping --ui",
	},
	SeeAlso: []string{"docs"},
}

// PingOptions returns configuration options for the ping command
func PingOptions() []config.ConfigOption {
	return []config.ConfigOption{
		{
			Key:          "app.ping.output_message",
			DefaultValue: "Pong",
			Description:  "Default message to display for the ping command",
			Type:         "string",
			ShortFlag:    "m", // Enables: -m "Hello" (shorthand for --message)
			Required:     false,
			Example:      "Hello World!",
		},
		{
			Key:          "app.ping.output_color",
			DefaultValue: "white",
			Description:  "Text color for ping command output (white, red, green, blue, cyan, yellow, magenta)",
			Type:         "string",
			ShortFlag:    "c", // Enables: -c green (shorthand for --color)
			Required:     false,
			Example:      "green",
		},
		{
			Key:          "app.ping.ui",
			DefaultValue: false,
			Description:  "Enable interactive UI for the ping command",
			Type:         "bool",
			// ShortFlag omitted - no short flag for this option
			Required: false,
			Example:  "true",
		},
	}
}

// Self-register ping options provider at init time
func init() {
	config.RegisterOptionsProvider(PingOptions)
}
