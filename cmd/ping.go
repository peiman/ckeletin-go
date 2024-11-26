// cmd/ping.go
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	message  string
	colorStr string
	uiFlag   bool
	logLevel string
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Responds with a pong",
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize logger with the current log level
		logger.Init()

		log.Info().Msg("Ping command invoked")

		// Get the message and color from configuration
		msg := viper.GetString("app.output_message")
		col := viper.GetString("app.output_color")

		// Check if UI is enabled
		if viper.GetBool("app.ui") {
			// Run the UI
			if err := ui.RunUI(msg, col); err != nil {
				log.Error().Err(err).Msg("Failed to start UI")
				fmt.Println("An error occurred while running the UI.")
			}
		} else {
			// Print the message with color
			if err := printColoredMessage(msg, col); err != nil {
				log.Error().Err(err).Msg("Failed to print message")
				fmt.Println("An error occurred while printing the message.")
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)

	// Define flags specific to the ping command
	pingCmd.Flags().StringVarP(&message, "message", "m", "", "Custom output message")
	pingCmd.Flags().StringVarP(&colorStr, "color", "c", "", "Output color")
	pingCmd.Flags().BoolVarP(&uiFlag, "ui", "u", false, "Enable UI")
	pingCmd.Flags().StringVarP(&logLevel, "log-level", "l", "", "Set the log level (debug, info, warn, error)")

	// Bind flags to Viper with error checking
	if err := viper.BindPFlag("app.output_message", pingCmd.Flags().Lookup("message")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'message'")
	}

	if err := viper.BindPFlag("app.output_color", pingCmd.Flags().Lookup("color")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'color'")
	}

	if err := viper.BindPFlag("app.ui", pingCmd.Flags().Lookup("ui")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'ui'")
	}

	if err := viper.BindPFlag("app.log_level", pingCmd.Flags().Lookup("log-level")); err != nil {
		log.Fatal().Err(err).Msg("Failed to bind flag 'log-level'")
	}
}

func printColoredMessage(message, col string) error {
	// Map color names to color attributes
	colorMap := map[string]color.Attribute{
		"black":   color.FgBlack,
		"red":     color.FgRed,
		"green":   color.FgGreen,
		"yellow":  color.FgYellow,
		"blue":    color.FgBlue,
		"magenta": color.FgMagenta,
		"cyan":    color.FgCyan,
		"white":   color.FgWhite,
	}

	attr, exists := colorMap[col]
	if !exists {
		return fmt.Errorf("invalid color: %s", col)
	}

	c := color.New(attr).Add(color.Bold)
	c.Println(message)
	return nil
}
