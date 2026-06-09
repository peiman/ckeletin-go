//go:build dev

// internal/progress/demo.go
//
// Demo sequences for the dev progress command (dev builds only).
// Shows spinner, progress bar, and multi-phase progress.

package progress

import (
	"context"
	"time"
)

// Default durations for demos (can be overridden via DemoOptions.Delay).
const (
	defaultSpinnerDuration = 2 * time.Second
	defaultStepDelay       = 500 * time.Millisecond
	defaultPhaseStepDelay  = 300 * time.Millisecond
)

// DemoOptions selects which demos to run and how fast they advance.
type DemoOptions struct {
	// SpinnerOnly runs only the spinner demo.
	SpinnerOnly bool

	// BarOnly runs only the progress bar demo.
	BarOnly bool

	// Delay overrides all demo durations when greater than zero.
	Delay time.Duration
}

// demoConfig holds configuration for demo functions.
type demoConfig struct {
	spinnerDuration time.Duration
	stepDelay       time.Duration
	phaseStepDelay  time.Duration
}

// newDemoConfig resolves demo durations from defaults and the optional override.
func newDemoConfig(delay time.Duration) demoConfig {
	cfg := demoConfig{
		spinnerDuration: defaultSpinnerDuration,
		stepDelay:       defaultStepDelay,
		phaseStepDelay:  defaultPhaseStepDelay,
	}
	if delay > 0 {
		cfg.spinnerDuration = delay
		cfg.stepDelay = delay
		cfg.phaseStepDelay = delay
	}
	return cfg
}

// RunDemo runs the demos selected by opts against the given reporter.
// With no selection it runs all demos: spinner, progress bar, and multi-phase.
func RunDemo(ctx context.Context, reporter *Reporter, opts DemoOptions) error {
	cfg := newDemoConfig(opts.Delay)

	runAll := !opts.SpinnerOnly && !opts.BarOnly

	if runAll || opts.SpinnerOnly {
		if err := demoSpinner(ctx, reporter, cfg); err != nil {
			return err
		}
	}

	if runAll || opts.BarOnly {
		if err := demoProgressBar(ctx, reporter, cfg); err != nil {
			return err
		}
	}

	if runAll {
		if err := demoMultiPhase(ctx, reporter, cfg); err != nil {
			return err
		}
	}

	return nil
}

// demoSpinner demonstrates indeterminate progress with a spinner.
func demoSpinner(ctx context.Context, reporter *Reporter, cfg demoConfig) error {
	reporter.SetPhase("spinner-demo")
	reporter.Start(ctx, "Simulating network request...")

	// Simulate work (respects context cancellation)
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(cfg.spinnerDuration):
	}

	reporter.Complete(ctx, "Network request completed")
	return nil
}

// demoProgressBar demonstrates determinate progress with a progress bar.
func demoProgressBar(ctx context.Context, reporter *Reporter, cfg demoConfig) error {
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(cfg.stepDelay):
		}
	}

	reporter.Complete(ctx, "All items processed successfully")
	return nil
}

// demoMultiPhase demonstrates multi-phase progress reporting.
func demoMultiPhase(ctx context.Context, reporter *Reporter, cfg demoConfig) error {
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
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cfg.phaseStepDelay):
			}
		}

		reporter.Complete(ctx, phase.desc+" complete")
	}

	return nil
}
