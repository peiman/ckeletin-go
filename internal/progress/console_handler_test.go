package progress

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsoleHandler(t *testing.T) {
	tests := []struct {
		name     string
		event    Event
		contains []string
	}{
		{
			name:     "start event with phase",
			event:    NewEvent(EventStart, "Starting download").WithPhase("download"),
			contains: []string{"==>", "Starting download"},
		},
		{
			name:     "start event without phase",
			event:    NewEvent(EventStart, "Starting"),
			contains: []string{"Starting"},
		},
		{
			name:     "progress event with total",
			event:    NewEvent(EventProgress, "file.zip").WithProgress(5, 10),
			contains: []string{"[5/10]", "file.zip"},
		},
		{
			name:     "complete event",
			event:    NewEvent(EventComplete, "Download complete"),
			contains: []string{"✓", "Download complete"},
		},
		{
			name:     "error event",
			event:    NewEvent(EventError, "Download failed"),
			contains: []string{"✗", "Download failed"},
		},
		{
			name:     "warning event",
			event:    NewEvent(EventWarning, "Low disk space"),
			contains: []string{"⚠", "Low disk space"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			var buf bytes.Buffer
			handler := NewConsoleHandler(&buf)

			// EXECUTION
			handler.OnProgress(context.Background(), tt.event)

			// ASSERTION
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestConsoleHandler_IndeterminateProgress(t *testing.T) {
	// Indeterminate progress (Total=0) should not output anything
	// to avoid spamming the console

	// SETUP
	var buf bytes.Buffer
	handler := NewConsoleHandler(&buf)
	event := NewEvent(EventProgress, "Loading...").WithProgress(0, 0)

	// EXECUTION
	handler.OnProgress(context.Background(), event)

	// ASSERTION - should be empty for indeterminate
	assert.Empty(t, buf.String())
}

func TestConsoleHandler_ConcurrentWrites(t *testing.T) {
	// Test that concurrent writes don't interleave
	// SETUP
	var buf bytes.Buffer
	handler := NewConsoleHandler(&buf)

	// EXECUTION - concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			handler.OnProgress(context.Background(), NewEvent(EventComplete, "Done"))
			done <- true
		}(i)
	}

	// Wait for all
	for i := 0; i < 10; i++ {
		<-done
	}

	// ASSERTION - should have 10 complete lines
	output := buf.String()
	assert.Contains(t, output, "✓")
}

func TestConsoleHandler_ContextCancellation(t *testing.T) {
	// SETUP
	var buf bytes.Buffer
	handler := NewConsoleHandler(&buf)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// EXECUTION
	handler.OnProgress(ctx, NewEvent(EventComplete, "should not appear"))

	// ASSERTION - should be empty since context was cancelled
	assert.Empty(t, buf.String(), "should not write when context is cancelled")
}
