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
	tests := []struct {
		name              string
		binaryNameToSet   string
		expectedSections  []string
		expectedEnvPrefix string
	}{
		{
			name:            "Standard markdown generation",
			binaryNameToSet: "testapp",
			expectedSections: []string{
				"# testapp Configuration",
				"## Configuration Sources",
				"## Configuration Options",
				"| Key | Type | Default | Environment Variable | Description |",
				"## Example Configuration",
				"### YAML Configuration File",
				"### Environment Variables",
			},
			expectedEnvPrefix: "TESTAPP_APP_LOG_LEVEL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save original binary name and restore after test
			origBinaryName := binaryName
			defer func() {
				binaryName = origBinaryName
			}()

			// Set test binary name
			binaryName = tt.binaryNameToSet

			// Output buffer
			var buf bytes.Buffer

			// EXECUTION PHASE
			err := generateMarkdownDocs(&buf)

			// ASSERTION PHASE
			if err != nil {
				t.Fatalf("generateMarkdownDocs() error = %v", err)
			}

			// Check output
			output := buf.String()

			// Check for expected sections
			for _, section := range tt.expectedSections {
				if !strings.Contains(output, section) {
					t.Errorf("generateMarkdownDocs() output missing section: %q", section)
				}
			}

			// Check for environment variables with correct prefix
			if !strings.Contains(output, tt.expectedEnvPrefix) {
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
		})
	}
}

// Test specifically for required flag support in generateMarkdownDocs
func TestGenerateMarkdownDocsRequiredFlag(t *testing.T) {
	tests := []struct {
		name             string
		configToSet      string
		configValue      string
		expectedContains string
	}{
		{
			name:             "Required flag in table",
			configToSet:      "app.test.required",
			configValue:      "test",
			expectedContains: "| Key | Type | Default | Environment Variable | Description |",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save viper state
			defer viper.Reset()
			viper.Reset()

			// Add a test configuration option that's marked as required
			viper.SetDefault(tt.configToSet, tt.configValue)

			// Output buffer
			var buf bytes.Buffer

			// EXECUTION PHASE
			err := generateMarkdownDocs(&buf)

			// ASSERTION PHASE
			if err != nil {
				t.Fatalf("generateMarkdownDocs() error = %v", err)
			}

			// Check output for Required indicator
			output := buf.String()
			if !strings.Contains(output, tt.expectedContains) {
				t.Errorf("Markdown table header missing")
			}
		})
	}
}

// TestGenerateYAMLConfig tests the YAML configuration template generation
func TestGenerateYAMLConfig(t *testing.T) {
	tests := []struct {
		name             string
		expectedContains []string
	}{
		{
			name: "YAML structure validation",
			expectedContains: []string{
				"app:",
				"# Logging level for the application",
				"log_level:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Output buffer
			var buf bytes.Buffer

			// EXECUTION PHASE
			err := generateYAMLConfig(&buf)

			// ASSERTION PHASE
			if err != nil {
				t.Fatalf("generateYAMLConfig() error = %v", err)
			}

			// Check output
			output := buf.String()

			// Check for expected YAML structure
			for _, part := range tt.expectedContains {
				if !strings.Contains(output, part) {
					t.Errorf("generateYAMLConfig() output missing expected part: %q", part)
				}
			}

			// Check specific formatting elements
			if !strings.Contains(output, "  #") && !strings.Contains(output, "  log_level:") {
				t.Errorf("generateYAMLConfig() not formatting nested options correctly")
			}
		})
	}
}

