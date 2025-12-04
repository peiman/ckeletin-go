// internal/config/commands/ping_config.go
//
// Ping command configuration: metadata + options
//
// This file is the single source of truth for the ping command configuration.
// It combines command metadata (Use, Short, Long, flags) with configuration options.

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

// Self-register ping options provider at init time
func init() {
	config.RegisterOptionsProvider(PingOptions)
}
