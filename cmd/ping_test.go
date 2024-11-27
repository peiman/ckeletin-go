package cmd

import (
	"bytes"
	"errors"
	"testing"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestPingCommand(t *testing.T) {
	// Save original logger and restore after the test
	originalLogger := log.Logger
	defer func() { log.Logger = originalLogger }()
	log.Logger = zerolog.New(bytes.NewBuffer([]byte{})) // Disable output for test

	tests := []struct {
		name       string
		args       []string
		setup      func()
		uiRunner   *ui.MockUIRunner
		wantErr    bool
		wantOutput string
	}{
		{
			name: "Default",
			args: []string{},
			setup: func() {
				// No setup needed; use defaults
			},
			uiRunner:   &ui.MockUIRunner{},
			wantErr:    false,
			wantOutput: "Pong\n",
		},
		{
			name: "Custom Message and Color",
			args: []string{"--message", "Hello, Test!", "--color", "red"},
			setup: func() {
				// No setup needed; flags will override defaults
			},
			uiRunner:   &ui.MockUIRunner{},
			wantErr:    false,
			wantOutput: "Hello, Test!\n",
		},
		{
			name: "UI Enabled",
			args: []string{"--ui"},
			setup: func() {
				// No setup needed; --ui flag enables UI
			},
			uiRunner: &ui.MockUIRunner{},
			wantErr:  false,
		},
		{
			name: "UI Enabled with Error",
			args: []string{"--ui"},
			setup: func() {
				// No setup needed; --ui flag enables UI
			},
			uiRunner: &ui.MockUIRunner{
				ReturnError: errors.New("UI error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset Viper to clear any previous configurations
			viper.Reset()

			// Initialize the command, which sets defaults and configurations
			cmd := NewPingCommand(tt.uiRunner)

			// Apply any test-specific setup
			if tt.setup != nil {
				tt.setup()
			}

			// Set up command execution
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)
			cmd.SilenceUsage = true
			cmd.SilenceErrors = true

			// Execute the command
			err := cmd.Execute()

			// Check for expected errors
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check for expected output
			if gotOutput := buf.String(); gotOutput != tt.wantOutput {
				t.Errorf("Output = %q, want %q", gotOutput, tt.wantOutput)
			}

			// Verify UIRunner calls for UI-enabled scenarios
			uiFlag := viper.GetBool("app.ping.ui")
			if uiFlag {
				if tt.uiRunner.CalledWithMessage != viper.GetString("app.ping.output_message") {
					t.Errorf("UIRunner called with wrong message: got %q, want %q",
						tt.uiRunner.CalledWithMessage, viper.GetString("app.ping.output_message"))
				}
				if tt.uiRunner.CalledWithColor != viper.GetString("app.ping.output_color") {
					t.Errorf("UIRunner called with wrong color: got %q, want %q",
						tt.uiRunner.CalledWithColor, viper.GetString("app.ping.output_color"))
				}
			}
		})
	}
}
