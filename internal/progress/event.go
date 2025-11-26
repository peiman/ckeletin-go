// Package progress provides a flexible progress reporting system with handler-based
// architecture for composable output handling (logging, console, Bubble Tea UI).
//
// The package follows ADR-012 (Structured Output and Shadow Logging) by ensuring
// all progress events are logged to the audit stream while providing visual feedback
// through the status stream.
package progress

import (
	"time"
)

// EventType represents the type of progress event.
type EventType int

const (
	// EventStart indicates the beginning of an operation.
	EventStart EventType = iota
	// EventProgress indicates progress update during an operation.
	EventProgress
	// EventComplete indicates successful completion.
	EventComplete
	// EventError indicates an error occurred.
	EventError
	// EventWarning indicates a non-fatal warning.
	EventWarning
)

// String returns the string representation of EventType.
func (e EventType) String() string {
	switch e {
	case EventStart:
		return "start"
	case EventProgress:
		return "progress"
	case EventComplete:
		return "complete"
	case EventError:
		return "error"
	case EventWarning:
		return "warning"
	default:
		return "unknown"
	}
}

// Event represents a progress event that handlers can observe.
// Events are immutable once created and should be passed by value.
type Event struct {
	// Type indicates what kind of event this is.
	Type EventType

	// Phase identifies the current phase of a multi-phase operation.
	// Examples: "downloading", "validating", "installing"
	Phase string

	// Task identifies the specific task within a phase.
	// Examples: "file1.zip", "package.json"
	Task string

	// Current progress value (0 to Total).
	Current int64

	// Total expected value (for percentage calculation).
	// If Total is 0, progress is indeterminate.
	Total int64

	// Message is a human-readable description of the current state.
	Message string

	// Error contains the error if Type is EventError.
	Error error

	// Timestamp when this event was created.
	Timestamp time.Time

	// Metadata contains arbitrary key-value pairs for extensibility.
	// Examples: {"bytes_per_second": 1024, "eta_seconds": 30}
	Metadata map[string]interface{}
}

// NewEvent creates a new Event with the given type and message.
// Timestamp is set to the current time.
func NewEvent(eventType EventType, message string) Event {
	return Event{
		Type:      eventType,
		Message:   message,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}
}

// WithPhase returns a new Event with the phase set.
func (e Event) WithPhase(phase string) Event {
	e.Phase = phase
	return e
}

// WithTask returns a new Event with the task set.
func (e Event) WithTask(task string) Event {
	e.Task = task
	return e
}

// WithProgress returns a new Event with current/total progress set.
func (e Event) WithProgress(current, total int64) Event {
	e.Current = current
	e.Total = total
	return e
}

// WithError returns a new Event with the error set.
func (e Event) WithError(err error) Event {
	e.Error = err
	return e
}

// WithMeta returns a new Event with an additional metadata key-value pair.
func (e Event) WithMeta(key string, value interface{}) Event {
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	// Create a copy of metadata to maintain immutability
	newMeta := make(map[string]interface{}, len(e.Metadata)+1)
	for k, v := range e.Metadata {
		newMeta[k] = v
	}
	newMeta[key] = value
	e.Metadata = newMeta
	return e
}

// Percentage returns the progress as a percentage (0-100).
// Returns -1 if progress is indeterminate (Total == 0).
func (e Event) Percentage() float64 {
	if e.Total == 0 {
		return -1
	}
	return float64(e.Current) / float64(e.Total) * 100
}

// IsIndeterminate returns true if progress cannot be calculated.
func (e Event) IsIndeterminate() bool {
	return e.Total == 0
}
