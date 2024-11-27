// internal/ui/ui_test.go

package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestGetLipglossColor(t *testing.T) {
	t.Run("Valid colors", func(t *testing.T) {
		for colorName, expected := range ColorMap {
			t.Run(colorName, func(t *testing.T) {
				got, err := GetLipglossColor(colorName)
				if err != nil {
					t.Errorf("GetLipglossColor(%q) returned unexpected error: %v", colorName, err)
				}
				if got != expected {
					t.Errorf("GetLipglossColor(%q) = %v, want %v", colorName, got, expected)
				}
			})
		}
	})

	t.Run("Invalid color", func(t *testing.T) {
		invalidColor := "not-a-color"
		got, err := GetLipglossColor(invalidColor)
		if err == nil {
			t.Errorf("GetLipglossColor(%q) did not return an error", invalidColor)
		}
		if got != "" {
			t.Errorf("GetLipglossColor(%q) = %v, want \"\"", invalidColor, got)
		}
	})
}

func TestRunUIWithMock(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		color      string
		mockError  error
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "Valid message and color",
			message:    "Hello, World!",
			color:      "red",
			mockError:  nil,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "Invalid color",
			message:    "Invalid Color Test",
			color:      "not-a-color",
			mockError:  errors.New("invalid color"),
			wantErr:    true,
			wantCalled: true,
		},
		{
			name:       "Empty message",
			message:    "",
			color:      "blue",
			mockError:  nil,
			wantErr:    false,
			wantCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRunner := &MockUIRunner{
				ReturnError: tt.mockError,
			}

			err := mockRunner.RunUI(tt.message, tt.color)

			// Check if RunUI was called
			if (mockRunner.CalledWithMessage != tt.message || mockRunner.CalledWithColor != tt.color) && tt.wantCalled {
				t.Errorf("RunUI() was not called with expected arguments. Got message=%q, color=%q, want message=%q, color=%q",
					mockRunner.CalledWithMessage, mockRunner.CalledWithColor, tt.message, tt.color)
			}

			// Validate the error returned
			if (err != nil) != tt.wantErr {
				t.Errorf("RunUI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestModelView(t *testing.T) {
	m := model{
		message:    "Test Message",
		colorStyle: lipgloss.NewStyle(),
	}

	expectedOutput := "Test Message\n\nPress 'q' or 'CTRL-C' to exit."

	if got := m.View(); got != expectedOutput {
		t.Errorf("View() = %q, want %q", got, expectedOutput)
	}
}

func TestModelUpdate(t *testing.T) {
	m := model{
		message:    "Test Message",
		colorStyle: lipgloss.NewStyle(),
		done:       false,
	}

	tests := []struct {
		name     string
		msg      tea.Msg
		wantDone bool
		wantCmd  bool
	}{
		{
			name:     "Key 'q' quits",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}},
			wantDone: false,
			wantCmd:  true,
		},
		{
			name:     "CTRL+C quits",
			msg:      tea.KeyMsg{Type: tea.KeyCtrlC},
			wantDone: false,
			wantCmd:  true,
		},
		{
			name:     "Unhandled key",
			msg:      tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}},
			wantDone: false,
			wantCmd:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := m.Update(tt.msg)
			if updatedModel.(model).done != tt.wantDone {
				t.Errorf("Update() done = %v, want %v", updatedModel.(model).done, tt.wantDone)
			}

			// Check if a command was returned
			if (cmd != nil) != tt.wantCmd {
				t.Errorf("Update() cmd returned = %v, want %v", cmd != nil, tt.wantCmd)
			}
		})
	}
}
