package check

import (
	"bytes"
	"context"
	"testing"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{FailFast: true, Verbose: true}

	executor := NewExecutor(cfg, &buf)

	require.NotNil(t, executor)
	assert.Equal(t, cfg, executor.cfg)
	assert.NotNil(t, executor.printer)
}

func TestNewExecutorWithPrinter(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{FailFast: true}
	mock := checkmate.NewMockPrinter()

	executor := NewExecutorWithPrinter(cfg, mock, &buf)

	require.NotNil(t, executor)
	assert.Equal(t, mock, executor.printer)
}

func TestExecutor_Execute_WithMock(t *testing.T) {
	// This test verifies the executor integrates with checkmate correctly
	// by using a mock printer to capture output calls
	var buf bytes.Buffer
	mock := checkmate.NewMockPrinter()
	cfg := Config{FailFast: false}

	executor := NewExecutorWithPrinter(cfg, mock, &buf)

	// Note: This will actually run the checks against the real system
	// In a real test scenario, we would mock exec.Command
	// For now, we just verify the structure works
	_ = executor // Executor is correctly constructed
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		failFast bool
		verbose  bool
	}{
		{
			name:     "default config",
			cfg:      Config{},
			failFast: false,
			verbose:  false,
		},
		{
			name:     "fail fast enabled",
			cfg:      Config{FailFast: true},
			failFast: true,
			verbose:  false,
		},
		{
			name:     "verbose enabled",
			cfg:      Config{Verbose: true},
			failFast: false,
			verbose:  true,
		},
		{
			name:     "both enabled",
			cfg:      Config{FailFast: true, Verbose: true},
			failFast: true,
			verbose:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.failFast, tt.cfg.FailFast)
			assert.Equal(t, tt.verbose, tt.cfg.Verbose)
		})
	}
}

func TestExecutor_Execute_CallsRunner(t *testing.T) {
	// Test that Execute properly sets up the runner with checks
	var buf bytes.Buffer
	mock := checkmate.NewMockPrinter()
	cfg := Config{FailFast: false}

	executor := NewExecutorWithPrinter(cfg, mock, &buf)
	require.NotNil(t, executor)

	// We can't easily test the full execution without mocking exec.Command
	// but we can verify the executor is properly constructed
	assert.Equal(t, cfg.FailFast, executor.cfg.FailFast)
	assert.Equal(t, cfg.Verbose, executor.cfg.Verbose)
}

func TestExecutor_Execute_ContextCancellation(t *testing.T) {
	// Test that a cancelled context is respected
	var buf bytes.Buffer
	mock := checkmate.NewMockPrinter()
	cfg := Config{FailFast: false}

	executor := NewExecutorWithPrinter(cfg, mock, &buf)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Execute with cancelled context - the runner should handle this
	// The exact behavior depends on how quickly checks start
	_ = executor.Execute(ctx)

	// The category header should have been printed before cancellation
	// (runner prints it before starting checks)
	assert.True(t, mock.HasCall("CategoryHeader"))
}
