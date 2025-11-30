package progress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNopHandler(t *testing.T) {
	h := NewNopHandler()
	assert.NotNil(t, h)
}

func TestNopHandler_OnProgress(t *testing.T) {
	h := NewNopHandler()
	ctx := context.Background()

	// All event types should work without panic
	eventTypes := []EventType{
		EventStart,
		EventProgress,
		EventComplete,
		EventError,
		EventWarning,
	}

	for _, et := range eventTypes {
		t.Run(et.String(), func(t *testing.T) {
			event := NewEvent(et, "test message")

			// Should not panic
			h.OnProgress(ctx, event)
		})
	}
}

func TestNopHandler_OnProgress_WithAllFields(t *testing.T) {
	h := NewNopHandler()
	ctx := context.Background()

	// Create a fully populated event
	event := NewEvent(EventProgress, "test").
		WithPhase("phase").
		WithTask("task").
		WithProgress(50, 100).
		WithMeta("key", "value")

	// Should handle without issues
	h.OnProgress(ctx, event)
}

func TestNopHandler_OnProgress_CanceledContext(t *testing.T) {
	h := NewNopHandler()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := NewEvent(EventStart, "test")

	// Should not panic with canceled context
	h.OnProgress(ctx, event)
}

func TestNopHandler_ImplementsHandler(t *testing.T) {
	var _ Handler = (*NopHandler)(nil)
}

func TestNopHandler_MultipleCallsSafe(t *testing.T) {
	h := NewNopHandler()
	ctx := context.Background()

	// Call many times - should be safe
	for i := 0; i < 1000; i++ {
		event := NewEvent(EventProgress, "test").WithProgress(int64(i), 1000)
		h.OnProgress(ctx, event)
	}
}
