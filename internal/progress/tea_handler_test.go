package progress

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTeaHandler(t *testing.T) {
	var buf bytes.Buffer
	h := NewTeaHandler(&buf)

	assert.NotNil(t, h)
	assert.NotNil(t, h.style)
	assert.NotNil(t, h.model)
	assert.NotNil(t, h.ready)
	assert.False(t, h.started)
}

func TestTeaHandler_Stop(t *testing.T) {
	var buf bytes.Buffer
	h := NewTeaHandler(&buf)

	// Stop before start should not panic
	h.Stop()
	assert.False(t, h.started)
	assert.Nil(t, h.program)
}

func TestTeaHandler_ImplementsHandler(t *testing.T) {
	var _ Handler = (*TeaHandler)(nil)
}

func TestTeaModel_Init(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)

	cmd := m.Init()
	assert.NotNil(t, cmd, "Init should return a tick command")
}

func TestTeaModel_Update_KeyMsg(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantDone bool
	}{
		{
			name:     "ctrl+c quits",
			key:      "ctrl+c",
			wantDone: true,
		},
		{
			name:     "q quits",
			key:      "q",
			wantDone: true,
		},
		{
			name:     "other key does nothing",
			key:      "a",
			wantDone: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultStyle()
			m := newTeaModel(style)

			msg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)}
			if tt.key == "ctrl+c" {
				msg = tea.KeyMsg{Type: tea.KeyCtrlC}
			}

			newModel, _ := m.Update(msg)
			updated := newModel.(*teaModel)

			assert.Equal(t, tt.wantDone, updated.done)
		})
	}
}

func TestTeaModel_Update_TickMsg(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)

	initialFrame := m.spinnerFrame

	// Send tick message
	msg := tickMsg(time.Now())
	newModel, cmd := m.Update(msg)
	updated := newModel.(*teaModel)

	// Should advance spinner frame
	expectedFrame := (initialFrame + 1) % len(style.SpinnerFrames)
	assert.Equal(t, expectedFrame, updated.spinnerFrame)

	// Should return another tick command
	assert.NotNil(t, cmd)
}

func TestTeaModel_Update_TickMsg_WhenDone(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.done = true

	initialFrame := m.spinnerFrame

	// Send tick message when done
	msg := tickMsg(time.Now())
	newModel, cmd := m.Update(msg)
	updated := newModel.(*teaModel)

	// Should not advance spinner frame
	assert.Equal(t, initialFrame, updated.spinnerFrame)

	// Should not return command
	assert.Nil(t, cmd)
}

func TestTeaModel_Update_ProgressEventMsg(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		wantDone  bool
	}{
		{
			name:      "start event does not mark done",
			eventType: EventStart,
			wantDone:  false,
		},
		{
			name:      "progress event does not mark done",
			eventType: EventProgress,
			wantDone:  false,
		},
		{
			name:      "complete event marks done",
			eventType: EventComplete,
			wantDone:  true,
		},
		{
			name:      "error event marks done",
			eventType: EventError,
			wantDone:  true,
		},
		{
			name:      "warning event does not mark done",
			eventType: EventWarning,
			wantDone:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultStyle()
			m := newTeaModel(style)

			event := NewEvent(tt.eventType, "test")
			msg := progressEventMsg{event: event}
			newModel, _ := m.Update(msg)
			updated := newModel.(*teaModel)

			assert.Equal(t, tt.wantDone, updated.done)
			assert.Equal(t, event.Message, updated.currentEvent.Message)
		})
	}
}

func TestTeaModel_View_EmptyMessage(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)

	view := m.View()
	assert.Equal(t, "", view)
}

func TestTeaModel_View_StartEvent(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventStart, "Loading...")

	view := m.View()

	// Should contain spinner frame
	assert.Contains(t, view, style.SpinnerFrames[0])
	assert.Contains(t, view, "Loading...")
	assert.True(t, strings.HasSuffix(view, "\n"))
}

func TestTeaModel_View_ProgressEvent_Indeterminate(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventProgress, "Processing...")

	view := m.View()

	// Should show spinner (indeterminate)
	assert.Contains(t, view, style.SpinnerFrames[0])
	assert.Contains(t, view, "Processing...")
}

func TestTeaModel_View_ProgressEvent_Determinate(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventProgress, "Downloading").WithProgress(50, 100)

	view := m.View()

	// Should show progress bar
	assert.Contains(t, view, "[")
	assert.Contains(t, view, "]")
	assert.Contains(t, view, "50/100")
	assert.Contains(t, view, "50%")
}

