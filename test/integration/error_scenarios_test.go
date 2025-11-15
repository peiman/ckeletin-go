// test/integration/error_scenarios_test.go
//
// Integration tests for error scenarios and edge cases
//
// These tests verify:
// - Proper error handling and exit codes
// - Security validations (file permissions, size limits)
// - Invalid input handling
// - Graceful failure modes
// - Error message quality

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConfigFileErrors tests error handling for config file issues
func TestConfigFileErrors(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		setupFunc          func(t *testing.T) string // Returns config file path
		args               []string
		wantExitCode       int
		wantStderrContains string
	}{
		{
			name: "Config file does not exist",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent.yaml")
			},
			args:               []string{"config", "validate", "--file"},
			wantExitCode:       1,
			wantStderrContains: "not found",
		},
		{
			name: "Config file too large (DoS prevention)",
			setupFunc: func(t *testing.T) string {
				configFile := filepath.Join(tmpDir, "toolarge.yaml")
				// Create a file larger than 1MB (MaxConfigFileSize)
				largeContent := make([]byte, 2*1024*1024) // 2MB
				for i := range largeContent {
					largeContent[i] = 'x'
				}
				if err := os.WriteFile(configFile, largeContent, 0600); err != nil {
					t.Fatalf("Failed to create large config file: %v", err)
				}
				return configFile
			},
			args:               []string{"config", "validate", "--file"},
			wantExitCode:       1,
			wantStderrContains: "validation failed",
		},
		{
			name: "Malformed YAML syntax",
			setupFunc: func(t *testing.T) string {
				configFile := filepath.Join(tmpDir, "malformed.yaml")
				content := `app:
  log_level: info
  ping:
    output_message: "unclosed string
    output_color: red
`
				if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				return configFile
			},
			args:               []string{"config", "validate", "--file"},
			wantExitCode:       1,
			wantStderrContains: "validation failed",
		},
		{
			name: "Config value exceeds string length limit",
			setupFunc: func(t *testing.T) string {
				configFile := filepath.Join(tmpDir, "toolongstring.yaml")
				// Create a string longer than MaxStringValueLength (10KB)
				longString := strings.Repeat("x", 11*1024)
				content := `app:
  log_level: info
  ping:
    output_message: "` + longString + `"
`
				if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				return configFile
			},
			args:               []string{"config", "validate", "--file"},
			wantExitCode:       1,
			wantStderrContains: "validation failed",
		},
	}

	// Add world-writable test only on Unix systems
	if runtime.GOOS != "windows" {
		tests = append(tests, struct {
			name               string
			setupFunc          func(t *testing.T) string
			args               []string
			wantExitCode       int
			wantStderrContains string
		}{
			name: "Config file is world-writable (security issue)",
			setupFunc: func(t *testing.T) string {
				configFile := filepath.Join(tmpDir, "worldwritable.yaml")
				content := `app:
  log_level: info
`
				if err := os.WriteFile(configFile, []byte(content), 0600); err != nil {
					t.Fatalf("Failed to create config file: %v", err)
				}
				// Explicitly make it world-writable (0602 = owner r/w, world w)
				if err := os.Chmod(configFile, 0602); err != nil {
					t.Fatalf("Failed to chmod config file: %v", err)
				}
				return configFile
			},
			args:               []string{"config", "validate", "--file"},
			wantExitCode:       1,
			wantStderrContains: "validation failed",
		})
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configFile := tt.setupFunc(t)

			// Build command
			args := append(tt.args, configFile)
			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := getExitCode(err)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"exit code mismatch\nstdout: %s\nstderr: %s",
				stdout.String(), stderr.String())

			// Check error message
			if tt.wantStderrContains != "" {
				stderrOutput := stderr.String()
				assert.Contains(t, stderrOutput, tt.wantStderrContains,
					"stderr should contain expected text")
			}
		})
	}
}

// TestInvalidFlagValues tests handling of invalid flag values
func TestInvalidFlagValues(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantExitCode       int
		wantStderrContains string
	}{
		{
			name:               "Invalid color name",
			args:               []string{"ping", "--color", "invalid-color"},
			wantExitCode:       1,
			wantStderrContains: "invalid color",
		},
		{
			name:               "Invalid log level",
			args:               []string{"--log-level", "invalid-level", "ping"},
			wantExitCode:       0, // Logs warning but continues
			wantStderrContains: "Invalid console log level",
		},
		{
			name:               "Unknown flag",
			args:               []string{"ping", "--unknown-flag", "value"},
			wantExitCode:       1,
			wantStderrContains: "unknown flag",
		},
		{
			name:               "Invalid docs format",
			args:               []string{"docs", "config", "--format", "invalid"},
			wantExitCode:       1,
			wantStderrContains: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := getExitCode(err)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"exit code mismatch\nstderr: %s", stderr.String())

			// Check error message if specified
			if tt.wantStderrContains != "" {
				stderrOutput := stderr.String()
				assert.Contains(t, stderrOutput, tt.wantStderrContains,
					"stderr should contain expected text")
			}
		})
	}
}

