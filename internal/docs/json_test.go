// internal/docs/json_test.go

package docs

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// decodeEnvelope parses a JSON envelope and re-decodes its data payload into a
// Result so assertions can use the typed fields.
func decodeEnvelope(t *testing.T, raw []byte) (output.JSONEnvelope, Result) {
	t.Helper()

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal(raw, &envelope),
		"GenerateJSON must emit a valid JSON envelope, got: %s", string(raw))

	data, err := json.Marshal(envelope.Data)
	require.NoError(t, err)
	var result Result
	require.NoError(t, json.Unmarshal(data, &result))
	return envelope, result
}

func TestGenerateJSON_StdoutContent(t *testing.T) {
	// SETUP PHASE
	output.SetCommandName("config")
	defer output.SetCommandName("")

	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		OutputFile:   "",
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.NoError(t, err)
	envelope, result := decodeEnvelope(t, buf.Bytes())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "config", envelope.Command)
	assert.Nil(t, envelope.Error)

	assert.Equal(t, FormatMarkdown, result.Format)
	assert.Empty(t, result.OutputFile, "no output file configured")
	assert.Contains(t, result.Content, "## Configuration Sources",
		"content should carry the generated markdown")
}

func TestGenerateJSON_OutputFile(t *testing.T) {
	// SETUP PHASE
	outputFile := filepath.Join(t.TempDir(), "docs.md")

	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		OutputFile:   outputFile,
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.NoError(t, err)
	envelope, result := decodeEnvelope(t, buf.Bytes())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, outputFile, result.OutputFile)
	assert.Empty(t, result.Content,
		"content stays out of the envelope when written to a file")

	written, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(written), "## Configuration Sources",
		"the documentation file must still be written in JSON mode")
}

func TestGenerateJSON_YAMLFormat(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: FormatYAML,
		OutputFile:   "",
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.NoError(t, err)
	_, result := decodeEnvelope(t, buf.Bytes())
	assert.Equal(t, FormatYAML, result.Format)
	assert.NotEmpty(t, result.Content)
}

// TestGenerateJSON_OutputFileOpenError verifies an output-file open failure
// surfaces as an error and leaves the envelope writer untouched.
func TestGenerateJSON_OutputFileOpenError(t *testing.T) {
	// SETUP PHASE
	origOpenOutputFile := openOutputFile
	defer func() { openOutputFile = origOpenOutputFile }()
	openOutputFile = func(path string) (io.WriteCloser, error) {
		return nil, errors.New("permission denied")
	}

	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		OutputFile:   "docs.md",
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output file")
	assert.Zero(t, buf.Len(),
		"no envelope may be written on error — main.go renders the error envelope")
}

// TestGenerateJSON_OutputFileWriteError verifies a write failure on the
// output file propagates as an error (no success envelope over a truncated
// artifact) and leaves the envelope writer untouched.
func TestGenerateJSON_OutputFileWriteError(t *testing.T) {
	// SETUP PHASE
	origOpenOutputFile := openOutputFile
	defer func() { openOutputFile = origOpenOutputFile }()
	openOutputFile = func(path string) (io.WriteCloser, error) {
		return &failAfterWriter{limit: 64}, nil
	}

	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: FormatMarkdown,
		OutputFile:   "docs.md",
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.Error(t, err, "a failing output file must surface an error")
	assert.ErrorIs(t, err, errWriteFailed)
	assert.Zero(t, buf.Len(),
		"no envelope may be written on error — main.go renders the error envelope")
}

func TestGenerateJSON_UnsupportedFormat(t *testing.T) {
	// SETUP PHASE
	var buf bytes.Buffer
	generator := NewGenerator(Config{
		Writer:       &buf,
		OutputFormat: "invalid",
		OutputFile:   "",
		Registry:     config.Registry,
	})

	// EXECUTION PHASE
	err := generator.GenerateJSON(&buf)

	// ASSERTION PHASE
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format")
	assert.Zero(t, buf.Len(),
		"no envelope may be written on error — main.go renders the error envelope")
}
