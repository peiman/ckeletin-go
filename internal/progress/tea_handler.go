package progress

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// TeaHandler renders progress events using Bubble Tea.
// It provides animated spinners, progress bars, and real-time updates.
//
// This handler is designed to be used as part of a CompositeHandler,
// typically alongside LogHandler for shadow logging (ADR-012 compliance).
type TeaHandler struct {
	out     io.Writer
	style   *Style
	mu      sync.Mutex
	program *tea.Program
	model   *teaModel
	started bool
	ready   chan struct{} // signals when program is ready to receive messages
}

// NewTeaHandler creates a new Bubble Tea based progress handler.
func NewTeaHandler(out io.Writer) *TeaHandler {
	style := DefaultStyle()
	model := newTeaModel(style)
	return &TeaHandler{
		out:   out,
		style: style,
		model: model,
		ready: make(chan struct{}),
	}
}

// OnProgress implements Handler by sending events to the Bubble Tea model.
// Respects context cancellation.
func (h *TeaHandler) OnProgress(ctx context.Context, event Event) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	h.mu.Lock()
	needsStart := !h.started
	h.mu.Unlock()

	// Start the program on first event if not started
	if needsStart {
		h.start()
	}

	// Wait for program to be ready (with context cancellation support)
	select {
	case <-ctx.Done():
		return
	case <-h.ready:
		// Program is ready
	}

	h.mu.Lock()
	program := h.program
	h.mu.Unlock()

	// Send the event to the model
	if program != nil {
		program.Send(progressEventMsg{event: event})

		// If this is a terminal event, signal completion
		if event.Type == EventComplete || event.Type == EventError {
			// Use a short timer instead of sleep to allow for cancellation
			timer := time.NewTimer(50 * time.Millisecond)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
			program.Send(tea.Quit())
		}
	}
}

// start initializes the Bubble Tea program.
func (h *TeaHandler) start() {
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	h.mu.Unlock()

	opts := []tea.ProgramOption{
		tea.WithOutput(h.out),
	}

	h.mu.Lock()
	h.program = tea.NewProgram(h.model, opts...)
	program := h.program
	h.mu.Unlock()

	// Run in goroutine so OnProgress doesn't block
	go func() {
		// Signal that the program is ready to receive messages
		// Close the channel to signal all waiting goroutines
		close(h.ready)
		_, _ = program.Run()
	}()
}

// Stop gracefully stops the Bubble Tea program.
func (h *TeaHandler) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.program != nil {
		h.program.Quit()
		h.program = nil
		h.started = false
	}
}

// Bubble Tea messages
type (
	// progressEventMsg wraps a progress event
	progressEventMsg struct {
		event Event
	}

	// tickMsg triggers animation updates
	tickMsg time.Time
)

// teaModel is the Bubble Tea model for progress display.
type teaModel struct {
	style        *Style
	currentEvent Event
	spinnerFrame int
	done         bool
}

func newTeaModel(style *Style) *teaModel {
	return &teaModel{
		style: style,
	}
}

// Init implements tea.Model.
func (m *teaModel) Init() tea.Cmd {
	return tickCmd(m.style.SpinnerInterval)
}

// tickCmd returns a command that sends tick messages for animation.
func tickCmd(interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update implements tea.Model.
func (m *teaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.done = true
			return m, tea.Quit
		}

	case tickMsg:
		if !m.done {
			m.spinnerFrame = (m.spinnerFrame + 1) % len(m.style.SpinnerFrames)
			return m, tickCmd(m.style.SpinnerInterval)
		}

	case progressEventMsg:
		m.currentEvent = msg.event
		if msg.event.Type == EventComplete || msg.event.Type == EventError {
			m.done = true
		}
		return m, nil
	}

	return m, nil
}

// View implements tea.Model.
func (m *teaModel) View() string {
	if m.currentEvent.Message == "" {
		return ""
	}

	var b strings.Builder

	// Show phase if present
	if m.currentEvent.Phase != "" {
		b.WriteString(m.style.PhaseStyle.Render(m.currentEvent.Phase))
		b.WriteString(" ")
	}

	switch m.currentEvent.Type {
	case EventStart:
		spinner := m.style.SpinnerFrames[m.spinnerFrame]
		b.WriteString(m.style.SpinnerStyle.Render(spinner))
		b.WriteString(" ")
		b.WriteString(m.currentEvent.Message)

	case EventProgress:
		if m.currentEvent.IsIndeterminate() {
			// Indeterminate: spinner only
			spinner := m.style.SpinnerFrames[m.spinnerFrame]
			b.WriteString(m.style.SpinnerStyle.Render(spinner))
			b.WriteString(" ")
			b.WriteString(m.currentEvent.Message)
		} else {
			// Determinate: progress bar
			spinner := m.style.SpinnerFrames[m.spinnerFrame]
			b.WriteString(m.style.SpinnerStyle.Render(spinner))
			b.WriteString(" ")
			b.WriteString(m.renderBar())
			b.WriteString(" ")
			b.WriteString(m.renderCounter())
			if m.currentEvent.Message != "" {
				b.WriteString(" ")
				b.WriteString(m.style.TaskStyle.Render(m.currentEvent.Message))
			}
		}

	case EventComplete:
		b.WriteString(m.style.SuccessStyle.Render("✓"))
		b.WriteString(" ")
		b.WriteString(m.currentEvent.Message)

	case EventError:
		b.WriteString(m.style.ErrorStyle.Render("✗"))
		b.WriteString(" ")
		b.WriteString(m.currentEvent.Message)

	case EventWarning:
		b.WriteString(m.style.WarningStyle.Render("⚠"))
		b.WriteString(" ")
		b.WriteString(m.currentEvent.Message)
	}

	b.WriteString("\n")
	return b.String()
}

// renderBar creates the progress bar visualization.
func (m *teaModel) renderBar() string {
	if m.currentEvent.Total <= 0 {
		return ""
	}

	percent := float64(m.currentEvent.Current) / float64(m.currentEvent.Total)
	filled := int(percent * float64(m.style.BarWidth))
	if filled > m.style.BarWidth {
		filled = m.style.BarWidth
	}

	var bar strings.Builder
	bar.WriteString("[")
	for i := 0; i < m.style.BarWidth; i++ {
		if i < filled {
			bar.WriteString(m.style.BarStyle.Render(m.style.BarChar))
		} else {
			bar.WriteString(m.style.BarEmptyStyle.Render(m.style.BarEmptyChar))
		}
	}
	bar.WriteString("]")

	return bar.String()
}

// renderCounter shows the X/Y (percent%) counter.
func (m *teaModel) renderCounter() string {
	if m.currentEvent.Total <= 0 {
		return ""
	}

	percent := m.currentEvent.Percentage()
	return m.style.CounterStyle.Render(
		fmt.Sprintf("%d/%d (%.0f%%)", m.currentEvent.Current, m.currentEvent.Total, percent),
	)
}
