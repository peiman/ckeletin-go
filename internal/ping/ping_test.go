// internal/ping/ping_test.go

package ping

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/progress"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

func TestExecutor_Execute_NonUIMode(t *testing.T) {
	// SETUP PHASE: Setup logging with cleanup to prevent race conditions
	_, cleanup := setupTestLogger(t)
	defer cleanup()

	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Successful output - white",
			cfg: Config{
				Message: "Test Message",
				Color:   "white",
				UI:      false,
			},
		},
		{
			name: "Successful output - red",
			cfg: Config{
				Message: "Red Message",
				Color:   "red",
				UI:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE: Create output buffer and executor
			mockRunner := &ui.MockUIRunner{}
			var rec progress.Recorder
			executor := NewExecutor(tt.cfg, mockRunner)

			// EXECUTION PHASE: Execute the command
			response, err := executor.Execute(&rec)

			// ASSERTION PHASE: Check results
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, tt.cfg.Message, response.Message)
			assert.Equal(t, "step", rec.Events[len(rec.Events)-1].Kind)

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
			var rec progress.Recorder
			executor := NewExecutor(tt.cfg, mockRunner)

			// EXECUTION PHASE: Execute the command
			response, err := executor.Execute(&rec)

			// ASSERTION PHASE: Check results
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, tt.cfg.Message, response.Message)
			}

			// Verify UI runner was called with correct parameters
			assert.Equal(t, tt.wantUIMessage, mockRunner.CalledWithMessage)
			assert.Equal(t, tt.wantUIColor, mockRunner.CalledWithColor)

			if tt.wantErr {
				assert.Equal(t, "warn", rec.Events[len(rec.Events)-1].Kind)
			}
		})
	}
}

func TestExecutor_Execute_NilReporter(t *testing.T) {
	executor := NewExecutor(Config{Message: "Test Message", Color: "white"}, &ui.MockUIRunner{})
	response, err := executor.Execute(nil)
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Message", response.Message)
}

func TestExecutor_Execute_InvalidColor(t *testing.T) {
	// SETUP PHASE: Create executor with invalid color
	cfg := Config{
		Message: "Test Message",
		Color:   "invalid_color",
		UI:      false,
	}
	mockRunner := &ui.MockUIRunner{}
	var rec progress.Recorder
	executor := NewExecutor(cfg, mockRunner)

	// EXECUTION PHASE: Execute the command
	_, err := executor.Execute(&rec)

	// ASSERTION PHASE: Check for expected error
	require.Error(t, err, "Execute() expected error for invalid color, got nil")
	assert.True(t, strings.Contains(err.Error(), "invalid color"),
		"Execute() error = %v, expected to contain 'invalid color'", err)
}
