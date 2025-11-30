package progress

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCompositeHandler(t *testing.T) {
	tests := []struct {
		name      string
		handlers  []Handler
		wantCount int
	}{
		{
			name:      "no handlers",
			handlers:  nil,
			wantCount: 0,
		},
		{
			name:      "empty handlers",
			handlers:  []Handler{},
			wantCount: 0,
		},
		{
			name:      "single handler",
			handlers:  []Handler{NewNopHandler()},
			wantCount: 1,
		},
		{
			name:      "multiple handlers",
			handlers:  []Handler{NewNopHandler(), NewNopHandler(), NewNopHandler()},
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCompositeHandler(tt.handlers...)
			assert.NotNil(t, h)
			assert.Equal(t, tt.wantCount, len(h.Handlers()))
		})
	}
}

func TestCompositeHandler_OnProgress(t *testing.T) {
	// Create mock handlers to track calls
	var called1, called2, called3 bool
	h1 := HandlerFunc(func(ctx context.Context, event Event) {
		called1 = true
	})
	h2 := HandlerFunc(func(ctx context.Context, event Event) {
		called2 = true
	})
	h3 := HandlerFunc(func(ctx context.Context, event Event) {
		called3 = true
	})

	composite := NewCompositeHandler(h1, h2, h3)
	ctx := context.Background()
	event := NewEvent(EventStart, "test")

	composite.OnProgress(ctx, event)

	assert.True(t, called1, "handler 1 should be called")
	assert.True(t, called2, "handler 2 should be called")
	assert.True(t, called3, "handler 3 should be called")
}

func TestCompositeHandler_OnProgress_ReceivesCorrectEvent(t *testing.T) {
	var receivedEvent Event
	h := HandlerFunc(func(ctx context.Context, event Event) {
		receivedEvent = event
	})

	composite := NewCompositeHandler(h)
	ctx := context.Background()
	event := NewEvent(EventProgress, "test message").
		WithPhase("test-phase").
		WithProgress(50, 100)

	composite.OnProgress(ctx, event)

	assert.Equal(t, EventProgress, receivedEvent.Type)
	assert.Equal(t, "test message", receivedEvent.Message)
	assert.Equal(t, "test-phase", receivedEvent.Phase)
	assert.Equal(t, int64(50), receivedEvent.Current)
	assert.Equal(t, int64(100), receivedEvent.Total)
}

func TestCompositeHandler_OnProgress_SkipsNilHandlers(t *testing.T) {
	var called bool
	h := HandlerFunc(func(ctx context.Context, event Event) {
		called = true
	})

	// Create composite with nil in the middle
	composite := &CompositeHandler{
		handlers: []Handler{nil, h, nil},
	}

	ctx := context.Background()
	event := NewEvent(EventStart, "test")

	// Should not panic
	composite.OnProgress(ctx, event)

	assert.True(t, called, "non-nil handler should be called")
}

func TestCompositeHandler_OnProgress_EmptyHandlers(t *testing.T) {
	composite := NewCompositeHandler()
	ctx := context.Background()
	event := NewEvent(EventStart, "test")

	// Should not panic
	composite.OnProgress(ctx, event)
}

func TestCompositeHandler_Add(t *testing.T) {
	composite := NewCompositeHandler()
	assert.Equal(t, 0, len(composite.Handlers()))

	h1 := NewNopHandler()
	composite.Add(h1)
	assert.Equal(t, 1, len(composite.Handlers()))

	h2 := NewNopHandler()
	composite.Add(h2)
	assert.Equal(t, 2, len(composite.Handlers()))
}

func TestCompositeHandler_Add_NilHandler(t *testing.T) {
	composite := NewCompositeHandler()
	composite.Add(nil)
	assert.Equal(t, 0, len(composite.Handlers()), "nil handler should not be added")
}

func TestCompositeHandler_RemoveAt(t *testing.T) {
	// Use HandlerFunc with state to verify removal
	var called1, called2, called3 int
	h1 := HandlerFunc(func(ctx context.Context, event Event) { called1++ })
	h2 := HandlerFunc(func(ctx context.Context, event Event) { called2++ })
	h3 := HandlerFunc(func(ctx context.Context, event Event) { called3++ })

	composite := NewCompositeHandler(h1, h2, h3)
	assert.Equal(t, 3, composite.Len())

	// Remove the middle handler (index 1)
	removed := composite.RemoveAt(1)
	assert.True(t, removed)
	assert.Equal(t, 2, composite.Len())

	// Verify by calling OnProgress and checking counts
	ctx := context.Background()
	event := NewEvent(EventStart, "test")
	composite.OnProgress(ctx, event)

	assert.Equal(t, 1, called1, "h1 should be called")
	assert.Equal(t, 0, called2, "h2 should not be called (was removed)")
	assert.Equal(t, 1, called3, "h3 should be called")
}

