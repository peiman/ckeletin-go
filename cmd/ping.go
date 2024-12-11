// cmd/ping.go

package cmd

import (
	"fmt"
	"strings"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

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
- Use --ui to launch an interactive Bubble Tea UI.`,
	RunE: runPing,
}

func init() {
	pingCmd.Flags().String("message", "", "Custom output message")
	pingCmd.Flags().String("color", "", "Output color")
	pingCmd.Flags().Bool("ui", false, "Enable UI")

	// Bind flags to Viper
	if err := viper.BindPFlag("app.ping.output_message", pingCmd.Flags().Lookup("message")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'message' flag")
	}
	if err := viper.BindPFlag("app.ping.output_color", pingCmd.Flags().Lookup("color")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'color' flag")
	}
	if err := viper.BindPFlag("app.ping.ui", pingCmd.Flags().Lookup("ui")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind 'ui' flag")
	}

	// Add pingCmd to RootCmd
	RootCmd.AddCommand(pingCmd)
}

func initPingConfig() {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("app.ping.output_message", "Pong")
	viper.SetDefault("app.ping.output_color", "white")
	viper.SetDefault("app.ping.ui", false)
}

func runPing(cmd *cobra.Command, args []string) error {
	log.Debug().Msg("Starting runPing execution")
	initPingConfig()

	// Get values from flags or Viper
	message := viper.GetString("app.ping.output_message")
	if cmd.Flags().Changed("message") {
		message, _ = cmd.Flags().GetString("message")
	}

	colorStr := viper.GetString("app.ping.output_color")
	if cmd.Flags().Changed("color") {
		colorStr, _ = cmd.Flags().GetString("color")
	}

	uiFlag := viper.GetBool("app.ping.ui")
	if cmd.Flags().Changed("ui") {
		uiFlag, _ = cmd.Flags().GetBool("ui")
	}

	log.Debug().
		Str("message", message).
		Str("color", colorStr).
		Bool("ui_enabled", uiFlag).
		Msg("Configuration loaded")

	writer := cmd.OutOrStdout()
	log.Debug().
		Str("writer_type", fmt.Sprintf("%T", writer)).
		Msg("Using writer")

	if uiFlag {
		log.Info().Str("message", message).Str("color", colorStr).Msg("Starting UI")
		if err := pingRunner.RunUI(message, colorStr); err != nil {
			log.Error().Err(err).Msg("Failed to run UI")
			return err
		}
		return nil
	}

	// Non-UI mode: print the message
	err := ui.PrintColoredMessage(writer, message, colorStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("message", message).
			Str("color", colorStr).
			Msg("Failed to print colored message")
		// Wrap the error to provide context
		return fmt.Errorf("failed to print colored message: %w", err)
	}

	log.Debug().Msg("runPing completed successfully")
	return nil
}
