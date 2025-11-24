package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "absolute path to relative",
			input:    "/Users/peiman/dev/cli/ckeletin-go/internal/config/validator.go",
			expected: "./internal/config/validator.go",
		},
		{
			name:     "absolute path with line number",
			input:    "/Users/peiman/dev/cli/ckeletin-go/cmd/root.go:42",
			expected: "./cmd/root.go:42",
		},
		{
			name:     "multiple paths in output",
			input:    "Error in /Users/peiman/dev/cli/ckeletin-go/internal/logger/logger.go:15 and /Users/peiman/dev/cli/ckeletin-go/cmd/version.go:8",
			expected: "Error in ./internal/logger/logger.go:15 and ./cmd/version.go:8",
		},
		{
			name:     "generic absolute path",
			input:    "/some/absolute/path/to/ckeletin-go/internal/config/keys.go",
			expected: "./internal/config/keys.go",
		},
		{
			name:     "no path to normalize",
			input:    "Just some regular text",
			expected: "Just some regular text",
		},
		{
			name:     "relative path unchanged",
			input:    "./internal/config/validator.go",
			expected: "./internal/config/validator.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizePaths(tt.input)
			assert.Equal(t, tt.expected, got, "NormalizePaths should normalize paths correctly")
		})
	}
}

func TestNormalizeTimings(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single timing",
			input:    "Completed in 1.23s",
			expected: "Completed in X.XXs",
		},
		{
			name:     "multiple timings",
			input:    "Test 1: 0.45s, Test 2: 12.89s, Test 3: 123.45s",
			expected: "Test 1: X.XXs, Test 2: X.XXs, Test 3: X.XXs",
		},
		{
			name:     "subsecond timing",
			input:    "Completed in 0.001s",
			expected: "Completed in X.XXs",
		},
		{
			name:     "took duration",
			input:    "took 45.2s",
			expected: "took X.XXs",
		},
		{
			name:     "elapsed duration",
			input:    "elapsed: 123.45s",
			expected: "elapsed: X.XXs",
		},
		{
			name:     "duration with text",
			input:    "Test completed, took 5.67s to finish",
			expected: "Test completed, took X.XXs to finish",
		},
		{
			name:     "no timing to normalize",
			input:    "Just some regular text",
			expected: "Just some regular text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTimings(tt.input)
			assert.Equal(t, tt.expected, got, "NormalizeTimings should normalize timings correctly")
		})
	}
}

func TestNormalizeTempPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "macOS temp directory",
			input:    "/var/folders/ab/cd1234567890/T/TestScaffoldInit1234567890/001",
			expected: "/tmp/TEMP_DIR/001",
		},
		{
			name:     "Linux temp directory",
			input:    "/tmp/TestScaffoldInit1234567890/001",
			expected: "/tmp/TEMP_DIR/001",
		},
		{
			name:     "Windows temp directory with backslashes",
			input:    `C:\Users\username\AppData\Local\Temp\TestScaffoldInit1234567890\001`,
			expected: "/tmp/TEMP_DIR/001",
		},
		{
			name:     "Windows temp directory with forward slashes",
			input:    "C:/Users/username/AppData/Local/Temp/TestScaffoldInit1234567890/001",
			expected: "/tmp/TEMP_DIR/001",
		},
		{
			name:     "Windows temp directory with different drive letter",
			input:    `D:\Users\testuser\AppData\Local\Temp\TestScaffoldInit9876543210\042`,
			expected: "/tmp/TEMP_DIR/042",
		},
		{
			name:     "multiple temp paths in output",
			input:    "/tmp/TestScaffoldInit111/001 and /var/folders/xy/z9876543/T/TestScaffoldInit222/002",
			expected: "/tmp/TEMP_DIR/001 and /tmp/TEMP_DIR/002",
		},
		{
			name:     "no temp path to normalize",
			input:    "Just some regular text",
			expected: "Just some regular text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeTempPaths(tt.input)
			assert.Equal(t, tt.expected, got, "NormalizeTempPaths should normalize temp paths correctly")
		})
	}
}

func TestNormalizeCheckOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "full check output with all normalizations",
			input: `/Users/peiman/dev/cli/ckeletin-go/internal/config/validator.go:26
✅ All checks passed (15/15)
Completed in 12.34s
Test took 5.67s to finish`,
			expected: `./internal/config/validator.go:26
✅ All checks passed (15/15)
Completed in X.XXs
Test took X.XXs to finish`,
		},
		{
			name: "multiple paths and timings",
			input: `Error in /Users/peiman/dev/cli/ckeletin-go/cmd/root.go:10
/some/other/path/to/ckeletin-go/internal/logger/logger.go:20
Tests completed in 45.2s, took 50.1s total`,
			expected: `Error in ./cmd/root.go:10
./internal/logger/logger.go:20
Tests completed in X.XXs, took X.XXs total`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "text with no normalizations needed",
			input:    "Just some regular output text",
			expected: "Just some regular output text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NormalizeCheckOutput(tt.input)
			assert.Equal(t, tt.expected, got, "NormalizeCheckOutput should apply all normalizations")
		})
	}
}
