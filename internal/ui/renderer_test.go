package ui

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderSuccess(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		data     interface{}
		contains []string
	}{
		{
			name:     "basic message",
			message:  "Operation completed",
			data:     nil,
			contains: []string{"✔", "Operation completed"},
		},
		{
			name:     "message with string data",
			message:  "File saved",
			data:     "output.txt",
			contains: []string{"✔", "File saved", "output.txt"},
		},
		{
			name:     "message with same string data",
			message:  "Done",
			data:     "Done",
			contains: []string{"✔", "Done"},
		},
		{
			name:     "message with non-string data",
			message:  "Count",
			data:     42,
			contains: []string{"✔", "Count"},
		},
		{
			name:     "message with empty string data",
			message:  "Completed",
			data:     "",
			contains: []string{"✔", "Completed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RenderSuccess(&buf, tt.message, tt.data)

			assert.NoError(t, err)
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestRenderError(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		err      error
		contains []string
	}{
		{
			name:     "basic error",
			message:  "Operation failed",
			err:      errors.New("connection timeout"),
			contains: []string{"✘", "Error:", "Operation failed"},
		},
		{
			name:     "nil error",
			message:  "Something went wrong",
			err:      nil,
			contains: []string{"✘", "Error:", "Something went wrong"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := RenderError(&buf, tt.message, tt.err)

			assert.NoError(t, err)
			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

// errorWriter is a writer that always returns an error
type errorWriter struct{}

func (e *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestRenderSuccess_WriteError(t *testing.T) {
	err := RenderSuccess(&errorWriter{}, "message", nil)
	assert.Error(t, err)
}

func TestRenderError_WriteError(t *testing.T) {
	err := RenderError(&errorWriter{}, "message", errors.New("test"))
	assert.Error(t, err)
}
