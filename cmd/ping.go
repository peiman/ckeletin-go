// cmd/ping.go

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// NewPingCommand creates a new `ping` command with a customizable UIRunner
func NewPingCommand(uiRunner ui.UIRunner) *cobra.Command {
	// Initialize command-specific defaults and configurations
	initPingConfig()

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Responds with a pong",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get values from flags or configuration
			message := viper.GetString("app.ping.output_message")
			colorStr := viper.GetString("app.ping.output_color")
			uiFlag := viper.GetBool("app.ping.ui")

			log.Info().
				Str("command", "ping").
				Bool("ui_enabled", uiFlag).
				Msg("Ping command invoked")

			// Log configuration details
			log.Debug().
				Str("message", message).
				Str("color", colorStr).
				Bool("ui_enabled", uiFlag).
				Msg("Command configuration loaded")

			if uiFlag {
				// Log that the UI is starting
				log.Info().
					Str("message", message).
					Str("color", colorStr).
					Msg("Starting UI")

				// Run the UI
				if err := uiRunner.RunUI(message, colorStr); err != nil {
					log.Error().
						Err(err).
						Str("message", message).
						Str("color", colorStr).
						Msg("Failed to run UI")
					return err
				}

				log.Info().
					Str("message", message).
					Str("color", colorStr).
					Msg("UI executed successfully")
			} else {
				// Log that we're printing the colored message
				log.Info().
					Str("message", message).
					Str("color", colorStr).
					Msg("Printing colored message")

				// Print the message
				if err := ui.PrintColoredMessage(cmd.OutOrStdout(), message, colorStr); err != nil {
					log.Error().
						Err(err).
						Str("message", message).
						Str("color", colorStr).
						Msg("Failed to print colored message")
					return err
				}

				log.Info().
					Str("message", message).
					Str("color", colorStr).
					Msg("Colored message printed successfully")
			}

			return nil
		},
	}

	// Define flags specific to the ping command
	cmd.Flags().String("message", "", "Custom output message")
	cmd.Flags().String("color", "", "Output color")
	cmd.Flags().Bool("ui", false, "Enable UI")

	// Bind flags to Viper
	if err := viper.BindPFlag("app.ping.output_message", cmd.Flags().Lookup("message")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'message': %v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("app.ping.output_color", cmd.Flags().Lookup("color")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'color': %v\n", err)
		os.Exit(1)
	}

	if err := viper.BindPFlag("app.ping.ui", cmd.Flags().Lookup("ui")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'ui': %v\n", err)
		os.Exit(1)
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
