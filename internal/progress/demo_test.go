//go:build dev

package progress

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// demoTestDelay keeps simulated demo work fast in tests.
const demoTestDelay = time.Millisecond

func TestRunDemo_Selection(t *testing.T) {
	tests := []struct {
		name         string
		opts         DemoOptions
		wantPhases   []string
		absentPhases []string
	}{
		{
			name:       "runs all demos by default",
			opts:       DemoOptions{Delay: demoTestDelay},
			wantPhases: []string{"spinner-demo", "progress-demo", "download", "compile", "package"},
		},
		{
			name:         "spinner only",
			opts:         DemoOptions{SpinnerOnly: true, Delay: demoTestDelay},
			wantPhases:   []string{"spinner-demo"},
			absentPhases: []string{"progress-demo", "download", "compile", "package"},
		},
		{
			name:         "bar only",
			opts:         DemoOptions{BarOnly: true, Delay: demoTestDelay},
			wantPhases:   []string{"progress-demo"},
			absentPhases: []string{"spinner-demo", "download", "compile", "package"},
		},
		{
			name:         "spinner and bar skip multi-phase",
			opts:         DemoOptions{SpinnerOnly: true, BarOnly: true, Delay: demoTestDelay},
			wantPhases:   []string{"spinner-demo", "progress-demo"},
			absentPhases: []string{"download", "compile", "package"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			mock := NewMockHandler()
			reporter := NewReporter(WithHandler(mock))

			// EXECUTION
			err := RunDemo(context.Background(), reporter, tt.opts)

			// ASSERTION
			require.NoError(t, err)
			phases := make(map[string]bool)
			for _, e := range mock.GetEvents() {
				phases[e.Phase] = true
			}
			for _, phase := range tt.wantPhases {
				assert.True(t, phases[phase], "phase %q should be reported", phase)
			}
			for _, phase := range tt.absentPhases {
				assert.False(t, phases[phase], "phase %q should not be reported", phase)
			}
		})
	}
}

func TestRunDemo_EventSequence(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))

	// EXECUTION
	err := RunDemo(context.Background(), reporter, DemoOptions{Delay: demoTestDelay})

	// ASSERTION
	require.NoError(t, err)
	// Spinner, progress bar, and three multi-phase phases each emit one Start/Complete pair.
	assert.Len(t, mock.EventsOfType(EventStart), 5, "spinner + bar + 3 phases should each start once")
	assert.Len(t, mock.EventsOfType(EventComplete), 5, "spinner + bar + 3 phases should each complete once")
	// 5 progress bar items + 3+4+2 multi-phase steps.
	assert.Len(t, mock.EventsOfType(EventProgress), 14, "bar items and phase steps should all report progress")
	assert.True(t, mock.HasEventWithMessage("Simulating network request..."), "spinner demo should report its start message")
	assert.True(t, mock.HasEventWithMessage("All items processed successfully"), "bar demo should report its completion message")
}

func TestRunDemo_ContextCancellation(t *testing.T) {
	tests := []struct {
		name string
		opts DemoOptions
	}{
		{name: "all demos", opts: DemoOptions{Delay: 100 * time.Millisecond}},
		{name: "spinner only", opts: DemoOptions{SpinnerOnly: true, Delay: 100 * time.Millisecond}},
		{name: "bar only", opts: DemoOptions{BarOnly: true, Delay: 100 * time.Millisecond}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			mock := NewMockHandler()
			reporter := NewReporter(WithHandler(mock))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// EXECUTION
			err := RunDemo(ctx, reporter, tt.opts)

			// ASSERTION
			assert.ErrorIs(t, err, context.Canceled, "cancelled context should abort the demo")
		})
	}
}

func TestDemoFuncs_ContextCancellation(t *testing.T) {
	tests := []struct {
		name string
		fn   func(context.Context, *Reporter, demoConfig) error
	}{
		{name: "demoSpinner", fn: demoSpinner},
		{name: "demoProgressBar", fn: demoProgressBar},
		{name: "demoMultiPhase", fn: demoMultiPhase},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			mock := NewMockHandler()
			reporter := NewReporter(WithHandler(mock))
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// EXECUTION
			err := tt.fn(ctx, reporter, newDemoConfig(100*time.Millisecond))

			// ASSERTION
			assert.ErrorIs(t, err, context.Canceled, "cancelled context should abort the demo")
		})
	}
}

func TestNewDemoConfig(t *testing.T) {
	tests := []struct {
		name  string
		delay time.Duration
		want  demoConfig
	}{
		{
			name:  "defaults when no override",
			delay: 0,
			want: demoConfig{
				spinnerDuration: defaultSpinnerDuration,
				stepDelay:       defaultStepDelay,
				phaseStepDelay:  defaultPhaseStepDelay,
			},
		},
		{
			name:  "override applies to all durations",
			delay: 25 * time.Millisecond,
			want: demoConfig{
				spinnerDuration: 25 * time.Millisecond,
				stepDelay:       25 * time.Millisecond,
				phaseStepDelay:  25 * time.Millisecond,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP + EXECUTION
			got := newDemoConfig(tt.delay)

			// ASSERTION
			assert.Equal(t, tt.want, got)
		})
	}
}
