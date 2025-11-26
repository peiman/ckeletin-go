package progress

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockHandler(t *testing.T) {
	t.Run("records events", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		ctx := context.Background()

		// EXECUTION
		mock.OnProgress(ctx, NewEvent(EventStart, "start"))
		mock.OnProgress(ctx, NewEvent(EventProgress, "progress"))
		mock.OnProgress(ctx, NewEvent(EventComplete, "complete"))

		// ASSERTION
		assert.Equal(t, 3, mock.EventCount())
	})

	t.Run("LastEvent returns last event", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		ctx := context.Background()

		// EXECUTION
		mock.OnProgress(ctx, NewEvent(EventStart, "first"))
		mock.OnProgress(ctx, NewEvent(EventComplete, "last"))

		// ASSERTION
		event, ok := mock.LastEvent()
		require.True(t, ok)
		assert.Equal(t, "last", event.Message)
	})

	t.Run("LastEvent returns false when empty", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()

		// EXECUTION & ASSERTION
		_, ok := mock.LastEvent()
		assert.False(t, ok)
	})

	t.Run("Reset clears events", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		mock.OnProgress(context.Background(), NewEvent(EventStart, "test"))
		require.Equal(t, 1, mock.EventCount())

		// EXECUTION
		mock.Reset()

		// ASSERTION
		assert.Equal(t, 0, mock.EventCount())
	})

	t.Run("EventsOfType filters correctly", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		ctx := context.Background()
		mock.OnProgress(ctx, NewEvent(EventStart, "start"))
		mock.OnProgress(ctx, NewEvent(EventProgress, "progress1"))
		mock.OnProgress(ctx, NewEvent(EventProgress, "progress2"))
		mock.OnProgress(ctx, NewEvent(EventComplete, "complete"))

		// EXECUTION
		progressEvents := mock.EventsOfType(EventProgress)

		// ASSERTION
		require.Len(t, progressEvents, 2)
		assert.Equal(t, "progress1", progressEvents[0].Message)
		assert.Equal(t, "progress2", progressEvents[1].Message)
	})

	t.Run("GetEvents returns copy", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		mock.OnProgress(context.Background(), NewEvent(EventStart, "test"))

		// EXECUTION
		events := mock.GetEvents()
		events[0] = Event{} // Modify the copy

		// ASSERTION - original should be unchanged
		original := mock.GetEvents()
		assert.Equal(t, "test", original[0].Message)
	})

	t.Run("HasEventWithMessage finds matching event", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		mock.OnProgress(context.Background(), NewEvent(EventStart, "unique message"))

		// ASSERTION
		assert.True(t, mock.HasEventWithMessage("unique message"))
		assert.False(t, mock.HasEventWithMessage("nonexistent"))
	})

	t.Run("concurrent access is safe", func(t *testing.T) {
		// SETUP
		mock := NewMockHandler()
		ctx := context.Background()

		// EXECUTION - concurrent writes
		done := make(chan bool, 100)
		for i := 0; i < 100; i++ {
			go func() {
				mock.OnProgress(ctx, NewEvent(EventProgress, "test"))
				done <- true
			}()
		}

		// Wait for all
		for i := 0; i < 100; i++ {
			<-done
		}

		// ASSERTION
		assert.Equal(t, 100, mock.EventCount())
	})
}
