// internal/ui/ui_test.go

package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func TestGetLipglossColor(t *testing.T) {
	tests := []struct {
		colorName string
		wantErr   bool
	}{
		{"red", false},
		{"green", false},
		{"invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.colorName, func(t *testing.T) {
			color, err := GetLipglossColor(tt.colorName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetLipglossColor() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil {
				if _, ok := ColorMap[tt.colorName]; !ok {
					t.Errorf("Color %s should be valid", tt.colorName)
				}
				expectedColor := ColorMap[tt.colorName]
				if color != expectedColor {
					t.Errorf("Expected color %v, got %v", expectedColor, color)
				}
			}
		})
	}
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

// TestModelInit tests the Init method of the model
func TestModelInit(t *testing.T) {
	m := model{
		message:    "Test Message",
		colorStyle: lipgloss.NewStyle(),
	}

	// Init should return nil as it's a no-op in this implementation
	cmd := m.Init()

	if cmd != nil {
		t.Errorf("Init() = %v, want nil", cmd)
	}
}

// TestRunUI tests the basic error path of RunUI
func TestRunUI(t *testing.T) {
	tests := []struct {
		name    string
		message string
		color   string
		wantErr bool
	}{
		{
			name:    "Invalid color error",
			message: "Test Message",
			color:   "invalid-color",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := NewDefaultUIRunner()
			err := runner.RunUI(tt.message, tt.color)

			if (err != nil) != tt.wantErr {
				t.Errorf("RunUI() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestRunUISuccessPath tests the success path of RunUI using the test runner
func TestRunUISuccessPath(t *testing.T) {
	// Create a test runner with nil program factory
	runner := NewTestUIRunner()

	// Valid color to pass the GetLipglossColor check
	err := runner.RunUI("Test Message", "blue")

	// Should not return an error
	if err != nil {
		t.Errorf("RunUI() error = %v, want nil", err)
	}
}

// TestDefaultUIRunnerCreation tests creating a DefaultUIRunner
func TestDefaultUIRunnerCreation(t *testing.T) {
	runner := NewDefaultUIRunner()
	if runner == nil {
		t.Fatalf("NewDefaultUIRunner() returned nil")
	}
	if runner.newProgram == nil {
		t.Errorf("NewDefaultUIRunner() returned a runner with nil newProgram")
	}
}

// TestNewTestUIRunner tests creating a test UI runner
func TestNewTestUIRunner(t *testing.T) {
	runner := NewTestUIRunner()
	if runner == nil {
		t.Fatalf("NewTestUIRunner() returned nil")
	}
	if runner.newProgram != nil {
		t.Errorf("NewTestUIRunner() returned a runner with non-nil newProgram")
	}
}
