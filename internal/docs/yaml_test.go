// internal/docs/yaml_test.go

package docs

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	cfg := Config{
		Writer:       &buf,
		OutputFormat: FormatYAML,
		OutputFile:   "",
		Registry:     mockRegistry,
	}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateYAMLDocs(&buf)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateYAMLDocs failed")

	output := buf.String()

	// Check for basic YAML structure
	expectedLines := []string{
		"app:",         // Top-level section
		"  log_level:", // Option
		"  ping:",      // Nested section
	}

	for _, line := range expectedLines {
		assert.True(t, strings.Contains(output, line), "Missing expected YAML line: %s", line)
	}

	// Check that options have descriptions
	assert.True(t, strings.Contains(output, "  # "), "Missing option description comments")

	// Check that we have proper indentation for nested options
	assert.True(t, strings.Contains(output, "    "), // 4-space indentation for nested options
		"Missing proper indentation for nested options")
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
	require.NoError(t, err, "generateYAMLContent failed")

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
		assert.True(t, strings.Contains(output, line), "Missing expected YAML content: %s", line)
	}

	// Verify proper indentation logic
	assert.False(t, strings.Contains(output, "app.simple"),
		"Improper key formatting - did not properly convert dots to nesting")
}

// TestGenerateYAMLContent_DeterministicOutput verifies the generator emits
// identical output across runs (map iteration order must not leak through)
func TestGenerateYAMLContent_DeterministicOutput(t *testing.T) {
	// SETUP PHASE
	// Multiple top-level groups and nested groups so randomized map iteration
	// order would be caught with near-certainty if it leaked into the output
	registry := []config.ConfigOption{
		{Key: "app.log_level", Description: "Log level", DefaultValue: "info", Type: "string"},
		{Key: "app.ping.enabled", Description: "Enable ping", DefaultValue: true, Type: "bool"},
		{Key: "app.ping.message", Description: "Ping message", DefaultValue: "pong", Type: "string"},
		{Key: "server.port", Description: "Server port", DefaultValue: 8080, Type: "int"},
		{Key: "server.tls.cert", Description: "TLS certificate", DefaultValue: "cert.pem", Type: "string"},
		{Key: "telemetry.enabled", Description: "Enable telemetry", DefaultValue: false, Type: "bool"},
		{Key: "standalone", Description: "A standalone option", DefaultValue: "x", Type: "string"},
	}

	// EXECUTION PHASE
	var first bytes.Buffer
	require.NoError(t, generateYAMLContent(&first, registry), "generateYAMLContent failed")

	// ASSERTION PHASE
	for i := 0; i < 20; i++ {
		var buf bytes.Buffer
		require.NoError(t, generateYAMLContent(&buf, registry), "generateYAMLContent failed on run %d", i+1)
		require.Equal(t, first.String(), buf.String(),
			"run %d produced different output - generation must be deterministic", i+1)
	}

	// Top-level groups should be emitted in sorted order
	output := first.String()
	appIdx := strings.Index(output, "app:")
	serverIdx := strings.Index(output, "server:")
	telemetryIdx := strings.Index(output, "telemetry:")
	assert.True(t, appIdx < serverIdx && serverIdx < telemetryIdx,
		"top-level groups should appear in sorted order, got app=%d server=%d telemetry=%d",
		appIdx, serverIdx, telemetryIdx)
}

// TestGenerateYAMLContent_WriteError verifies write failures propagate
// instead of being silently swallowed.
func TestGenerateYAMLContent_WriteError(t *testing.T) {
	// SETUP PHASE
	registry := []config.ConfigOption{
		{Key: "app.log_level", Description: "Log level", DefaultValue: "info", Type: "string"},
		{Key: "app.ping.enabled", Description: "Enable ping", DefaultValue: true, Type: "bool"},
	}

	// Fails partway through the document
	w := &failAfterWriter{limit: 16}

	// EXECUTION PHASE
	err := generateYAMLContent(w, registry)

	// ASSERTION PHASE
	require.Error(t, err, "a failing writer must surface an error")
	assert.ErrorIs(t, err, errWriteFailed)
}

// TestGenerateYAMLDocs_WriteError verifies the write error reaches the
// public entry point.
func TestGenerateYAMLDocs_WriteError(t *testing.T) {
	// SETUP PHASE
	cfg := Config{OutputFormat: FormatYAML, Registry: config.Registry}
	generator := NewGenerator(cfg)
	w := &failAfterWriter{limit: 16}

	// EXECUTION PHASE
	err := generator.GenerateYAMLDocs(w)

	// ASSERTION PHASE
	require.Error(t, err, "a failing writer must surface an error")
	assert.ErrorIs(t, err, errWriteFailed)
}

// TestGenerateYAMLDocs_EmptyRegistry tests handling of an empty registry
func TestGenerateYAMLDocs_EmptyRegistry(t *testing.T) {
	// SETUP PHASE
	// Create output buffer
	var buf bytes.Buffer

	// Create generator with empty registry
	cfg := Config{
		Writer:       &buf,
		OutputFormat: FormatYAML,
		OutputFile:   "",
		Registry: func() []config.ConfigOption {
			return []config.ConfigOption{}
		},
	}
	generator := NewGenerator(cfg)

	// EXECUTION PHASE
	err := generator.GenerateYAMLDocs(&buf)

	// ASSERTION PHASE
	require.NoError(t, err, "GenerateYAMLDocs failed with empty registry")

	// For an empty registry, we expect an empty output (or just whitespace)
	output := buf.String()
	trimmed := strings.TrimSpace(output)
	assert.Empty(t, trimmed, "Expected empty output for empty registry, got: %q", output)
}
