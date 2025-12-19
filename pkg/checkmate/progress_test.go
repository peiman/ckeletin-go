package checkmate

import (
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

func TestGetErrorSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short single line",
			input:    "error occurred",
			expected: "error occurred",
		},
		{
			name:     "multiline takes first",
			input:    "first line\nsecond line\nthird line",
			expected: "first line",
		},
		{
			name:     "long line truncated",
			input:    "this is a very long error message that should be truncated because it exceeds the maximum allowed length",
			expected: "this is a very long error message that should b...",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: "",
		},
		{
			name:     "exactly 50 chars",
			input:    "12345678901234567890123456789012345678901234567890",
			expected: "12345678901234567890123456789012345678901234567890",
		},
		{
			name:     "51 chars truncated",
			input:    "123456789012345678901234567890123456789012345678901",
			expected: "12345678901234567890123456789012345678901234567...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorSummary(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatErrorDetails(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected []string
	}{
		{
			name:     "short single line",
			input:    "error",
			maxWidth: 60,
			expected: []string{"error"},
		},
		{
			name:     "multiline",
			input:    "line one\nline two\nline three",
			maxWidth: 60,
			expected: []string{"line one", "line two", "line three"},
		},
		{
			name:     "empty lines filtered",
			input:    "line one\n\nline two\n\n\nline three",
			maxWidth: 60,
			expected: []string{"line one", "line two", "line three"},
		},
		{
			name:     "long line wrapped",
			input:    "this is a very long line that should be wrapped at word boundaries",
			maxWidth: 30,
			expected: []string{"this is a very long line that", "should be wrapped at word", "boundaries"},
		},
		{
			name:     "empty string",
			input:    "",
			maxWidth: 60,
			expected: nil,
		},
		{
			name:     "whitespace trimmed",
			input:    "  line one  \n  line two  ",
			maxWidth: 60,
			expected: []string{"line one", "line two"},
		},
		{
			name:     "truncated at maxLines",
			input:    "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12",
			maxWidth: 60,
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "... (truncated)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorDetails(tt.input, tt.maxWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckProgress_Fields(t *testing.T) {
	cp := CheckProgress{
		Name:        "test",
		Status:      CheckFailed,
		Progress:    0.5,
		Error:       assert.AnError,
		Remediation: "fix it",
	}

	assert.Equal(t, "test", cp.Name)
	assert.Equal(t, CheckFailed, cp.Status)
	assert.Equal(t, 0.5, cp.Progress)
	assert.Error(t, cp.Error)
	assert.Equal(t, "fix it", cp.Remediation)
}

func TestCheckUpdateMsg_Fields(t *testing.T) {
	msg := CheckUpdateMsg{
		Index:       1,
		Status:      CheckPassed,
		Progress:    1.0,
		Error:       nil,
		Remediation: "run fix",
	}

	assert.Equal(t, 1, msg.Index)
	assert.Equal(t, CheckPassed, msg.Status)
	assert.Equal(t, 1.0, msg.Progress)
	assert.Nil(t, msg.Error)
	assert.Equal(t, "run fix", msg.Remediation)
}

func TestCheckStatus_Constants(t *testing.T) {
	assert.Equal(t, CheckStatus(0), CheckPending)
	assert.Equal(t, CheckStatus(1), CheckRunning)
	assert.Equal(t, CheckStatus(2), CheckPassed)
	assert.Equal(t, CheckStatus(3), CheckFailed)
}

func TestNewProgressModel(t *testing.T) {
	checks := []string{"format", "lint", "test"}

	t.Run("creates model with correct checks", func(t *testing.T) {
		m := NewProgressModel("Quality", checks)
		assert.Equal(t, "Quality", m.title)
		assert.Len(t, m.checks, 3)
		for i, name := range checks {
			assert.Equal(t, name, m.checks[i].Name)
			assert.Equal(t, CheckPending, m.checks[i].Status)
		}
	})

	t.Run("applies WithSkipSummary option", func(t *testing.T) {
		m := NewProgressModel("Test", checks, WithSkipSummary())
		assert.True(t, m.skipSummary)
	})

	t.Run("empty checks creates empty model", func(t *testing.T) {
		m := NewProgressModel("Empty", []string{})
		assert.Len(t, m.checks, 0)
	})
}

func TestProgressModel_Update(t *testing.T) {
	checks := []string{"format", "lint", "test"}
	model := NewProgressModel("Test", checks)

	t.Run("CheckUpdateMsg updates check state", func(t *testing.T) {
		msg := CheckUpdateMsg{Index: 1, Status: CheckPassed, Progress: 1.0}
		updated, _ := model.Update(msg)
		m := updated.(ProgressModel)
		assert.Equal(t, CheckPassed, m.checks[1].Status)
		assert.Equal(t, 1.0, m.checks[1].Progress)
	})

	t.Run("CheckUpdateMsg with error sets error", func(t *testing.T) {
		msg := CheckUpdateMsg{Index: 0, Status: CheckFailed, Error: assert.AnError, Remediation: "fix it"}
		updated, _ := model.Update(msg)
		m := updated.(ProgressModel)
		assert.Equal(t, CheckFailed, m.checks[0].Status)
		assert.Equal(t, assert.AnError, m.checks[0].Error)
		assert.Equal(t, "fix it", m.checks[0].Remediation)
	})

	t.Run("CheckUpdateMsg ignores invalid index", func(t *testing.T) {
		msg := CheckUpdateMsg{Index: 99, Status: CheckPassed}
		updated, _ := model.Update(msg)
		m := updated.(ProgressModel)
		// Should not panic, just ignore
		assert.Len(t, m.checks, 3)
	})

	t.Run("DoneMsg sets done and quits", func(t *testing.T) {
		updated, cmd := model.Update(DoneMsg{})
		m := updated.(ProgressModel)
		assert.True(t, m.done)
		assert.NotNil(t, cmd) // Should return tea.Quit
	})

	t.Run("CoverageMsg updates coverage", func(t *testing.T) {
		msg := CoverageMsg{Coverage: 85.5}
		updated, _ := model.Update(msg)
		m := updated.(ProgressModel)
		assert.Equal(t, 85.5, m.coverage)
	})
}

func TestProgressModel_View(t *testing.T) {
	checks := []string{"format", "lint"}
	model := NewProgressModel("Quality", checks)

	t.Run("contains title", func(t *testing.T) {
		view := model.View()
		assert.Contains(t, view, "Quality")
	})

	t.Run("contains check names", func(t *testing.T) {
		view := model.View()
		assert.Contains(t, view, "format")
		assert.Contains(t, view, "lint")
	})

	t.Run("shows waiting for pending checks", func(t *testing.T) {
		view := model.View()
		assert.Contains(t, view, "waiting")
	})

	t.Run("shows checkmark for passed", func(t *testing.T) {
		model.checks[0].Status = CheckPassed
		view := model.View()
		assert.Contains(t, view, "✓")
	})

	t.Run("shows x for failed", func(t *testing.T) {
		model.checks[1].Status = CheckFailed
		view := model.View()
		assert.Contains(t, view, "✗")
	})

	t.Run("shows error summary for failed check", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test"})
		m.checks[0].Status = CheckFailed
		m.checks[0].Error = fmt.Errorf("something went wrong")
		view := m.View()
		assert.Contains(t, view, "something went wrong")
	})

	t.Run("shows duration for completed check", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test"})
		m.checks[0].Status = CheckPassed
		m.checks[0].Duration = 500 * time.Millisecond
		view := m.View()
		assert.Contains(t, view, "500ms")
	})

	t.Run("shows running indicator", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test"})
		m.checks[0].Status = CheckRunning
		m.checks[0].Progress = 0.5
		view := m.View()
		// Should show progress bar (filled blocks)
		assert.Contains(t, view, "█")
	})
}

