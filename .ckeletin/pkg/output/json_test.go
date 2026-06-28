package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderJSON_Success(t *testing.T) {
	var buf bytes.Buffer
	env := JSONEnvelope{
		Status:  "success",
		Command: "ping",
		Data:    map[string]string{"message": "hello"},
		Error:   nil,
	}

	err := RenderJSON(&buf, env)
	require.NoError(t, err)

	var decoded JSONEnvelope
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, "success", decoded.Status)
	assert.Equal(t, "ping", decoded.Command)
	assert.Nil(t, decoded.Error)
	assert.NotNil(t, decoded.Data)
}

func TestRenderJSON_Error(t *testing.T) {
	var buf bytes.Buffer
	code := "CONFIG_VALIDATION"
	env := JSONEnvelope{
		Status:  "error",
		Command: "ping",
		Data:    nil,
		Error:   &JSONError{Message: "invalid color", Code: &code},
	}

	err := RenderJSON(&buf, env)
	require.NoError(t, err)

	var decoded JSONEnvelope
	err = json.Unmarshal(buf.Bytes(), &decoded)
	require.NoError(t, err)

	assert.Equal(t, "error", decoded.Status)
	assert.Nil(t, decoded.Data)
	assert.NotNil(t, decoded.Error)
	assert.Equal(t, "invalid color", decoded.Error.Message)
	require.NotNil(t, decoded.Error.Code)
	assert.Equal(t, "CONFIG_VALIDATION", *decoded.Error.Code)
}

// TestRenderJSON_ErrorCodeNullWhenAbsent locks the CKSPEC-OUT-003 total
// error-object contract (#40): an error with no classification MUST still emit
// the `code` key as JSON null — never omit it.
func TestRenderJSON_ErrorCodeNullWhenAbsent(t *testing.T) {
	var buf bytes.Buffer
	env := JSONEnvelope{
		Status:  "error",
		Command: "ping",
		Data:    nil,
		Error:   &JSONError{Message: "boom"}, // no Code
	}

	err := RenderJSON(&buf, env)
	require.NoError(t, err)

	// The raw JSON MUST contain "code":null — present, not omitted.
	assert.Contains(t, buf.String(), `"code": null`,
		"error object must emit code as present-and-null, never omit it")
}

func TestRenderJSON_NilData(t *testing.T) {
	var buf bytes.Buffer
	env := JSONEnvelope{
		Status:  "success",
		Command: "test",
		Data:    nil,
		Error:   nil,
	}

	err := RenderJSON(&buf, env)
	require.NoError(t, err)

	var raw map[string]json.RawMessage
	err = json.Unmarshal(buf.Bytes(), &raw)
	require.NoError(t, err)

	assert.Equal(t, "null", string(raw["data"]))
	assert.Equal(t, "null", string(raw["error"]))
}

// errorWriter is a writer that always returns an error.
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestRenderJSON_WriteError(t *testing.T) {
	env := JSONEnvelope{Status: "success", Command: "test"}
	err := RenderJSON(&errorWriter{}, env)
	assert.Error(t, err)
}

type mockJSONResponder struct {
	custom map[string]string
}

func (m *mockJSONResponder) JSONResponse() interface{} {
	return m.custom
}

func TestRenderJSON_JSONResponder(t *testing.T) {
	responder := &mockJSONResponder{custom: map[string]string{"key": "value"}}
	data := ResolveJSONData(responder)

	result, ok := data.(map[string]string)
	require.True(t, ok)
	assert.Equal(t, "value", result["key"])
}

func TestRenderJSON_NonJSONResponder(t *testing.T) {
	plain := map[string]int{"count": 42}
	data := ResolveJSONData(plain)

	result, ok := data.(map[string]int)
	require.True(t, ok)
	assert.Equal(t, 42, result["count"])
}

func TestOutputMode_Default(t *testing.T) {
	SetOutputMode("")
	assert.Equal(t, "text", OutputMode())
	assert.False(t, IsJSONMode())
}

func TestOutputMode_JSON(t *testing.T) {
	SetOutputMode("json")
	defer SetOutputMode("")

	assert.Equal(t, "json", OutputMode())
	assert.True(t, IsJSONMode())
}

func TestCommandName(t *testing.T) {
	SetCommandName("ping")
	defer SetCommandName("")

	assert.Equal(t, "ping", CommandName())
}
