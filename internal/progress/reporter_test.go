package progress

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReporter(t *testing.T) {
	t.Run("creates with default handler", func(t *testing.T) {
		reporter := NewReporter()
		assert.NotNil(t, reporter)
	})

	t.Run("creates with custom handler", func(t *testing.T) {
		mock := NewMockHandler()
		reporter := NewReporter(WithHandler(mock))

		reporter.Start(context.Background(), "test")

		assert.Equal(t, 1, mock.EventCount())
	})
}

func TestReporter_Start(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.Start(ctx, "Starting operation")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, EventStart, event.Type)
	assert.Equal(t, "Starting operation", event.Message)
}

func TestReporter_Progress(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.Progress(ctx, 50, 100, "Halfway there")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, EventProgress, event.Type)
	assert.Equal(t, int64(50), event.Current)
	assert.Equal(t, int64(100), event.Total)
	assert.Equal(t, "Halfway there", event.Message)
}

func TestReporter_Complete(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.Complete(ctx, "Operation complete")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, EventComplete, event.Type)
	assert.Equal(t, "Operation complete", event.Message)
}

func TestReporter_Error(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()
	testErr := errors.New("test error")

	// EXECUTION
	reporter.Error(ctx, testErr, "Operation failed")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, EventError, event.Type)
	assert.Equal(t, testErr, event.Error)
	assert.Equal(t, "Operation failed", event.Message)
}

func TestReporter_Warning(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.Warning(ctx, "Low disk space")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, EventWarning, event.Type)
	assert.Equal(t, "Low disk space", event.Message)
}

func TestReporter_SetPhase(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.SetPhase("download")
	reporter.Start(ctx, "Downloading")

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, "download", event.Phase)
	assert.Equal(t, "download", reporter.Phase())
}

func TestReporter_Emit(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	customEvent := NewEvent(EventProgress, "Custom").
		WithPhase("custom-phase").
		WithMeta("custom-key", "custom-value")
	reporter.Emit(ctx, customEvent)

	// ASSERTION
	require.Equal(t, 1, mock.EventCount())
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, "custom-phase", event.Phase)
	assert.Equal(t, "custom-value", event.Metadata["custom-key"])
}

func TestReporter_ProgressFunc(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	progressFn := reporter.ProgressFunc(ctx, "Downloading")
	progressFn(25, 100)
	progressFn(50, 100)
	progressFn(100, 100)

	// ASSERTION
	events := mock.EventsOfType(EventProgress)
	require.Len(t, events, 3)
	assert.Equal(t, int64(25), events[0].Current)
	assert.Equal(t, int64(50), events[1].Current)
	assert.Equal(t, int64(100), events[2].Current)
}

func TestReporter_SetHandler(t *testing.T) {
	// SETUP
	mock1 := NewMockHandler()
	mock2 := NewMockHandler()
	reporter := NewReporter(WithHandler(mock1))
	ctx := context.Background()

	// EXECUTION
	reporter.Start(ctx, "test1")
	reporter.SetHandler(mock2)
	reporter.Start(ctx, "test2")

	// ASSERTION
	assert.Equal(t, 1, mock1.EventCount())
	assert.Equal(t, 1, mock2.EventCount())
}

func TestReporter_PhaseAppliedToEvents(t *testing.T) {
	// Test that phase is applied to events that don't have one

	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.SetPhase("phase1")
	reporter.Start(ctx, "test")

	// ASSERTION
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, "phase1", event.Phase)
}

func TestReporter_EventPhaseNotOverwritten(t *testing.T) {
	// Test that event's own phase is not overwritten by reporter's phase

	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.SetPhase("reporter-phase")
	customEvent := NewEvent(EventStart, "test").WithPhase("event-phase")
	reporter.Emit(ctx, customEvent)

	// ASSERTION
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.Equal(t, "event-phase", event.Phase)
}

func TestReporter_TimestampIsSet(t *testing.T) {
	// SETUP
	mock := NewMockHandler()
	reporter := NewReporter(WithHandler(mock))
	ctx := context.Background()

	// EXECUTION
	reporter.Start(ctx, "test")

	// ASSERTION
	event, ok := mock.LastEvent()
	require.True(t, ok)
	assert.False(t, event.Timestamp.IsZero())
}
