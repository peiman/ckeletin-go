// internal/config/command_metadata.go
//
// This file defines the CommandMetadata type used for command metadata.
// Command metadata provides all the information needed to create a Cobra command
// and generate documentation automatically.

package config

// CommandMetadata holds all metadata for a command
// This is the single source of truth for command documentation and configuration
type CommandMetadata struct {
	// Use is the one-line usage message (e.g., "ping", "docs config")
	Use string

	// Short is a short description shown in the 'help' output
	Short string

	// Long is the long message shown in the 'help <this-command>' output
	Long string

	// ConfigPrefix is the configuration key prefix for this command (e.g., "app.ping")
	ConfigPrefix string

	// FlagOverrides maps config keys to custom flag names
	// Example: map[string]string{"app.ping.output_message": "message"}
	// This allows short flag names instead of using the full key suffix
	FlagOverrides map[string]string

	// Examples are example command invocations for documentation
	// Example: []string{"ping", "ping --message 'Hello'"}
	Examples []string

	// SeeAlso lists related commands for documentation
	// Example: []string{"docs", "version"}
	SeeAlso []string

	// Hidden indicates if the command should be hidden from help output
	Hidden bool
}
