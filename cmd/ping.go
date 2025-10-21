// cmd/ping.go

package cmd

import (
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/spf13/cobra"
)

// uiRunnerFactory is a function that creates a new UI runner
// This allows tests to inject a mock runner
var uiRunnerFactory = func() ui.UIRunner {
	return ui.NewDefaultUIRunner()
}

// pingCmd represents the ping command
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
	// All defaults MUST be defined in internal/config/ping_options.go
	// See internal/config/ping_options.go for all configuration options
}

// runPing is a thin CLI wrapper that delegates to internal/ping.Executor
func runPing(cmd *cobra.Command, args []string) error {
	// Get configuration values from Viper
	message := getConfigValue[string](cmd, "message", "app.ping.output_message")
	color := getConfigValue[string](cmd, "color", "app.ping.output_color")
	enableUI := getConfigValue[bool](cmd, "ui", "app.ping.ui")

	// Create configuration for ping executor
	cfg := ping.NewConfig(message, color, enableUI)

	// Create executor with dependencies (dependency injection)
	executor := ping.NewExecutor(
		cfg,
		uiRunnerFactory(), // UI runner dependency (uses factory for testability)
		cmd.OutOrStdout(), // Output writer dependency
	)

	// Execute business logic
	return executor.Execute()
}
