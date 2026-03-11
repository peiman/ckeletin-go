package checkmate

import (
	"context"
	"time"
)

// Status represents the outcome of a check or operation.
type Status string

const (
	// StatusSuccess indicates a successful check.
	StatusSuccess Status = "success"
	// StatusFailure indicates a failed check.
	StatusFailure Status = "failure"
)

// PrinterInterface defines the contract for check output.
// Use this interface for dependency injection in your code,
// allowing easy substitution of MockPrinter in tests.
type PrinterInterface interface {
	// CategoryHeader displays a category header with decorative separators.
	// Example output: "─── Code Quality ────────────────────────"
	CategoryHeader(title string)

	// CheckHeader displays a check-in-progress message.
	// Example output: "🔍 Checking formatting..."
	CheckHeader(message string)

	// CheckSuccess displays a success message.
	// Example output: "✅ All files properly formatted"
	CheckSuccess(message string)

	// CheckFailure displays a failure with details and remediation guidance.
	// Example output:
	//   "❌ Format check failed"
	//   "Details:"
	//   "  <details>"
	//   "How to fix:"
	//   "  • <remediation>"
	CheckFailure(title, details, remediation string)

	// CheckSummary displays a summary box with status and items.
	// Example output:
	//   "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	//   "✅ All checks passed"
	//   ""
	//   "• Item 1"
	//   "• Item 2"
	//   "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	CheckSummary(status Status, title string, items ...string)

	// CheckInfo displays indented informational lines.
	// Example output: "   Tool: go-licenses"
	CheckInfo(lines ...string)

	// CheckNote displays an informational note.
	// Example output: "Note: This is informational"
	CheckNote(message string)

	// CheckLine displays a single-line check result with duration.
	// Used in non-TTY mode to mimic TUI output structure.
	// Example output: "format .......................... [OK] 1.451s"
	CheckLine(name string, status Status, duration time.Duration)
}

// RunnerInterface defines the contract for running checks.
// Use this interface for dependency injection in your code,
// allowing easy substitution of MockRunner in tests.
type RunnerInterface interface {
	// Add adds a check to the runner. Returns the runner for chaining.
	Add(check Check) RunnerInterface

	// AddFunc is a convenience for adding simple checks.
	AddFunc(name string, fn func(ctx context.Context) error) RunnerInterface

	// WithRemediation sets remediation text for the last added check.
	WithRemediation(text string) RunnerInterface

	// WithDetails sets details text for the last added check.
	WithDetails(text string) RunnerInterface

	// Run executes all checks and returns results.
	Run(ctx context.Context) RunResult
}
