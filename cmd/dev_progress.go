//go:build dev

// ckeletin:allow-custom-command — dev-only demo wiring with local flag switches,
// not a config-registry-driven command.
// cmd/dev_progress.go
//
// Progress package demonstration command (dev-only).
// Thin wiring over internal/progress demo logic (ADR-001).

package cmd

import (
	"context"

	"github.com/peiman/ckeletin-go/internal/progress"
	"github.com/spf13/cobra"
)

var devProgressCmd = &cobra.Command{
	Use:   "progress",
	Short: "Demonstrate progress reporting capabilities",
	Long: `Demonstrate the progress reporting package with various examples:

  - Spinner (indeterminate progress)
  - Progress bar (determinate progress)
  - Multi-phase operations

This command is useful for testing progress UI in different terminal environments.

Examples:
  dev progress              # Run all demos (non-interactive)
  dev progress --ui         # Run with Bubble Tea interactive UI
  dev progress --spinner    # Run only spinner demo
  dev progress --bar        # Run only progress bar demo`,
	RunE: runDevProgress,
}

func init() {
	devCmd.AddCommand(devProgressCmd)

	devProgressCmd.Flags().Bool("ui", false, "Use interactive Bubble Tea UI")
	devProgressCmd.Flags().Bool("spinner", false, "Run only spinner demo")
	devProgressCmd.Flags().Bool("bar", false, "Run only progress bar demo")
	devProgressCmd.Flags().Duration("delay", 0, "Override step delay duration (e.g., 100ms for fast demo)")
}

func runDevProgress(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	useUI, _ := cmd.Flags().GetBool("ui")
	spinnerOnly, _ := cmd.Flags().GetBool("spinner")
	barOnly, _ := cmd.Flags().GetBool("bar")
	delay, _ := cmd.Flags().GetDuration("delay")

	reporter := progress.NewReporter(
		progress.WithOutput(cmd.ErrOrStderr(), useUI),
	)

	return progress.RunDemo(ctx, reporter, progress.DemoOptions{
		SpinnerOnly: spinnerOnly,
		BarOnly:     barOnly,
		Delay:       delay,
	})
}
