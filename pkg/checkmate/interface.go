package checkmate

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
}
