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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var binaryPath string

// getExitCode extracts the exit code from a command execution error.
// Returns 0 if err is nil, the exit code if err is an ExitError, or -1 otherwise.
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode()
	}
	return -1
}

// TestMain builds the binary before running tests
func TestMain(m *testing.M) {
	// Build the binary with platform-specific name
	binaryName := "ckeletin-go-test"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}

	cmd := exec.Command("go", "build", "-o", binaryName, "../../main.go")
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to build test binary: %v\n", err)
		os.Exit(1)
	}
	binaryPath = "./" + binaryName

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
			exitCode := getExitCode(err)

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
			exitCode := getExitCode(err)

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

// TestConfigLoading tests configuration loading from files
func TestConfigLoading(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name               string
		configContent      string
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name: "Load config with custom message",
			configContent: `app:
  log_level: info
  ping:
    output_message: "Custom Config Message"
    output_color: "blue"
`,
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Custom Config Message",
		},
		{
			name: "Config file with defaults",
			configContent: `app:
  log_level: debug
`,
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Pong", // Default message
		},
		{
			name: "Config file with complex nested structure",
			configContent: `app:
  log_level: warn
  ping:
    output_message: "Nested Config Test"
    output_color: "green"
    ui: false
`,
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Nested Config Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config file
			configFile := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configFile, []byte(tt.configContent), 0600); err != nil {
				t.Fatalf("Failed to create config file: %v", err)
			}

			// Run command with config file
			args := append([]string{"--config", configFile}, tt.args...)
			cmd := exec.Command(binaryPath, args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			// Check exit code
			exitCode := getExitCode(err)

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nstdout: %s\nstderr: %s",
					exitCode, tt.wantExitCode, stdout.String(), stderr.String())
			}

			// Check output
			output := stdout.String()
			if tt.wantOutputContains != "" && !strings.Contains(output, tt.wantOutputContains) {
				t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContains, output)
			}

			// Cleanup for next test
			os.Remove(configFile)
		})
	}
}

// TestEnvironmentVariables tests configuration via environment variables
func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name               string
		envVars            map[string]string
		args               []string
		wantExitCode       int
		wantOutputContains string
	}{
		{
			name: "Set message via env var",
			envVars: map[string]string{
				"CKELETIN_GO_APP_PING_OUTPUT_MESSAGE": "Env Var Message",
			},
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Env Var Message",
		},
		{
			name: "Set log level via env var",
			envVars: map[string]string{
				"CKELETIN_GO_APP_LOG_LEVEL": "debug",
			},
			args:         []string{"ping"},
			wantExitCode: 0,
			// Debug messages should appear in stderr
		},
		{
			name: "Multiple env vars",
			envVars: map[string]string{
				"CKELETIN_GO_APP_PING_OUTPUT_MESSAGE": "Multi Env Test",
				"CKELETIN_GO_APP_PING_OUTPUT_COLOR":   "cyan",
			},
			args:               []string{"ping"},
			wantExitCode:       0,
			wantOutputContains: "Multi Env Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			// Set environment variables
			cmd.Env = os.Environ()
			for k, v := range tt.envVars {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			err := cmd.Run()

			// Check exit code
			exitCode := getExitCode(err)

			if exitCode != tt.wantExitCode {
				t.Errorf("Exit code = %d, want %d\nstdout: %s\nstderr: %s",
					exitCode, tt.wantExitCode, stdout.String(), stderr.String())
			}

			// Check output
			if tt.wantOutputContains != "" {
				output := stdout.String()
				if !strings.Contains(output, tt.wantOutputContains) {
					t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContains, output)
				}
			}
		})
	}
}

