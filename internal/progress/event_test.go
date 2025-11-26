package progress

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name     string
		et       EventType
		expected string
	}{
		{"start", EventStart, "start"},
		{"progress", EventProgress, "progress"},
		{"complete", EventComplete, "complete"},
		{"error", EventError, "error"},
		{"warning", EventWarning, "warning"},
		{"unknown", EventType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.et.String())
		})
	}
}

func TestNewEvent(t *testing.T) {
	// SETUP
	msg := "Test message"

	// EXECUTION
	event := NewEvent(EventStart, msg)

	// ASSERTION
	assert.Equal(t, EventStart, event.Type)
	assert.Equal(t, msg, event.Message)
	assert.NotZero(t, event.Timestamp)
	assert.NotNil(t, event.Metadata)
	assert.Empty(t, event.Metadata)
}

func TestEvent_WithPhase(t *testing.T) {
	// SETUP
	event := NewEvent(EventProgress, "test")

	// EXECUTION
	result := event.WithPhase("downloading")

	// ASSERTION
	assert.Equal(t, "downloading", result.Phase)
	assert.Empty(t, event.Phase) // Original unchanged
}

func TestEvent_WithTask(t *testing.T) {
	// SETUP
	event := NewEvent(EventProgress, "test")

	// EXECUTION
	result := event.WithTask("file.zip")

	// ASSERTION
	assert.Equal(t, "file.zip", result.Task)
	assert.Empty(t, event.Task) // Original unchanged
}

func TestEvent_WithProgress(t *testing.T) {
	// SETUP
	event := NewEvent(EventProgress, "test")

	// EXECUTION
	result := event.WithProgress(50, 100)

	// ASSERTION
	assert.Equal(t, int64(50), result.Current)
	assert.Equal(t, int64(100), result.Total)
	assert.Zero(t, event.Current) // Original unchanged
}

func TestEvent_WithError(t *testing.T) {
	// SETUP
	event := NewEvent(EventError, "test")
	testErr := errors.New("test error")

	// EXECUTION
	result := event.WithError(testErr)

	// ASSERTION
	assert.Equal(t, testErr, result.Error)
	assert.Nil(t, event.Error) // Original unchanged
}

func TestEvent_WithMeta(t *testing.T) {
	// SETUP
	event := NewEvent(EventProgress, "test")

	// EXECUTION
	result := event.WithMeta("key1", "value1").WithMeta("key2", 42)

	// ASSERTION
	assert.Equal(t, "value1", result.Metadata["key1"])
	assert.Equal(t, 42, result.Metadata["key2"])
	assert.Empty(t, event.Metadata) // Original unchanged
}

func TestEvent_WithMeta_NilMetadata(t *testing.T) {
	// SETUP
	event := Event{Type: EventProgress, Message: "test"}
	// Metadata is nil

	// EXECUTION
	result := event.WithMeta("key", "value")

	// ASSERTION
	assert.NotNil(t, result.Metadata)
	assert.Equal(t, "value", result.Metadata["key"])
}

func TestEvent_Percentage(t *testing.T) {
	tests := []struct {
		name     string
		current  int64
		total    int64
		expected float64
	}{
		{"zero progress", 0, 100, 0},
		{"half progress", 50, 100, 50},
		{"full progress", 100, 100, 100},
		{"indeterminate", 50, 0, -1},
		{"over 100%", 150, 100, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(EventProgress, "test").WithProgress(tt.current, tt.total)
			assert.Equal(t, tt.expected, event.Percentage())
		})
	}
}

func TestEvent_IsIndeterminate(t *testing.T) {
	tests := []struct {
		name     string
		total    int64
		expected bool
	}{
		{"total zero", 0, true},
		{"total positive", 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := NewEvent(EventProgress, "test").WithProgress(50, tt.total)
			assert.Equal(t, tt.expected, event.IsIndeterminate())
		})
	}
}

func TestEvent_Chaining(t *testing.T) {
	// Test that methods can be chained
	event := NewEvent(EventProgress, "Processing file").
		WithPhase("download").
		WithTask("file.zip").
		WithProgress(50, 100).
		WithMeta("speed", 1024)

	assert.Equal(t, EventProgress, event.Type)
	assert.Equal(t, "Processing file", event.Message)
	assert.Equal(t, "download", event.Phase)
	assert.Equal(t, "file.zip", event.Task)
	assert.Equal(t, int64(50), event.Current)
	assert.Equal(t, int64(100), event.Total)
	assert.Equal(t, 1024, event.Metadata["speed"])
}

func TestEvent_TimestampIsSet(t *testing.T) {
	before := time.Now()
	event := NewEvent(EventStart, "test")
	after := time.Now()

	require.False(t, event.Timestamp.Before(before), "timestamp should be >= before")
	require.False(t, event.Timestamp.After(after), "timestamp should be <= after")
}
