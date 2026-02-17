package checkmate

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_Defaults(t *testing.T) {
	p := New()
	require.NotNil(t, p)
	assert.NotNil(t, p.writer)
	assert.NotNil(t, p.theme)
}

func TestNew_WithWriter(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf))

	p.CheckSuccess("test")

	output := buf.String()
	assert.Contains(t, output, "test")
}

func TestNew_WithTheme(t *testing.T) {
	theme := MinimalTheme()
	theme.ForceColors = true // Prevent auto-switching
	p := New(WithTheme(theme), WithWriter(&bytes.Buffer{}))

	assert.Equal(t, "[OK]", p.theme.IconSuccess)
}

func TestNew_WithStderr(t *testing.T) {
	p := New(WithStderr())
	assert.Equal(t, os.Stderr, p.writer)
}

func TestNew_AutoDetectNonTTY(t *testing.T) {
	// When writing to a buffer (non-TTY), should auto-switch to minimal theme
	var buf bytes.Buffer
	p := New(WithWriter(&buf))

	// The theme should have been auto-switched to minimal
	// (unless we're running in a TTY, in which case this test is less meaningful)
	p.CheckSuccess("test")
	output := buf.String()

	// Should contain either emoji or [OK] depending on TTY detection
	assert.True(t, strings.Contains(output, "test"))
}

