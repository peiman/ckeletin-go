// internal/ui/validation_test.go

package ui

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/validator"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderValidationJSON_Success(t *testing.T) {
	output.SetCommandName("validate")
	defer output.SetCommandName("")

	result := &validator.Result{
		Valid:      true,
		ConfigFile: "/tmp/config.yaml",
	}

	var buf bytes.Buffer
	err := RenderValidationJSON(&buf, result, nil)
	require.NoError(t, err)

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "validate", envelope.Command)
	assert.Nil(t, envelope.Error)

	dataMap, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, true, dataMap["valid"])
	assert.Equal(t, "/tmp/config.yaml", dataMap["config_file"])
	assert.Empty(t, dataMap["errors"])
}

func TestRenderValidationJSON_Failure(t *testing.T) {
	output.SetCommandName("validate")
	defer output.SetCommandName("")

	result := &validator.Result{
		Valid:      false,
		ConfigFile: "/tmp/config.yaml",
		Errors:     []error{errors.New("bad key"), errors.New("bad value")},
		Warnings:   []string{"unknown key: extra"},
	}
	exitErr := validator.ExitCodeForResult(result)
	require.Error(t, exitErr)

	var buf bytes.Buffer
	err := RenderValidationJSON(&buf, result, exitErr)
	require.NoError(t, err, "a rendered failure envelope is not a render error")

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal(buf.Bytes(), &envelope))

	assert.Equal(t, "error", envelope.Status)
	assert.Equal(t, "validate", envelope.Command)
	require.NotNil(t, envelope.Error)
	assert.Equal(t, exitErr.Error(), envelope.Error.Message)

	dataMap, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, dataMap["valid"])
	assert.Equal(t, []interface{}{"bad key", "bad value"}, dataMap["errors"])
	assert.Equal(t, []interface{}{"unknown key: extra"}, dataMap["warnings"])
}

func TestRenderValidationJSON_WriteError(t *testing.T) {
	result := &validator.Result{Valid: true, ConfigFile: "/tmp/config.yaml"}

	err := RenderValidationJSON(&errorWriter{}, result, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to write JSON output")
}
