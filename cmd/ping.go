// cmd/ping.go - Ping command implementation
package cmd

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/peiman/ckeletin-go/internal/errors"
	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// pingOptions holds the command's options
type pingOptions struct {
	count int
}

func newPingCommand() *cobra.Command {
	opts := &pingOptions{}

	cmd := &cobra.Command{
		Use:   "ping",
		Short: "Responds with pong",
		Long: `A demonstration command that shows how to implement new commands
using the framework's features like logging, configuration, and error handling.

The ping command demonstrates how to use Viper configuration:
- Default count can be set in config file (ping.defaultCount)
- Output message can be customized (ping.outputMessage)
- Colored output can be enabled (ping.coloredOutput)

Example config (ckeletin-go.json):
{
  "ping": {
    "defaultCount": 3,
    "outputMessage": "pong",
    "coloredOutput": true
  }
}

Example usage:
  ckeletin-go ping            # Outputs using configured defaults
  ckeletin-go ping --count 3  # Outputs configured message three times`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			// Get logger instance
			logger := infrastructure.GetLogger()

			// Get config values
			outputMessage := viper.GetString("ping.outputMessage")
			useColor := viper.GetBool("ping.coloredOutput")

			// If count wasn't specified via flag, use config default
			if !cmd.Flags().Changed("count") {
				opts.count = viper.GetInt("ping.defaultCount")
			}

			// Validate count
			if opts.count <= 0 {
				err := errors.NewAppError("INVALID_COUNT", "count flag must be greater than 0", nil)
				logger.Error().Err(err).Int("count", opts.count).Msg("Invalid count value provided")
				return err
			}

			// Log command execution
			logger.Debug().
				Int("count", opts.count).
				Str("message", outputMessage).
				Bool("colored", useColor).
				Msg("Executing ping command")

			// Prepare colored output if enabled
			output := outputMessage
			if useColor {
				output = color.GreenString(output)
			}

			// Output the configured number of times
			for i := 0; i < opts.count; i++ {
				fmt.Fprintln(cmd.OutOrStdout(), output)
			}

			logger.Info().
				Int("count", opts.count).
				Str("message", outputMessage).
				Msg("Ping command completed successfully")
			return nil
		},
	}

	// Add flags with default from config
	defaultCount := viper.GetInt("ping.defaultCount")
	if defaultCount == 0 {
		defaultCount = infrastructure.DefaultPingCount
	}
	cmd.Flags().IntVarP(&opts.count, "count", "c", defaultCount, "number of times to ping")

	return cmd
}

func init() {
	rootCmd.AddCommand(newPingCommand())
}
