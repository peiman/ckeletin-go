package cmd

import (
	"strings"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewPingCommand creates a new `ping` command with a customizable UIRunner
func NewPingCommand(uiRunner ui.UIRunner) *cobra.Command {
	var message, colorStr string
	var uiFlag bool

	// Initialize command-specific defaults and configurations
	initPingConfig()

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Responds with a pong",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Info().
				Str("command", "ping").
				Bool("ui_enabled", uiFlag).
				Msg("Ping command invoked")

			// Get message and color from configuration
			msg := viper.GetString("app.ping.output_message")
			col := viper.GetString("app.ping.output_color")
			uiFlag := viper.GetBool("app.ping.ui")

			// Log configuration details
			log.Debug().
				Str("message", msg).
				Str("color", col).
				Bool("ui_enabled", uiFlag).
				Msg("Command configuration loaded")

			if uiFlag {
				// Log that the UI is starting
				log.Info().
					Str("message", msg).
					Str("color", col).
					Msg("Starting UI")

				// Run the UI
				if err := uiRunner.RunUI(msg, col); err != nil {
					log.Error().
						Err(err).
						Str("message", msg).
						Str("color", col).
						Msg("Failed to run UI")
					return err
				}

				log.Info().
					Str("message", msg).
					Str("color", col).
					Msg("UI executed successfully")
			} else {
				// Log that we're printing the colored message
				log.Info().
					Str("message", msg).
					Str("color", col).
					Msg("Printing colored message")

				// Print the message
				if err := ui.PrintColoredMessage(cmd.OutOrStdout(), msg, col); err != nil {
					log.Error().
						Err(err).
						Str("message", msg).
						Str("color", col).
						Msg("Failed to print colored message")
					return err
				}

				log.Info().
					Str("message", msg).
					Str("color", col).
					Msg("Colored message printed successfully")
			}

			return nil
		},
	}

	// Define flags specific to the ping command
	cmd.Flags().StringVarP(&message, "message", "m", "", "Custom output message")
	cmd.Flags().StringVarP(&colorStr, "color", "c", "", "Output color")
	cmd.Flags().BoolVarP(&uiFlag, "ui", "", false, "Enable UI")

	// Bind flags
	if err := viper.BindPFlag("app.ping.output_message", cmd.Flags().Lookup("message")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'message'")
	}

	if err := viper.BindPFlag("app.ping.output_color", cmd.Flags().Lookup("color")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'color'")
	}

	if err := viper.BindPFlag("app.ping.ui", cmd.Flags().Lookup("ui")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'ui'")
	}

	return cmd
}

func initPingConfig() {
	// Handle environment variables specific to the ping command
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Set default values for ping command configurations
	viper.SetDefault("app.ping.output_message", "Pong")
	viper.SetDefault("app.ping.output_color", "white")
	viper.SetDefault("app.ping.ui", false)
}
