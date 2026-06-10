package ui

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/peiman/ckeletin-go/pkg/checkmate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCheckPrinter(t *testing.T) {
	p := NewCheckPrinter()
	require.NotNil(t, p)
}

func TestNewCheckPrinter_WithOptions(t *testing.T) {
	var buf bytes.Buffer
	p := NewCheckPrinter(
		checkmate.WithWriter(&buf),
		checkmate.WithTheme(checkmate.MinimalTheme()),
	)
	require.NotNil(t, p)

	p.CheckSuccess("test")
	assert.Contains(t, buf.String(), "[OK]")
	assert.Contains(t, buf.String(), "test")
}

func TestNewCheckPrinterWithWriter(t *testing.T) {
	var buf bytes.Buffer
	p := NewCheckPrinterWithWriter(&buf)
	require.NotNil(t, p)

	p.CheckSuccess("works")
	assert.Contains(t, buf.String(), "works")
}

func TestNewCheckPrinterWithWriter_AdditionalOptions(t *testing.T) {
	var buf bytes.Buffer
	p := NewCheckPrinterWithWriter(&buf, checkmate.WithTheme(checkmate.MinimalTheme()))
	require.NotNil(t, p)

	// CheckHeader skips output in non-TTY mode
	p.CheckHeader("testing")
	assert.Empty(t, buf.String(), "CheckHeader should skip in non-TTY")

	// CheckSuccess should work
	p.CheckSuccess("testing")
	assert.Contains(t, buf.String(), "[OK]") // Minimal theme success icon
}

func TestStdoutCheckPrinter(t *testing.T) {
	// SETUP: swap os.Stdout for a pipe so the factory captures it
	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = origStdout })

	// EXECUTION
	p := StdoutCheckPrinter()
	require.NotNil(t, p)
	p.CheckSuccess("stdout-bound")
	require.NoError(t, w.Close())

	// ASSERTION: output landed on the swapped stdout
	captured, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Contains(t, string(captured), "stdout-bound")
}

func TestStderrCheckPrinter(t *testing.T) {
	// SETUP: swap os.Stderr for a pipe so the factory captures it
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	// EXECUTION
	p := StderrCheckPrinter()
	require.NotNil(t, p)
	p.CheckSuccess("stderr-bound")
	require.NoError(t, w.Close())

	// ASSERTION: output landed on the swapped stderr
	captured, err := io.ReadAll(r)
	require.NoError(t, err)
	assert.Contains(t, string(captured), "stderr-bound")
}

func TestCheckPrinter_Integration(t *testing.T) {
	var buf bytes.Buffer
	p := NewCheckPrinterWithWriter(&buf, checkmate.WithTheme(checkmate.MinimalTheme()))

	// Test a typical check workflow
	p.CategoryHeader("Tests")
	// CheckHeader skips in non-TTY mode
	p.CheckHeader("Running unit tests")
	p.CheckSuccess("All tests passed")

	output := buf.String()
	assert.Contains(t, output, "Tests")
	// CheckHeader skips output in non-TTY, only result shows
	assert.NotContains(t, output, "Running unit tests")
	assert.Contains(t, output, "All tests passed")
	assert.Contains(t, output, "[OK]")
}
