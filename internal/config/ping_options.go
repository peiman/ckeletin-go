// internal/config/ping_options.go
//
// Ping command configuration options
//
// This file contains configuration options specific to the 'ping' command.
// These settings only affect the behavior of the ping command and are not used elsewhere.

package config

// PingOptions returns configuration options for the ping command
func PingOptions() []ConfigOption {
	return []ConfigOption{
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
	RegisterOptionsProvider(PingOptions)
}