func TestCompositeHandler_RemoveAt_OutOfBounds(t *testing.T) {
	composite := NewCompositeHandler(NewNopHandler())

	assert.False(t, composite.RemoveAt(-1), "negative index should return false")
	assert.False(t, composite.RemoveAt(1), "out of bounds index should return false")
	assert.False(t, composite.RemoveAt(100), "large index should return false")
	assert.Equal(t, 1, composite.Len(), "handler should not be removed")
}

func TestCompositeHandler_RemoveAt_Empty(t *testing.T) {
	composite := NewCompositeHandler()
	assert.False(t, composite.RemoveAt(0), "empty composite should return false")
}

func TestCompositeHandler_Len(t *testing.T) {
	composite := NewCompositeHandler()
	assert.Equal(t, 0, composite.Len())

	composite.Add(NewNopHandler())
	assert.Equal(t, 1, composite.Len())

	composite.Add(NewNopHandler())
	assert.Equal(t, 2, composite.Len())

	composite.RemoveAt(0)
	assert.Equal(t, 1, composite.Len())
}

func TestCompositeHandler_Handlers_ReturnsCopy(t *testing.T) {
	h1 := NewNopHandler()
	h2 := NewNopHandler()

	composite := NewCompositeHandler(h1, h2)

	handlers := composite.Handlers()
	handlers[0] = nil // Modify the copy

	// Original should be unchanged
	originalHandlers := composite.Handlers()
	assert.NotNil(t, originalHandlers[0], "original handler should not be modified")
}

func TestCompositeHandler_Concurrent_OnProgress(t *testing.T) {
	var callCount int64
	h := HandlerFunc(func(ctx context.Context, event Event) {
		atomic.AddInt64(&callCount, 1)
	})

	composite := NewCompositeHandler(h)
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			event := NewEvent(EventProgress, "test").WithProgress(int64(i), 100)
			composite.OnProgress(ctx, event)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, int64(100), atomic.LoadInt64(&callCount))
}

func TestCompositeHandler_Concurrent_AddRemove(t *testing.T) {
	composite := NewCompositeHandler()
	ctx := context.Background()

	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			composite.Add(NewNopHandler())
		}()
	}

	// Concurrent OnProgress calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			event := NewEvent(EventStart, "test")
			composite.OnProgress(ctx, event)
		}()
	}

	wg.Wait()

	// Should have 50 handlers
	assert.Equal(t, 50, len(composite.Handlers()))
}

func TestCompositeHandler_OrderPreserved(t *testing.T) {
	var order []int
	var mu sync.Mutex

	createHandler := func(id int) Handler {
		return HandlerFunc(func(ctx context.Context, event Event) {
			mu.Lock()
			order = append(order, id)
			mu.Unlock()
		})
	}

	composite := NewCompositeHandler(
		createHandler(1),
		createHandler(2),
		createHandler(3),
	)

	ctx := context.Background()
	event := NewEvent(EventStart, "test")
	composite.OnProgress(ctx, event)

	assert.Equal(t, []int{1, 2, 3}, order, "handlers should be called in order")
}

func TestCompositeHandler_ImplementsHandler(t *testing.T) {
	var _ Handler = (*CompositeHandler)(nil)
}

func TestCompositeHandler_NestedComposite(t *testing.T) {
	var called1, called2, called3 bool
	h1 := HandlerFunc(func(ctx context.Context, event Event) { called1 = true })
	h2 := HandlerFunc(func(ctx context.Context, event Event) { called2 = true })
	h3 := HandlerFunc(func(ctx context.Context, event Event) { called3 = true })

	// Create nested composite
	inner := NewCompositeHandler(h1, h2)
	outer := NewCompositeHandler(inner, h3)

	ctx := context.Background()
	event := NewEvent(EventStart, "test")
	outer.OnProgress(ctx, event)

	assert.True(t, called1, "nested handler 1 should be called")
	assert.True(t, called2, "nested handler 2 should be called")
	assert.True(t, called3, "outer handler 3 should be called")
}

func TestCompositeHandler_WithMockHandler(t *testing.T) {
	mock := &MockHandler{}
	composite := NewCompositeHandler(mock)

	ctx := context.Background()
	event := NewEvent(EventComplete, "done")
	composite.OnProgress(ctx, event)

	require.Len(t, mock.Events, 1)
	assert.Equal(t, EventComplete, mock.Events[0].Type)
	assert.Equal(t, "done", mock.Events[0].Message)
}
