package progress

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsInteractive(t *testing.T) {
	t.Run("bytes.Buffer is not interactive", func(t *testing.T) {
		var buf bytes.Buffer
		assert.False(t, IsInteractive(&buf))
	})

	t.Run("nil is not interactive", func(t *testing.T) {
		assert.False(t, IsInteractive(nil))
	})

	// Note: Testing actual TTY detection is difficult in unit tests
	// because it depends on the test environment. In CI, stdout/stderr
	// are typically not TTYs.
}

func TestIsStderrInteractive(t *testing.T) {
	// This test documents the behavior rather than asserting a specific value
	// because the result depends on the test environment
	result := IsStderrInteractive()
	t.Logf("IsStderrInteractive: %v", result)
	// In CI environments, this is typically false
	// In interactive terminals, this is typically true
}

func TestIsStdoutInteractive(t *testing.T) {
	// This test documents the behavior rather than asserting a specific value
	result := IsStdoutInteractive()
	t.Logf("IsStdoutInteractive: %v", result)
}

func TestIsInteractive_WithFile(t *testing.T) {
	// Create a temp file to test file descriptor behavior
	f, err := os.CreateTemp("", "test")
	if err != nil {
		t.Skip("Could not create temp file")
	}
	defer os.Remove(f.Name())
	defer f.Close()

	// A regular file is not a terminal
	assert.False(t, IsInteractive(f))
}
