package progress

import (
	"bytes"
	"strings"
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
