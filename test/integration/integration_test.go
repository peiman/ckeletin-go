// test/integration/integration_test.go
//
// Integration tests for full command execution
//
// These tests execute actual commands end-to-end to verify:
// - Complete command workflows
// - Flag parsing and precedence
// - Configuration loading
// - Output generation
// - Exit codes

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

// TestMain builds the binary before running tests
func TestMain(m *testing.M) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "ckeletin-go-test", "../../main.go")
	if err := cmd.Run(); err != nil {
		panic("Failed to build binary: " + err.Error())
	}
	binaryPath = "./ckeletin-go-test"

	// Run tests
	code := m.Run()

	// Cleanup
	os.Remove(binaryPath)

	os.Exit(code)
}

func TestPingCommand(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name:               "Default ping",
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Pong",
		},
		{
			name:               "Ping with custom message",
			args:               []string{"ping", "--message", "Hello World"},
			wantExitCode:       0,
			wantOutputContains: "Hello World",
		},
		{
			name:               "Ping with color flag",
			args:               []string{"ping", "--color", "green"},
			wantExitCode:       0,
			wantOutputContains: "", // Output varies by terminal
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
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nstdout: %s\nstderr: %s",
					exitCode, tt.wantExitCode, stdout.String(), stderr.String())
			}

			// Check output if specified
			if tt.wantOutputContains != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContains, output)
				}
			}
		})
	}
}

func TestConfigValidateCommand(t *testing.T) {
	// Create temp directory for test config files
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		configContent      string
		configPerms        os.FileMode
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name: "Valid config",
			configContent: `app:
  log_level: debug
`,
			configPerms:        0600,
			wantExitCode:       0,
			wantOutputContains: "Configuration is valid",
		},
		{
			name: "Invalid YAML",
			configContent: `app:
  invalid: [unclosed
`,
			configPerms:        0600,
			wantExitCode:       1,
			wantOutputContains: "Configuration is invalid",
		},
		{
			name: "Unknown keys (warning)",
			configContent: `app:
  log_level: info
  unknown_key: value
`,
			configPerms:        0600,
			wantExitCode:       1, // Warnings cause exit 1
			wantOutputContains: "Unknown configuration key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test config file
			configFile := filepath.Join(tmpDir, tt.name+".yaml")
			if err := os.WriteFile(configFile, []byte(tt.configContent), tt.configPerms); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Build command
			args := append([]string{"config", "validate", "--file", configFile}, tt.args...)
			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nstdout: %s\nstderr: %s",
					exitCode, tt.wantExitCode, stdout.String(), stderr.String())
			}

			// Check output
			combinedOutput := stdout.String() + stderr.String()
			if tt.wantOutputContains != "" && !strings.Contains(combinedOutput, tt.wantOutputContains) {
				t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContains, combinedOutput)
			}
		})
	}
}

func TestDocsCommand(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name:               "Generate markdown docs",
			args:               []string{"docs", "config"},
			wantExitCode:       0,
			wantOutputContains: "Configuration Options",
		},
		{
			name:               "Generate YAML docs",
			args:               []string{"docs", "config", "--format", "yaml"},
			wantExitCode:       0,
			wantOutputContains: "app:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nstderr: %s",
					exitCode, tt.wantExitCode, stderr.String())
			}

			if tt.wantOutputContains != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("Output doesn't contain %q", tt.wantOutputContains)
				}
			}
		})
	}
}

func TestHelpCommand(t *testing.T) {
	tests := []struct {
		name               string
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name:               "Root help",
			args:               []string{"--help"},
			wantExitCode:       0,
			wantOutputContains: "Available Commands",
		},
		{
			name:               "Ping help",
			args:               []string{"ping", "--help"},
			wantExitCode:       0,
			wantOutputContains: "ping",
		},
		{
			name:               "Config validate help",
			args:               []string{"config", "validate", "--help"},
			wantExitCode:       0,
			wantOutputContains: "Validate a configuration file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				}
			}

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d", exitCode, tt.wantExitCode)
			}

			if tt.wantOutputContains != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("Output doesn't contain %q", tt.wantOutputContains)
				}
			}
		})
	}
}

func TestVersionFlag(t *testing.T) {
	cmd := exec.Command(binaryPath, "--version")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Version command failed: %v", err)
	}

	output := stdout.String()
	// Version output should contain version info
	if !strings.Contains(output, "version") && !strings.Contains(output, "dev") {
		t.Errorf("Version output unexpected: %s", output)
	}
}
