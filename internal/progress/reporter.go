package progress

import (
	"context"
	"io"
	"os"
	"sync"
	"time"
)

// Reporter provides a convenient API for emitting progress events.
// It manages handlers and provides helper methods for common patterns.
//
// Example usage:
//
//	reporter := progress.NewReporter(
//	    progress.WithOutput(cmd.ErrOrStderr()),
//	    progress.WithInteractive(useUI),
//	)
//	reporter.Start(ctx, "Processing files")
//	for i, file := range files {
//	    reporter.Progress(ctx, int64(i+1), int64(len(files)), file.Name)
//	}
//	reporter.Complete(ctx, "Done")
type Reporter struct {
	handler Handler
	phase   string
	mu      sync.RWMutex
}

// Option configures a Reporter.
type Option func(*Reporter)

// WithHandler sets the handler for the reporter.
func WithHandler(h Handler) Option {
	return func(r *Reporter) {
		r.handler = h
	}
}

// WithOutput configures the reporter to output to the given writer.
// If interactive is true, uses TeaHandler; otherwise uses ConsoleHandler.
// LogHandler is always included for shadow logging (ADR-012 compliance).
func WithOutput(w io.Writer, interactive bool) Option {
	return func(r *Reporter) {
		var displayHandler Handler
		if interactive && IsInteractive(w) {
			displayHandler = NewTeaHandler(w)
		} else {
			displayHandler = NewConsoleHandler(w)
		}
		r.handler = NewCompositeHandler(
			NewLogHandler(),
			displayHandler,
		)
	}
}

// WithInteractive is a convenience option that uses stderr with the given
// interactivity setting.
func WithInteractive(interactive bool) Option {
	return func(r *Reporter) {
		WithOutput(os.Stderr, interactive)(r)
	}
}

// NewReporter creates a new Reporter with the given options.
// If no handler is provided, a default non-interactive handler is used.
func NewReporter(opts ...Option) *Reporter {
	r := &Reporter{}
	for _, opt := range opts {
		opt(r)
	}
	// Default handler if none provided
	if r.handler == nil {
		r.handler = NewCompositeHandler(
			NewLogHandler(),
			NewConsoleHandler(os.Stderr),
		)
	}
	return r
}

// SetHandler changes the handler (thread-safe).
func (r *Reporter) SetHandler(h Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handler = h
}

// emit dispatches an event to the handler.
func (r *Reporter) emit(ctx context.Context, event Event) {
	r.mu.RLock()
	handler := r.handler
	phase := r.phase
	r.mu.RUnlock()

	// Apply defaults if not set on event
	if event.Phase == "" && phase != "" {
		event.Phase = phase
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	if handler != nil {
		handler.OnProgress(ctx, event)
	}
}

// SetPhase sets the current phase for subsequent events.
func (r *Reporter) SetPhase(phase string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.phase = phase
}

// Phase returns the current phase.
func (r *Reporter) Phase() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.phase
}

// Start emits a start event.
func (r *Reporter) Start(ctx context.Context, message string) {
	r.emit(ctx, NewEvent(EventStart, message))
}

// Progress emits a progress event with current/total values.
func (r *Reporter) Progress(ctx context.Context, current, total int64, message string) {
	r.emit(ctx, NewEvent(EventProgress, message).WithProgress(current, total))
}

// Complete emits a completion event.
func (r *Reporter) Complete(ctx context.Context, message string) {
	r.emit(ctx, NewEvent(EventComplete, message))
}

// Error emits an error event.
func (r *Reporter) Error(ctx context.Context, err error, message string) {
	r.emit(ctx, NewEvent(EventError, message).WithError(err))
}

// Warning emits a warning event.
func (r *Reporter) Warning(ctx context.Context, message string) {
	r.emit(ctx, NewEvent(EventWarning, message))
}

// Emit allows emitting custom events.
func (r *Reporter) Emit(ctx context.Context, event Event) {
	r.emit(ctx, event)
}

// ProgressFunc returns a function suitable for passing to operations
// that report progress via callback.
//
// Example:
//
//	progressFn := reporter.ProgressFunc(ctx, "Downloading")
//	download(url, progressFn) // calls progressFn(bytesRead, totalBytes)
func (r *Reporter) ProgressFunc(ctx context.Context, message string) func(current, total int64) {
	return func(current, total int64) {
		r.Progress(ctx, current, total, message)
	}
}
