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
	writer        io.Writer
	theme         *Theme
	themeExplicit bool // Theme was chosen via WithTheme: never auto-degrade it
	renderer      *lipgloss.Renderer
	isTerminal    bool // Whether writer is an interactive terminal
	mu            sync.Mutex
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
	p := &Printer{}
	for _, opt := range opts {
		opt(p)
	}
	p.ensureInit()
	return p
}

// ensureInit applies New's defaults, making a zero-value Printer (or one
// configured with nil options like WithTheme(nil)) safe to use.
// Callers other than New must hold p.mu.
func (p *Printer) ensureInit() {
	if p.writer == nil {
		p.writer = os.Stdout
	}
	if p.theme == nil {
		p.theme = DefaultTheme()
	}
	if p.renderer == nil {
		// Create a lipgloss renderer for this writer to enable colors
		p.renderer = lipgloss.NewRenderer(p.writer)

		// Detect if we're writing to a terminal
		p.isTerminal = IsTerminal(p.writer)

		// Degrade to MinimalTheme on non-TTY writers, but only when the
		// theme was not explicitly chosen (WithTheme) and ForceColors is
		// not set - an explicit choice is always honored
		if !p.themeExplicit && !p.theme.ForceColors && !p.isTerminal {
			p.theme = MinimalTheme()
		}
	}
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

// WithTheme sets the theme. A non-nil theme is treated as an explicit
// choice and is always honored - it is never replaced by MinimalTheme
// on non-TTY writers. Passing nil keeps the default behavior
// (DefaultTheme on terminals, MinimalTheme otherwise).
//
// Example:
//
//	p := checkmate.New(checkmate.WithTheme(checkmate.MinimalTheme()))
func WithTheme(t *Theme) Option {
	return func(p *Printer) {
		p.theme = t
		p.themeExplicit = t != nil
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

// CategoryHeader displays a category header surrounded by blank lines,
// with a colored background on terminals.
// Example: "Code Quality"
func (p *Printer) CategoryHeader(title string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCategoryHeader(title)
}

// CheckHeader displays a check-in-progress line on terminals only,
// without a trailing newline so the final result overwrites it.
// Non-TTY writers get no output.
// Example: "├── ○ format"
func (p *Printer) CheckHeader(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckHeader(message)
}

// CheckSuccess displays a success message.
// Example: "├── ✓ All files formatted"
func (p *Printer) CheckSuccess(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckSuccess(message)
}

// CheckFailure displays a failure line ("├── ✗ title") followed by
// "Details:" and "How to fix:" sections indented under │ connectors.
// Pass empty strings for details or remediation to omit those sections.
func (p *Printer) CheckFailure(title, details, remediation string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckFailure(title, details, remediation)
}

// CheckSummary displays a bordered summary box (rounded ╭─╮ corners,
// or +-+ ASCII with MinimalTheme) with a centered title and one
// icon-prefixed line per item.
// status should be StatusSuccess or StatusFailure.
func (p *Printer) CheckSummary(status Status, title string, items ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckSummary(status, title, items)
}

// CheckInfo displays indented informational lines.
// Example: "   Tool: go-licenses"
func (p *Printer) CheckInfo(lines ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckInfo(lines)
}

// CheckNote displays an informational note.
// Example: "Note: This is informational"
func (p *Printer) CheckNote(message string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckNote(message)
}

// CheckLine displays a single-line check result with duration.
// Used in non-TTY mode to mimic TUI output structure; on terminals it
// prints nothing (CheckHeader/CheckSuccess/CheckFailure handle output).
// Example (MinimalTheme): "format .................................. [OK] 1.451s"
func (p *Printer) CheckLine(name string, status Status, duration time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ensureInit()
	p.renderCheckLine(name, status, duration)
}

// style renders text with a theme style using the printer's renderer.
// This ensures proper color output by binding the style to our renderer.
func (p *Printer) style(s lipgloss.Style, text string) string {
	return p.renderer.NewStyle().Inherit(s).Render(text)
}

// Ensure Printer implements PrinterInterface.
var _ PrinterInterface = (*Printer)(nil)
