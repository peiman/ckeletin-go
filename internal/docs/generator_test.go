// internal/docs/generator_test.go

package docs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockWriteCloser is a simple implementation of io.WriteCloser for testing
type MockWriteCloser struct {
	*bytes.Buffer
	closeErr error
	onClose  func() // Function to call on Close() for verifying it was called
}

func (m *MockWriteCloser) Close() error {
	if m.onClose != nil {
		m.onClose()
	}
	return m.closeErr
}

func NewMockWriteCloser(content string, closeErr error) *MockWriteCloser {
	return &MockWriteCloser{
		Buffer:   bytes.NewBufferString(content),
		closeErr: closeErr,
	}
}

func TestNewGenerator(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}
	cfg := Config{Writer: writer, OutputFormat: FormatMarkdown, OutputFile: "", Registry: config.Registry}

	// EXECUTION PHASE
	generator := NewGenerator(cfg)

	// ASSERTION PHASE
	assert.Equal(t, writer, generator.cfg.Writer, "Generator did not store the config correctly")
}

func TestSetAppInfo(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}
	cfg := Config{Writer: writer, OutputFormat: FormatMarkdown, OutputFile: "", Registry: config.Registry}
	generator := NewGenerator(cfg)

	appInfo := AppInfo{
		BinaryName: "test-app",
		EnvPrefix:  "TEST_APP",
	}
	appInfo.ConfigPaths.DefaultPath = "/path/to/config"
	appInfo.ConfigPaths.DefaultFullName = "config.yaml"

	// EXECUTION PHASE
	generator.SetAppInfo(appInfo)

	// ASSERTION PHASE
	assert.Equal(t, "test-app", generator.appInfo.BinaryName)
	assert.Equal(t, "TEST_APP", generator.appInfo.EnvPrefix)
	assert.Equal(t, "/path/to/config", generator.appInfo.ConfigPaths.DefaultPath)
}

func TestGenerate_UnsupportedFormat(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}
	cfg := Config{Writer: writer, OutputFormat: "invalid", OutputFile: "", Registry: config.Registry}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.Generate()

	// ASSERTION PHASE
	require.Error(t, err, "Expected error for unsupported format, got nil")
	assert.True(t, strings.Contains(err.Error(), "unsupported format"),
		"Expected error to contain 'unsupported format', got %s", err.Error())
}

func TestGenerate_FileError(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}
	cfg := Config{Writer: writer, OutputFormat: FormatMarkdown, OutputFile: "test.md", Registry: config.Registry}
	generator := NewGenerator(cfg)

	// Mock file opening to simulate error
	origOpenOutputFile := openOutputFile
	defer func() { openOutputFile = origOpenOutputFile }()

	// Simulate file opening error
	openErr := errors.New("failed to open file")
	openOutputFile = func(path string) (io.WriteCloser, error) {
		return nil, openErr
	}

	// EXECUTION PHASE
	err := generator.Generate()

	// ASSERTION PHASE
	require.Error(t, err, "Expected error when file cannot be opened, got nil")
	assert.True(t, strings.Contains(err.Error(), "failed to create output file"),
		"Expected 'failed to create output file' in error, got %s", err.Error())
}

