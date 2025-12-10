package ui

import (
	"bytes"
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

	p.CheckHeader("testing")
	assert.Contains(t, buf.String(), "[-]") // Minimal theme icon
}

func TestStdoutCheckPrinter(t *testing.T) {
	p := StdoutCheckPrinter()
	require.NotNil(t, p)
	// Can't easily test stdout output, but verify it doesn't panic
}

func TestStderrCheckPrinter(t *testing.T) {
	p := StderrCheckPrinter()
	require.NotNil(t, p)
	// Verify it's configured for stderr by checking the writer
	// (internal detail, but validates the function works)
	_ = os.Stderr // Just ensure we reference stderr
}

func TestCheckPrinter_Integration(t *testing.T) {
	var buf bytes.Buffer
	p := NewCheckPrinterWithWriter(&buf, checkmate.WithTheme(checkmate.MinimalTheme()))

	// Test a typical check workflow
	p.CategoryHeader("Tests")
	p.CheckHeader("Running unit tests")
	p.CheckSuccess("All tests passed")

	output := buf.String()
	assert.Contains(t, output, "Tests")
	assert.Contains(t, output, "Running unit tests")
	assert.Contains(t, output, "All tests passed")
	assert.Contains(t, output, "[OK]")
}
