// cmd/ping.go

package cmd

import (
	"github.com/peiman/ckeletin-go/internal/config/commands"
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/spf13/cobra"
)

// uiRunnerFactory allows tests to inject a mock runner
var uiRunnerFactory = func() ui.UIRunner {
	return ui.NewDefaultUIRunner()
}

var pingCmd = NewCommand(commands.PingMetadata, runPing)

func init() {
	MustAddToRoot(pingCmd)
}

func runPing(cmd *cobra.Command, args []string) error {
	cfg := ping.Config{
		Message: getConfigValue[string](cmd, "message", "app.ping.output_message"),
		Color:   getConfigValue[string](cmd, "color", "app.ping.output_color"),
		UI:      getConfigValue[bool](cmd, "ui", "app.ping.ui"),
	}
	return ping.NewExecutor(cfg, uiRunnerFactory(), cmd.OutOrStdout()).Execute()
}
