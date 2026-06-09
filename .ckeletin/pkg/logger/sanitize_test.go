// .ckeletin/pkg/logger/sanitize_test.go

package logger

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
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
			assert.Equal(t, tt.expected, got)
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
			assert.Equal(t, tt.expected, got)
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
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestSetMaxLogLength(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	// Save and restore original value
	original := maxLogStringLength.Load()
	defer maxLogStringLength.Store(original)

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
			maxLogStringLength.Store(original)

			SetMaxLogLength(tt.length)

			if tt.wantApply {
				assert.Equal(t, int64(tt.length), maxLogStringLength.Load(),
					"SetMaxLogLength(%d) did not apply, got %d", tt.length, maxLogStringLength.Load())
			} else {
				assert.Equal(t, original, maxLogStringLength.Load(),
					"SetMaxLogLength(%d) should not apply, but changed from %d to %d",
					tt.length, original, maxLogStringLength.Load())
			}
		})
	}
}

func TestSetMaxLogLength_ConcurrentWithSanitizeLogString(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	original := maxLogStringLength.Load()
	defer maxLogStringLength.Store(original)

	// SETUP: input long enough that every SanitizeLogString call reads the limit
	input := strings.Repeat("x", 2000)

	// EXECUTION: hammer the setter and the sanitizer concurrently;
	// `go test -race` flags unsynchronized access to the shared limit
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(2)
		go func(offset int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				SetMaxLogLength(100 + offset*200 + j)
			}
		}(i)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				got := SanitizeLogString(input)

				// ASSERTION: result is always input truncated at some applied limit
				assert.True(t, strings.HasSuffix(got, "...[truncated]"))
			}
		}()
	}
	wg.Wait()
}

func TestSanitizeLogString_MultibyteTruncation(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	original := maxLogStringLength.Load()
	defer maxLogStringLength.Store(original)

	tests := []struct {
		name     string
		maxLen   int
		input    string
		expected string
	}{
		{
			name:     "ASCII truncated exactly at limit",
			maxLen:   10,
			input:    strings.Repeat("a", 15),
			expected: strings.Repeat("a", 10) + "...[truncated]",
		},
		{
			name:     "Emoji straddling the cut is dropped, not split",
			maxLen:   10,
			input:    strings.Repeat("a", 8) + "\U0001F600\U0001F600", // 8 + 4 + 4 bytes
			expected: strings.Repeat("a", 8) + "...[truncated]",
		},
		{
			name:     "CJK character straddling the cut is dropped, not split",
			maxLen:   10,
			input:    "你好你好", // 4 runes x 3 bytes; byte 10 splits the 4th rune
			expected: "你好你" + "...[truncated]",
		},
		{
			name:     "Emoji ending exactly on the cut is kept",
			maxLen:   8,
			input:    "\U0001F600\U0001F600\U0001F600", // 12 bytes; bytes 0-7 are two whole emoji
			expected: "\U0001F600\U0001F600" + "...[truncated]",
		},
		{
			name:     "Multibyte string within limit unchanged",
			maxLen:   12,
			input:    "你好你好", // exactly 12 bytes, not above the limit
			expected: "你好你好",
		},
		{
			name:     "Single rune wider than limit truncates to empty",
			maxLen:   2,
			input:    "\U0001F600abc", // 4-byte emoji exceeds the 2-byte limit
			expected: "...[truncated]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			SetMaxLogLength(tt.maxLen)

			// EXECUTION
			got := SanitizeLogString(tt.input)

			// ASSERTION
			assert.Equal(t, tt.expected, got)
			assert.True(t, utf8.ValidString(got), "output must be valid UTF-8")
		})
	}
}

func TestSanitizeLogString_Truncation(t *testing.T) {
	// Note: Cannot use t.Parallel() because this test modifies package-level variable
	// Test with custom max length
	original := maxLogStringLength.Load()
	defer maxLogStringLength.Store(original)

	SetMaxLogLength(50)

	input := strings.Repeat("x", 100)
	expected := strings.Repeat("x", 50) + "...[truncated]"

	got := SanitizeLogString(input)
	assert.Equal(t, expected, got)
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
			assert.Equal(t, tt.expectedLen, got,
				"initMaxLogLength() = %d, want %d", got, tt.expectedLen)
		})
	}
}
