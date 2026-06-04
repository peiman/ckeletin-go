// cmd/config_validate_json_test.go

package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runConfigValidateJSON drives `config validate --file <path> --output json` and
// returns the captured stdout.
func runConfigValidateJSON(t *testing.T, cfgYAML string) string {
	t.Helper()

	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)
	defer func() { validateConfigFile = "" }()

	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(cfgPath, []byte(cfgYAML), 0o600))

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"config", "validate", "--file", cfgPath, "--output", "json"})

	// Exit status is communicated via the JSON envelope (like `check`), so the
	// returned error is not asserted here.
	_ = RootCmd.Execute()
	return stdout.String()
}

// TestConfigValidateJSON_WarningsEmitSingleEnvelope is the regression test for the
// JSON-contract bug: a config that produces warnings must yield exactly ONE JSON
// envelope on stdout — no human-readable text from FormatResult leaking in front
// of it.
func TestConfigValidateJSON_WarningsEmitSingleEnvelope(t *testing.T) {
	// Valid config, but with an unknown key → a warning (the path that leaked text).
	out := runConfigValidateJSON(t, "app:\n  log_level: info\nunknown_key_xyz: 1\n")

	assert.NotContains(t, out, "Validating:", "no human text may precede the JSON envelope")
	assert.NotContains(t, out, "⚠️", "no human warning text may leak to stdout")

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal([]byte(out), &envelope),
		"stdout must be exactly one JSON envelope, got: %q", out)
	assert.Equal(t, "validate", envelope.Command)
	assert.Equal(t, "error", envelope.Status, "warnings map to a non-zero result")
	require.NotNil(t, envelope.Error)
}

// TestConfigValidateJSON_ValidEmitsSingleEnvelope verifies the success path emits a
// single JSON envelope too (previously it emitted only human text and no envelope).
func TestConfigValidateJSON_ValidEmitsSingleEnvelope(t *testing.T) {
	out := runConfigValidateJSON(t, "app:\n  log_level: info\n")

	assert.NotContains(t, out, "Validating:", "no human text may precede the JSON envelope")

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal([]byte(out), &envelope),
		"stdout must be exactly one JSON envelope, got: %q", out)
	assert.Equal(t, "validate", envelope.Command)
	assert.NotNil(t, envelope.Data)
}
