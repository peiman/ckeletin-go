package checkmate

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Printer renders check output to a writer.
// All methods are thread-safe for concurrent use.
type Printer struct {
	writer     io.Writer
	theme      *Theme
	renderer   *lipgloss.Renderer
	isTerminal bool // Whether writer is an interactive terminal
	mu         sync.Mutex
}

// Option configures a Printer.
type Option func(*Printer)

// New creates a new Printer with the given options.
// Default: writes to stdout with DefaultTheme, auto-detecting TTY.
//
// Example:
//
//	p := checkmate.New()
//	p.CheckHeader("Running tests")
//	p.CheckSuccess("All tests passed")
func New(opts ...Option) *Printer {
	p := &Printer{
		writer: os.Stdout,
		theme:  DefaultTheme(),
	}
	for _, opt := range opts {
		opt(p)
	}
	// Create a lipgloss renderer for this writer to enable colors
	p.renderer = lipgloss.NewRenderer(p.writer)

	// Detect if we're writing to a terminal
	p.isTerminal = IsTerminal(p.writer)

	// Auto-detect TTY and switch to minimal theme if needed
	// (unless ForceColors is set)
	if !p.theme.ForceColors && !p.isTerminal {
		p.theme = MinimalTheme()
	}
	return p
}

// WithWriter sets the output writer.
//
// Example:
//
//	var buf bytes.Buffer
//	p := checkmate.New(checkmate.WithWriter(&buf))
func WithWriter(w io.Writer) Option {
	return func(p *Printer) {
		p.writer = w
	}
}

// WithTheme sets the theme.
//
// Example:
//
//	p := checkmate.New(checkmate.WithTheme(checkmate.MinimalTheme()))
func WithTheme(t *Theme) Option {
	return func(p *Printer) {
		p.theme = t
	}
}

// WithStderr is a convenience option to write to stderr.
//
// Example:
//
//	p := checkmate.New(checkmate.WithStderr())
func WithStderr() Option {
	return WithWriter(os.Stderr)
}

// CategoryHeader displays a category header with decorative separators.
// Example: "─── Code Quality ────────────────────────"
func (p *Printer) CategoryHeader(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCategoryHeader(title)
}

// CheckHeader displays a check-in-progress message.
// Example: "🔍 Checking formatting..."
func (p *Printer) CheckHeader(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckHeader(message)
}

// CheckSuccess displays a success message.
// Example: "✅ All files properly formatted"
func (p *Printer) CheckSuccess(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckSuccess(message)
}

// CheckFailure displays a failure with details and remediation guidance.
// Pass empty strings for details or remediation to omit those sections.
func (p *Printer) CheckFailure(title, details, remediation string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckFailure(title, details, remediation)
}

// CheckSummary displays a summary box with status and items.
// status should be StatusSuccess or StatusFailure.
func (p *Printer) CheckSummary(status Status, title string, items ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckSummary(status, title, items)
}

// CheckInfo displays indented informational lines.
// Example: "   Tool: go-licenses"
func (p *Printer) CheckInfo(lines ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckInfo(lines)
}

// CheckNote displays an informational note.
// Example: "Note: This is informational"
func (p *Printer) CheckNote(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckNote(message)
}

// CheckLine displays a single-line check result with duration.
// Used in non-TTY mode to mimic TUI output structure.
// Example: "format .......................... [OK] 1.451s"
func (p *Printer) CheckLine(name string, status Status, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.renderCheckLine(name, status, duration)
}

// style renders text with a theme style using the printer's renderer.
// This ensures proper color output by binding the style to our renderer.
func (p *Printer) style(s lipgloss.Style, text string) string {
	return p.renderer.NewStyle().Inherit(s).Render(text)
}

// Ensure Printer implements PrinterInterface.
var _ PrinterInterface = (*Printer)(nil)
