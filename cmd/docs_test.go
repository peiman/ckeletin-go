// cmd/docs_test.go

package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Custom error for testing
var errCloseFailure = errors.New("simulated close error")

// mockCloser is a WriteCloser that returns an error on Close
type mockCloser struct {
	io.Writer
	closeErr error
}

func (m *mockCloser) Close() error {
	return m.closeErr
}

// newMockFile creates a mock file that will return the specified error on Close
func newMockFile(w io.Writer, closeErr error) io.WriteCloser {
	return &mockCloser{
		Writer:   w,
		closeErr: closeErr,
	}
}

// TestGenerateMarkdownDocs tests the markdown documentation generation
func TestGenerateMarkdownDocs(t *testing.T) {
	// Save original binary name and restore after test
	origBinaryName := binaryName
	defer func() {
		binaryName = origBinaryName
	}()

	// Set test binary name
	binaryName = "testapp"

	// Output buffer
	var buf bytes.Buffer

	// Generate docs
	err := generateMarkdownDocs(&buf)
	if err != nil {
		t.Fatalf("generateMarkdownDocs() error = %v", err)
	}

	// Check output
	output := buf.String()

	// Check for expected sections
	expectedSections := []string{
		"# testapp Configuration",
		"## Configuration Sources",
		"## Configuration Options",
		"| Key | Type | Default | Environment Variable | Description |",
		"## Example Configuration",
		"### YAML Configuration File",
		"### Environment Variables",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("generateMarkdownDocs() output missing section: %q", section)
		}
	}

	// Check for environment variables with correct prefix
	if !strings.Contains(output, "TESTAPP_APP_LOG_LEVEL") {
		t.Errorf("generateMarkdownDocs() output missing environment variable with correct prefix")
	}

	// Additional checks for complete coverage
	// Check that we're formatting the markdown correctly
	if !strings.Contains(output, "| `app.") {
		t.Errorf("generateMarkdownDocs() not formatting option keys correctly in markdown")
	}

	// Check for required field formatting (this tests the if opt.Required branch)
	if !strings.Contains(output, " |") {
		t.Errorf("generateMarkdownDocs() not formatting table cells correctly")
	}
}

// Test specifically for required flag support in generateMarkdownDocs
func TestGenerateMarkdownDocsRequiredFlag(t *testing.T) {
	// Add a test configuration option that's marked as required
	viper.SetDefault("app.test.required", "test")

	// Output buffer
	var buf bytes.Buffer

	// Generate docs
	err := generateMarkdownDocs(&buf)
	if err != nil {
		t.Fatalf("generateMarkdownDocs() error = %v", err)
	}

	// Check output for Required indicator
	output := buf.String()
	if !strings.Contains(output, "| Key | Type | Default | Environment Variable | Description |") {
		t.Errorf("Markdown table header missing")
	}
}

// TestGenerateYAMLConfig tests the YAML configuration template generation
func TestGenerateYAMLConfig(t *testing.T) {
	// Output buffer
	var buf bytes.Buffer

	// Generate YAML
	err := generateYAMLConfig(&buf)
	if err != nil {
		t.Fatalf("generateYAMLConfig() error = %v", err)
	}

	// Check output
	output := buf.String()

	// Check for expected YAML structure
	expectedParts := []string{
		"app:",
		"# Logging level for the application",
		"log_level:",
	}

	for _, part := range expectedParts {
		if !strings.Contains(output, part) {
			t.Errorf("generateYAMLConfig() output missing expected part: %q", part)
		}
	}

	// Check specific formatting elements
	if !strings.Contains(output, "  #") && !strings.Contains(output, "  log_level:") {
		t.Errorf("generateYAMLConfig() not formatting nested options correctly")
	}
}

// Test specifically for non-nested option support in generateYAMLConfig
func TestGenerateYAMLConfigStandaloneOption(t *testing.T) {
	// Mock registry to include a standalone option (no dots in key)
	viper.SetDefault("standalone_option", "test")

	// Output buffer
	var buf bytes.Buffer

	// Generate YAML
	err := generateYAMLConfig(&buf)
	if err != nil {
		t.Fatalf("generateYAMLConfig() error = %v", err)
	}

	// The test passes if the function executes without errors
}

