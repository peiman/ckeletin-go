// internal/docs/markdown_test.go

package docs

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	cfg := Config{Writer: &buf, OutputFormat: FormatMarkdown, OutputFile: "", Registry: config.Registry}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateMarkdownDocs failed")

	output := buf.String()

	// Check header
	assert.True(t, strings.Contains(output, "# testapp Configuration"), "Missing header in output")

	// Check sections
	expectedSections := []string{
		"## Configuration Sources",
		"## Configuration Options",
		"## Example Configuration",
		"### YAML Configuration File",
		"### Environment Variables",
	}

	for _, section := range expectedSections {
		assert.True(t, strings.Contains(output, section), "Missing section: %s", section)
	}

	// Check configuration sources
	assert.True(t, strings.Contains(output, "Environment variables (with prefix `TESTAPP_`)"),
		"Missing environment variable prefix")

	// Path should be sanitized to use ~ instead of /home/user
	assert.True(t, strings.Contains(output, "Configuration file (~/.testapp.yaml)"),
		"Missing sanitized config file path")

	// Check table headers and basic structure
	tableHeaders := "| Key | Type | Default | Environment Variable | Description |"
	assert.True(t, strings.Contains(output, tableHeaders), "Missing table headers")

	// Check YAML section existence
	assert.True(t, strings.Contains(output, "```yaml"), "Missing YAML code block")

	// Check environment variables section
	assert.True(t, strings.Contains(output, "```bash"), "Missing bash code block for environment variables")
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

	generator := NewGenerator(Config{Writer: &buf, OutputFormat: FormatMarkdown, OutputFile: "", Registry: config.Registry})

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	require.Error(t, err, "Expected error for YAML generation, got nil")
	assert.True(t, strings.Contains(err.Error(), expectedErr.Error()),
		"Expected error to contain %q, got %q", expectedErr, err.Error())
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
	cfg := Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		OutputFile:   "",
		Registry: func() []config.ConfigOption {
			return []config.ConfigOption{}
		},
	}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateMarkdownDocs failed with empty registry")

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
		assert.True(t, strings.Contains(output, section), "Missing section with empty registry: %s", section)
	}

	// Check table headers still exist
	tableHeaders := "| Key | Type | Default | Environment Variable | Description |"
	assert.True(t, strings.Contains(output, tableHeaders), "Missing table headers with empty registry")

	// Check that the blocks are properly closed
	assert.True(t, strings.Contains(output, "```yaml") && strings.Contains(output, "```bash"),
		"Missing code blocks with empty registry")
}

// TestMarkdownGenerationNoUserPaths tests that generated docs don't contain user-specific paths
// This test will initially FAIL - generated docs contain actual user paths like /Users/peiman/
func TestMarkdownGenerationNoUserPaths(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer

	// Simulate a user-specific path
	appInfo := AppInfo{
		BinaryName: "test-app",
		EnvPrefix:  "TEST_APP",
	}
	appInfo.ConfigPaths.DefaultPath = "/Users/someuser/.test-app.yaml"
	appInfo.ConfigPaths.DefaultFullName = ".test-app.yaml"

	cfg := Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		Registry:     func() []config.ConfigOption { return []config.ConfigOption{} },
	}

	gen := NewGenerator(cfg)

	// EXECUTION PHASE
	err := gen.GenerateMarkdownDocs(&buf, appInfo)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateMarkdownDocs failed")

	output := buf.String()

	// Should NOT contain user-specific paths
	assert.False(t, strings.Contains(output, "/Users/"),
		"Generated markdown should not contain /Users/ paths")
	assert.False(t, strings.Contains(output, "someuser"),
		"Generated markdown should not contain usernames")

	// SHOULD contain generic placeholder
	assert.True(t, strings.Contains(output, "$HOME") || strings.Contains(output, "~"),
		"Generated markdown should use $HOME or ~ for home directory")
}