func TestProgressModel_ViewDone(t *testing.T) {
	t.Run("shows summary when done", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test1", "test2"})
		m.checks[0].Status = CheckPassed
		m.checks[1].Status = CheckPassed
		m.done = true
		view := m.View()
		assert.Contains(t, view, "All Checks Passed")
	})

	t.Run("skips summary when skipSummary is set", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test"}, WithSkipSummary())
		m.checks[0].Status = CheckPassed
		m.done = true
		view := m.View()
		assert.NotContains(t, view, "All Checks Passed")
	})

	t.Run("shows failed count when checks fail", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test1", "test2"})
		m.checks[0].Status = CheckPassed
		m.checks[1].Status = CheckFailed
		m.done = true
		view := m.View()
		assert.Contains(t, view, "1/2 Checks Failed")
	})

	t.Run("shows coverage when available", func(t *testing.T) {
		m := NewProgressModel("Test", []string{"test"})
		m.checks[0].Status = CheckPassed
		m.coverage = 85.5
		m.done = true
		view := m.View()
		assert.Contains(t, view, "Coverage")
	})
}

func TestProgressModel_Init(t *testing.T) {
	m := NewProgressModel("Test", []string{"test"})
	cmd := m.Init()
	assert.NotNil(t, cmd)
}

func TestProgressModel_UpdateWindowSize(t *testing.T) {
	m := NewProgressModel("Test", []string{"test"})
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	um := updated.(ProgressModel)
	assert.Equal(t, 100, um.width)
}

func TestProgressModel_UpdateKeyMsg(t *testing.T) {
	m := NewProgressModel("Test", []string{"test"})

	t.Run("q quits", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
		assert.NotNil(t, cmd)
	})

	t.Run("ctrl+c quits", func(t *testing.T) {
		_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
		assert.NotNil(t, cmd)
	})
}

func TestProgressModel_UpdateSpinnerTick(t *testing.T) {
	m := NewProgressModel("Test", []string{"test"})
	// Trigger a spinner tick (the model uses spinner.Tick)
	updated, cmd := m.Update(spinner.TickMsg{Time: time.Now()})
	assert.NotNil(t, updated)
	assert.NotNil(t, cmd)
}

func TestProgressModel_UpdateTickMsg(t *testing.T) {
	m := NewProgressModel("Test", []string{"test"})

	t.Run("continues ticking when not done", func(t *testing.T) {
		_, cmd := m.Update(tickMsg(time.Now()))
		assert.NotNil(t, cmd) // Should return another tick command
	})

	t.Run("stops ticking when done", func(t *testing.T) {
		m.done = true
		_, cmd := m.Update(tickMsg(time.Now()))
		assert.Nil(t, cmd)
	})
}
