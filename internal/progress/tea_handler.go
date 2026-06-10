package progress

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rs/zerolog/log"
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

	// extraTeaOpts is appended to the default program options. Tests use it
	// to run the program without a TTY (tea.WithInput(nil)); nil in production.
	extraTeaOpts []tea.ProgramOption
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

	// Snapshot the channel under the mutex: Stop() replaces it for reuse
	h.mu.Lock()
	ready := h.ready
	h.mu.Unlock()

	// Wait for program to be ready (with context cancellation support)
	select {
	case <-ctx.Done():
		return
	case <-ready:
		// Program is ready
	}

	h.mu.Lock()
	program := h.program
	h.mu.Unlock()

	// Send the event to the model
	if program == nil {
		log.Debug().
			Str("event_type", event.Type.String()).
			Msg("progress event dropped: no active Bubble Tea program")
		return
	}

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

// start initializes the Bubble Tea program.
func (h *TeaHandler) start() {
	// started and program must be published in one critical section:
	// a Stop() between them would see a nil program, no-op, and the program
	// would launch anyway, voiding Stop's wait-for-shutdown guarantee.
	// tea.NewProgram only constructs the program (Run starts it), so it is
	// safe to call while holding the mutex.
	h.mu.Lock()
	if h.started {
		h.mu.Unlock()
		return
	}
	h.started = true
	opts := append([]tea.ProgramOption{tea.WithOutput(h.out)}, h.extraTeaOpts...)
	program := tea.NewProgram(h.model, opts...)
	h.program = program
	ready := h.ready
	h.mu.Unlock()

	// Run in goroutine so OnProgress doesn't block
	go func() {
		// Signal that the program is ready to receive messages
		// Close the channel to signal all waiting goroutines
		close(ready)
		if _, err := program.Run(); err != nil {
			// Without a reset, started would stay true forever and every
			// later OnProgress would silently no-op (dead handler).
			log.Debug().Err(err).Msg("Bubble Tea program failed; resetting progress handler so it can retry")
			h.resetAfterRunFailure(program)
		}
	}()
}

// resetAfterRunFailure clears handler state after Run() fails so a later
// OnProgress can start a fresh program. It only resets if program is still
// the current one: a Stop() or a newer start() must not be clobbered.
func (h *TeaHandler) resetAfterRunFailure(program *tea.Program) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.program != program {
		return
	}
	h.program = nil
	h.started = false
	// start() closed the old ready channel; closing it twice would panic
	h.ready = make(chan struct{})
}

// Stop gracefully stops the Bubble Tea program.
// The handler can be reused after Stop: the next OnProgress starts a new program.
func (h *TeaHandler) Stop() {
	h.mu.Lock()
	program := h.program
	if program != nil {
		h.program = nil
		h.started = false
		// Recreate the ready channel so the handler can be restarted;
		// start() closed the old one, and closing twice would panic
		h.ready = make(chan struct{})
	}
	h.mu.Unlock()

	if program != nil {
		program.Quit()
		// Wait for Run() to return so the renderer goroutine has stopped
		// writing to h.out before the handler (or caller) reuses the writer.
		// Waiting outside the mutex keeps OnProgress/start from blocking.
		// Quit() sends on an unbuffered channel, so it blocks until Run's
		// event loop is receiving (or the program has died) — that is what
		// keeps Wait() from hanging when Stop races a just-started program
		// whose Run has not entered its loop yet.
		program.Wait()
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
