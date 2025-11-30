package progress

import (
	"context"
	"sync"
)

// CompositeHandler combines multiple handlers into one.
// Events are dispatched to all handlers sequentially.
// This enables patterns like: log + render + metrics simultaneously.
//
// Example:
//
//	handler := NewCompositeHandler(
//	    NewLogHandler(),      // Shadow logging (always)
//	    NewConsoleHandler(os.Stderr), // Simple output
//	)
type CompositeHandler struct {
	handlers []Handler
	mu       sync.RWMutex
}

// NewCompositeHandler creates a new CompositeHandler with the given handlers.
func NewCompositeHandler(handlers ...Handler) *CompositeHandler {
	return &CompositeHandler{
		handlers: handlers,
	}
}

// OnProgress implements Handler by dispatching to all handlers.
func (h *CompositeHandler) OnProgress(ctx context.Context, event Event) {
	h.mu.RLock()
	handlers := h.handlers
	h.mu.RUnlock()

	// Dispatch to all handlers sequentially
	// This ensures predictable ordering and avoids race conditions
	for _, handler := range handlers {
		if handler != nil {
			handler.OnProgress(ctx, event)
		}
	}
}

// Add adds a handler to the composite (thread-safe).
func (h *CompositeHandler) Add(handler Handler) {
	if handler == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers = append(h.handlers, handler)
}

// RemoveAt removes a handler at the given index (thread-safe).
// Returns true if the handler was removed, false if index out of bounds.
func (h *CompositeHandler) RemoveAt(index int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if index < 0 || index >= len(h.handlers) {
		return false
	}
	h.handlers = append(h.handlers[:index], h.handlers[index+1:]...)
	return true
}

// Len returns the number of handlers (thread-safe).
func (h *CompositeHandler) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.handlers)
}

// Handlers returns a copy of the handlers slice (thread-safe).
func (h *CompositeHandler) Handlers() []Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Handler, len(h.handlers))
	copy(result, h.handlers)
	return result
}
