//go:build dev

// ckeletin:allow-custom-command
// cmd/dev_progress.go
//
// Progress package demonstration command (dev-only).
// Shows spinner, progress bar, and multi-phase progress.

package cmd

import (
	"context"
	"time"

	"github.com/peiman/ckeletin-go/internal/progress"
	"github.com/rs/zerolog/log"
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

	log.Debug().Msg("Dev progress command registered")
}

func runDevProgress(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	useUI, _ := cmd.Flags().GetBool("ui")
	spinnerOnly, _ := cmd.Flags().GetBool("spinner")
	barOnly, _ := cmd.Flags().GetBool("bar")

	// Create reporter with appropriate output mode
	reporter := progress.NewReporter(
		progress.WithOutput(cmd.ErrOrStderr(), useUI),
	)

	// Determine which demos to run
	runAll := !spinnerOnly && !barOnly

	if runAll || spinnerOnly {
		if err := demoSpinner(ctx, reporter); err != nil {
			return err
		}
	}

	if runAll || barOnly {
		if err := demoProgressBar(ctx, reporter); err != nil {
			return err
		}
	}

	if runAll {
		if err := demoMultiPhase(ctx, reporter); err != nil {
			return err
		}
	}

	return nil
}

// demoSpinner demonstrates indeterminate progress with a spinner.
func demoSpinner(ctx context.Context, reporter *progress.Reporter) error {
	reporter.SetPhase("spinner-demo")
	reporter.Start(ctx, "Simulating network request...")

	// Simulate work
	time.Sleep(2 * time.Second)

	reporter.Complete(ctx, "Network request completed")
	return nil
}

// demoProgressBar demonstrates determinate progress with a progress bar.
func demoProgressBar(ctx context.Context, reporter *progress.Reporter) error {
	reporter.SetPhase("progress-demo")
	reporter.Start(ctx, "Processing items")

	items := []string{
		"Loading configuration",
		"Validating schema",
		"Processing data",
		"Generating output",
		"Finalizing results",
	}

	total := int64(len(items))
	for i, item := range items {
		reporter.Progress(ctx, int64(i+1), total, item)
		time.Sleep(500 * time.Millisecond)
	}

	reporter.Complete(ctx, "All items processed successfully")
	return nil
}

// demoMultiPhase demonstrates multi-phase progress reporting.
func demoMultiPhase(ctx context.Context, reporter *progress.Reporter) error {
	phases := []struct {
		name  string
		steps int
		desc  string
	}{
		{"download", 3, "Downloading dependencies"},
		{"compile", 4, "Compiling source code"},
		{"package", 2, "Creating package"},
	}

	for _, phase := range phases {
		reporter.SetPhase(phase.name)
		reporter.Start(ctx, phase.desc)

		for i := 0; i < phase.steps; i++ {
			reporter.Progress(ctx, int64(i+1), int64(phase.steps), "Step")
			time.Sleep(300 * time.Millisecond)
		}

		reporter.Complete(ctx, phase.desc+" complete")
	}

	return nil
}
