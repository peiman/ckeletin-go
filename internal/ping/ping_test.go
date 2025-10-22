// internal/ping/ping_test.go

package ping

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// mockUIRunner is a mock implementation of ui.UIRunner for testing
type mockUIRunner struct {
	CalledWithMessage string
	CalledWithColor   string
	ReturnError       error
}

func (m *mockUIRunner) RunUI(message, col string) error {
	m.CalledWithMessage = message
	m.CalledWithColor = col
	return m.ReturnError
}

// errorWriter always returns an error on Write
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

// setupTestLogger creates a test logger and returns a cleanup function
// that restores the original logger state. This prevents race conditions
// when tests run in parallel.
func setupTestLogger(t *testing.T) (*bytes.Buffer, func()) {
	t.Helper()

	// Save original logger
	oldLogger := log.Logger

	// Create test logger
	logBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(logBuf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	// Return buffer and cleanup function
	cleanup := func() {
		log.Logger = oldLogger
	}

	return logBuf, cleanup
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		expected Config
	}{
		{
			name: "Basic config",
			cfg: Config{
				Message: "Hello",
				Color:   "white",
				UI:      false,
			},
			expected: Config{
				Message: "Hello",
				Color:   "white",
				UI:      false,
			},
		},
		{
			name: "Custom config",
			cfg: Config{
				Message: "Custom",
				Color:   "red",
				UI:      true,
			},
			expected: Config{
				Message: "Custom",
				Color:   "red",
				UI:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.cfg.Message != tt.expected.Message {
				t.Errorf("Message = %v, want %v", tt.cfg.Message, tt.expected.Message)
			}
			if tt.cfg.Color != tt.expected.Color {
				t.Errorf("Color = %v, want %v", tt.cfg.Color, tt.expected.Color)
			}
			if tt.cfg.UI != tt.expected.UI {
				t.Errorf("UI = %v, want %v", tt.cfg.UI, tt.expected.UI)
			}
		})
	}
}

func TestExecutor_Execute_NonUIMode(t *testing.T) {
	// SETUP PHASE: Setup logging with cleanup to prevent race conditions
	_, cleanup := setupTestLogger(t)
	defer cleanup()

	tests := []struct {
		name       string
		cfg        Config
		wantOutput string
		wantErr    bool
	}{
		{
			name: "Successful output - white",
			cfg: Config{
				Message: "Test Message",
				Color:   "white",
				UI:      false,
			},
			wantOutput: "Test Message\n",
			wantErr:    false,
		},
		{
			name: "Successful output - red",
			cfg: Config{
				Message: "Red Message",
				Color:   "red",
				UI:      false,
			},
			wantOutput: "Red Message\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE: Create output buffer and executor
			outBuf := &bytes.Buffer{}
			mockRunner := &mockUIRunner{}
			executor := NewExecutor(tt.cfg, mockRunner, outBuf)

			// EXECUTION PHASE: Execute the command
			err := executor.Execute()

			// ASSERTION PHASE: Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			got := outBuf.String()
			if got != tt.wantOutput {
				t.Errorf("Execute() output = %q, want %q", got, tt.wantOutput)
			}

			// Verify UI runner was not called
			if mockRunner.CalledWithMessage != "" || mockRunner.CalledWithColor != "" {
				t.Error("UI runner should not be called in non-UI mode")
			}
		})
	}
}

func TestExecutor_Execute_UIMode(t *testing.T) {
	// SETUP PHASE: Setup logging with cleanup to prevent race conditions
	_, cleanup := setupTestLogger(t)
	defer cleanup()

	tests := []struct {
		name          string
		cfg           Config
		uiRunnerError error
		wantErr       bool
		wantUIMessage string
		wantUIColor   string
	}{
		{
			name: "Successful UI execution",
			cfg: Config{
				Message: "UI Message",
				Color:   "blue",
				UI:      true,
			},
			uiRunnerError: nil,
			wantErr:       false,
			wantUIMessage: "UI Message",
			wantUIColor:   "blue",
		},
		{
			name: "UI execution error",
			cfg: Config{
				Message: "UI Message",
				Color:   "red",
				UI:      true,
			},
			uiRunnerError: fmt.Errorf("ui error"),
			wantErr:       true,
			wantUIMessage: "UI Message",
			wantUIColor:   "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE: Create mock UI runner and executor
			mockRunner := &mockUIRunner{ReturnError: tt.uiRunnerError}
			outBuf := &bytes.Buffer{}
			executor := NewExecutor(tt.cfg, mockRunner, outBuf)

			// EXECUTION PHASE: Execute the command
			err := executor.Execute()

			// ASSERTION PHASE: Check results
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify UI runner was called with correct parameters
			if mockRunner.CalledWithMessage != tt.wantUIMessage {
				t.Errorf("UI runner called with message = %q, want %q", mockRunner.CalledWithMessage, tt.wantUIMessage)
			}
			if mockRunner.CalledWithColor != tt.wantUIColor {
				t.Errorf("UI runner called with color = %q, want %q", mockRunner.CalledWithColor, tt.wantUIColor)
			}

			// Verify nothing was written to output in UI mode
			if outBuf.Len() > 0 {
				t.Errorf("Output buffer should be empty in UI mode, got: %q", outBuf.String())
			}
		})
	}
}

func TestExecutor_Execute_WriteError(t *testing.T) {
	// SETUP PHASE: Setup error writer
	writer := &errorWriter{}
	cfg := Config{
		Message: "Test Message",
		Color:   "white",
		UI:      false,
	}
	mockRunner := &mockUIRunner{}
	executor := NewExecutor(cfg, mockRunner, writer)

	// EXECUTION PHASE: Execute the command
	err := executor.Execute()

	// ASSERTION PHASE: Check for expected error
	if err == nil {
		t.Error("Execute() expected error, got nil")
		return
	}
	if !strings.Contains(err.Error(), "failed to print colored message") {
		t.Errorf("Execute() error = %v, expected to contain 'failed to print colored message'", err)
	}
}

func TestExecutor_Execute_InvalidColor(t *testing.T) {
	// SETUP PHASE: Create executor with invalid color
	outBuf := &bytes.Buffer{}
	cfg := Config{
		Message: "Test Message",
		Color:   "invalid_color",
		UI:      false,
	}
	mockRunner := &mockUIRunner{}
	executor := NewExecutor(cfg, mockRunner, outBuf)

	// EXECUTION PHASE: Execute the command
	err := executor.Execute()

	// ASSERTION PHASE: Check for expected error
	if err == nil {
		t.Error("Execute() expected error for invalid color, got nil")
		return
	}
	if !strings.Contains(err.Error(), "failed to print colored message") {
		t.Errorf("Execute() error = %v, expected to contain 'failed to print colored message'", err)
	}
}
