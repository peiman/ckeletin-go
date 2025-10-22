// cmd/ping.go

package cmd

import (
	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/peiman/ckeletin-go/internal/config/commands"
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/spf13/cobra"
)

var pingCmd = NewCommand(commands.PingMetadata, runPing)

func init() {
	MustAddToRoot(pingCmd)
}

func runPing(cmd *cobra.Command, args []string) error {
	return runPingWithUIRunner(cmd, args, ui.NewDefaultUIRunner())
}

// runPingWithUIRunner is the internal implementation that allows dependency injection for testing
func runPingWithUIRunner(cmd *cobra.Command, args []string, uiRunner ui.UIRunner) error {
	cfg := ping.Config{
		Message: getConfigValue[string](cmd, "message", config.KeyAppPingOutputMessage),
		Color:   getConfigValue[string](cmd, "color", config.KeyAppPingOutputColor),
		UI:      getConfigValue[bool](cmd, "ui", config.KeyAppPingUi),
	}
	return ping.NewExecutor(cfg, uiRunner, cmd.OutOrStdout()).Execute()
}
