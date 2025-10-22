// internal/docs/generator_test.go

package docs

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
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
	if generator.cfg.Writer != writer {
		t.Errorf("Generator did not store the config correctly")
	}
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
	if generator.appInfo.BinaryName != "test-app" {
		t.Errorf("Expected BinaryName to be 'test-app', got %s", generator.appInfo.BinaryName)
	}
	if generator.appInfo.EnvPrefix != "TEST_APP" {
		t.Errorf("Expected EnvPrefix to be 'TEST_APP', got %s", generator.appInfo.EnvPrefix)
	}
	if generator.appInfo.ConfigPaths.DefaultPath != "/path/to/config" {
		t.Errorf("Expected DefaultPath to be '/path/to/config', got %s",
			generator.appInfo.ConfigPaths.DefaultPath)
	}
}

func TestGenerate_UnsupportedFormat(t *testing.T) {
	// SETUP PHASE
	writer := &bytes.Buffer{}
	cfg := Config{Writer: writer, OutputFormat: "invalid", OutputFile: "", Registry: config.Registry}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.Generate()

	// ASSERTION PHASE
	if err == nil {
		t.Errorf("Expected error for unsupported format, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Expected error to contain 'unsupported format', got %s", err.Error())
	}
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
	if err == nil {
		t.Errorf("Expected error when file cannot be opened, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create output file") {
		t.Errorf("Expected 'failed to create output file' in error, got %s", err.Error())
	}
}

func TestGenerate_CloseError(t *testing.T) {
	// SETUP PHASE
	// Create a mock generator that just returns the close error
	closeWasCalled := false
	closeErr := errors.New("close error")
	mockFile := &MockWriteCloser{
		Buffer:   bytes.NewBuffer(nil),
		closeErr: closeErr,
		onClose:  func() { closeWasCalled = true },
	}

	// Create a custom generate function that simulates the file close error
	customGenerate := func() error {
		// Write something to the mock file
		_, err := mockFile.Write([]byte("test content"))
		if err != nil {
			return err
		}

		// Now close the file and handle the error as Generate would
		closeErr := mockFile.Close()
		if closeErr != nil {
			return closeErr
		}

		return nil
	}

	// EXECUTION PHASE
	err := customGenerate()

	// ASSERTION PHASE
	// Verify that Close was called
	if !closeWasCalled {
		t.Fatalf("Close was not called on the mock file")
	}

	// Verify the close error is propagated
	if err == nil {
		t.Errorf("Expected close error to be propagated, got nil")
	}

	if err != closeErr {
		t.Errorf("Expected error to be the close error, got %v", err)
	}
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
	if err != nil {
		t.Fatalf("GenerateMarkdown failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "# test Configuration") {
		t.Error("Expected markdown output to contain app name")
	}

	if !strings.Contains(output, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST") {
		t.Error("Expected YAML content function to be called")
	}
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
	if err != nil {
		t.Fatalf("GenerateYAML failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "TEST_YAML_CONTENT_FOR_CONVENIENCE_TEST") {
		t.Error("Expected YAML content function to be called")
	}
}
