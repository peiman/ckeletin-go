package progress

import "context"

// Handler is the core interface for receiving progress events.
// Implementations can render to terminal, log to file, emit metrics, etc.
//
// Design follows the Observer pattern: handlers subscribe to events emitted
// by the Reporter, decoupling event production from consumption.
//
// Implementations should be safe for concurrent calls from multiple goroutines.
type Handler interface {
	// OnProgress is called for each progress event.
	// Context can be used for cancellation or deadline propagation.
	OnProgress(ctx context.Context, event Event)
}

// HandlerFunc is an adapter to allow ordinary functions as handlers.
// This follows the same pattern as http.HandlerFunc.
type HandlerFunc func(ctx context.Context, event Event)

// OnProgress implements Handler interface by calling the function.
func (f HandlerFunc) OnProgress(ctx context.Context, event Event) {
	f(ctx, event)
}
