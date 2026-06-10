package checkmate

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/charmbracelet/lipgloss"
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
	p := New(WithTheme(theme), WithWriter(&bytes.Buffer{}))

	assert.Equal(t, "[OK]", p.theme.IconSuccess)
}

func TestNew_WithTheme_ExplicitThemeRetainedOnNonTTY(t *testing.T) {
	var buf bytes.Buffer
	theme := DefaultTheme()

	p := New(WithWriter(&buf), WithTheme(theme))

	assert.Same(t, theme, p.theme,
		"a theme explicitly chosen via WithTheme must not be replaced on non-TTY writers")
}

func TestEnsureInit_NonExplicitTheme_NonTTY(t *testing.T) {
	// The auto-degrade only applies when no theme was explicitly chosen.
	// Printer fields are unexported, so a pre-set theme without WithTheme
	// can only happen in-package (e.g. a zero-value Printer).
	t.Run("degrades to MinimalTheme by default", func(t *testing.T) {
		var buf bytes.Buffer
		p := Printer{writer: &buf, theme: DefaultTheme()}

		p.CheckSuccess("test")

		assert.Equal(t, "[OK]", p.theme.IconSuccess,
			"non-explicit theme should degrade to MinimalTheme on non-TTY writers")
	})

	t.Run("ForceColors retains the theme", func(t *testing.T) {
		var buf bytes.Buffer
		theme := DefaultTheme()
		theme.ForceColors = true
		p := Printer{writer: &buf, theme: theme}

		p.CheckSuccess("test")

		assert.Same(t, theme, p.theme,
			"ForceColors should retain a non-explicit theme on non-TTY writers")
	})
}

func TestNew_WithStderr(t *testing.T) {
	p := New(WithStderr())
	assert.Equal(t, os.Stderr, p.writer)
}

func TestNew_WithNilTheme(t *testing.T) {
	var buf bytes.Buffer

	var p *Printer
	assert.NotPanics(t, func() {
		p = New(WithTheme(nil), WithWriter(&buf))
	})

	p.CheckSuccess("still works")
	assert.Contains(t, buf.String(), "still works")
}

func TestPrinter_ZeroValue(t *testing.T) {
	// A zero-value Printer (constructed without New) must not panic;
	// it lazily applies the same defaults New provides.
	var buf bytes.Buffer
	p := Printer{writer: &buf}

	assert.NotPanics(t, func() {
		p.CheckSuccess("zero value")
		p.CheckSummary(StatusSuccess, "Done")
	})
	assert.Contains(t, buf.String(), "zero value")
	assert.Contains(t, buf.String(), "Done")
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
			theme:    DefaultTheme(),
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
			theme:       DefaultTheme(),
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
			theme:    DefaultTheme(),
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

func TestCheckSummary_LongTitle_NoPanic(t *testing.T) {
	tests := []struct {
		name  string
		theme *Theme
		title string
	}{
		{
			name:  "title longer than summary width",
			theme: MinimalTheme(),
			title: strings.Repeat("x", 60),
		},
		{
			name:  "title just over inner width",
			theme: MinimalTheme(),
			title: strings.Repeat("x", 43),
		},
		{
			name: "tiny summary width",
			theme: func() *Theme {
				th := MinimalTheme()
				th.SummaryWidth = 1
				return th
			}(),
			title: "Done",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			assert.NotPanics(t, func() {
				p.CheckSummary(StatusFailure, tt.title, "item")
			})
			assert.Contains(t, buf.String(), tt.title)
		})
	}
}

func TestCheckSummary_BoxAlignment(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		title  string
		items  []string
	}{
		{
			name:   "success box lines are flush",
			status: StatusSuccess,
			title:  "All checks passed",
			items:  []string{"Build", "Test"},
		},
		{
			name:   "failure box lines are flush",
			status: StatusFailure,
			title:  "2 checks failed",
			items:  []string{"Build", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			theme := MinimalTheme()
			p := New(WithWriter(&buf), WithTheme(theme))

			p.CheckSummary(tt.status, tt.title, tt.items...)

			// Minimal theme output is pure ASCII with no escape codes,
			// so every box line must be exactly SummaryWidth chars wide.
			for _, line := range strings.Split(buf.String(), "\n") {
				if line == "" {
					continue
				}
				assert.Len(t, line, theme.SummaryWidth, "line %q", line)
			}
		})
	}
}

func TestCheckSummary_BoxAlignment_UnicodeTheme(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		title  string
		items  []string
	}{
		{
			name:   "success box lines are flush with multibyte icons",
			status: StatusSuccess,
			title:  "All checks passed",
			items:  []string{"Build", "Test"},
		},
		{
			name:   "failure box lines are flush with multibyte icons",
			status: StatusFailure,
			title:  "2 checks failed",
			items:  []string{"Build", "Test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP: default theme uses multibyte icons (✓/✗) whose byte
			// length differs from their single-column display width;
			// WithTheme retains the unicode theme on a non-TTY buffer
			var buf bytes.Buffer
			theme := DefaultTheme()
			p := New(WithWriter(&buf), WithTheme(theme))

			// EXECUTION
			p.CheckSummary(tt.status, tt.title, tt.items...)

			// ASSERTION: every box line must occupy exactly SummaryWidth
			// display columns (lipgloss.Width ignores ANSI codes)
			for _, line := range strings.Split(buf.String(), "\n") {
				if line == "" {
					continue
				}
				assert.Equal(t, theme.SummaryWidth, lipgloss.Width(line), "line %q", line)
			}
		})
	}
}

