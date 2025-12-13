package checkmate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetErrorSummary(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short single line",
			input:    "error occurred",
			expected: "error occurred",
		},
		{
			name:     "multiline takes first",
			input:    "first line\nsecond line\nthird line",
			expected: "first line",
		},
		{
			name:     "long line truncated",
			input:    "this is a very long error message that should be truncated because it exceeds the maximum allowed length",
			expected: "this is a very long error message that should b...",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only newlines",
			input:    "\n\n\n",
			expected: "",
		},
		{
			name:     "exactly 50 chars",
			input:    "12345678901234567890123456789012345678901234567890",
			expected: "12345678901234567890123456789012345678901234567890",
		},
		{
			name:     "51 chars truncated",
			input:    "123456789012345678901234567890123456789012345678901",
			expected: "12345678901234567890123456789012345678901234567...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getErrorSummary(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatErrorDetails(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxWidth int
		expected []string
	}{
		{
			name:     "short single line",
			input:    "error",
			maxWidth: 60,
			expected: []string{"error"},
		},
		{
			name:     "multiline",
			input:    "line one\nline two\nline three",
			maxWidth: 60,
			expected: []string{"line one", "line two", "line three"},
		},
		{
			name:     "empty lines filtered",
			input:    "line one\n\nline two\n\n\nline three",
			maxWidth: 60,
			expected: []string{"line one", "line two", "line three"},
		},
		{
			name:     "long line wrapped",
			input:    "this is a very long line that should be wrapped at word boundaries",
			maxWidth: 30,
			expected: []string{"this is a very long line that", "should be wrapped at word", "boundaries"},
		},
		{
			name:     "empty string",
			input:    "",
			maxWidth: 60,
			expected: nil,
		},
		{
			name:     "whitespace trimmed",
			input:    "  line one  \n  line two  ",
			maxWidth: 60,
			expected: []string{"line one", "line two"},
		},
		{
			name:     "truncated at maxLines",
			input:    "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n11\n12",
			maxWidth: 60,
			expected: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "... (truncated)"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorDetails(tt.input, tt.maxWidth)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckProgress_Fields(t *testing.T) {
	cp := CheckProgress{
		Name:        "test",
		Status:      CheckFailed,
		Progress:    0.5,
		Error:       assert.AnError,
		Remediation: "fix it",
	}

	assert.Equal(t, "test", cp.Name)
	assert.Equal(t, CheckFailed, cp.Status)
	assert.Equal(t, 0.5, cp.Progress)
	assert.Error(t, cp.Error)
	assert.Equal(t, "fix it", cp.Remediation)
}

func TestCheckUpdateMsg_Fields(t *testing.T) {
	msg := CheckUpdateMsg{
		Index:       1,
		Status:      CheckPassed,
		Progress:    1.0,
		Error:       nil,
		Remediation: "run fix",
	}

	assert.Equal(t, 1, msg.Index)
	assert.Equal(t, CheckPassed, msg.Status)
	assert.Equal(t, 1.0, msg.Progress)
	assert.Nil(t, msg.Error)
	assert.Equal(t, "run fix", msg.Remediation)
}

func TestCheckStatus_Constants(t *testing.T) {
	assert.Equal(t, CheckStatus(0), CheckPending)
	assert.Equal(t, CheckStatus(1), CheckRunning)
	assert.Equal(t, CheckStatus(2), CheckPassed)
	assert.Equal(t, CheckStatus(3), CheckFailed)
}
