// cmd/ping_test.go

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type mockUIRunner struct {
	CalledWithMessage string
	CalledWithColor   string
	ReturnError       error
}

func (m *mockUIRunner) RunUI(message, col string) error {
	m.CalledWithMessage = message
	m.CalledWithColor = col
	return m.ReturnError
}

type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (int, error) {
	return 0, fmt.Errorf("write error")
}

// TestConfigDefaultsForPing ensures the default values for ping command are properly set
func TestConfigDefaultsForPing(t *testing.T) {
	// Reset viper for a clean test
	viper.Reset()

	// Apply defaults from registry
	config.SetDefaults()

	// Check that defaults are set correctly
	if viper.GetString("app.ping.output_message") != "Pong" {
		t.Errorf("Expected default message to be 'Pong', got '%s'", viper.GetString("app.ping.output_message"))
	}

	if viper.GetString("app.ping.output_color") != "white" {
		t.Errorf("Expected default color to be 'white', got '%s'", viper.GetString("app.ping.output_color"))
	}

	if viper.GetBool("app.ping.ui") != false {
		t.Errorf("Expected default ui to be false, got %v", viper.GetBool("app.ping.ui"))
	}
}

// TestGetConfigValue tests our new helper function for retrieving config values
func TestGetConfigValue(t *testing.T) {
	// Setup test config in viper
	viper.Reset()
	viper.Set("app.ping.output_message", "ConfigMessage")
	viper.Set("app.ping.output_color", "blue")
	viper.Set("app.ping.ui", true)

	// Create a command for testing
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("message", "", "Custom output message")
	cmd.Flags().String("color", "", "Output color")
	cmd.Flags().Bool("ui", false, "Enable UI")

	// Test string value without flag set (should use viper value)
	message := getConfigValue[string](cmd, "message", "app.ping.output_message")
	if message != "ConfigMessage" {
		t.Errorf("Expected message to be 'ConfigMessage', got '%s'", message)
	}

	// Test when flag is set (should override viper value)
	if err := cmd.Flags().Set("message", "FlagMessage"); err != nil {
		t.Fatalf("Failed to set message flag: %v", err)
	}

	message = getConfigValue[string](cmd, "message", "app.ping.output_message")
	if message != "FlagMessage" {
		t.Errorf("Expected message to be 'FlagMessage' when flag is set, got '%s'", message)
	}

	// Test bool value without flag set (should use viper value)
	uiFlag := getConfigValue[bool](cmd, "ui", "app.ping.ui")
	if uiFlag != true {
		t.Errorf("Expected ui to be true, got %v", uiFlag)
	}

	// Test when flag is set (should override viper value)
	if err := cmd.Flags().Set("ui", "false"); err != nil {
		t.Fatalf("Failed to set ui flag: %v", err)
	}

	uiFlag = getConfigValue[bool](cmd, "ui", "app.ping.ui")
	if uiFlag != false {
		t.Errorf("Expected ui to be false when flag is set, got %v", uiFlag)
	}
}

// TestPingCommandFlags tests that flags are correctly processed including from config file
func TestPingCommandFlags(t *testing.T) {
	// Setup test config in viper
	viper.Reset()
	viper.Set("app.ping.output_message", "ConfigMessage")
	viper.Set("app.ping.output_color", "blue")
	viper.Set("app.ping.ui", true)

	// Create a command for testing
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("message", "", "Custom output message")
	cmd.Flags().String("color", "", "Output color")
	cmd.Flags().Bool("ui", false, "Enable UI")

	// Test when flag is not set (should use viper value)
	message := viper.GetString("app.ping.output_message")
	if cmd.Flags().Changed("message") {
		message, _ = cmd.Flags().GetString("message")
	}

	if message != "ConfigMessage" {
		t.Errorf("Expected message to be 'ConfigMessage', got '%s'", message)
	}

	// Test when flag is set (should override viper value)
	if err := cmd.Flags().Set("message", "FlagMessage"); err != nil {
		t.Fatalf("Failed to set message flag: %v", err)
	}

	message = viper.GetString("app.ping.output_message")
	if cmd.Flags().Changed("message") {
		message, _ = cmd.Flags().GetString("message")
	}

	if message != "FlagMessage" {
		t.Errorf("Expected message to be 'FlagMessage' when flag is set, got '%s'", message)
	}

	// Test the same for color
	colorStr := viper.GetString("app.ping.output_color")
	if cmd.Flags().Changed("color") {
		colorStr, _ = cmd.Flags().GetString("color")
	}

	if colorStr != "blue" {
		t.Errorf("Expected color to be 'blue', got '%s'", colorStr)
	}

	if err := cmd.Flags().Set("color", "red"); err != nil {
		t.Fatalf("Failed to set color flag: %v", err)
	}

	colorStr = viper.GetString("app.ping.output_color")
	if cmd.Flags().Changed("color") {
		colorStr, _ = cmd.Flags().GetString("color")
	}

	if colorStr != "red" {
		t.Errorf("Expected color to be 'red' when flag is set, got '%s'", colorStr)
	}

	// Test the same for ui flag
	uiFlag := viper.GetBool("app.ping.ui")
	if cmd.Flags().Changed("ui") {
		uiFlag, _ = cmd.Flags().GetBool("ui")
	}

	if uiFlag != true {
		t.Errorf("Expected ui to be true, got %v", uiFlag)
	}

	if err := cmd.Flags().Set("ui", "false"); err != nil {
		t.Fatalf("Failed to set ui flag: %v", err)
	}

	uiFlag = viper.GetBool("app.ping.ui")
	if cmd.Flags().Changed("ui") {
		uiFlag, _ = cmd.Flags().GetBool("ui")
	}

	if uiFlag != false {
		t.Errorf("Expected ui to be false when flag is set, got %v", uiFlag)
	}
}