func TestTeaModel_View_CompleteEvent(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventComplete, "Done!")

	view := m.View()

	assert.Contains(t, view, "✓")
	assert.Contains(t, view, "Done!")
}

func TestTeaModel_View_ErrorEvent(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventError, "Failed!")

	view := m.View()

	assert.Contains(t, view, "✗")
	assert.Contains(t, view, "Failed!")
}

func TestTeaModel_View_WarningEvent(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventWarning, "Warning!")

	view := m.View()

	assert.Contains(t, view, "⚠")
	assert.Contains(t, view, "Warning!")
}

func TestTeaModel_View_WithPhase(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)
	m.currentEvent = NewEvent(EventStart, "Loading").WithPhase("download")

	view := m.View()

	assert.Contains(t, view, "download")
	assert.Contains(t, view, "Loading")
}

func TestTeaModel_RenderBar(t *testing.T) {
	tests := []struct {
		name    string
		current int64
		total   int64
		wantBar bool
	}{
		{
			name:    "zero total returns empty",
			current: 0,
			total:   0,
			wantBar: false,
		},
		{
			name:    "negative total returns empty",
			current: 0,
			total:   -1,
			wantBar: false,
		},
		{
			name:    "valid progress returns bar",
			current: 50,
			total:   100,
			wantBar: true,
		},
		{
			name:    "overflow capped",
			current: 150,
			total:   100,
			wantBar: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultStyle()
			m := newTeaModel(style)
			m.currentEvent = NewEvent(EventProgress, "test").WithProgress(tt.current, tt.total)

			bar := m.renderBar()

			if tt.wantBar {
				assert.Contains(t, bar, "[")
				assert.Contains(t, bar, "]")
			} else {
				assert.Equal(t, "", bar)
			}
		})
	}
}

func TestTeaModel_RenderCounter(t *testing.T) {
	tests := []struct {
		name        string
		current     int64
		total       int64
		wantCounter bool
		wantText    string
	}{
		{
			name:        "zero total returns empty",
			current:     0,
			total:       0,
			wantCounter: false,
		},
		{
			name:        "valid progress shows counter",
			current:     25,
			total:       100,
			wantCounter: true,
			wantText:    "25/100",
		},
		{
			name:        "complete shows 100%",
			current:     100,
			total:       100,
			wantCounter: true,
			wantText:    "100%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			style := DefaultStyle()
			m := newTeaModel(style)
			m.currentEvent = NewEvent(EventProgress, "test").WithProgress(tt.current, tt.total)

			counter := m.renderCounter()

			if tt.wantCounter {
				assert.NotEmpty(t, counter)
				if tt.wantText != "" {
					assert.Contains(t, counter, tt.wantText)
				}
			} else {
				assert.Equal(t, "", counter)
			}
		})
	}
}

func TestTeaModel_SpinnerAnimation(t *testing.T) {
	style := DefaultStyle()
	m := newTeaModel(style)

	// Cycle through all spinner frames
	frames := make([]int, len(style.SpinnerFrames))
	for i := 0; i < len(style.SpinnerFrames); i++ {
		frames[i] = m.spinnerFrame
		m.Update(tickMsg(time.Now()))
	}

	// Should have cycled through all frames
	for i, frame := range frames {
		assert.Equal(t, i, frame, "frame %d should be %d", i, i)
	}

	// Should wrap around
	_, _ = m.Update(tickMsg(time.Now()))
	assert.Equal(t, 1, m.spinnerFrame, "should wrap to frame 1 (after 0)")
}

func TestTickCmd(t *testing.T) {
	interval := 100 * time.Millisecond
	cmd := tickCmd(interval)

	// The command should not be nil
	require.NotNil(t, cmd)
}

func TestProgressEventMsg(t *testing.T) {
	event := NewEvent(EventStart, "test")
	msg := progressEventMsg{event: event}

	assert.Equal(t, event.Type, msg.event.Type)
	assert.Equal(t, event.Message, msg.event.Message)
}

func TestTeaHandler_OnProgress_ContextCancellation(t *testing.T) {
	var buf bytes.Buffer
	h := NewTeaHandler(&buf)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := NewEvent(EventStart, "test")

	// Should return immediately without panic
	h.OnProgress(ctx, event)

	// Handler should not have started since context was cancelled
	assert.False(t, h.started)
}

