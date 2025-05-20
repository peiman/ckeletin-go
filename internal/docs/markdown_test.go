// internal/docs/markdown_test.go

package docs

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
)

// TestGenerateMarkdownDocs tests the basic structure of generated markdown documentation
func TestGenerateMarkdownDocs(t *testing.T) {
	// SETUP PHASE
	// Create output buffer
	var buf bytes.Buffer

	// Create test app info
	appInfo := AppInfo{
		BinaryName: "testapp",
		EnvPrefix:  "TESTAPP",
	}
	appInfo.ConfigPaths.DefaultPath = "/home/user/.testapp.yaml"
	appInfo.ConfigPaths.DefaultFullName = ".testapp.yaml"

	// Create generator
	cfg := NewConfig(&buf, WithOutputFormat(FormatMarkdown))
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("GenerateMarkdownDocs failed: %v", err)
	}

	output := buf.String()

	// Check header
	if !strings.Contains(output, "# testapp Configuration") {
		t.Errorf("Missing header in output")
	}

	// Check sections
	expectedSections := []string{
		"## Configuration Sources",
		"## Configuration Options",
		"## Example Configuration",
		"### YAML Configuration File",
		"### Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Missing section: %s", section)
		}
	}

	// Check configuration sources
	if !strings.Contains(output, "Environment variables (with prefix `TESTAPP_`)") {
		t.Errorf("Missing environment variable prefix")
	}

	if !strings.Contains(output, "Configuration file (/home/user/.testapp.yaml)") {
		t.Errorf("Missing config file path")
	}

	// Check table headers and basic structure
	tableHeaders := "| Key | Type | Default | Environment Variable | Description |"
	if !strings.Contains(output, tableHeaders) {
		t.Errorf("Missing table headers")
	}

	// Check YAML section existence
	if !strings.Contains(output, "```yaml") {
		t.Errorf("Missing YAML code block")
	}

	// Check environment variables section
	if !strings.Contains(output, "```bash") {
		t.Errorf("Missing bash code block for environment variables")
	}
}

// TestGenerateMarkdownDocs_YAMLError tests how markdown generation handles YAML errors
func TestGenerateMarkdownDocs_YAMLError(t *testing.T) {
	// SETUP PHASE
	// Create a test app info
	appInfo := AppInfo{
		BinaryName: "testapp",
		EnvPrefix:  "TESTAPP",
	}

	// Create a buffer
	var buf bytes.Buffer

	// Create a generator with a custom generator function
	expectedErr := errors.New("yaml generation error")

	// Store the original function
	origGenerateYAMLContent := generateYAMLContentFunc

	// Replace with a mock implementation that returns an error
	generateYAMLContentFunc = func(w io.Writer, registry []config.ConfigOption) error {
		return expectedErr
	}

	// Restore the original function after the test
	defer func() {
		generateYAMLContentFunc = origGenerateYAMLContent
	}()

	generator := NewGenerator(NewConfig(&buf))

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	if err == nil {
		t.Errorf("Expected error for YAML generation, got nil")
	}

	if !strings.Contains(err.Error(), expectedErr.Error()) {
		t.Errorf("Expected error to contain %q, got %q", expectedErr, err.Error())
	}
}

// TestGenerateMarkdownDocs_EmptyRegistry tests how the markdown generator handles an empty registry
func TestGenerateMarkdownDocs_EmptyRegistry(t *testing.T) {
	// SETUP PHASE
	// Create test app info
	appInfo := AppInfo{
		BinaryName: "testapp",
		EnvPrefix:  "TESTAPP",
	}
	appInfo.ConfigPaths.DefaultPath = "/home/user/.testapp.yaml"
	appInfo.ConfigPaths.DefaultFullName = ".testapp.yaml"

	// Create buffer
	var buf bytes.Buffer

	// Create a generator config with a custom registry function that returns empty registry
	cfg := NewConfig(&buf, WithOutputFormat(FormatMarkdown), WithRegistryFunc(func() []config.ConfigOption {
		return []config.ConfigOption{}
	}))
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("GenerateMarkdownDocs failed with empty registry: %v", err)
	}

	output := buf.String()

	// Check the document still has structure
	expectedSections := []string{
		"# testapp Configuration",
		"## Configuration Sources",
		"## Configuration Options",
		"## Example Configuration",
		"### YAML Configuration File",
		"### Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Missing section with empty registry: %s", section)
		}
	}

	// Check table headers still exist
	tableHeaders := "| Key | Type | Default | Environment Variable | Description |"
	if !strings.Contains(output, tableHeaders) {
		t.Errorf("Missing table headers with empty registry")
	}

	// Check that the blocks are properly closed
	if !strings.Contains(output, "```yaml") || !strings.Contains(output, "```bash") {
		t.Errorf("Missing code blocks with empty registry")
	}
}
