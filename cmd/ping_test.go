package cmd

import (
	"bytes"
	"errors"
	"testing"

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

func TestPingCommand(t *testing.T) {
	originalLogger := log.Logger
	defer func() { log.Logger = originalLogger }()
	log.Logger = zerolog.New(bytes.NewBuffer([]byte{}))

	tests := []struct {
		name       string
		args       []string
		setup      func()
		uiRunner   *mockUIRunner
		wantErr    bool
		wantOutput string
	}{
		{
			name:       "Default",
			args:       []string{},
			setup:      func() {},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "Pong\n",
		},
		{
			name:       "Custom Message and Color",
			args:       []string{"--message", "Hello, Test!", "--color", "red"},
			setup:      func() {},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "Hello, Test!\n",
		},
		{
			name:       "UI Enabled",
			args:       []string{"--ui"},
			setup:      func() {},
			uiRunner:   &mockUIRunner{},
			wantErr:    false,
			wantOutput: "", // UI mode no direct output
		},
		{
			name:  "UI Enabled with Error",
			args:  []string{"--ui"},
			setup: func() {},
			uiRunner: &mockUIRunner{
				ReturnError: errors.New("UI error"),
			},
			wantErr:    true,
			wantOutput: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Don't reset viper, to keep flag bindings
			if tt.setup != nil {
				tt.setup()
			}

			// Inject mock UI runner
			originalRunner := pingRunner
			pingRunner = tt.uiRunner
			defer func() { pingRunner = originalRunner }()

			testRoot := &cobra.Command{Use: "test"}
			testRoot.AddCommand(pingCmd)
			testRoot.SetArgs(append([]string{"ping"}, tt.args...))

			buf := new(bytes.Buffer)
			testRoot.SetOut(buf)
			testRoot.SetErr(buf)
			testRoot.SilenceUsage = true
			testRoot.SilenceErrors = true

			err := testRoot.Execute()
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			gotOutput := buf.String()
			if gotOutput != tt.wantOutput {
				t.Errorf("Output = %q, want %q", gotOutput, tt.wantOutput)
			}

			uiFlag := viper.GetBool("app.ping.ui")
			if uiFlag && tt.uiRunner != nil && !tt.wantErr {
				// Check UI Runner calls
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