func TestTeaHandler_OnProgress_StartsProgram(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	// Initially not started
	assert.False(t, h.started)

	// Send a start event - this triggers the start() method
	event := NewEvent(EventStart, "Starting operation")

	// Run in a goroutine since OnProgress may block waiting for program
	done := make(chan struct{})
	go func() {
		h.OnProgress(ctx, event)
		close(done)
	}()

	// Give it time to start
	time.Sleep(100 * time.Millisecond)

	h.mu.Lock()
	started := h.started
	h.mu.Unlock()

	// Program should have started
	assert.True(t, started, "program should have started after OnProgress")

	// Stop the handler
	h.Stop()

	// Wait for goroutine to complete
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// Timeout is acceptable since program may be blocked
	}
}

func TestTeaHandler_OnProgress_SendsEvent(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	// Start by sending multiple events
	done := make(chan struct{})
	go func() {
		// Send start event
		h.OnProgress(ctx, NewEvent(EventStart, "Starting"))

		// Send progress events
		h.OnProgress(ctx, NewEvent(EventProgress, "Working").WithProgress(1, 3))
		h.OnProgress(ctx, NewEvent(EventProgress, "Working").WithProgress(2, 3))
		h.OnProgress(ctx, NewEvent(EventProgress, "Working").WithProgress(3, 3))

		// Complete event should trigger quit
		h.OnProgress(ctx, NewEvent(EventComplete, "Done"))
		close(done)
	}()

	// Wait for events to be sent
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		// Timeout acceptable
	}

	// Stop handler
	h.Stop()
}

func TestTeaHandler_OnProgress_ErrorEvent(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		// Send start then error
		h.OnProgress(ctx, NewEvent(EventStart, "Starting"))
		h.OnProgress(ctx, NewEvent(EventError, "Failed"))
		close(done)
	}()

	// Wait for events
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
	}

	h.Stop()
}

func TestTeaHandler_OnProgress_ContextCancelDuringWait(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)

	ctx, cancel := context.WithCancel(context.Background())

	// Start the program first
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel() // Cancel while waiting for ready
	}()

	event := NewEvent(EventStart, "test")

	// OnProgress should handle context cancellation while waiting
	h.OnProgress(ctx, event)

	h.Stop()
}

func TestTeaHandler_OnProgress_MultipleStart(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		// Multiple calls should only start once
		h.OnProgress(ctx, NewEvent(EventStart, "First"))
		h.OnProgress(ctx, NewEvent(EventStart, "Second"))
		h.OnProgress(ctx, NewEvent(EventComplete, "Done"))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
	}

	h.Stop()
}

func TestTeaHandler_Stop_AfterStart(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	// Start the handler
	done := make(chan struct{})
	go func() {
		h.OnProgress(ctx, NewEvent(EventStart, "Starting"))
		close(done)
	}()

	// Wait a bit for start
	time.Sleep(100 * time.Millisecond)

	// Stop should work after starting
	h.Stop()

	// Verify stopped
	h.mu.Lock()
	assert.False(t, h.started)
	assert.Nil(t, h.program)
	h.mu.Unlock()

	<-done
}

func TestTeaHandler_OnProgress_AfterStop_Restarts(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)
	ctx := context.Background()

	// First lifecycle: start, then stop
	first := make(chan struct{})
	go func() {
		h.OnProgress(ctx, NewEvent(EventStart, "first run"))
		close(first)
	}()
	select {
	case <-first:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("first OnProgress did not complete")
	}
	h.Stop()

	// EXECUTION PHASE
	// Reusing the handler after Stop must restart cleanly; closing the
	// already-closed ready channel would panic
	second := make(chan struct{})
	go func() {
		h.OnProgress(ctx, NewEvent(EventStart, "second run"))
		close(second)
	}()
	select {
	case <-second:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("OnProgress after Stop did not complete; handler is not reusable")
	}

	// ASSERTION PHASE
	h.mu.Lock()
	started := h.started
	program := h.program
	h.mu.Unlock()
	assert.True(t, started, "handler should have restarted after Stop")
	assert.NotNil(t, program, "program should be running again after restart")

	h.Stop()
}

// postStopWriter is a goroutine-safe io.Writer that records whether any
// write arrives after the test flips the stopped flag.
type postStopWriter struct {
	mu       sync.Mutex
	buf      bytes.Buffer
	stopped  atomic.Bool
	postStop atomic.Bool
}

func (w *postStopWriter) Write(p []byte) (int, error) {
	if w.stopped.Load() {
		w.postStop.Store(true)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *postStopWriter) Len() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Len()
}