// TestConfigPrecedence tests the precedence of config sources (flags > env > config file > defaults)
func TestConfigPrecedence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a config file
	configContent := `app:
  log_level: info
  ping:
    output_message: "Config File Message"
    output_color: "blue"
`
	configFile := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	tests := []struct {
		name               string
		envVars            map[string]string
		args               []string
		wantOutputContains string
	}{
		{
			name:               "Config file only",
			envVars:            map[string]string{},
			args:               []string{"--config", configFile, "ping"},
			wantOutputContains: "Config File Message",
		},
		{
			name: "Env var overrides config file",
			envVars: map[string]string{
				"CKELETIN_GO_APP_PING_OUTPUT_MESSAGE": "Env Var Message",
			},
			args:               []string{"--config", configFile, "ping"},
			wantOutputContains: "Env Var Message",
		},
		{
			name: "CLI flag overrides env var and config file",
			envVars: map[string]string{
				"CKELETIN_GO_APP_PING_OUTPUT_MESSAGE": "Env Var Message",
			},
			args:               []string{"--config", configFile, "ping", "--message", "CLI Flag Message"},
			wantOutputContains: "CLI Flag Message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			// Set environment variables
			cmd.Env = os.Environ()
			for k, v := range tt.envVars {
				cmd.Env = append(cmd.Env, k+"="+v)
			}

			if err := cmd.Run(); err != nil {
				t.Fatalf("Command failed: %v", err)
			}

			output := stdout.String()
			if !strings.Contains(output, tt.wantOutputContains) {
				t.Errorf("Output doesn't contain %q\nGot: %s", tt.wantOutputContains, output)
			}
		})
	}
}

// TestMultiCommandWorkflow tests complex multi-command workflows
func TestMultiCommandWorkflow(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Create config, validate it, then use it", func(t *testing.T) {
		// Step 1: Create a config file
		configContent := `app:
  log_level: debug
  ping:
    output_message: "Workflow Test"
    output_color: "green"
`
		configFile := filepath.Join(tmpDir, "workflow-config.yaml")
		if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// Step 2: Validate the config
		validateCmd := exec.Command(binaryPath, "config", "validate", "--file", configFile)
		var validateStdout bytes.Buffer
		validateCmd.Stdout = &validateStdout

		if err := validateCmd.Run(); err != nil {
			t.Fatalf("Config validation failed: %v\nOutput: %s", err, validateStdout.String())
		}

		if !strings.Contains(validateStdout.String(), "valid") {
			t.Errorf("Validation output doesn't confirm config is valid: %s", validateStdout.String())
		}

		// Step 3: Use the validated config
		pingCmd := exec.Command(binaryPath, "--config", configFile, "ping")
		var pingStdout bytes.Buffer
		pingCmd.Stdout = &pingStdout

		if err := pingCmd.Run(); err != nil {
			t.Fatalf("Ping command failed: %v", err)
		}

		if !strings.Contains(pingStdout.String(), "Workflow Test") {
			t.Errorf("Ping output doesn't contain expected message: %s", pingStdout.String())
		}
	})

	t.Run("Generate docs then validate config", func(t *testing.T) {
		// Step 1: Generate documentation
		docsCmd := exec.Command(binaryPath, "docs", "config")
		var docsStdout bytes.Buffer
		docsCmd.Stdout = &docsStdout

		if err := docsCmd.Run(); err != nil {
			t.Fatalf("Docs generation failed: %v", err)
		}

		if !strings.Contains(docsStdout.String(), "Configuration") {
			t.Errorf("Docs don't contain configuration info: %s", docsStdout.String())
		}

		// Step 2: Create a config based on docs
		configContent := `app:
  log_level: info
`
		configFile := filepath.Join(tmpDir, "docs-based-config.yaml")
		if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		// Step 3: Validate the config
		validateCmd := exec.Command(binaryPath, "config", "validate", "--file", configFile)
		var validateStdout bytes.Buffer
		validateCmd.Stdout = &validateStdout

		if err := validateCmd.Run(); err != nil {
			t.Fatalf("Config validation failed: %v\nOutput: %s", err, validateStdout.String())
		}
	})
}
