// cmd/check.go

package cmd

import (
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/internal/check"
	"github.com/peiman/ckeletin-go/internal/config/commands"
	"github.com/spf13/cobra"
)

var checkCmd = MustNewCommand(commands.CheckMetadata, runCheck)

func init() {
	MustAddToRoot(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg := check.Config{
		FailFast: getConfigValueWithFlags[bool](cmd, "fail-fast", config.KeyAppCheckFailFast),
		Verbose:  getConfigValueWithFlags[bool](cmd, "verbose", config.KeyAppCheckVerbose),
	}
	return check.NewExecutor(cfg, cmd.OutOrStdout()).Execute(cmd.Context())
}
