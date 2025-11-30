package progress

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// LogHandler writes progress events to the structured log (Audit Stream).
// This implements ADR-012's shadow logging pattern, ensuring all progress
// events are recorded for debugging and auditing purposes.
type LogHandler struct {
	// Logger to use (defaults to global log.Logger)
	Logger zerolog.Logger

	// Component name for log entries
	Component string
}

// LogHandlerOption configures a LogHandler.
type LogHandlerOption func(*LogHandler)

// WithLogger sets a custom zerolog.Logger.
func WithLogger(l zerolog.Logger) LogHandlerOption {
	return func(h *LogHandler) {
		h.Logger = l
	}
}

// WithComponent sets the component name for log entries.
func WithComponent(name string) LogHandlerOption {
	return func(h *LogHandler) {
		h.Component = name
	}
}

// NewLogHandler creates a new LogHandler with the given options.
func NewLogHandler(opts ...LogHandlerOption) *LogHandler {
	h := &LogHandler{
		Logger:    log.Logger,
		Component: "progress",
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

// OnProgress implements Handler by logging the event to the audit stream.
// Respects context cancellation.
func (h *LogHandler) OnProgress(ctx context.Context, event Event) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return
	default:
	}

	// Create base log event with appropriate level
	var logEvent *zerolog.Event
	switch event.Type {
	case EventError:
		logEvent = h.Logger.Error()
	case EventWarning:
		logEvent = h.Logger.Warn()
	case EventProgress:
		// Use Debug for frequent progress updates to avoid log spam
		logEvent = h.Logger.Debug()
	default:
		logEvent = h.Logger.Info()
	}

	// Add structured fields
	logEvent.
		Str("component", h.Component).
		Str("event_type", event.Type.String()).
		Time("event_time", event.Timestamp)

	// Add optional fields only if present
	if event.Phase != "" {
		logEvent.Str("phase", event.Phase)
	}
	if event.Task != "" {
		logEvent.Str("task", event.Task)
	}

	// Add progress info
	if event.Total > 0 {
		logEvent.
			Int64("current", event.Current).
			Int64("total", event.Total).
			Float64("percentage", event.Percentage())
	}

	// Add error if present
	if event.Error != nil {
		logEvent.Err(event.Error)
	}

	// Add metadata if present
	if len(event.Metadata) > 0 {
		logEvent.Interface("metadata", event.Metadata)
	}

	// Log the message
	logEvent.Msg(event.Message)
}