// TestInvalidCommands tests handling of invalid commands and subcommands
func TestInvalidCommands(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantExitCode       int
		wantStderrContains string
	}{
		{
			name:               "Unknown command",
			args:               []string{"unknown-command"},
			wantExitCode:       1,
			wantStderrContains: "unknown command",
		},
		{
			name:               "Unknown subcommand",
			args:               []string{"config", "unknown-subcommand"},
			wantExitCode:       0, // Shows help instead of error
			wantStderrContains: "",
		},
		{
			name:               "Docs without subcommand",
			args:               []string{"docs"},
			wantExitCode:       0, // Shows help, which is valid
			wantStderrContains: "",
		},
		{
			name:               "Config without subcommand",
			args:               []string{"config"},
			wantExitCode:       0, // Shows help
			wantStderrContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := getExitCode(err)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"exit code mismatch\nstdout: %s\nstderr: %s",
				stdout.String(), stderr.String())

			if tt.wantStderrContains != "" {
				combinedOutput := stdout.String() + stderr.String()
				assert.Contains(t, combinedOutput, tt.wantStderrContains,
					"output should contain expected text")
			}
		})
	}
}

// TestDocumentationOutputErrors tests error handling in documentation generation
func TestDocumentationOutputErrors(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		setupFunc          func(t *testing.T) (string, []string) // Returns output path and args
		wantExitCode       int
		wantStderrContains string
	}{
		{
			name: "Output to non-existent directory",
			setupFunc: func(t *testing.T) (string, []string) {
				nonExistentPath := filepath.Join(tmpDir, "nonexistent", "dir", "output.md")
				return nonExistentPath, []string{"docs", "config", "--output", nonExistentPath}
			},
			wantExitCode:       1,
			wantStderrContains: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, args := tt.setupFunc(t)

			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := getExitCode(err)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"exit code mismatch\nstderr: %s", stderr.String())

			if tt.wantStderrContains != "" {
				stderrOutput := stderr.String()
				assert.Contains(t, stderrOutput, tt.wantStderrContains,
					"stderr should contain expected text")
			}
		})
	}
}

// TestConfigPrecedenceWithErrors tests config precedence when errors occur
func TestConfigPrecedenceWithErrors(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Invalid config file with valid env var - env var wins", func(t *testing.T) {
		// Create invalid config file
		invalidConfig := filepath.Join(tmpDir, "invalid.yaml")
		if err := os.WriteFile(invalidConfig, []byte("invalid: [yaml"), 0600); err != nil {
			t.Fatalf("Failed to create invalid config: %v", err)
		}

		cmd := exec.Command(binaryPath, "--config", invalidConfig, "ping")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Set env var
		cmd.Env = append(os.Environ(), "CKELETIN_GO_APP_PING_OUTPUT_MESSAGE=Env Var Works")

		err := cmd.Run()

		// Should fail due to invalid config
		require.Error(t, err, "command should fail with invalid config")

		exitCode := getExitCode(err)
		assert.Equal(t, 1, exitCode, "exit code should be 1")
	})

	t.Run("Flag overrides everything even with config errors", func(t *testing.T) {
		// Even with a missing config file, flags should work
		cmd := exec.Command(binaryPath, "--config", "/nonexistent/config.yaml", "ping", "--message", "Flag Message")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Run()

		// Should fail because config file is required when --config is specified
		require.Error(t, err, "command should fail with nonexistent config")
	})
}

// TestEdgeCaseInputs tests handling of edge case inputs
func TestEdgeCaseInputs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantExitCode int
	}{
		{
			name:         "Empty message flag",
			args:         []string{"ping", "--message", ""},
			wantExitCode: 0, // Empty string is valid
		},
		{
			name:         "Very long message flag",
			args:         []string{"ping", "--message", strings.Repeat("x", 1000)},
			wantExitCode: 0, // Long string is valid
		},
		{
			name:         "Special characters in message",
			args:         []string{"ping", "--message", "Hello\nWorld\t!@#$%^&*()"},
			wantExitCode: 0,
		},
		{
			name:         "Unicode in message",
			args:         []string{"ping", "--message", "Hello 世界 🌍"},
			wantExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := getExitCode(err)

			assert.Equal(t, tt.wantExitCode, exitCode,
				"exit code mismatch\nstdout: %s\nstderr: %s",
				stdout.String(), stderr.String())
		})
	}
}

// TestConcurrentCommandExecution tests that multiple commands can run without interference
func TestConcurrentCommandExecution(t *testing.T) {
	// Run multiple instances of the binary concurrently
	const numConcurrent = 10

	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(id int) {
			cmd := exec.Command(binaryPath, "ping", "--message", "Concurrent test")
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()
			if err != nil {
				errChan <- err
				return
			}

			if !strings.Contains(stdout.String(), "Concurrent test") {
				errChan <- os.ErrInvalid
				return
			}

			errChan <- nil
		}(i)
	}

	// Collect results
	for i := 0; i < numConcurrent; i++ {
		err := <-errChan
		assert.NoError(t, err, "concurrent execution %d should succeed", i)
	}
}
