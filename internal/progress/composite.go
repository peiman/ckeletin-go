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

// Remove removes a handler from the composite (thread-safe).
// Uses pointer comparison for identity.
func (h *CompositeHandler) Remove(handler Handler) {
	if handler == nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for i, hdlr := range h.handlers {
		if hdlr == handler {
			h.handlers = append(h.handlers[:i], h.handlers[i+1:]...)
			return
		}
	}
}

// Handlers returns a copy of the handlers slice (thread-safe).
func (h *CompositeHandler) Handlers() []Handler {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]Handler, len(h.handlers))
	copy(result, h.handlers)
	return result
}
