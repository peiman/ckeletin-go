// cmd/ping_test.go

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// mockUIRunner is a mock implementation of ui.UIRunner for testing
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

// TestPingCommand tests the ping command integration with configuration
func TestPingCommand(t *testing.T) {
	// SETUP PHASE: Setup debug logging
	logBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(logBuf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	tests := []struct {
		name            string
		testFixturePath string
		args            []string
		wantErr         bool
		wantOutput      string
		writer          *bytes.Buffer
		mockRunner      *mockUIRunner
	}{
		{
			name:            "Default Configuration",
			testFixturePath: "../testdata/config.yaml",
			args:            []string{},
			wantErr:         false,
			wantOutput:      "",
			writer:          &bytes.Buffer{},
			mockRunner:      &mockUIRunner{},
		},
		{
			name:            "JSON Configuration",
			testFixturePath: "../testdata/config.json",
			args:            []string{},
			wantErr:         false,
			wantOutput:      "JSON Config Message\n",
			writer:          &bytes.Buffer{},
			mockRunner:      &mockUIRunner{},
		},
		{
			name:            "CLI Args Override Configuration",
			testFixturePath: "../testdata/config.yaml",
			args:            []string{"--message", "CLI Message", "--color", "cyan"},
			wantErr:         false,
			wantOutput:      "",
			writer:          &bytes.Buffer{},
			mockRunner:      &mockUIRunner{},
		},
		{
			name:            "Partial Configuration",
			testFixturePath: "../testdata/partial_config.yaml",
			args:            []string{"--color", "white"},
			wantErr:         false,
			wantOutput:      "Partial Config Message\n",
			writer:          &bytes.Buffer{},
			mockRunner:      &mockUIRunner{},
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

			// Create a new command instance for each test to avoid state pollution
			cmd := &cobra.Command{
				Use: "ping",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Use the injected mock runner for this test
					return runPingWithUIRunner(cmd, args, tt.mockRunner)
				},
			}

			// Register flags
			if err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
				"app.ping.output_message": "message",
				"app.ping.output_color":   "color",
				"app.ping.ui":             "ui",
			}); err != nil {
				t.Fatalf("Failed to register flags: %v", err)
			}

			cmd.SetOut(tt.writer)
			cmd.SetArgs(tt.args)
			if err := cmd.ParseFlags(tt.args); err != nil {
				t.Fatalf("Failed to parse flags: %v", err)
			}

			tt.writer.Reset()

			// EXECUTION PHASE: Run the ping command
			err := runPingWithUIRunner(cmd, []string{}, tt.mockRunner)

			// ASSERTION PHASE: Check error and output
			if (err != nil) != tt.wantErr {
				t.Errorf("runPing() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !viper.GetBool("app.ping.ui") {
				got := tt.writer.String()
				if got != tt.wantOutput {
					t.Errorf("runPing() output = %q, want %q", got, tt.wantOutput)
				}
			}
		})
	}
}

// TestPingCommandFlags tests flag precedence over configuration
func TestPingCommandFlags(t *testing.T) {
	// SETUP PHASE: Set viper config
	viper.Reset()
	viper.Set("app.ping.output_message", "ConfigMessage")
	viper.Set("app.ping.output_color", "blue")
	viper.Set("app.ping.ui", false)

	// Create mock UI runner
	mockRunner := &mockUIRunner{}

	// Create command
	cmd := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPingWithUIRunner(cmd, args, mockRunner)
		},
	}

	// Register flags
	if err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	}); err != nil {
		t.Fatalf("Failed to register flags: %v", err)
	}

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	// EXECUTION PHASE: No flags set, should use viper values
	err := runPingWithUIRunner(cmd, []string{}, mockRunner)
	if err != nil {
		t.Fatalf("runPing() failed: %v", err)
	}

	// ASSERTION PHASE: Check that viper values were used
	got := outBuf.String()
	if !strings.Contains(got, "ConfigMessage") {
		t.Errorf("Expected output to contain 'ConfigMessage', got %q", got)
	}

	// EXECUTION PHASE: Set flags, should override viper
	if err := cmd.Flags().Set("message", "FlagMessage"); err != nil {
		t.Fatalf("Failed to set message flag: %v", err)
	}
	if err := cmd.Flags().Set("color", "red"); err != nil {
		t.Fatalf("Failed to set color flag: %v", err)
	}

	outBuf.Reset()
	err = runPingWithUIRunner(cmd, []string{}, mockRunner)
	if err != nil {
		t.Fatalf("runPing() with flags failed: %v", err)
	}

	// ASSERTION PHASE: Check that flag values were used
	got = outBuf.String()
	if !strings.Contains(got, "FlagMessage") {
		t.Errorf("Expected output to contain 'FlagMessage', got %q", got)
	}
}

// TestPingConfigDefaults ensures the default values from config registry are used
func TestPingConfigDefaults(t *testing.T) {
	// SETUP PHASE: Reset viper and apply config defaults
	viper.Reset()
	config.SetDefaults()

	// Create mock UI runner
	mockRunner := &mockUIRunner{}

	// Create command
	cmd := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPingWithUIRunner(cmd, args, mockRunner)
		},
	}

	// Register flags
	if err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	}); err != nil {
		t.Fatalf("Failed to register flags: %v", err)
	}

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	// EXECUTION PHASE: Run with defaults
	err := runPingWithUIRunner(cmd, []string{}, mockRunner)

	// ASSERTION PHASE: Check that defaults were used
	if err != nil {
		t.Fatalf("runPing() failed: %v", err)
	}

	got := outBuf.String()
	// Default message is "Pong" as defined in ping_options.go
	if !strings.Contains(got, "Pong") {
		t.Errorf("Expected output to contain default message 'Pong', got %q", got)
	}
}
