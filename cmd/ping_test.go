// cmd/ping_test.go

package cmd

import (
	"bytes"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/ui"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPingCommand tests the ping command integration with configuration
func TestPingCommand(t *testing.T) {
	// SETUP PHASE: Save logger state and restore after test
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Setup debug logging
	logBuf := &bytes.Buffer{}
	log.Logger = zerolog.New(logBuf).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	tests := []struct {
		name            string
		testFixturePath string
		args            []string
		wantErr         bool
		wantOutput      string
		writer          *bytes.Buffer
		mockRunner      *ui.MockUIRunner
	}{
		{
			name:            "Default Configuration",
			testFixturePath: "../testdata/config/valid.yaml",
			args:            []string{},
			wantErr:         false,
			wantOutput:      "",
			writer:          &bytes.Buffer{},
			mockRunner:      &ui.MockUIRunner{},
		},
		{
			name:            "JSON Configuration",
			testFixturePath: "../testdata/config/valid.json",
			args:            []string{},
			wantErr:         false,
			wantOutput:      "✔ Pong! JSON Config Message\n",
			writer:          &bytes.Buffer{},
			mockRunner:      &ui.MockUIRunner{},
		},
		{
			name:            "CLI Args Override Configuration",
			testFixturePath: "../testdata/config/valid.yaml",
			args:            []string{"--message", "CLI Message", "--color", "cyan"},
			wantErr:         false,
			wantOutput:      "",
			writer:          &bytes.Buffer{},
			mockRunner:      &ui.MockUIRunner{},
		},
		{
			name:            "Partial Configuration",
			testFixturePath: "../testdata/config/partial.yaml",
			args:            []string{"--color", "white"},
			wantErr:         false,
			wantOutput:      "✔ Pong! Partial Config Message\n",
			writer:          &bytes.Buffer{},
			mockRunner:      &ui.MockUIRunner{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE: Reset viper, load fixture, and setup command
			viper.Reset()
			viper.SetConfigFile(tt.testFixturePath)
			err := viper.ReadInConfig()
			require.NoError(t, err, "Failed to load test fixture %s", tt.testFixturePath)

			// Create a new command instance for each test to avoid state pollution
			cmd := &cobra.Command{
				Use: "ping",
				RunE: func(cmd *cobra.Command, args []string) error {
					// Use the injected mock runner for this test
					return runPingWithUIRunner(cmd, args, tt.mockRunner)
				},
			}

			// Register flags
			err = RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
				"app.ping.output_message": "message",
				"app.ping.output_color":   "color",
				"app.ping.ui":             "ui",
			})
			require.NoError(t, err, "Failed to register flags")

			cmd.SetOut(tt.writer)
			cmd.SetArgs(tt.args)
			err = cmd.ParseFlags(tt.args)
			require.NoError(t, err, "Failed to parse flags")

			tt.writer.Reset()

			// EXECUTION PHASE: Run the ping command
			err = runPingWithUIRunner(cmd, []string{}, tt.mockRunner)

			// ASSERTION PHASE: Check error and output
			if tt.wantErr {
				assert.Error(t, err, "runPing() should return error")
			} else {
				assert.NoError(t, err, "runPing() should not return error")
			}

			if !tt.wantErr && !viper.GetBool("app.ping.ui") {
				got := tt.writer.String()
				assert.Equal(t, tt.wantOutput, got, "runPing() output mismatch")
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
	mockRunner := &ui.MockUIRunner{}

	// Create command
	cmd := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPingWithUIRunner(cmd, args, mockRunner)
		},
	}

	// Register flags
	err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	})
	require.NoError(t, err, "Failed to register flags")

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	// EXECUTION PHASE: No flags set, should use viper values
	err = runPingWithUIRunner(cmd, []string{}, mockRunner)
	require.NoError(t, err, "runPing() failed")

	// ASSERTION PHASE: Check that viper values were used
	got := outBuf.String()
	assert.Contains(t, got, "ConfigMessage", "Expected output to contain 'ConfigMessage'")

	// EXECUTION PHASE: Set flags, should override viper
	err = cmd.Flags().Set("message", "FlagMessage")
	require.NoError(t, err, "Failed to set message flag")
	err = cmd.Flags().Set("color", "red")
	require.NoError(t, err, "Failed to set color flag")

	outBuf.Reset()
	err = runPingWithUIRunner(cmd, []string{}, mockRunner)
	require.NoError(t, err, "runPing() with flags failed")

	// ASSERTION PHASE: Check that flag values were used
	got = outBuf.String()
	assert.Contains(t, got, "FlagMessage", "Expected output to contain 'FlagMessage'")
}

// TestPingConfigDefaults ensures the default values from config registry are used
func TestPingConfigDefaults(t *testing.T) {
	// SETUP PHASE: Reset viper and apply config defaults
	viper.Reset()
	config.SetDefaults()

	// Create mock UI runner
	mockRunner := &ui.MockUIRunner{}

	// Create command
	cmd := &cobra.Command{
		Use: "ping",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPingWithUIRunner(cmd, args, mockRunner)
		},
	}

	// Register flags
	err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	})
	require.NoError(t, err, "Failed to register flags")

	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	// EXECUTION PHASE: Run with defaults
	err = runPingWithUIRunner(cmd, []string{}, mockRunner)

	// ASSERTION PHASE: Check that defaults were used
	require.NoError(t, err, "runPing() failed")

	got := outBuf.String()
	// Default message is "Pong" as defined in ping_options.go
	assert.Contains(t, got, "Pong", "Expected output to contain default message 'Pong'")
}

// TestRunPingWrapper tests the public runPing wrapper function to ensure 100% coverage.
// This test verifies that the wrapper correctly instantiates the default UIRunner
// and delegates to runPingWithUIRunner.
func TestRunPingWrapper(t *testing.T) {
	// SETUP PHASE: Reset viper and set defaults
	viper.Reset()
	config.SetDefaults()

	// Create a command instance
	cmd := &cobra.Command{
		Use:  "ping",
		RunE: runPing, // Use the actual wrapper, not the DI version
	}

	// Register flags
	err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", map[string]string{
		"app.ping.output_message": "message",
		"app.ping.output_color":   "color",
		"app.ping.ui":             "ui",
	})
	require.NoError(t, err, "Failed to register flags")

	// Set output buffer
	outBuf := &bytes.Buffer{}
	cmd.SetOut(outBuf)

	// Set UI to false to avoid terminal UI in tests
	err = cmd.Flags().Set("ui", "false")
	require.NoError(t, err, "Failed to set ui flag")

	// EXECUTION PHASE: Call runPing directly (not the DI version)
	// This tests that the wrapper creates ui.NewDefaultUIRunner() correctly
	err = runPing(cmd, []string{})

	// ASSERTION PHASE: Should complete without error
	assert.NoError(t, err, "runPing() wrapper should not return error")
	assert.Contains(t, outBuf.String(), "Pong", "Output should contain default message")
}
