package checkmate

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsTerminal_Buffer(t *testing.T) {
	var buf bytes.Buffer
	assert.False(t, IsTerminal(&buf), "Buffer should not be a terminal")
}

func TestIsTerminal_File(t *testing.T) {
	// Create a temp file - files are not terminals
	f, err := os.CreateTemp("", "test")
	if err != nil {
		t.Skip("Could not create temp file")
	}
	defer os.Remove(f.Name())
	defer f.Close()

	assert.False(t, IsTerminal(f), "Regular file should not be a terminal")
}

func TestIsStdoutTerminal(t *testing.T) {
	// This test's result depends on how the test is run
	// In CI (piped output), it should be false
	// In a terminal, it might be true
	// We just verify it doesn't panic
	_ = IsStdoutTerminal()
}

func TestIsStderrTerminal(t *testing.T) {
	// This test's result depends on how the test is run
	// We just verify it doesn't panic
	_ = IsStderrTerminal()
}
