package progress

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/charmbracelet/lipgloss"
)

// ConsoleHandler writes simple progress output to a writer (typically stderr).
// This provides non-interactive progress feedback for terminals that don't
// support Bubble Tea or when running in CI/piped environments.
type ConsoleHandler struct {
	writer io.Writer
	mu     sync.Mutex

	// Styles for output
	successStyle lipgloss.Style
	errorStyle   lipgloss.Style
	warningStyle lipgloss.Style
	phaseStyle   lipgloss.Style
}

// NewConsoleHandler creates a new ConsoleHandler writing to the given writer.
func NewConsoleHandler(w io.Writer) *ConsoleHandler {
	return &ConsoleHandler{
		writer:       w,
		successStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),  // Green
		errorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true), // Red
		warningStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true), // Orange
		phaseStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true),  // Blue
	}
}

// OnProgress implements Handler by writing formatted progress to the writer.
// Respects context cancellation.
func (h *ConsoleHandler) OnProgress(ctx context.Context, event Event) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	switch event.Type {
	case EventStart:
		if event.Phase != "" {
			_, _ = fmt.Fprintf(h.writer, "%s %s\n", h.phaseStyle.Render("==>"), event.Message)
		} else {
			_, _ = fmt.Fprintf(h.writer, "  %s\n", event.Message)
		}

	case EventProgress:
		if event.Total > 0 {
			// Determinate progress: show X of Y
			_, _ = fmt.Fprintf(h.writer, "  [%d/%d] %s\n", event.Current, event.Total, event.Message)
		}
		// Indeterminate progress: skip (too noisy for console)

	case EventComplete:
		_, _ = fmt.Fprintf(h.writer, "  %s %s\n", h.successStyle.Render("✓"), event.Message)

	case EventError:
		_, _ = fmt.Fprintf(h.writer, "  %s %s\n", h.errorStyle.Render("✗"), event.Message)

	case EventWarning:
		_, _ = fmt.Fprintf(h.writer, "  %s %s\n", h.warningStyle.Render("⚠"), event.Message)
	}
}
