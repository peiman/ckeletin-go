// internal/logger/sanitize_test.go

package logger

import (
	"errors"
	"strings"
	"testing"
)

func TestSanitizeLogString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string unchanged",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Newline removed",
			input:    "Line1\nLine2",
			expected: "Line1Line2",
		},
		{
			name:     "Tab removed",
			input:    "Col1\tCol2",
			expected: "Col1Col2",
		},
		{
			name:     "Carriage return removed",
			input:    "Text\rText",
			expected: "TextText",
		},
		{
			name:     "Multiple control characters",
			input:    "Text\n\r\tMore\x00Text",
			expected: "TextMoreText",
		},
		{
			name:     "ANSI escape codes removed",
			input:    "\x1b[31mRed Text\x1b[0m",
			expected: "[31mRed Text[0m",
		},
		{
			name:     "Very long string truncated",
			input:    strings.Repeat("a", 1500),
			expected: strings.Repeat("a", 1000) + "...[truncated]",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Log injection attempt",
			input:    "legit log\n2025-01-01 FAKE ERROR: injected message",
			expected: "legit log2025-01-01 FAKE ERROR: injected message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeLogString(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeLogString() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	// Note: Cannot use t.Parallel() when using t.Setenv()
	// Use t.Setenv for automatic cleanup
	testHome := "/home/testuser"
	t.Setenv("HOME", testHome)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Home directory replaced",
			input:    "/home/testuser/config.yaml",
			expected: "~/config.yaml",
		},
		{
			name:     "Non-home path unchanged",
			input:    "/etc/config.yaml",
			expected: "/etc/config.yaml",
		},
		{
			name:     "Path with control characters",
			input:    "/home/testuser/file\nname.txt",
			expected: "~/filename.txt",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
		{
			name:     "Relative path",
			input:    "../config.yaml",
			expected: "../config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizePath(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizePath() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSanitizeError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "Nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "Normal error",
			err:      errors.New("something went wrong"),
			expected: "something went wrong",
		},
		{
			name:     "Error with newline",
			err:      errors.New("error on line1\nand line2"),
			expected: "error on line1and line2",
		},
		{
			name:     "Error with control chars",
			err:      errors.New("error\twith\ttabs"),
			expected: "errorwithtabs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := SanitizeError(tt.err)
			if got != tt.expected {
				t.Errorf("SanitizeError() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSetMaxLogLength(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	// Save and restore original value
	original := maxLogStringLength
	defer func() { maxLogStringLength = original }()

	tests := []struct {
		name      string
		length    int
		wantApply bool
	}{
		{
			name:      "Positive value applied",
			length:    500,
			wantApply: true,
		},
		{
			name:      "Zero value ignored",
			length:    0,
			wantApply: false,
		},
		{
			name:      "Negative value ignored",
			length:    -100,
			wantApply: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset to original for each test
			maxLogStringLength = original

			SetMaxLogLength(tt.length)

			if tt.wantApply {
				if maxLogStringLength != tt.length {
					t.Errorf("SetMaxLogLength(%d) did not apply, got %d", tt.length, maxLogStringLength)
				}
			} else {
				if maxLogStringLength != original {
					t.Errorf("SetMaxLogLength(%d) should not apply, but changed from %d to %d",
						tt.length, original, maxLogStringLength)
				}
			}
		})
	}
}

func TestSanitizeLogString_Truncation(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	// Test with custom max length
	original := maxLogStringLength
	defer func() { maxLogStringLength = original }()

	SetMaxLogLength(50)

	input := strings.Repeat("x", 100)
	expected := strings.Repeat("x", 50) + "...[truncated]"

	got := SanitizeLogString(input)
	if got != expected {
		t.Errorf("SanitizeLogString() with custom length = %q, want %q", got, expected)
	}
}

func TestInitMaxLogLength(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test uses t.Setenv()
	tests := []struct {
		name        string
		envValue    string
		expectedLen int
	}{
		{
			name:        "Default value when env not set",
			envValue:    "",
			expectedLen: 1000,
		},
		{
			name:        "Valid positive value",
			envValue:    "500",
			expectedLen: 500,
		},
		{
			name:        "Large valid value",
			envValue:    "5000",
			expectedLen: 5000,
		},
		{
			name:        "Invalid negative value uses default",
			envValue:    "-100",
			expectedLen: 1000,
		},
		{
			name:        "Invalid zero value uses default",
			envValue:    "0",
			expectedLen: 1000,
		},
		{
			name:        "Invalid non-numeric value uses default",
			envValue:    "invalid",
			expectedLen: 1000,
		},
		{
			name:        "Invalid float value uses default",
			envValue:    "100.5",
			expectedLen: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup environment variable
			if tt.envValue != "" {
				t.Setenv("LOG_TRUNCATE_LIMIT", tt.envValue)
			}

			// Call the initialization function
			got := initMaxLogLength()

			// Verify result
			if got != tt.expectedLen {
				t.Errorf("initMaxLogLength() = %d, want %d", got, tt.expectedLen)
			}
		})
	}
}