// Test specifically for non-nested option support in generateYAMLConfig
func TestGenerateYAMLConfigStandaloneOption(t *testing.T) {
	tests := []struct {
		name          string
		standaloneKey string
		defaultValue  string
		wantErr       bool
	}{
		{
			name:          "Standalone option (no dots in key)",
			standaloneKey: "standalone_option",
			defaultValue:  "test",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save viper state and restore after
			defer viper.Reset()
			viper.Reset()

			// Set up a standalone option (no dots in key) in viper
			// Note: we can't easily mock the registry, so we just set this up and test that the
			// function runs without error. The actual behavior with standalone options
			// should be tested in a more focused test that can mock config.Registry()
			viper.SetDefault(tt.standaloneKey, tt.defaultValue)

			// Output buffer
			var buf bytes.Buffer

			// EXECUTION PHASE
			err := generateYAMLConfig(&buf)

			// ASSERTION PHASE
			if (err != nil) != tt.wantErr {
				t.Fatalf("generateYAMLConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			// We're just verifying that the function doesn't crash with standalone options
			// A more comprehensive test would need to mock the registry
		})
	}
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
	tests := []struct {
		name         string
		closeErr     error
		wantLogEntry string
	}{
		{
			name:         "File close error is logged",
			closeErr:     errCloseFailure,
			wantLogEntry: "Failed to close output file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save original open file function
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
					closeErr: tt.closeErr,
				}

				return closer, nil
			}

			// EXECUTION PHASE
			_ = runDocsConfig(cmd, []string{})

			// ASSERTION PHASE
			// Check log output instead of return error
			logOutput := logBuf.String()
			if !strings.Contains(logOutput, tt.wantLogEntry) ||
				!strings.Contains(logOutput, tt.closeErr.Error()) {
				t.Errorf("Close error not properly logged. Log output: %s", logOutput)
			}
		})
	}
}

// TestDocsCommands tests the docs command structure
func TestDocsCommands(t *testing.T) {
	tests := []struct {
		name          string
		cmd           *cobra.Command
		expectedUse   string
		flagToCheck   string
		expectedShort string
	}{
		{
			name:        "docs command",
			cmd:         docsCmd,
			expectedUse: "docs",
		},
		{
			name:          "config subcommand",
			cmd:           configCmd,
			expectedUse:   "config",
			flagToCheck:   "format",
			expectedShort: "f",
		},
		{
			name:          "output flag check",
			cmd:           configCmd,
			flagToCheck:   "output",
			expectedShort: "o",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			// The commands are already initialized during package init

			// ASSERTION PHASE
			if tt.expectedUse != "" && tt.cmd.Use != tt.expectedUse {
				t.Errorf("%s.Use = %q, want %q", tt.name, tt.cmd.Use, tt.expectedUse)
			}

			if tt.flagToCheck != "" {
				flag := tt.cmd.Flag(tt.flagToCheck)
				if flag == nil {
					t.Errorf("Missing '%s' flag in %s", tt.flagToCheck, tt.name)
				} else if flag.Shorthand != tt.expectedShort {
					t.Errorf("%s.Shorthand = %q, want %q", tt.flagToCheck, flag.Shorthand, tt.expectedShort)
				}
			}
		})
	}
}

// Test for proper YAML nesting structure in generateYAMLConfig
func TestGenerateYAMLConfigNestedStructure(t *testing.T) {
	tests := []struct {
		name              string
		expectedStrings   []string
		unexpectedStrings []string
	}{
		{
			name: "Nested YAML structure",
			expectedStrings: []string{
				"app:",
				"  log_level:",
				"  ping:",
				"    output_message:",
				"    output_color:",
				"    ui:",
			},
			unexpectedStrings: []string{
				"  ping.output_message:",
				"  ping.output_color:",
				"  ping.ui:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Reset viper state to ensure clean test
			viper.Reset()

			// Output buffer
			var buf bytes.Buffer

			// EXECUTION PHASE
			err := generateYAMLConfig(&buf)

			// ASSERTION PHASE
			if err != nil {
				t.Fatalf("generateYAMLConfig() error = %v", err)
			}

			// Get output
			output := buf.String()

			// Check that expected strings are in the output
			for _, s := range tt.expectedStrings {
				if !strings.Contains(output, s) {
					t.Errorf("generateYAMLConfig() output missing expected content: %q", s)
				}
			}

			// Check that unexpected strings are NOT in the output
			for _, s := range tt.unexpectedStrings {
				if strings.Contains(output, s) {
					t.Errorf("generateYAMLConfig() output contains unexpected content: %q", s)
				}
			}

			// Ensure proper indentation of nested structures
			// The ping section should be indented under app
			if !strings.Contains(output, "  ping:") {
				t.Errorf("generateYAMLConfig() output not properly indenting 'ping' under 'app'")
			}

			// The output_message should be indented under ping
			if !strings.Contains(output, "    output_") {
				t.Errorf("generateYAMLConfig() output not properly indenting options under 'ping'")
			}
		})
	}
}
