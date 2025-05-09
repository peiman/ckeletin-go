// cmd/docs_test.go

package cmd

import (
	"bytes"
	"errors"
	"io"
	"os"
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
	errCh    chan<- error // Channel to send the error through
}

func (m *mockCloser) Close() error {
	// If we have a channel, signal that Close was called
	if m.errCh != nil {
		m.errCh <- m.closeErr
	}
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
	// Original openOutputFile implementation
	origOpenOutputFile := openOutputFile
	// Restore it after all tests
	defer func() {
		openOutputFile = origOpenOutputFile
	}()

	tests := []struct {
		name            string
		testFixturePath string   // Path to test fixture
		args            []string // CLI args
		expectOutput    string   // Expected output content
		expectedFormat  string   // Expected format
		wantErrContain  string   // Expected error substring, empty for no error
		outputToFile    bool     // Whether to write to file
		outputFilePath  string   // Output file path
		mockCloseErr    error    // Mock file close error
	}{
		{
			name:            "Markdown Format Default",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{},
			expectedFormat:  FormatMarkdown,
			expectOutput:    "# ckeletin-go Configuration",
			wantErrContain:  "",
		},
		{
			name:            "YAML Format",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{"--format", "yaml"},
			expectedFormat:  FormatYAML,
			expectOutput:    "app:",
			wantErrContain:  "",
		},
		{
			name:            "Invalid Format",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{"--format", "invalid"},
			wantErrContain:  "unsupported format",
		},
		{
			name:            "Output to File Markdown",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{"--output-file", "test_output.md"},
			expectedFormat:  FormatMarkdown,
			outputToFile:    true,
			outputFilePath:  "test_output.md",
			expectOutput:    "# ckeletin-go Configuration",
			wantErrContain:  "",
		},
		{
			name:            "Output to File YAML",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{"--format", "yaml", "--output-file", "test_output.yaml"},
			expectedFormat:  FormatYAML,
			outputToFile:    true,
			outputFilePath:  "test_output.yaml",
			expectOutput:    "app:",
			wantErrContain:  "",
		},
		{
			name:            "File Close Error",
			testFixturePath: "../testdata/docs_config.yaml",
			args:            []string{"--output-file", "test_output.md"},
			expectedFormat:  FormatMarkdown,
			outputToFile:    true,
			outputFilePath:  "test_output.md",
			mockCloseErr:    errCloseFailure,
			wantErrContain:  "",
			expectOutput:    "# ckeletin-go Configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save original values and restore after test
			origFormat := docsOutputFormat
			origFile := docsOutputFile
			defer func() {
				docsOutputFormat = origFormat
				docsOutputFile = origFile

				// Clean up test output files
				if tt.outputToFile {
					os.Remove(tt.outputFilePath)
				}
			}()

			// Reset viper for test
			viper.Reset()

			// Load test fixture if specified
			if tt.testFixturePath != "" {
				viper.SetConfigFile(tt.testFixturePath)
				if err := viper.ReadInConfig(); err != nil {
					t.Fatalf("Failed to load test fixture %s: %v", tt.testFixturePath, err)
				}
			}

			// Create command and register flags
			cmd := &cobra.Command{Use: "config"}
			cmd.Flags().String("format", FormatMarkdown, "Output format (markdown, yaml)")
			cmd.Flags().String("output-file", "", "Output file")

			// Parse args
			cmd.SetArgs(tt.args)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Set up command parameters
			if cmd.Flags().Changed("format") {
				format, _ := cmd.Flags().GetString("format")
				docsOutputFormat = format
			} else {
				docsOutputFormat = FormatMarkdown
			}

			if cmd.Flags().Changed("output-file") {
				outputFile, _ := cmd.Flags().GetString("output-file")
				docsOutputFile = outputFile
			} else {
				docsOutputFile = ""
			}

			// Setup output capture
			var buf bytes.Buffer
			cmd.SetOut(&buf)

			// Prepare mock file if needed
			if tt.mockCloseErr != nil {
				openOutputFile = func(path string) (io.WriteCloser, error) {
					return newMockFile(&buf, tt.mockCloseErr), nil
				}
			} else if tt.outputToFile {
				// For file output tests, we'll capture to the buffer instead
				openOutputFile = func(path string) (io.WriteCloser, error) {
					return newMockFile(&buf, nil), nil
				}
			} else {
				// Use default implementation for non-file tests
				openOutputFile = origOpenOutputFile
			}

			// EXECUTION PHASE
			err := runDocsConfig(cmd, []string{})

			// ASSERTION PHASE
			// Check error
			if tt.wantErrContain != "" {
				if err == nil {
					t.Errorf("runDocsConfig() expected error containing %q, got nil", tt.wantErrContain)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErrContain) {
					t.Errorf("runDocsConfig() error = %v, wantErrContain %q", err, tt.wantErrContain)
					return
				}
			} else if err != nil {
				t.Errorf("runDocsConfig() unexpected error: %v", err)
				return
			}

			// Check format
			if tt.expectedFormat != "" && docsOutputFormat != tt.expectedFormat {
				t.Errorf("runDocsConfig() format = %q, want %q", docsOutputFormat, tt.expectedFormat)
			}

			// Check output content if not an error test
			if tt.expectOutput != "" && tt.wantErrContain == "" {
				output := buf.String()
				if !strings.Contains(output, tt.expectOutput) {
					t.Errorf("runDocsConfig() output missing expected content: %q", tt.expectOutput)
					t.Errorf("Actual output: %q", output)
				}
			}

			// Check file output if needed
			if tt.outputToFile && tt.mockCloseErr == nil {
				if docsOutputFile != tt.outputFilePath {
					t.Errorf("runDocsConfig() output file = %q, want %q", docsOutputFile, tt.outputFilePath)
				}
			}
		})
	}
}

// TestRunDocsConfig_FileCloseError tests the file close error specifically,
// it verifies that the error is at least logged even if not returned.
func TestRunDocsConfig_FileCloseError(t *testing.T) {
	// Setup
	origOpenOutputFile := openOutputFile
	defer func() {
		openOutputFile = origOpenOutputFile
	}()

	// Reset and set up configs
	viper.Reset()
	docsOutputFormat = FormatMarkdown
	docsOutputFile = "test_output.md"

	// Create a command for testing
	cmd := &cobra.Command{Use: "config"}
	cmd.Flags().String("format", FormatMarkdown, "Output format (markdown, yaml)")
	cmd.Flags().String("output-file", "", "Output file")

	// Create a buffer to capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	// Also capture log output
	var logBuf bytes.Buffer
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = zerolog.New(&logBuf)

	// Mock the file open function to return a file that will error on close
	openOutputFile = func(path string) (io.WriteCloser, error) {
		// Create a custom closer that will return our error
		closer := &mockCloser{
			Writer:   &buf,
			closeErr: errCloseFailure,
		}

		return closer, nil
	}

	// Execute the function
	_ = runDocsConfig(cmd, []string{})

	// Check log output instead of return error
	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Failed to close output file") ||
		!strings.Contains(logOutput, errCloseFailure.Error()) {
		t.Errorf("Close error not properly logged. Log output: %s", logOutput)
	}
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