func TestPingCommand(t *testing.T) {
	// Setup debug logging for tests
	logBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(logBuf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	originalRunner := pingRunner
	defer func() { pingRunner = originalRunner }()

	tests := []struct {
		name            string
		testFixturePath string   // Path to test fixture
		args            []string // CLI args
		uiRunner        *mockUIRunner
		wantErr         bool
		wantOutput      string
		writer          io.Writer
	}{
		{
			name:            "Default Configuration",
			testFixturePath: "../testdata/config.yaml",
			args:            []string{},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "Config Message\n", // From testdata/config.yaml
			writer:          &bytes.Buffer{},
		},
		{
			name:            "JSON Configuration",
			testFixturePath: "../testdata/config.json",
			args:            []string{},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "JSON Config Message\n", // From testdata/config.json
			writer:          &bytes.Buffer{},
		},
		{
			name:            "CLI Args Override Configuration",
			testFixturePath: "../testdata/config.yaml",
			args:            []string{"--message", "CLI Message", "--color", "cyan"},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "CLI Message\n",
			writer:          &bytes.Buffer{},
		},
		{
			name:            "Partial Configuration",
			testFixturePath: "../testdata/partial_config.yaml",
			args:            []string{"--color", "white"},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "Partial Config Message\n", // From testdata/partial_config.yaml
			writer:          &bytes.Buffer{},
		},
		{
			name:            "UI Enabled",
			testFixturePath: "../testdata/ui_test_config.yaml",
			args:            []string{},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "", // No standard output when UI is enabled
			writer:          &bytes.Buffer{},
		},
		{
			name:            "UI Error",
			testFixturePath: "../testdata/ui_test_config.yaml",
			args:            []string{},
			uiRunner:        &mockUIRunner{ReturnError: fmt.Errorf("ui error")},
			wantErr:         true,
			wantOutput:      "",
			writer:          &bytes.Buffer{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Reset viper for each test
			viper.Reset()

			// Load test fixture
			viper.SetConfigFile(tt.testFixturePath)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("Failed to load test fixture %s: %v", tt.testFixturePath, err)
			}

			// Create command and register flags
			cmd := &cobra.Command{Use: "ping"}
			cmd.Flags().String("message", "", "Custom output message")
			cmd.Flags().String("color", "", "Output color")
			cmd.Flags().Bool("ui", false, "Enable UI")
			cmd.SetOut(tt.writer)

			// Parse args
			cmd.SetArgs(tt.args)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			// Setup mock UI runner
			pingRunner = tt.uiRunner

			// Prepare output buffer
			outBuf, isBuffer := tt.writer.(*bytes.Buffer)

			// Clear the output buffer to ensure we're only capturing the current test output
			if isBuffer {
				outBuf.Reset()
			}

			// EXECUTION PHASE
			err := runPing(cmd, []string{})

			// ASSERTION PHASE
			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("runPing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check UI runner was called with correct parameters for UI tests
			if viper.GetBool("app.ping.ui") {
				expectedMessage := viper.GetString("app.ping.output_message")
				// If message flag was set, it should override viper config
				if cmd.Flags().Changed("message") {
					expectedMessage, _ = cmd.Flags().GetString("message")
				}

				expectedColor := viper.GetString("app.ping.output_color")
				// If color flag was set, it should override viper config
				if cmd.Flags().Changed("color") {
					expectedColor, _ = cmd.Flags().GetString("color")
				}

				if tt.uiRunner.CalledWithMessage != expectedMessage {
					t.Errorf("UI runner called with wrong message, got: %s, want: %s",
						tt.uiRunner.CalledWithMessage, expectedMessage)
				}

				if tt.uiRunner.CalledWithColor != expectedColor {
					t.Errorf("UI runner called with wrong color, got: %s, want: %s",
						tt.uiRunner.CalledWithColor, expectedColor)
				}
			}

			// Check output
			if isBuffer && !tt.wantErr && !viper.GetBool("app.ping.ui") {
				got := outBuf.String()
				if got != tt.wantOutput {
					t.Errorf("runPing() output = %q, want %q", got, tt.wantOutput)
				}
			}
		})
	}
}

// TestPingCommand_WriteError tests the error handling when writing fails
func TestPingCommand_WriteError(t *testing.T) {
	// Setup
	writer := &errorWriter{}
	cmd := &cobra.Command{Use: "ping"}
	cmd.SetOut(writer)

	// Make sure UI is disabled and color is set to a valid value
	viper.Reset()
	viper.Set("app.ping.ui", false)
	viper.Set("app.ping.output_message", "Test Message")
	viper.Set("app.ping.output_color", "white")

	// Execute
	err := runPing(cmd, []string{})

	// Assert
	if err == nil {
		t.Error("runPing() expected error, got nil")
		return
	}

	if !strings.Contains(err.Error(), "failed to print colored message") {
		t.Errorf("runPing() error = %v, expected to contain 'failed to print colored message'", err)
	}
}