func TestCheckSummary_UsesSummaryChar(t *testing.T) {
	tests := []struct {
		name     string
		theme    *Theme
		contains string
	}{
		{
			name:     "minimal theme summary char",
			theme:    MinimalTheme(),
			contains: "+" + strings.Repeat("=", 43) + "+",
		},
		{
			name: "custom summary char",
			theme: func() *Theme {
				th := MinimalTheme()
				th.SummaryChar = "~"
				return th
			}(),
			contains: "+" + strings.Repeat("~", 43) + "+",
		},
		{
			name: "empty summary char falls back to ASCII default",
			theme: func() *Theme {
				th := MinimalTheme()
				th.SummaryChar = ""
				return th
			}(),
			contains: "+" + strings.Repeat("-", 43) + "+",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			p := New(WithWriter(&buf), WithTheme(tt.theme))

			p.CheckSummary(StatusSuccess, "Done")

			assert.Contains(t, buf.String(), tt.contains)
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

// newTTYPrinter creates a Printer that simulates TTY mode for testing
// the terminal escape code branches.
func newTTYPrinter(buf *bytes.Buffer) *Printer {
	p := New(WithWriter(buf), WithTheme(DefaultTheme()))
	p.isTerminal = true
	return p
}

func TestCheckHeader_TTY(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckHeader("Running lint")

	output := buf.String()
	// TTY mode should produce output (unlike non-TTY which skips)
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "Running lint")
}

func TestCheckSuccess_TTY(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckSuccess("format passed")

	output := buf.String()
	assert.Contains(t, output, "format passed")
	// TTY mode uses escape codes for line clearing
	assert.Contains(t, output, "\r")
	assert.Contains(t, output, "\033[K")
}

func TestCheckFailure_TTY(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckFailure("lint failed", "2 errors found", "Run: task format")

	output := buf.String()
	assert.Contains(t, output, "lint failed")
	assert.Contains(t, output, "2 errors found")
	assert.Contains(t, output, "Run: task format")
	// TTY mode uses escape codes for line clearing
	assert.Contains(t, output, "\r")
	assert.Contains(t, output, "\033[K")
}

func TestCheckFailure_TTY_NoDetails(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckFailure("check failed", "", "Fix it")

	output := buf.String()
	assert.Contains(t, output, "check failed")
	assert.NotContains(t, output, "Details:")
	assert.Contains(t, output, "Fix it")
}

func TestCheckFailure_TTY_NoRemediation(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckFailure("check failed", "something broke", "")

	output := buf.String()
	assert.Contains(t, output, "check failed")
	assert.Contains(t, output, "something broke")
	assert.NotContains(t, output, "How to fix:")
}

func TestCheckLine_TTY_SkipsOutput(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckLine("format", StatusSuccess, 500*time.Millisecond)

	// TTY mode should skip CheckLine output (animated output handles it)
	assert.Empty(t, buf.String())
}

func TestCheckLine_NonTTY(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckLine("format", StatusSuccess, 1451*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "format")
	assert.Contains(t, output, "[OK]")
	assert.Contains(t, output, "1.451s")
	assert.Contains(t, output, ".") // dot padding
}

func TestCheckLine_NonTTY_Failure(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckLine("lint", StatusFailure, 3*time.Second)

	output := buf.String()
	assert.Contains(t, output, "lint")
	assert.Contains(t, output, "[FAIL]")
}

func TestCheckInfo_MultipleLines(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckInfo("Line 1", "Line 2", "Line 3")

	output := buf.String()
	assert.Contains(t, output, "Line 1")
	assert.Contains(t, output, "Line 2")
	assert.Contains(t, output, "Line 3")
}

func TestCheckNote_WithMessage(t *testing.T) {
	var buf bytes.Buffer
	p := New(WithWriter(&buf), WithTheme(MinimalTheme()))

	p.CheckNote("Remember to run task format")

	output := buf.String()
	assert.Contains(t, output, "Note:")
	assert.Contains(t, output, "Remember to run task format")
}

func TestCheckSummary_TTY_Success(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckSummary(StatusSuccess, "All 5 Checks Passed", "lint", "format", "test")

	output := buf.String()
	assert.Contains(t, output, "All 5 Checks Passed")
	assert.Contains(t, output, "lint")
	assert.Contains(t, output, "format")
	assert.Contains(t, output, "test")
	// TTY mode uses Unicode box drawing
	assert.Contains(t, output, "╭")
	assert.Contains(t, output, "╰")
}

func TestCheckSummary_TTY_Failure(t *testing.T) {
	var buf bytes.Buffer
	p := newTTYPrinter(&buf)

	p.CheckSummary(StatusFailure, "2 Checks Failed", "lint", "test")

	output := buf.String()
	assert.Contains(t, output, "2 Checks Failed")
	assert.Contains(t, output, "lint")
	assert.Contains(t, output, "test")
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{50 * time.Millisecond, "50ms"},
		{1 * time.Second, "1.000s"},
		{1500 * time.Millisecond, "1.500s"},
		{10 * time.Second, "10.000s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, formatDuration(tt.duration))
		})
	}
}