func TestGenerate_CloseError(t *testing.T) {
	// REGRESSION TEST: Verify that close errors are properly propagated
	// Bug: The deferred closeErr assignment happens after the function returns,
	// so checking closeErr in the function body always sees nil.
	//
	// This test ensures that file.Close() errors are properly returned to the caller.

	// SETUP PHASE
	closeWasCalled := false
	closeErr := errors.New("disk full - close failed")

	// Mock openOutputFile to return our mock file that will fail on close
	origOpenOutputFile := openOutputFile
	defer func() { openOutputFile = origOpenOutputFile }()

	openOutputFile = func(path string) (io.WriteCloser, error) {
		mockFile := &MockWriteCloser{
			Buffer:   bytes.NewBuffer(nil),
			closeErr: closeErr,
			onClose:  func() { closeWasCalled = true },
		}
		return mockFile, nil
	}

	// Create generator with output file (triggers file creation)
	writer := &bytes.Buffer{}
	cfg := Config{
		Writer:       writer,
		OutputFormat: FormatMarkdown,
		OutputFile:   "test-output.md", // Triggers file creation path
		Registry:     config.Registry,
	}
	generator := NewGenerator(cfg)
	generator.SetAppInfo(AppInfo{BinaryName: "test"})

	// EXECUTION PHASE
	err := generator.Generate()

	// ASSERTION PHASE
	// Verify that Close was called
	require.True(t, closeWasCalled, "Close was not called on the file - defer didn't run")

	// CRITICAL: Verify the close error is propagated to caller
	// This will FAIL with the current buggy implementation because closeErr
	// is checked before the deferred function assigns it
	require.Error(t, err,
		"Expected close error to be propagated, got nil\n"+
			"This indicates the close-error aggregation bug:\n"+
			"The defer sets closeErr AFTER the function returns, so checks for\n"+
			"'if closeErr != nil' in the function body always see nil",
	)

	assert.True(t, strings.Contains(err.Error(), "close"),
		"Expected error message to mention 'close', got: %v", err)
}

func TestGenerate_BothGenerationAndCloseErrors(t *testing.T) {
	// Test that both generation error AND close error are properly handled
	// This tests the error aggregation path where both operations fail

	// SETUP PHASE
	closeErr := errors.New("disk full - close failed")

	// Mock openOutputFile to return our mock file that will fail on close
	origOpenOutputFile := openOutputFile
	defer func() { openOutputFile = origOpenOutputFile }()

	openOutputFile = func(path string) (io.WriteCloser, error) {
		mockFile := &MockWriteCloser{
			Buffer:   bytes.NewBuffer(nil),
			closeErr: closeErr,
		}
		return mockFile, nil
	}

	// Create generator with unsupported format to trigger generation error
	writer := &bytes.Buffer{}
	cfg := Config{
		Writer:       writer,
		OutputFormat: "invalid-format", // This will cause generation error
		OutputFile:   "test-output.md",
		Registry:     config.Registry,
	}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.Generate()

	// ASSERTION PHASE
	require.Error(t, err, "Expected both generation and close errors to be returned")

	// Verify error message contains both errors
	// With errors.Join, both errors are included in the multi-error
	errMsg := err.Error()
	assert.True(t, strings.Contains(errMsg, "generation failed"),
		"Expected error to mention 'generation failed', got: %v", errMsg)
	assert.True(t, strings.Contains(errMsg, "file close failed"),
		"Expected error to mention file close failure, got: %v", errMsg)
}

// TestGenerateMarkdownConvenience tests the convenience function for generating markdown
func TestGenerateMarkdownConvenience(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	appInfo := AppInfo{BinaryName: "test"}

	// Store original functions
	origYAMLFunc := generateYAMLContentFunc

	// Replace with a test version that adds a recognizable marker
	generateYAMLContentFunc = func(w io.Writer, registry []config.ConfigOption) error {
		_, err := fmt.Fprintln(w, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST")
		return err
	}

	// Restore after test
	defer func() {
		generateYAMLContentFunc = origYAMLFunc
	}()

	// EXECUTION PHASE
	err := GenerateMarkdown(&buf, appInfo)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateMarkdown failed")

	output := buf.String()
	assert.True(t, strings.Contains(output, "# test Configuration"),
		"Expected markdown output to contain app name")

	assert.True(t, strings.Contains(output, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST"),
		"Expected YAML content function to be called")
}

// TestGenerateYAMLConvenience tests the convenience function for generating YAML
func TestGenerateYAMLConvenience(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer

	// Store original functions
	origYAMLFunc := generateYAMLContentFunc

	// Replace with a test version that adds a recognizable marker
	generateYAMLContentFunc = func(w io.Writer, registry []config.ConfigOption) error {
		_, err := fmt.Fprintln(w, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST")
		return err
	}

	// Restore after test
	defer func() {
		generateYAMLContentFunc = origYAMLFunc
	}()

	// EXECUTION PHASE
	err := GenerateYAML(&buf)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateYAML failed")

	output := buf.String()
	assert.True(t, strings.Contains(output, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST"),
		"Expected YAML content function to be called")
}
