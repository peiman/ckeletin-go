package progress

import (
	"context"
	"sync"
)

// MockHandler records all events for testing.
// This follows the same pattern as internal/ui/mock.go (ADR-003 compliant).
type MockHandler struct {
	Events      []Event
	mu          sync.Mutex
	ReturnError error
}

// NewMockHandler creates a new MockHandler.
func NewMockHandler() *MockHandler {
	return &MockHandler{
		Events: make([]Event, 0),
	}
}

// OnProgress implements Handler by recording the event.
func (h *MockHandler) OnProgress(_ context.Context, event Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Events = append(h.Events, event)
}

// Reset clears recorded events.
func (h *MockHandler) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.Events = make([]Event, 0)
}

// LastEvent returns the most recent event.
// Returns empty Event and false if no events recorded.
func (h *MockHandler) LastEvent() (Event, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.Events) == 0 {
		return Event{}, false
	}
	return h.Events[len(h.Events)-1], true
}

// EventCount returns the number of recorded events.
func (h *MockHandler) EventCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.Events)
}

// EventsOfType returns all events of a specific type.
func (h *MockHandler) EventsOfType(t EventType) []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	var result []Event
	for _, e := range h.Events {
		if e.Type == t {
			result = append(result, e)
		}
	}
	return result
}

// GetEvents returns a copy of all events (thread-safe).
func (h *MockHandler) GetEvents() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	result := make([]Event, len(h.Events))
	copy(result, h.Events)
	return result
}

// HasEventWithMessage checks if any event contains the given message.
func (h *MockHandler) HasEventWithMessage(message string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, e := range h.Events {
		if e.Message == message {
			return true
		}
	}
	return false
}
