package check

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractShellError(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{"empty output", "", ""},
		// Success indicators get filtered, but if that's all there is, fallback returns last lines
		{"success indicators only", "✅ All checks passed\n✅ Done", "✅ All checks passed\n✅ Done"},
		// Error lines keep their prefixes
		{"error with cross mark", "✅ Step 1 passed\n❌ Step 2 failed: something went wrong", "❌ Step 2 failed: something went wrong"},
		{"error with checkmark x", "✗ validation failed", "✗ validation failed"},
		// Keyword matching
		{"error keyword", "Some output\nerror: command not found\nMore output", "error: command not found"},
		{"Error keyword capitalized", "Error loading config", "Error loading config"},
		{"failed keyword", "test failed with exit code 1", "test failed with exit code 1"},
		{"FAIL keyword", "FAIL: something broke", "FAIL: something broke"},
		// Multiple errors joined
		{"multiple errors", "❌ Error 1\n❌ Error 2\n❌ Error 3", "❌ Error 1\n❌ Error 2\n❌ Error 3"},
		// Fallback to last lines
		{"fallback to last lines", "line 1\nline 2\nline 3\nline 4", "line 2\nline 3\nline 4"},
		// ANSI stripping keeps emoji
		{"strips ANSI codes", "\033[31m❌ error message\033[0m", "❌ error message"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractShellError(tt.output))
		})
	}
}

func TestStripANSI(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"no ANSI codes", "plain text", "plain text"},
		{"red color", "\033[31mred text\033[0m", "red text"},
		{"bold", "\033[1mbold text\033[0m", "bold text"},
		{"multiple codes", "\033[1m\033[31mbold red\033[0m normal", "bold red normal"},
		{"256 color", "\033[38;5;196mcolored\033[0m", "colored"},
		{"empty string", "", ""},
		{"complex sequence", "\033[0;32m✅\033[0m check passed", "✅ check passed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, stripANSI(tt.input))
		})
	}
}

func TestIsLetter(t *testing.T) {
	// Letters should return true
	for _, b := range []byte{'a', 'z', 'A', 'Z', 'm', 'M'} {
		t.Run(string(b), func(t *testing.T) {
			assert.True(t, isLetter(b), "expected %c to be a letter", b)
		})
	}

	// Non-letters should return false
	for _, b := range []byte{'0', '9', ' ', ';', '['} {
		t.Run(string(b), func(t *testing.T) {
			assert.False(t, isLetter(b), "expected %c to not be a letter", b)
		})
	}
}

func TestFilterLintOutput(t *testing.T) {
	tests := []struct {
		name, output, tool, expected string
	}{
		{"no issues found", "", "golangci-lint", "golangci-lint found issues"},
		{"single issue", "main.go:10:5: undefined: foo", "go vet", "go vet found 1 issue(s):\n  • main.go:10:5: undefined: foo"},
		{"multiple issues", "main.go:10:5: error 1\nutils.go:20:3: error 2", "golangci-lint", "golangci-lint found 2 issue(s):\n  • main.go:10:5: error 1\n  • utils.go:20:3: error 2"},
		{"filters package headers", "# github.com/example/pkg\nmain.go:10:5: error here", "go vet", "go vet found 1 issue(s):\n  • main.go:10:5: error here"},
		{"filters level= lines", "level=warning msg=\"some warning\"\nmain.go:5:1: actual error", "golangci-lint", "golangci-lint found 1 issue(s):\n  • main.go:5:1: actual error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, filterLintOutput(tt.output, tt.tool))
		})
	}
}

func TestFilterTestOutput(t *testing.T) {
	tests := []struct {
		name        string
		output      string
		shouldMatch []string
	}{
		{"empty output", "", []string{"tests failed (unknown error)"}},
		{"passing tests filtered out", "ok  \tgithub.com/example/pkg\t0.5s", []string{"tests failed (unknown error)"}},
		{"skipped tests filtered out", "?   \tgithub.com/example/pkg\t[no test files]", []string{"tests failed (unknown error)"}},
		{
			"test failure captured",
			"--- FAIL: TestSomething (0.00s)\n    test.go:10: expected true\nFAIL\tgithub.com/example/pkg\t0.5s",
			[]string{"1 test(s) failed", "TestSomething", "--- FAIL: TestSomething"},
		},
		{
			"compile error captured",
			"# github.com/example/pkg\n./main.go:10:5: undefined: foo\nFAIL\tgithub.com/example/pkg [build failed]",
			[]string{"failed to compile", "# github.com/example/pkg", "./main.go:10:5: undefined: foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterTestOutput(tt.output)
			for _, match := range tt.shouldMatch {
				assert.Contains(t, result, match)
			}
		})
	}
}

func TestExtractTestName(t *testing.T) {
	tests := []struct {
		line, expected string
	}{
		{"--- FAIL: TestSomething (0.00s)", "TestSomething"},
		{"--- FAIL: TestSomething/subtest (0.00s)", "TestSomething/subtest"},
		{"--- FAIL: TestSomething/with_underscore (0.00s)", "TestSomething/with_underscore"},
		{"--- FAIL: TestSlowTest (12.345s)", "TestSlowTest"},
		{"--- FAIL: TestNoDuration", "TestNoDuration"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, extractTestName(tt.line))
		})
	}
}

func TestParseCoverage(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected float64
	}{
		{"empty output", "", 0},
		{"no coverage info", "ok  \tgithub.com/example/pkg\t0.5s", 0},
		{"single package coverage", "ok  \tgithub.com/example/pkg\t0.5s\tcoverage: 80.0% of statements", 80.0},
		{"multiple packages averaged", "ok  \tpkg1\t0.1s\tcoverage: 60.0% of statements\nok  \tpkg2\t0.1s\tcoverage: 80.0% of statements", 70.0},
		{"100% coverage", "ok  \tpkg\t0.1s\tcoverage: 100.0% of statements", 100.0},
		{"0% coverage", "ok  \tpkg\t0.1s\tcoverage: 0.0% of statements", 0.0},
		{"decimal coverage", "ok  \tpkg\t0.1s\tcoverage: 75.5% of statements", 75.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.expected, parseCoverage(tt.output), 0.01)
		})
	}
}
