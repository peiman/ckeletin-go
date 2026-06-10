//go:build dev

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
	categories := check.ParseCategories(
		getConfigValueWithFlags[string](cmd, "category", config.KeyAppCheckCategory))
	if len(categories) > 0 {
		if err := check.ValidateCategories(categories); err != nil {
			return err
		}
	}

	cfg := check.Config{
		FailFast:   getConfigValueWithFlags[bool](cmd, "fail-fast", config.KeyAppCheckFailFast),
		Verbose:    getConfigValueWithFlags[bool](cmd, "verbose", config.KeyAppCheckVerbose),
		Parallel:   getConfigValueWithFlags[bool](cmd, "parallel", config.KeyAppCheckParallel),
		Categories: categories,
		ShowTiming: getConfigValueWithFlags[bool](cmd, "timing", config.KeyAppCheckTiming),
		BinaryName: binaryName,
	}

	return check.NewExecutor(cfg, cmd.OutOrStdout()).Execute(cmd.Context())
}
