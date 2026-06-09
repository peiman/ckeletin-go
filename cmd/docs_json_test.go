// cmd/docs_json_test.go

package cmd

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/internal/docs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestOutputJSON_DocsConfigCommand verifies `docs config --output json` emits
// the standard success envelope (CKSPEC-OUT-002) with the generated
// documentation wrapped as structured data instead of a raw markdown stream.
func TestOutputJSON_DocsConfigCommand(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"docs", "config", "--output", "json"})

	err := RootCmd.Execute()
	require.NoError(t, err)

	var envelope output.JSONEnvelope
	err = json.Unmarshal(stdout.Bytes(), &envelope)
	require.NoError(t, err, "stdout should contain valid JSON, got: %s", stdout.String())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "config", envelope.Command)
	assert.Nil(t, envelope.Error)

	data, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok, "data payload should be an object, got %T", envelope.Data)
	assert.Equal(t, docs.FormatMarkdown, data["format"])
	content, _ := data["content"].(string)
	assert.Contains(t, content, "## Configuration Sources",
		"data.content should carry the generated markdown")
	assert.NotContains(t, data, "output_file",
		"output_file is omitted when docs go to stdout")
}
