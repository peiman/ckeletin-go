// internal/check/check.go
//
// Check command executor using pkg/checkmate for beautiful output.

package check

import (
	"context"
	"fmt"
	"io"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/rs/zerolog/log"
)

// Config holds configuration for the check command
type Config struct {
	FailFast bool
	Verbose  bool
}

// Executor handles the execution of the check command
type Executor struct {
	cfg     Config
	printer checkmate.PrinterInterface
	writer  io.Writer
}

// NewExecutor creates a new check command executor
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	printer := checkmate.New(checkmate.WithWriter(writer))
	return &Executor{
		cfg:     cfg,
		printer: printer,
		writer:  writer,
	}
}

// NewExecutorWithPrinter creates an executor with a custom printer (for testing)
func NewExecutorWithPrinter(cfg Config, printer checkmate.PrinterInterface, writer io.Writer) *Executor {
	return &Executor{
		cfg:     cfg,
		printer: printer,
		writer:  writer,
	}
}

// Execute runs all quality checks
func (e *Executor) Execute(ctx context.Context) error {
	log.Debug().
		Bool("fail_fast", e.cfg.FailFast).
		Bool("verbose", e.cfg.Verbose).
		Msg("Starting quality checks")

	opts := []checkmate.RunnerOption{
		checkmate.WithCategory("Code Quality"),
	}
	if e.cfg.FailFast {
		opts = append(opts, checkmate.WithFailFast())
	}

	runner := checkmate.NewRunner(e.printer, opts...)

	// Register all checks
	runner.
		AddFunc("format", e.checkFormat).WithRemediation("Run: task format").
		AddFunc("lint", e.checkLint).WithRemediation("Run: task lint").
		AddFunc("test", e.checkTest).WithRemediation("Fix failing tests").
		AddFunc("deps", e.checkDeps).WithRemediation("Run: go mod tidy").
		AddFunc("vuln", e.checkVuln).WithRemediation("Update vulnerable dependencies")

	result := runner.Run(ctx)

	log.Info().
		Int("passed", result.Passed).
		Int("failed", result.Failed).
		Int("total", result.Total).
		Dur("duration", result.Duration).
		Msg("Check run completed")

	if !result.Success() {
		return fmt.Errorf("%d/%d checks failed", result.Failed, result.Total)
	}
	return nil
}
