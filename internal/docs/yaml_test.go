// internal/docs/yaml_test.go

package docs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
)

// TestGenerateYAMLDocs tests the YAML document generation
func TestGenerateYAMLDocs(t *testing.T) {
	// SETUP PHASE
	// Create output buffer
	var buf bytes.Buffer

	// Create a mock registry function for testing
	mockRegistry := func() []config.ConfigOption {
		return []config.ConfigOption{
			{
				Key:          "app.log_level",
				Type:         "string",
				DefaultValue: "info",
				Description:  "Application log level",
			},
			{
				Key:          "app.ping.enabled",
				Type:         "bool",
				DefaultValue: true,
				Description:  "Enable ping endpoint",
			},
		}
	}

	// Create generator with mock registry
	cfg := NewConfig(&buf, WithOutputFormat(FormatYAML), WithRegistryFunc(mockRegistry))
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateYAMLDocs(&buf)

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("GenerateYAMLDocs failed: %v", err)
	}

	output := buf.String()

	// Check for basic YAML structure
	expectedLines := []string{
		"app:",         // Top-level section
		"  log_level:", // Option
		"  ping:",      // Nested section
	}

	for _, line := range expectedLines {
		if !strings.Contains(output, line) {
			t.Errorf("Missing expected YAML line: %s", line)
		}
	}

	// Check that options have descriptions
	if !strings.Contains(output, "  # ") {
		t.Errorf("Missing option description comments")
	}

	// Check that we have proper indentation for nested options
	if !strings.Contains(output, "    ") { // 4-space indentation for nested options
		t.Errorf("Missing proper indentation for nested options")
	}
}

// TestGenerateYAMLContent tests the YAML content generator
func TestGenerateYAMLContent(t *testing.T) {
	// SETUP PHASE
	// Create a simple mock config registry for testing
	mockOptions := []struct {
		key         string
		description string
	}{
		{"app.simple", "A simple option"},
		{"app.nested.option", "A nested option"},
		{"standalone", "A standalone option"},
	}

	// Build mock ConfigOptions from the simple data
	mockConfigOptions := make([]config.ConfigOption, 0, len(mockOptions))
	for _, opt := range mockOptions {
		mockConfigOptions = append(mockConfigOptions, config.ConfigOption{
			Key:          opt.key,
			Description:  opt.description,
			DefaultValue: "test-value",
			Type:         "string",
		})
	}

	// Create buffer for output
	var buf bytes.Buffer

	// EXECUTION PHASE
	err := generateYAMLContent(&buf, mockConfigOptions)

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("generateYAMLContent failed: %v", err)
	}

	output := buf.String()

	// Check basic structure
	expectedStructure := []string{
		"app:",
		"  # A simple option",
		"  simple: test-value",
		"  nested:",
		"    # A nested option",
		"    option: test-value",
		"# A standalone option",
		"standalone: test-value",
	}

	for _, line := range expectedStructure {
		if !strings.Contains(output, line) {
			t.Errorf("Missing expected YAML content: %s", line)
		}
	}

	// Verify proper indentation logic
	if strings.Contains(output, "app.simple") {
		t.Errorf("Improper key formatting - did not properly convert dots to nesting")
	}
}

// TestGenerateYAMLDocs_EmptyRegistry tests handling of an empty registry
func TestGenerateYAMLDocs_EmptyRegistry(t *testing.T) {
	// SETUP PHASE
	// Create output buffer
	var buf bytes.Buffer

	// Create generator with empty registry
	cfg := NewConfig(&buf, WithOutputFormat(FormatYAML), WithRegistryFunc(func() []config.ConfigOption {
		return []config.ConfigOption{}
	}))
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateYAMLDocs(&buf)

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("GenerateYAMLDocs failed with empty registry: %v", err)
	}

	// For an empty registry, we expect an empty output (or just whitespace)
	output := buf.String()
	trimmed := strings.TrimSpace(output)
	if len(trimmed) > 0 {
		t.Errorf("Expected empty output for empty registry, got: %q", output)
	}
}
