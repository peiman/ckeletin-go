package progress

import "context"

// NopHandler is a handler that does nothing.
// Useful for testing or when progress reporting should be disabled.
//
// This implements the Null Object pattern to avoid nil checks.
type NopHandler struct{}

// NewNopHandler creates a new NopHandler.
func NewNopHandler() *NopHandler {
	return &NopHandler{}
}

// OnProgress implements Handler by doing nothing.
func (h *NopHandler) OnProgress(_ context.Context, _ Event) {
	// Intentionally empty - used for disabling progress output
}
