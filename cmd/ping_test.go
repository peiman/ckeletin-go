// cmd/ping_test.go

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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

func (e errorWriter) Write(p []byte) (int, error) {
	log.Debug().Msg("errorWriter.Write called - generating error")
	return 0, fmt.Errorf("write error")
}

// setupTestViper initializes a clean viper instance for testing
func setupTestViper(ui bool, message, color string) {
	viper.Reset()

	// Instead of calling viper.SetDefault directly, set values directly
	// for testing purposes only
	viper.Set("app.ping.output_message", message)
	viper.Set("app.ping.output_color", color)
	viper.Set("app.ping.ui", ui)
}

// TestInitPingConfig ensures the default values are properly set
func TestInitPingConfig(t *testing.T) {
	// Reset viper for a clean test
	viper.Reset()

	// Apply defaults from registry
	config.SetDefaults()

	// Call the function
	initPingConfig()

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
		name         string
		args         []string
		uiRunner     *mockUIRunner
		wantErr      bool
		wantOutput   string
		mockPrintErr bool
		writer       io.Writer
		viperUI      bool
		viperMsg     string
		viperColor   string
	}{
		{
			name:       "Default",
			args:       []string{},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "Pong\n",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:       "Custom Message and Color",
			args:       []string{"--message", "Hello, Test!", "--color", "red"},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "Hello, Test!\n",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:       "UI Enabled",
			args:       []string{"--ui"},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:       "UI Enabled with Error",
			args:       []string{"--ui"},
			uiRunner:   &mockUIRunner{ReturnError: errors.New("UI error")},
			wantErr:    true,
			wantOutput: "",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:         "PrintColoredMessage Error",
			args:         []string{"--message", "Should fail print"},
			uiRunner:     &mockUIRunner{},
			wantErr:      true,
			wantOutput:   "",
			mockPrintErr: true,
			writer:       &errorWriter{},
			viperUI:      false,
			viperMsg:     "Should fail print",
			viperColor:   "white",
		},
		{
			name:       "Empty Message",
			args:       []string{"--message", ""},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "\n",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:       "Invalid Color",
			args:       []string{"--color", "invalidcolor"},
			uiRunner:   &mockUIRunner{},
			wantErr:    true,
			wantOutput: "",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
		{
			name:       "Special Characters",
			args:       []string{"--message", "Hello世界!@#$%^&*()"},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "Hello世界!@#$%^&*()\n",
			writer:     &bytes.Buffer{},
			viperUI:    false,
			viperMsg:   "Pong",
			viperColor: "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logBuf.Reset() // Clear the log buffer for each test
			log.Debug().Str("test_case", tt.name).Msg("Starting test case")

			// Set up a clean viper instance for this test
			setupTestViper(tt.viperUI, tt.viperMsg, tt.viperColor)

			pingRunner = tt.uiRunner

			// Create a new root command for each test
			RootCmd = &cobra.Command{Use: binaryName}
			pingCmd = &cobra.Command{
				Use:   "ping",
				Short: "Responds with a pong",
				RunE:  runPing,
			}

			// Set up the command flags
			pingCmd.Flags().String("message", "", "Custom output message")
			pingCmd.Flags().String("color", "", "Output color")
			pingCmd.Flags().Bool("ui", false, "Enable UI")

			RootCmd.AddCommand(pingCmd)
			RootCmd.SetArgs(append([]string{"ping"}, tt.args...))
			RootCmd.SetOut(tt.writer)
			RootCmd.SetErr(tt.writer)
			RootCmd.SilenceUsage = true
			RootCmd.SilenceErrors = true

			log.Debug().
				Bool("viper_ui", viper.GetBool("app.ping.ui")).
				Str("viper_msg", viper.GetString("app.ping.output_message")).
				Str("viper_color", viper.GetString("app.ping.output_color")).
				Msg("Viper config before execution")

			err := RootCmd.Execute()

			log.Debug().
				Err(err).
				Bool("expected_error", tt.wantErr).
				Bool("got_error", err != nil).
				Msg("Command execution completed")

			if (err != nil) != tt.wantErr {
				t.Logf("Debug logs:\n%s", logBuf.String())
				t.Errorf("%s: Execute() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			}

			if !tt.mockPrintErr {
				if buf, ok := tt.writer.(*bytes.Buffer); ok {
					gotOutput := buf.String()
					if gotOutput != tt.wantOutput {
						t.Errorf("%s: Output = %q, want %q", tt.name, gotOutput, tt.wantOutput)
					}
				}
			}
		})
	}
}
