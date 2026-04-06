// cmd/output_json_test.go

package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetOutputJSONTestState resets all global state that integration tests modify.
// Must be called via defer at the start of each test.
func resetOutputJSONTestState(origCfgFile string, origStatus string, origUsed string) {
	ui.SetOutputMode("")
	ui.SetCommandName("")
	viper.Reset()
	cfgFile = origCfgFile
	configFileStatus = origStatus
	configFileUsed = origUsed
	RootCmd.SetArgs(nil)
	RootCmd.SetOut(nil)
	RootCmd.SetErr(nil)

	// Reset the --output persistent flag to its default value.
	// Without this, Cobra retains the flag value from a previous Execute() call,
	// causing subsequent tests to inherit the prior test's --output setting.
	if f := RootCmd.PersistentFlags().Lookup("output"); f != nil {
		f.Value.Set("text") //nolint:errcheck // resetting to known-good default
		f.Changed = false
	}
}

func TestOutputJSON_PingCommand(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"ping", "--output", "json"})

	err := RootCmd.Execute()
	require.NoError(t, err)

	var envelope ui.JSONEnvelope
	err = json.Unmarshal(stdout.Bytes(), &envelope)
	require.NoError(t, err, "stdout should contain valid JSON, got: %s", stdout.String())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "ping", envelope.Command)
	assert.Nil(t, envelope.Error)
	assert.NotNil(t, envelope.Data)
}

func TestOutputJSON_DefaultIsText(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"ping"})

	err := RootCmd.Execute()
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "Pong", "text mode should contain default ping message")

	var envelope ui.JSONEnvelope
	err = json.Unmarshal(stdout.Bytes(), &envelope)
	assert.Error(t, err, "text mode output should not be valid JSON")
}

func TestOutputJSON_InvalidFormat(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout, stderr bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&stderr)
	RootCmd.SetArgs([]string{"ping", "--output", "xml"})

	err := RootCmd.Execute()
	assert.Error(t, err, "invalid output format should cause an error")
}

func TestOutputJSON_StderrSilent(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout, stderr bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&stderr)
	RootCmd.SetArgs([]string{"ping", "--output", "json"})

	err := RootCmd.Execute()
	require.NoError(t, err)

	// Stderr should be empty in JSON mode (zerolog disabled)
	assert.Empty(t, stderr.String(), "stderr should be empty in JSON mode")

	// Stdout should have valid JSON
	var envelope ui.JSONEnvelope
	err = json.Unmarshal(stdout.Bytes(), &envelope)
	assert.NoError(t, err, "stdout should be valid JSON")
}

func TestOutputJSON_EnvelopeStructure(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"ping", "--output", "json"})

	err := RootCmd.Execute()
	require.NoError(t, err)

	// Parse as raw JSON to check exact field presence
	var raw map[string]json.RawMessage
	err = json.Unmarshal(stdout.Bytes(), &raw)
	require.NoError(t, err)

	// All four fields should be present
	assert.Contains(t, raw, "status")
	assert.Contains(t, raw, "command")
	assert.Contains(t, raw, "data")
	assert.Contains(t, raw, "error")

	// error should be null on success
	assert.Equal(t, "null", string(raw["error"]))
}
