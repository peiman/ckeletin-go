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

// UIRunner interface for testing and mocking UI
type UIRunner interface {
	RunUI(message, col string) error
}

var (
	pingRunner UIRunner = &ui.DefaultUIRunner{} // default UI runner, can be replaced in tests
)

var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Responds with a pong",
	Long: `The ping command demonstrates configuration, logging, and optional Bubble Tea UI.
- Without arguments, prints "Pong".
- Use --message and --color to override defaults.
- Use --ui to launch an interactive Bubble Tea UI instead of printing.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		initPingConfig()

		message := viper.GetString("app.ping.output_message")
		colorStr := viper.GetString("app.ping.output_color")
		uiFlag := viper.GetBool("app.ping.ui")

		log.Info().
			Str("command", "ping").
			Bool("ui_enabled", uiFlag).
			Msg("Ping command invoked")

		log.Debug().
			Str("message", message).
			Str("color", colorStr).
			Bool("ui_enabled", uiFlag).
			Msg("Command configuration loaded")

		if uiFlag {
			log.Info().
				Str("message", message).
				Str("color", colorStr).
				Msg("Starting UI")

			if err := pingRunner.RunUI(message, colorStr); err != nil {
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

			// UI mode doesn't print directly, so no output
			return nil
		}

		// Print the message directly to stdout
		if err := ui.PrintColoredMessage(cmd.OutOrStdout(), message, colorStr); err != nil {
			log.Error().
				Err(err).
				Str("message", message).
				Str("color", colorStr).
				Msg("Failed to print colored message")
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	pingCmd.Flags().String("message", "", "Custom output message")
	pingCmd.Flags().String("color", "", "Output color")
	pingCmd.Flags().Bool("ui", false, "Enable UI")

	if err := viper.BindPFlag("app.ping.output_message", pingCmd.Flags().Lookup("message")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'message': %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("app.ping.output_color", pingCmd.Flags().Lookup("color")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'color': %v\n", err)
		os.Exit(1)
	}
	if err := viper.BindPFlag("app.ping.ui", pingCmd.Flags().Lookup("ui")); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to bind flag 'ui': %v\n", err)
		os.Exit(1)
	}
}

func initPingConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("app.ping.output_message", "Pong")
	viper.SetDefault("app.ping.output_color", "white")
	viper.SetDefault("app.ping.ui", false)
}
