// internal/ui/message_test.go

package ui

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintColoredMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     string
		color       string
		wantErr     bool
		wantContain string
	}{
		{
			name:        "Green message",
			message:     "Test Message",
			color:       "green",
			wantErr:     false,
			wantContain: "Test Message",
		},
		{
			name:        "Red message",
			message:     "Error Message",
			color:       "red",
			wantErr:     false,
			wantContain: "Error Message",
		},
		{
			name:        "Invalid color",
			message:     "Test with invalid color",
			color:       "invalid-color",
			wantErr:     true,
			wantContain: "",
		},
		{
			name:        "Empty message",
			message:     "",
			color:       "blue",
			wantErr:     false,
			wantContain: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE
			buf := new(bytes.Buffer)

			// EXECUTION PHASE
			err := PrintColoredMessage(buf, tt.message, tt.color)

			// ASSERTION PHASE
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			if tt.wantContain != "" {
				output := buf.String()
				assert.True(t, bytes.Contains([]byte(output), []byte(tt.wantContain)),
					"Expected output to contain %q, got %q", tt.wantContain, output)
			}
		})
	}
}