func TestTeaHandler_Stop_NoWritesAfterStopReturns(t *testing.T) {
	// SETUP PHASE
	w := &postStopWriter{}
	h := newTestTeaHandler(w)
	ctx := context.Background()

	// Drive the program with real events so the renderer is actively painting
	h.OnProgress(ctx, NewEvent(EventStart, "working"))
	h.OnProgress(ctx, NewEvent(EventProgress, "step").WithProgress(1, 2))

	require.Eventually(t, func() bool { return w.Len() > 0 },
		2*time.Second, 5*time.Millisecond,
		"renderer never wrote a frame; cannot exercise the shutdown race")

	// EXECUTION PHASE
	h.Stop()
	w.stopped.Store(true)

	// Give a leaked renderer goroutine time to write its shutdown sequence
	// (flush + erase-line happen after Quit is processed)
	time.Sleep(100 * time.Millisecond)

	// ASSERTION PHASE
	assert.False(t, w.postStop.Load(),
		"output written after Stop() returned; Stop must Wait() for Run to finish")
}

func TestTeaHandler_start_DoubleStart(t *testing.T) {
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)

	// Call start directly multiple times
	h.start()
	h.start() // Should be a no-op

	h.mu.Lock()
	assert.True(t, h.started)
	h.mu.Unlock()

	h.Stop()
}

// newTestTeaHandler returns a handler whose program runs without a TTY so
// Run() genuinely succeeds under go test/CI (bubbletea's default input
// expects a terminal).
func newTestTeaHandler(out io.Writer) *TeaHandler {
	h := NewTeaHandler(out)
	h.extraTeaOpts = []tea.ProgramOption{tea.WithInput(nil)}
	return h
}

func TestTeaHandler_start_StartedImpliesProgram(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	h := newTestTeaHandler(&buf)

	// A concurrent observer hunts for the TOCTOU window where started is
	// already true but program is still nil: a Stop() landing in that window
	// no-ops, the program launches anyway, and Stop's "no writes after
	// return" guarantee is void.
	var violated atomic.Bool
	checkerDone := make(chan struct{})
	stopChecker := make(chan struct{})
	go func() {
		defer close(checkerDone)
		for {
			select {
			case <-stopChecker:
				return
			default:
			}
			h.mu.Lock()
			if h.started && h.program == nil {
				violated.Store(true)
			}
			h.mu.Unlock()
		}
	}()

	// EXECUTION PHASE
	for i := 0; i < 50; i++ {
		h.start()
		h.Stop()
	}
	close(stopChecker)
	<-checkerDone

	// ASSERTION PHASE
	assert.False(t, violated.Load(),
		"observed started=true with nil program; start() must publish started and program in one critical section")
}

func TestTeaHandler_RunFailure_ResetsStateForRetry(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	h := NewTeaHandler(&buf)
	// A pre-cancelled program context makes Run() return ErrProgramKilled
	// immediately — the same failure shape as "no TTY available" in CI.
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	h.extraTeaOpts = []tea.ProgramOption{tea.WithInput(nil), tea.WithContext(cancelled)}

	// EXECUTION PHASE
	h.OnProgress(context.Background(), NewEvent(EventStart, "doomed"))

	// ASSERTION PHASE
	// Without the reset, started stays true forever and every later
	// OnProgress silently no-ops (dead handler).
	assert.Eventually(t, func() bool {
		h.mu.Lock()
		defer h.mu.Unlock()
		return !h.started && h.program == nil
	}, 2*time.Second, 5*time.Millisecond,
		"Run() failure must reset handler state so a later OnProgress can retry")
}

func TestTeaHandler_resetAfterRunFailure_IgnoresStaleProgram(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	h := NewTeaHandler(&buf)
	// tea.NewProgram only constructs; neither program is run
	stale := tea.NewProgram(newTeaModel(DefaultStyle()), tea.WithInput(nil), tea.WithOutput(&buf))
	current := tea.NewProgram(newTeaModel(DefaultStyle()), tea.WithInput(nil), tea.WithOutput(&buf))
	h.mu.Lock()
	h.started = true
	h.program = current
	readyBefore := h.ready
	h.mu.Unlock()

	// EXECUTION PHASE: a stale program's failure must not clobber the
	// current program's state
	h.resetAfterRunFailure(stale)

	// ASSERTION PHASE
	h.mu.Lock()
	assert.True(t, h.started, "stale reset must not clear started")
	assert.Same(t, current, h.program, "stale reset must not clear the current program")
	assert.True(t, readyBefore == h.ready, "stale reset must not replace the ready channel")
	h.mu.Unlock()

	// The current program's failure resets everything
	h.resetAfterRunFailure(current)

	h.mu.Lock()
	assert.False(t, h.started, "reset must clear started so OnProgress can retry")
	assert.Nil(t, h.program, "reset must clear the program")
	assert.True(t, readyBefore != h.ready, "reset must recreate the ready channel; the old one gets closed by start()")
	h.mu.Unlock()
}
