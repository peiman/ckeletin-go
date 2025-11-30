package progress

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogHandler(t *testing.T) {
	tests := []struct {
		name     string
		opts     []LogHandlerOption
		wantComp string
	}{
		{
			name:     "default options",
			opts:     nil,
			wantComp: "progress",
		},
		{
			name:     "custom component",
			opts:     []LogHandlerOption{WithComponent("custom")},
			wantComp: "custom",
		},
		{
			name:     "multiple options",
			opts:     []LogHandlerOption{WithComponent("test")},
			wantComp: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewLogHandler(tt.opts...)
			assert.NotNil(t, h)
			assert.Equal(t, tt.wantComp, h.Component)
		})
	}
}

func TestLogHandler_WithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf)

	h := NewLogHandler(WithLogger(logger))

	// Verify the logger is set
	ctx := context.Background()
	event := NewEvent(EventStart, "test message")
	h.OnProgress(ctx, event)

	assert.Contains(t, buf.String(), "test message")
}

func TestLogHandler_OnProgress_EventTypes(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		wantLevel string
	}{
		{
			name:      "start event logs info",
			eventType: EventStart,
			wantLevel: "info",
		},
		{
			name:      "progress event logs debug",
			eventType: EventProgress,
			wantLevel: "debug",
		},
		{
			name:      "complete event logs info",
			eventType: EventComplete,
			wantLevel: "info",
		},
		{
			name:      "error event logs error",
			eventType: EventError,
			wantLevel: "error",
		},
		{
			name:      "warning event logs warn",
			eventType: EventWarning,
			wantLevel: "warn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

			h := NewLogHandler(WithLogger(logger))
			ctx := context.Background()
			event := NewEvent(tt.eventType, "test message")
			h.OnProgress(ctx, event)

			// Parse the JSON log entry
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err)

			assert.Equal(t, tt.wantLevel, logEntry["level"])
			assert.Equal(t, "test message", logEntry["message"])
			assert.Equal(t, "progress", logEntry["component"])
			assert.Equal(t, tt.eventType.String(), logEntry["event_type"])
		})
	}
}

func TestLogHandler_OnProgress_OptionalFields(t *testing.T) {
	tests := []struct {
		name       string
		event      Event
		wantFields []string
		dontWant   []string
	}{
		{
			name:       "basic event has required fields only",
			event:      NewEvent(EventStart, "basic"),
			wantFields: []string{"component", "event_type", "event_time", "message"},
			dontWant:   []string{"phase", "task", "current", "total"},
		},
		{
			name:       "event with phase",
			event:      NewEvent(EventStart, "with phase").WithPhase("test-phase"),
			wantFields: []string{"component", "event_type", "event_time", "message", "phase"},
			dontWant:   []string{"task"},
		},
		{
			name:       "event with task",
			event:      NewEvent(EventStart, "with task").WithTask("test-task"),
			wantFields: []string{"component", "event_type", "event_time", "message", "task"},
			dontWant:   []string{"phase"},
		},
		{
			name:       "event with progress",
			event:      NewEvent(EventProgress, "with progress").WithProgress(50, 100),
			wantFields: []string{"component", "event_type", "event_time", "message", "current", "total", "percentage"},
			dontWant:   []string{"phase"},
		},
		{
			name:       "event with error",
			event:      NewEvent(EventError, "with error").WithError(errors.New("test error")),
			wantFields: []string{"component", "event_type", "event_time", "message", "error"},
			dontWant:   []string{"current"},
		},
		{
			name:       "event with metadata",
			event:      NewEvent(EventStart, "with metadata").WithMeta("key", "value"),
			wantFields: []string{"component", "event_type", "event_time", "message", "metadata"},
			dontWant:   []string{"phase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

			h := NewLogHandler(WithLogger(logger))
			ctx := context.Background()
			h.OnProgress(ctx, tt.event)

			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			require.NoError(t, err)

			for _, field := range tt.wantFields {
				assert.Contains(t, logEntry, field, "should have field %s", field)
			}

			for _, field := range tt.dontWant {
				assert.NotContains(t, logEntry, field, "should not have field %s", field)
			}
		})
	}
}

func TestLogHandler_OnProgress_ProgressValues(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

	h := NewLogHandler(WithLogger(logger))
	ctx := context.Background()
	event := NewEvent(EventProgress, "progress").WithProgress(25, 100)
	h.OnProgress(ctx, event)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	assert.Equal(t, float64(25), logEntry["current"])
	assert.Equal(t, float64(100), logEntry["total"])
	assert.Equal(t, float64(25), logEntry["percentage"])
}

func TestLogHandler_OnProgress_Metadata(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

	h := NewLogHandler(WithLogger(logger))
	ctx := context.Background()
	event := NewEvent(EventStart, "with meta").
		WithMeta("key1", "value1").
		WithMeta("key2", 42)
	h.OnProgress(ctx, event)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	metadata, ok := logEntry["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata should be a map")
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, float64(42), metadata["key2"])
}

func TestLogHandler_OnProgress_Timestamp(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

	h := NewLogHandler(WithLogger(logger))
	ctx := context.Background()

	before := time.Now().Add(-time.Second) // Give buffer for timing
	event := NewEvent(EventStart, "test")
	h.OnProgress(ctx, event)
	after := time.Now().Add(time.Second)

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	require.NoError(t, err)

	// Parse event_time
	eventTimeStr, ok := logEntry["event_time"].(string)
	require.True(t, ok, "event_time should be a string")

	eventTime, err := time.Parse(time.RFC3339Nano, eventTimeStr)
	require.NoError(t, err)

	assert.True(t, eventTime.After(before), "event time should be after before")
	assert.True(t, eventTime.Before(after), "event time should be before after")
}

func TestLogHandler_ImplementsHandler(t *testing.T) {
	var _ Handler = (*LogHandler)(nil)
}

func TestLogHandler_OnProgress_ContextCancellation(t *testing.T) {
	var buf bytes.Buffer
	logger := zerolog.New(&buf).Level(zerolog.DebugLevel)

	h := NewLogHandler(WithLogger(logger))

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := NewEvent(EventStart, "should not log")
	h.OnProgress(ctx, event)

	// Buffer should be empty since context was cancelled
	assert.Empty(t, buf.String(), "should not log when context is cancelled")
}