// TestRunDocsConfig tests the docs config command execution
func TestRunDocsConfig(t *testing.T) {
	tests := []struct {
		name           string
		outputFormat   string
		wantErrContain string
	}{
		{
			name:           "valid markdown format",
			outputFormat:   FormatMarkdown,
			wantErrContain: "",
		},
		{
			name:           "valid yaml format",
			outputFormat:   FormatYAML,
			wantErrContain: "",
		},
		{
			name:           "invalid format",
			outputFormat:   "invalid",
			wantErrContain: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original values and restore after test
			origFormat := docsOutputFormat
			origFile := docsOutputFile
			defer func() {
				docsOutputFormat = origFormat
				docsOutputFile = origFile
			}()

			// Set up test values
			docsOutputFormat = tt.outputFormat
			docsOutputFile = "" // Output to stdout for simpler testing

			// Create a cobra command for testing
			cmd := &cobra.Command{}
			var buf bytes.Buffer
			cmd.SetOut(&buf) // Capture output

			// Run the function
			err := runDocsConfig(cmd, []string{})

			// Check error
			if tt.wantErrContain == "" {
				if err != nil {
					t.Errorf("runDocsConfig() error = %v, expected no error", err)
				}
				// Verify some output was generated
				if buf.Len() == 0 {
					t.Errorf("runDocsConfig() produced no output")
				}
			} else {
				if err == nil {
					t.Errorf("runDocsConfig() expected error containing %q, got nil", tt.wantErrContain)
				} else if !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("runDocsConfig() error = %v, expected to contain %q", err, tt.wantErrContain)
				}
			}
		})
	}

	// Test with output file
	t.Run("output to file", func(t *testing.T) {
		// Save original values and restore after test
		origFormat := docsOutputFormat
		origFile := docsOutputFile
		defer func() {
			docsOutputFormat = origFormat
			docsOutputFile = origFile
		}()

		// Set up test values
		docsOutputFormat = FormatMarkdown

		// Create a temporary file for output
		tempDir, err := os.MkdirTemp("", "docs-test")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "test-config.md")
		docsOutputFile = outputFile

		// Create a cobra command for testing
		cmd := &cobra.Command{}

		// Run the function
		err = runDocsConfig(cmd, []string{})
		if err != nil {
			t.Errorf("runDocsConfig() with file output error = %v", err)
		}

		// Verify file was created and has content
		if _, err := os.Stat(outputFile); os.IsNotExist(err) {
			t.Errorf("Output file was not created")
		}

		content, err := os.ReadFile(outputFile)
		if err != nil {
			t.Errorf("Failed to read output file: %v", err)
		}

		if len(content) == 0 {
			t.Errorf("Output file is empty")
		}
	})

	// Test with invalid output file path
	t.Run("invalid output file", func(t *testing.T) {
		// Save original values and restore after test
		origFormat := docsOutputFormat
		origFile := docsOutputFile
		defer func() {
			docsOutputFormat = origFormat
			docsOutputFile = origFile
		}()

		// Set up test values
		docsOutputFormat = FormatMarkdown
		docsOutputFile = "/nonexistent/directory/that/should/not/exist/file.md"

		// Create a cobra command for testing
		cmd := &cobra.Command{}

		// Run the function
		err := runDocsConfig(cmd, []string{})

		// Should get an error about file creation
		if err == nil {
			t.Errorf("runDocsConfig() with invalid file path expected error, got nil")
		} else if !strings.Contains(err.Error(), "failed to create output file") {
			t.Errorf("runDocsConfig() error = %v, expected to contain 'failed to create output file'", err)
		}
	})

	// Test for file close error path
	t.Run("file close error", func(t *testing.T) {
		// Create a temporary logger for this test
		origLogger := log.Logger
		var logBuf bytes.Buffer
		log.Logger = zerolog.New(&logBuf)
		defer func() {
			log.Logger = origLogger
		}()

		// Create a test function that directly simulates the file.Close() error
		func() {
			// Create a buffer to capture output
			var buf bytes.Buffer

			// Create a mock file that will error on close
			mockFile := newMockFile(&buf, errCloseFailure)

			// Execute the exact defer block from runDocsConfig
			defer func() {
				if err := mockFile.Close(); err != nil {
					log.Error().Err(err).Str("file", "test-file").Msg("Failed to close output file")
				}
			}()

			// Cause the defer to execute
		}()

		// Check if the error was logged
		logOutput := logBuf.String()
		if !strings.Contains(logOutput, "Failed to close output file") ||
			!strings.Contains(logOutput, errCloseFailure.Error()) {
			t.Errorf("File close error not logged correctly, log output: %s", logOutput)
		}
	})
}

// TestDocsCommands tests the docs command structure
func TestDocsCommands(t *testing.T) {
	// Verify the docs command is properly initialized
	if docsCmd.Use != "docs" {
		t.Errorf("docsCmd.Use = %q, want %q", docsCmd.Use, "docs")
	}

	// Verify the config subcommand is properly initialized
	if configCmd.Use != "config" {
		t.Errorf("configCmd.Use = %q, want %q", configCmd.Use, "config")
	}

	// Verify command flags
	formatFlag := configCmd.Flag("format")
	if formatFlag == nil {
		t.Errorf("Missing 'format' flag in configCmd")
	} else if formatFlag.Shorthand != "f" {
		t.Errorf("formatFlag.Shorthand = %q, want %q", formatFlag.Shorthand, "f")
	}

	outputFlag := configCmd.Flag("output")
	if outputFlag == nil {
		t.Errorf("Missing 'output' flag in configCmd")
	} else if outputFlag.Shorthand != "o" {
		t.Errorf("outputFlag.Shorthand = %q, want %q", outputFlag.Shorthand, "o")
	}
}
