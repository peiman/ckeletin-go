// cmd/ping.go

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/internal/config/commands"
	"github.com/peiman/ckeletin-go/internal/ping"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/spf13/cobra"
)

var pingCmd = MustNewCommand(commands.PingMetadata, runPing)

func init() {
	MustAddToRoot(pingCmd)
}

func runPing(cmd *cobra.Command, args []string) error {
	return runPingWithUIRunner(cmd, args, ui.NewDefaultUIRunner())
}

// runPingWithUIRunner is the internal implementation that allows dependency injection for testing
func runPingWithUIRunner(cmd *cobra.Command, args []string, uiRunner ui.UIRunner) error {
	cfg := ping.Config{
		Message: getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage),
		Color:   getConfigValueWithFlags[string](cmd, "color", config.KeyAppPingOutputColor),
		UI:      getConfigValueWithFlags[bool](cmd, "ui", config.KeyAppPingUi),
	}
	response, err := ping.NewExecutor(cfg, uiRunner).Execute(output.NewReporter(cmd))
	if err != nil || cfg.UI {
		return err
	}
	return ui.RenderSuccess(cmd.OutOrStdout(), fmt.Sprintf("Pong! %s", response.Message), response)
}
