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
	// SETUP PHASE: Reset viper and apply config defaults
	viper.Reset()
	config.SetDefaults()

	// EXECUTION PHASE: Build config from defaults
	cfg := NewPingConfig(&cobra.Command{Use: "ping"})

	// ASSERTION PHASE: Check that defaults are set correctly
	if cfg.Message != "Pong" {
		t.Errorf("Expected default message to be 'Pong', got '%s'", cfg.Message)
	}
	if cfg.Color != "white" {
		t.Errorf("Expected default color to be 'white', got '%s'", cfg.Color)
	}
	if cfg.UI != false {
		t.Errorf("Expected default ui to be false, got %v", cfg.UI)
	}
}

// TestPingConfigOptions tests the functional options for PingConfig
func TestPingConfigOptions(t *testing.T) {
	// SETUP PHASE: No setup needed for direct options

	// EXECUTION PHASE: Build config with options
	cfg := NewPingConfig(&cobra.Command{Use: "ping"}, WithPingMessage("TestMsg"), WithPingColor("red"), WithPingUI(true))

	// ASSERTION PHASE: Check that options override defaults
	if cfg.Message != "TestMsg" {
		t.Errorf("Expected message to be 'TestMsg', got '%s'", cfg.Message)
	}
	if cfg.Color != "red" {
		t.Errorf("Expected color to be 'red', got '%s'", cfg.Color)
	}
	if cfg.UI != true {
		t.Errorf("Expected UI to be true, got %v", cfg.UI)
	}
}

func TestPingCommandFlags(t *testing.T) {
	// SETUP PHASE: Set viper config and create command
	viper.Reset()
	viper.Set("app.ping.output_message", "ConfigMessage")
	viper.Set("app.ping.output_color", "blue")
	viper.Set("app.ping.ui", true)

	cmd := &cobra.Command{Use: "ping"}
	cmd.Flags().String("message", "", "Custom output message")
	cmd.Flags().String("color", "", "Output color")
	cmd.Flags().Bool("ui", false, "Enable UI")

	// EXECUTION PHASE: No flag set, should use viper value
	cfg := NewPingConfig(cmd)

	// ASSERTION PHASE: Check viper values
	if cfg.Message != "ConfigMessage" {
		t.Errorf("Expected message to be 'ConfigMessage', got '%s'", cfg.Message)
	}
	if cfg.Color != "blue" {
		t.Errorf("Expected color to be 'blue', got '%s'", cfg.Color)
	}
	if cfg.UI != true {
		t.Errorf("Expected UI to be true, got %v", cfg.UI)
	}

	// EXECUTION PHASE: Set flags, should override viper
	if err := cmd.Flags().Set("message", "FlagMessage"); err != nil {
		t.Fatalf("Failed to set message flag: %v", err)
	}
	if err := cmd.Flags().Set("color", "red"); err != nil {
		t.Fatalf("Failed to set color flag: %v", err)
	}
	if err := cmd.Flags().Set("ui", "false"); err != nil {
		t.Fatalf("Failed to set ui flag: %v", err)
	}
	cfg = NewPingConfig(cmd)

	// ASSERTION PHASE: Check flag values
	if cfg.Message != "FlagMessage" {
		t.Errorf("Expected message to be 'FlagMessage', got '%s'", cfg.Message)
	}
	if cfg.Color != "red" {
		t.Errorf("Expected color to be 'red', got '%s'", cfg.Color)
	}
	if cfg.UI != false {
		t.Errorf("Expected UI to be false, got %v", cfg.UI)
	}
}

func TestPingCommand(t *testing.T) {
	// SETUP PHASE: Setup debug logging and test table
	logBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(logBuf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	originalRunner := pingRunner
	defer func() { pingRunner = originalRunner }()

	tests := []struct {
		name            string
		testFixturePath string
		args            []string
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
			wantOutput:      "Config Message\n",
			writer:          &bytes.Buffer{},
		},
		{
			name:            "JSON Configuration",
			testFixturePath: "../testdata/config.json",
			args:            []string{},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "JSON Config Message\n",
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
			wantOutput:      "Partial Config Message\n",
			writer:          &bytes.Buffer{},
		},
		{
			name:            "UI Enabled",
			testFixturePath: "../testdata/ui_test_config.yaml",
			args:            []string{},
			uiRunner:        &mockUIRunner{},
			wantErr:         false,
			wantOutput:      "",
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
			// SETUP PHASE: Reset viper, load fixture, and setup command
			viper.Reset()
			viper.SetConfigFile(tt.testFixturePath)
			if err := viper.ReadInConfig(); err != nil {
				t.Fatalf("Failed to load test fixture %s: %v", tt.testFixturePath, err)
			}

			cmd := &cobra.Command{Use: "ping"}
			cmd.Flags().String("message", "", "Custom output message")
			cmd.Flags().String("color", "", "Output color")
			cmd.Flags().Bool("ui", false, "Enable UI")
			cmd.SetOut(tt.writer)
			cmd.SetArgs(tt.args)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			pingRunner = tt.uiRunner

			outBuf, isBuffer := tt.writer.(*bytes.Buffer)
			if isBuffer {
				outBuf.Reset()
			}

			// EXECUTION PHASE: Run the ping command
			err := runPing(cmd, []string{})

			// ASSERTION PHASE: Check error and output
			if (err != nil) != tt.wantErr {
				t.Errorf("runPing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if viper.GetBool("app.ping.ui") {
				expectedMessage := viper.GetString("app.ping.output_message")
				if cmd.Flags().Changed("message") {
					expectedMessage, _ = cmd.Flags().GetString("message")
				}
				expectedColor := viper.GetString("app.ping.output_color")
				if cmd.Flags().Changed("color") {
					expectedColor, _ = cmd.Flags().GetString("color")
				}
				if tt.uiRunner.CalledWithMessage != expectedMessage {
					t.Errorf("UI runner called with wrong message, got: %s, want: %s", tt.uiRunner.CalledWithMessage, expectedMessage)
				}
				if tt.uiRunner.CalledWithColor != expectedColor {
					t.Errorf("UI runner called with wrong color, got: %s, want: %s", tt.uiRunner.CalledWithColor, expectedColor)
				}
			}

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
	// SETUP PHASE: Setup error writer and viper config
	writer := &errorWriter{}
	cmd := &cobra.Command{Use: "ping"}
	cmd.SetOut(writer)
	viper.Reset()
	viper.Set("app.ping.ui", false)
	viper.Set("app.ping.output_message", "Test Message")
	viper.Set("app.ping.output_color", "white")

	// EXECUTION PHASE: Run the ping command
	err := runPing(cmd, []string{})

	// ASSERTION PHASE: Check for expected error
	if err == nil {
		t.Error("runPing() expected error, got nil")
		return
	}
	if !strings.Contains(err.Error(), "failed to print colored message") {
		t.Errorf("runPing() error = %v, expected to contain 'failed to print colored message'", err)
	}
}
