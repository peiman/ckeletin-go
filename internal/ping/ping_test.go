// internal/ping/ping_test.go

package ping

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			assert.Equal(t, tt.expected.Message, tt.cfg.Message)
			assert.Equal(t, tt.expected.Color, tt.cfg.Color)
			assert.Equal(t, tt.expected.UI, tt.cfg.UI)
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
			wantOutput: "✔ Pong! Test Message\n",
			wantErr:    false,
		},
		{
			name: "Successful output - red",
			cfg: Config{
				Message: "Red Message",
				Color:   "red",
				UI:      false,
			},
			wantOutput: "✔ Pong! Red Message\n",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE: Create output buffer and executor
			outBuf := &bytes.Buffer{}
			mockRunner := &ui.MockUIRunner{}
			executor := NewExecutor(tt.cfg, mockRunner, outBuf)

			// EXECUTION PHASE: Execute the command
			err := executor.Execute()

			// ASSERTION PHASE: Check results
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			got := outBuf.String()
			assert.Equal(t, tt.wantOutput, got)

			// Verify UI runner was not called
			assert.Empty(t, mockRunner.CalledWithMessage, "UI runner should not be called in non-UI mode")
			assert.Empty(t, mockRunner.CalledWithColor, "UI runner should not be called in non-UI mode")
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
			mockRunner := &ui.MockUIRunner{ReturnError: tt.uiRunnerError}
			outBuf := &bytes.Buffer{}
			executor := NewExecutor(tt.cfg, mockRunner, outBuf)

			// EXECUTION PHASE: Execute the command
			err := executor.Execute()

			// ASSERTION PHASE: Check results
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify UI runner was called with correct parameters
			assert.Equal(t, tt.wantUIMessage, mockRunner.CalledWithMessage)
			assert.Equal(t, tt.wantUIColor, mockRunner.CalledWithColor)

			// Verify nothing was written to output in UI mode (UNLESS there was an error)
			if !tt.wantErr {
				assert.Empty(t, outBuf.String(), "Output buffer should be empty in UI mode")
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
	mockRunner := &ui.MockUIRunner{}
	executor := NewExecutor(cfg, mockRunner, writer)

	// EXECUTION PHASE: Execute the command
	err := executor.Execute()

	// ASSERTION PHASE: Check for expected error
	require.Error(t, err, "Execute() expected error, got nil")
	assert.True(t, strings.Contains(err.Error(), "failed to write output"),
		"Execute() error = %v, expected to contain 'failed to write output'", err)
}

func TestExecutor_Execute_InvalidColor(t *testing.T) {
	// SETUP PHASE: Create executor with invalid color
	outBuf := &bytes.Buffer{}
	cfg := Config{
		Message: "Test Message",
		Color:   "invalid_color",
		UI:      false,
	}
	mockRunner := &ui.MockUIRunner{}
	executor := NewExecutor(cfg, mockRunner, outBuf)

	// EXECUTION PHASE: Execute the command
	err := executor.Execute()

	// ASSERTION PHASE: Check for expected error
	require.Error(t, err, "Execute() expected error for invalid color, got nil")
	assert.True(t, strings.Contains(err.Error(), "invalid color"),
		"Execute() error = %v, expected to contain 'invalid color'", err)
}
