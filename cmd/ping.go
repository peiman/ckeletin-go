// cmd/ping.go

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	pingRunner ui.UIRunner = ui.NewDefaultUIRunner() // default UI runner, can be replaced in tests
)

// PingConfig holds all configuration for the ping command
// This struct is built using the Options Pattern for testability and clarity
// Defaults are set in internal/config/registry.go and loaded via Viper
// Use functional options to override values as needed
type PingConfig struct {
	Message string
	Color   string
	UI      bool
}

type PingOption func(*PingConfig)

func WithPingMessage(msg string) PingOption {
	return func(cfg *PingConfig) { cfg.Message = msg }
}
func WithPingColor(color string) PingOption {
	return func(cfg *PingConfig) { cfg.Color = color }
}
func WithPingUI(ui bool) PingOption {
	return func(cfg *PingConfig) { cfg.UI = ui }
}

// NewPingConfig builds a PingConfig from options, with values loaded from Viper/flags by default
func NewPingConfig(cmd *cobra.Command, opts ...PingOption) PingConfig {
	cfg := PingConfig{
		Message: getConfigValue[string](cmd, "message", "app.ping.output_message"),
		Color:   getConfigValue[string](cmd, "color", "app.ping.output_color"),
		UI:      getConfigValue[bool](cmd, "ui", "app.ping.ui"),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return cfg
}

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Responds with a pong",
	Long: `The ping command demonstrates configuration, logging, and optional Bubble Tea UI.
- Without arguments, prints "Pong".
- Supports overriding its output and an optional interactive UI.`,
	RunE: runPing,
}

func init() {
	// Auto-register flags from config registry for app.ping.* keys.
	RegisterFlagsForPrefixWithOverrides(pingCmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	})

	// Add pingCmd to RootCmd
	RootCmd.AddCommand(pingCmd)

	// Setup command configuration inheritance
	setupCommandConfig(pingCmd)

	// IMPORTANT: Never set defaults directly with viper.SetDefault() here.
	// All defaults MUST be defined in internal/config/registry.go
	// See internal/config/registry.go for all configuration options
}

func runPing(cmd *cobra.Command, args []string) error {
	log.Debug().Msg("Starting runPing execution")

	cfg := NewPingConfig(cmd)

	log.Debug().
		Str("message", cfg.Message).
		Str("color", cfg.Color).
		Bool("ui_enabled", cfg.UI).
		Msg("Configuration loaded")

	writer := cmd.OutOrStdout()
	log.Debug().
		Str("writer_type", fmt.Sprintf("%T", writer)).
		Msg("Using writer")

	if cfg.UI {
		log.Info().Str("message", cfg.Message).Str("color", cfg.Color).Msg("Starting UI")
		if err := pingRunner.RunUI(cfg.Message, cfg.Color); err != nil {
			log.Error().Err(err).Msg("Failed to run UI")
			return err
		}
		return nil
	}

	// Non-UI mode: print the message
	err := ui.PrintColoredMessage(writer, cfg.Message, cfg.Color)
	if err != nil {
		log.Error().
			Err(err).
			Str("message", cfg.Message).
			Str("color", cfg.Color).
			Msg("Failed to print colored message")
		// Wrap the error to provide context
		return fmt.Errorf("failed to print colored message: %w", err)
	}

	log.Debug().Msg("runPing completed successfully")
	return nil
}
