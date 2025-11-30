package progress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandlerFunc(t *testing.T) {
	// SETUP
	var receivedEvent Event
	var receivedCtx context.Context

	handlerFn := HandlerFunc(func(ctx context.Context, event Event) {
		receivedCtx = ctx
		receivedEvent = event
	})

	ctx := context.Background()
	event := NewEvent(EventStart, "test")

	// EXECUTION
	handlerFn.OnProgress(ctx, event)

	// ASSERTION
	assert.Equal(t, ctx, receivedCtx)
	assert.Equal(t, event.Type, receivedEvent.Type)
	assert.Equal(t, event.Message, receivedEvent.Message)
}

func TestNopHandler(t *testing.T) {
	// SETUP
	handler := NewNopHandler()
	ctx := context.Background()
	event := NewEvent(EventStart, "test")

	// EXECUTION - should not panic
	handler.OnProgress(ctx, event)

	// ASSERTION - NopHandler is a no-op, nothing to assert
	// Just verify it doesn't panic
	assert.NotNil(t, handler)
}

func TestCompositeHandler(t *testing.T) {
	t.Run("dispatches to all handlers", func(t *testing.T) {
		// SETUP
		mock1 := NewMockHandler()
		mock2 := NewMockHandler()
		composite := NewCompositeHandler(mock1, mock2)

		ctx := context.Background()
		event := NewEvent(EventStart, "test")

		// EXECUTION
		composite.OnProgress(ctx, event)

		// ASSERTION
		assert.Equal(t, 1, mock1.EventCount())
		assert.Equal(t, 1, mock2.EventCount())
	})

	t.Run("handles nil handlers gracefully", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		composite := NewCompositeHandler(nil, mock, nil)

		ctx := context.Background()
		event := NewEvent(EventStart, "test")

		// EXECUTION - should not panic
		composite.OnProgress(ctx, event)

		// ASSERTION
		assert.Equal(t, 1, mock.EventCount())
	})

	t.Run("Add handler", func(t *testing.T) {
		// SETUP
		composite := NewCompositeHandler()
		mock := NewMockHandler()

		// EXECUTION
		composite.Add(mock)
		composite.OnProgress(context.Background(), NewEvent(EventStart, "test"))

		// ASSERTION
		assert.Equal(t, 1, mock.EventCount())
		assert.Len(t, composite.Handlers(), 1)
	})

	t.Run("Add nil handler is no-op", func(t *testing.T) {
		// SETUP
		composite := NewCompositeHandler()

		// EXECUTION
		composite.Add(nil)

		// ASSERTION
		assert.Len(t, composite.Handlers(), 0)
	})

	t.Run("RemoveAt handler", func(t *testing.T) {
		// SETUP
		mock1 := NewMockHandler()
		mock2 := NewMockHandler()
		composite := NewCompositeHandler(mock1, mock2)

		// EXECUTION - remove first handler (index 0)
		removed := composite.RemoveAt(0)
		composite.OnProgress(context.Background(), NewEvent(EventStart, "test"))

		// ASSERTION
		assert.True(t, removed)
		assert.Equal(t, 0, mock1.EventCount())
		assert.Equal(t, 1, mock2.EventCount())
		assert.Len(t, composite.Handlers(), 1)
	})

	t.Run("RemoveAt out of bounds is no-op", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		composite := NewCompositeHandler(mock)

		// EXECUTION
		removed := composite.RemoveAt(5)

		// ASSERTION
		assert.False(t, removed)
		assert.Len(t, composite.Handlers(), 1)
	})

	t.Run("Handlers returns copy", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		composite := NewCompositeHandler(mock)

		// EXECUTION
		handlers := composite.Handlers()
		handlers[0] = nil // Modify the copy

		// ASSERTION - original should be unchanged
		assert.NotNil(t, composite.Handlers()[0])
	})
}