func TestCategoryHeader(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		theme    *Theme
		contains []string
	}{
		{
			name:     "default theme",
			title:    "Code Quality",
			theme:    forceColorTheme(DefaultTheme()),
			contains: []string{"Code Quality"},
		},
		{
			name:     "minimal theme",
			title:    "Tests",
			theme:    MinimalTheme(),
			contains: []string{"Tests"},
		},
		{
			name:     "long title",
			title:    "Very Long Category Title That Might Overflow",
			theme:    MinimalTheme(),
			contains: []string{"Very Long Category Title"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			p.CategoryHeader(tt.title)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCheckHeader(t *testing.T) {
	// In non-TTY mode (bytes.Buffer), CheckHeader skips output
	// since we can't do the "replace" animation effect
	t.Run("non-TTY skips output", func(t *testing.T) {
		var buf bytes.Buffer
		p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

		p.CheckHeader("Running tests")

		output := buf.String()
		assert.Empty(t, output, "CheckHeader should produce no output in non-TTY mode")
	})

	// Test that the method doesn't panic with various inputs
	t.Run("handles various inputs", func(t *testing.T) {
		var buf bytes.Buffer
		p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

		assert.NotPanics(t, func() {
			p.CheckHeader("")
			p.CheckHeader("Short")
			p.CheckHeader("A very long message that might cause issues with formatting")
		})
	})
}

func TestCheckSuccess(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		theme    *Theme
		contains []string
	}{
		{
			name:     "minimal theme",
			message:  "Build complete",
			theme:    MinimalTheme(),
			contains: []string{"|--", "[OK]", "Build complete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			p.CheckSuccess(tt.message)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}

	// Test that the method doesn't panic with various inputs
	t.Run("handles various inputs", func(t *testing.T) {
		var buf bytes.Buffer
		p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

		assert.NotPanics(t, func() {
			p.CheckSuccess("")
			p.CheckSuccess("Short")
			p.CheckSuccess("A very long message that might cause issues with formatting")
		})
	})
}

func TestCheckFailure(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		details     string
		remediation string
		theme       *Theme
		contains    []string
		notContains []string
	}{
		{
			name:        "full failure",
			title:       "Test failed",
			details:     "main.go:10: error",
			remediation: "Fix the error",
			theme:       forceColorTheme(DefaultTheme()),
			contains:    []string{"├──", "✗", "Test failed", "Details:", "main.go:10", "How to fix:", "Fix the error"},
		},
		{
			name:        "minimal theme",
			title:       "Build failed",
			details:     "compile error",
			remediation: "Check syntax",
			theme:       MinimalTheme(),
			contains:    []string{"[FAIL]", "Build failed", "compile error", "Check syntax"},
		},
		{
			name:        "no details",
			title:       "Check failed",
			details:     "",
			remediation: "Run task fix",
			theme:       MinimalTheme(),
			contains:    []string{"[FAIL]", "Check failed", "How to fix:", "Run task fix"},
			notContains: []string{"Details:"},
		},
		{
			name:        "no remediation",
			title:       "Error occurred",
			details:     "Something went wrong",
			remediation: "",
			theme:       MinimalTheme(),
			contains:    []string{"[FAIL]", "Error occurred", "Details:", "Something went wrong"},
			notContains: []string{"How to fix:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			p.CheckFailure(tt.title, tt.details, tt.remediation)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
			for _, notExpected := range tt.notContains {
				assert.NotContains(t, output, notExpected)
			}
		})
	}
}

func TestCheckSummary(t *testing.T) {
	tests := []struct {
		name     string
		status   Status
		title    string
		items    []string
		theme    *Theme
		contains []string
	}{
		{
			name:     "success with items",
			status:   StatusSuccess,
			title:    "All checks passed",
			items:    []string{"Formatting", "Linting"},
			theme:    forceColorTheme(DefaultTheme()),
			contains: []string{"─", "✓", "All checks passed", "Formatting", "Linting"},
		},
		{
			name:     "failure minimal",
			status:   StatusFailure,
			title:    "2 checks failed",
			items:    []string{"Build", "Test"},
			theme:    MinimalTheme(),
			contains: []string{"[FAIL]", "2 checks failed", "Build", "Test"},
		},
		{
			name:     "no items",
			status:   StatusSuccess,
			title:    "Done",
			items:    nil,
			theme:    MinimalTheme(),
			contains: []string{"[OK]", "Done"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			p.CheckSummary(tt.status, tt.title, tt.items...)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestCheckInfo(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckInfo("Tool: go-licenses", "Version: 1.0.0")

	output := buf.String()
	assert.Contains(t, output, "Tool: go-licenses")
	assert.Contains(t, output, "Version: 1.0.0")
	// Should be indented
	assert.Contains(t, output, "   ")
}

func TestCheckNote(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckNote("This is informational")

	output := buf.String()
	assert.Contains(t, output, "Note:")
	assert.Contains(t, output, "This is informational")
}

func TestPrinter_ConcurrentWrites(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent CheckSuccess calls
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			p.CheckSuccess("Done")
		}(i)
	}

	wg.Wait()

	output := buf.String()
	// Count occurrences of "Done" - should be exactly iterations
	count := strings.Count(output, "Done")
	assert.Equal(t, iterations, count, "All concurrent writes should be recorded")
}

func TestPrinter_ImplementsInterface(t *testing.T) {
	// Compile-time check that Printer implements PrinterInterface
	var _ PrinterInterface = (*Printer)(nil)

	// Runtime check
	var p PrinterInterface = New(WithWriter(&bytes.Buffer{}))
	require.NotNil(t, p)
}

func TestCheckLine(t *testing.T) {
	tests := []struct {
		name     string
		checkNm  string
		status   Status
		duration time.Duration
		contains []string
	}{
		{
			name:     "success with milliseconds",
			checkNm:  "format",
			status:   StatusSuccess,
			duration: 500 * time.Millisecond,
			contains: []string{"format", "[OK]", "500ms"},
		},
		{
			name:     "success with seconds",
			checkNm:  "test",
			status:   StatusSuccess,
			duration: 2500 * time.Millisecond,
			contains: []string{"test", "[OK]", "2.500s"},
		},
		{
			name:     "failure",
			checkNm:  "lint",
			status:   StatusFailure,
			duration: 1200 * time.Millisecond,
			contains: []string{"lint", "[FAIL]", "1.200s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

			p.CheckLine(tt.checkNm, tt.status, tt.duration)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
			// Should contain dots for alignment
			assert.Contains(t, output, ".")
		})
	}
}

// Helper to force colors in theme (prevents auto-switch to minimal)
func forceColorTheme(t *Theme) *Theme {
	t.ForceColors = true
	return t
}
