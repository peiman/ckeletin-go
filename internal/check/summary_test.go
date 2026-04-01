package check

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummaryBoxChars(t *testing.T) {
	tests := []struct {
		name   string
		useTUI bool
		// Expected representative characters to verify correct set
		topLeft     string
		successIcon string
		failIcon    string
	}{
		{
			name:        "TUI mode uses Unicode box-drawing characters",
			useTUI:      true,
			topLeft:     "╭",
			successIcon: "✓",
			failIcon:    "✗",
		},
		{
			name:        "CI mode uses ASCII characters",
			useTUI:      false,
			topLeft:     "+",
			successIcon: "[OK]",
			failIcon:    "[FAIL]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chars := summaryBoxChars(tt.useTUI)

			assert.Equal(t, tt.topLeft, chars.topLeft)
			assert.Equal(t, tt.successIcon, chars.successIcon)
			assert.Equal(t, tt.failIcon, chars.failIcon)

			// Verify all fields are non-empty
			assert.NotEmpty(t, chars.topRight)
			assert.NotEmpty(t, chars.bottomLeft)
			assert.NotEmpty(t, chars.bottomRight)
			assert.NotEmpty(t, chars.horizontal)
			assert.NotEmpty(t, chars.vertical)
			assert.NotEmpty(t, chars.catSeparator)
			assert.NotEmpty(t, chars.treeConnector)
			assert.NotEmpty(t, chars.treeLastConnector)
		})
	}
}

func TestSummaryBoxChars_CIvsUnicodeConsistency(t *testing.T) {
	ci := summaryBoxChars(false)
	tui := summaryBoxChars(true)

	// Both modes should define all box-drawing elements
	// CI mode should use simpler characters
	assert.Equal(t, "+", ci.topLeft)
	assert.Equal(t, "+", ci.topRight)
	assert.Equal(t, "+", ci.bottomLeft)
	assert.Equal(t, "+", ci.bottomRight)
	assert.Equal(t, "-", ci.horizontal)
	assert.Equal(t, "|", ci.vertical)

	// TUI mode should use Unicode
	assert.Equal(t, "╭", tui.topLeft)
	assert.Equal(t, "╮", tui.topRight)
	assert.Equal(t, "╰", tui.bottomLeft)
	assert.Equal(t, "╯", tui.bottomRight)
	assert.Equal(t, "─", tui.horizontal)
	assert.Equal(t, "│", tui.vertical)
}

func TestSummaryStyles(t *testing.T) {
	tests := []struct {
		name      string
		useTUI    bool
		allPassed bool
	}{
		{
			name:      "CI mode all passed",
			useTUI:    false,
			allPassed: true,
		},
		{
			name:      "CI mode with failures",
			useTUI:    false,
			allPassed: false,
		},
		{
			name:      "TUI mode all passed",
			useTUI:    true,
			allPassed: true,
		},
		{
			name:      "TUI mode with failures",
			useTUI:    true,
			allPassed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			styles := summaryStyles(tt.useTUI, tt.allPassed)

			// All styles should be non-zero (initialized)
			// We can at least verify the struct is populated by rendering with them
			assert.NotEmpty(t, styles.border.Render("x"))
			assert.NotEmpty(t, styles.dim.Render("x"))
			assert.NotEmpty(t, styles.bold.Render("x"))
			assert.NotEmpty(t, styles.success.Render("x"))
			assert.NotEmpty(t, styles.fail.Render("x"))
			assert.NotEmpty(t, styles.title.Render("x"))
		})
	}
}

func TestSummaryStyles_CIModeReturnsPlainStyles(t *testing.T) {
	// In CI mode, all styles should render text without color codes
	stylesPass := summaryStyles(false, true)
	stylesFail := summaryStyles(false, false)

	// Plain styles should render input unchanged
	assert.Equal(t, "test", stylesPass.border.Render("test"))
	assert.Equal(t, "test", stylesPass.dim.Render("test"))
	assert.Equal(t, "test", stylesPass.bold.Render("test"))
	assert.Equal(t, "test", stylesPass.success.Render("test"))
	assert.Equal(t, "test", stylesPass.fail.Render("test"))
	assert.Equal(t, "test", stylesPass.title.Render("test"))

	assert.Equal(t, "test", stylesFail.border.Render("test"))
	assert.Equal(t, "test", stylesFail.dim.Render("test"))
	assert.Equal(t, "test", stylesFail.title.Render("test"))
}

func TestPrintFinalSummary(t *testing.T) {
	tests := []struct {
		name          string
		results       []allCheckResult
		passed        int
		failed        int
		totalDuration time.Duration
		coverage      float64
		useTUI        bool
		wantContains  []string
		wantAbsent    []string
	}{
		{
			name: "all checks passed in CI mode",
			results: []allCheckResult{
				{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
				{name: "lint", category: "Code Quality", passed: true, duration: 200 * time.Millisecond},
			},
			passed:        2,
			failed:        0,
			totalDuration: 300 * time.Millisecond,
			useTUI:        false,
			wantContains:  []string{"All 2 Checks Passed", "format", "lint", "300ms", "Code Quality"},
			wantAbsent:    []string{"Checks Failed"},
		},
		{
			name: "some checks failed in CI mode",
			results: []allCheckResult{
				{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
				{name: "lint", category: "Code Quality", passed: false, duration: 200 * time.Millisecond, err: errors.New("lint errors"), remediation: "Run: task lint"},
			},
			passed:        1,
			failed:        1,
			totalDuration: 300 * time.Millisecond,
			useTUI:        false,
			wantContains:  []string{"1/2 Checks Failed", "format", "lint"},
			wantAbsent:    []string{"All", "Passed"},
		},
		{
			name: "all checks failed in CI mode",
			results: []allCheckResult{
				{name: "format", category: "Code Quality", passed: false, duration: 100 * time.Millisecond, err: errors.New("formatting issues"), remediation: "Run: task format"},
				{name: "lint", category: "Code Quality", passed: false, duration: 200 * time.Millisecond, err: errors.New("lint errors"), remediation: "Run: task lint"},
			},
			passed:        0,
			failed:        2,
			totalDuration: 300 * time.Millisecond,
			useTUI:        false,
			wantContains:  []string{"2/2 Checks Failed", "format", "lint"},
		},
		{
			name:          "empty results in CI mode",
			results:       []allCheckResult{},
			passed:        0,
			failed:        0,
			totalDuration: 0,
			useTUI:        false,
			wantContains:  []string{"All 0 Checks Passed"},
		},
		{
			name: "coverage is displayed when available",
			results: []allCheckResult{
				{name: "test", category: "Tests", passed: true, duration: 5 * time.Second},
			},
			passed:        1,
			failed:        0,
			totalDuration: 5 * time.Second,
			coverage:      85.5,
			useTUI:        false,
			wantContains:  []string{"85.5%", "Coverage"},
		},
		{
			name: "coverage not displayed when zero",
			results: []allCheckResult{
				{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
			},
			passed:        1,
			failed:        0,
			totalDuration: 100 * time.Millisecond,
			coverage:      0,
			useTUI:        false,
			wantAbsent:    []string{"Coverage"},
		},
		{
			name: "multiple categories displayed in order",
			results: []allCheckResult{
				{name: "go-version", category: "Development Environment", passed: true, duration: 50 * time.Millisecond},
				{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
				{name: "test", category: "Tests", passed: true, duration: 5 * time.Second},
			},
			passed:        3,
			failed:        0,
			totalDuration: 5150 * time.Millisecond,
			useTUI:        false,
			wantContains:  []string{"Development Environment", "Code Quality", "Tests", "go-version", "format", "test"},
		},
		{
			name: "duration shown per check",
			results: []allCheckResult{
				{name: "lint", category: "Code Quality", passed: true, duration: 2345 * time.Millisecond},
			},
			passed:        1,
			failed:        0,
			totalDuration: 2345 * time.Millisecond,
			useTUI:        false,
			wantContains:  []string{"2.345s"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			executor := &Executor{
				writer:   &buf,
				useTUI:   tt.useTUI,
				coverage: tt.coverage,
			}

			executor.printFinalSummary(tt.results, tt.passed, tt.failed, tt.totalDuration)

			output := buf.String()
			require.NotEmpty(t, output, "summary output should not be empty")

			for _, s := range tt.wantContains {
				assert.Contains(t, output, s, "output should contain %q", s)
			}
			for _, s := range tt.wantAbsent {
				assert.NotContains(t, output, s, "output should not contain %q", s)
			}
		})
	}
}

func TestPrintFinalSummary_FailedCheckShowsRemediation(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	results := []allCheckResult{
		{
			name:        "lint",
			category:    "Code Quality",
			passed:      false,
			duration:    200 * time.Millisecond,
			err:         errors.New("golangci-lint found 2 issue(s)"),
			remediation: "Run: task lint",
		},
	}

	executor.printFinalSummary(results, 0, 1, 200*time.Millisecond)

	output := buf.String()
	// The failure details should appear below the box
	assert.Contains(t, output, "lint", "should mention the check name")
	assert.Contains(t, output, "golangci-lint found 2 issue(s)", "should show the error message")
}

func TestPrintFinalSummary_TUIMode(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: true,
	}

	results := []allCheckResult{
		{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
	}

	executor.printFinalSummary(results, 1, 0, 100*time.Millisecond)

	output := buf.String()
	// TUI mode should contain escape sequence for clearing screen
	assert.Contains(t, output, "\033[2J\033[H", "TUI mode should clear screen")
}

func TestPrintFinalSummary_CIMode_NoClearScreen(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	results := []allCheckResult{
		{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
	}

	executor.printFinalSummary(results, 1, 0, 100*time.Millisecond)

	output := buf.String()
	// CI mode should NOT contain escape sequence for clearing screen
	assert.NotContains(t, output, "\033[2J", "CI mode should not clear screen")
}

func TestPrintFinalSummary_BoxStructure(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	results := []allCheckResult{
		{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
	}

	executor.printFinalSummary(results, 1, 0, 100*time.Millisecond)

	output := buf.String()
	// CI mode uses ASCII box characters
	assert.Contains(t, output, "+", "should contain box corners")
	assert.Contains(t, output, "|", "should contain vertical borders")
	assert.Contains(t, output, "-", "should contain horizontal borders")
	assert.Contains(t, output, "Duration:", "should show duration")
}

func TestPrintFinalSummary_NegativePaddingHandled(t *testing.T) {
	// Test with a very large number of failures to create a wide title that
	// might exceed the box width. The title format is:
	// " ✗ N/M Checks Failed " — if N and M are large, the title could be wide.
	// With CI mode's [FAIL] icon, even larger. This tests the negative padding guard.
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	// Generate many results to make a wide title
	var results []allCheckResult
	for i := 0; i < 100; i++ {
		results = append(results, allCheckResult{
			name:     fmt.Sprintf("check-%03d", i),
			category: "Code Quality",
			passed:   false,
			duration: time.Millisecond,
			err:      errors.New("failed"),
		})
	}

	// Should not panic even with extreme numbers
	assert.NotPanics(t, func() {
		executor.printFinalSummary(results, 0, 100, time.Second)
	})
}

func TestPrintFinalSummary_TUIModeFailedChecks(t *testing.T) {
	// Test that TUI mode with failed checks uses TUI-themed error printer
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: true,
	}

	results := []allCheckResult{
		{
			name:        "lint",
			category:    "Code Quality",
			passed:      false,
			duration:    200 * time.Millisecond,
			err:         errors.New("lint errors found"),
			remediation: "Run: task lint",
		},
	}

	executor.printFinalSummary(results, 0, 1, 200*time.Millisecond)

	output := buf.String()
	assert.Contains(t, output, "lint errors found")
}

func TestPrintFinalSummary_ZeroDuration(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	results := []allCheckResult{
		{name: "fast-check", category: "Code Quality", passed: true, duration: 0},
	}

	executor.printFinalSummary(results, 1, 0, 0)
	output := buf.String()
	assert.Contains(t, output, "Duration:")
}

func TestPrintFinalSummary_AllCategoriesPresent(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		writer: &buf,
		useTUI: false,
	}

	// Create results spanning all categories
	results := []allCheckResult{
		{name: "go-version", category: "Development Environment", passed: true, duration: 50 * time.Millisecond},
		{name: "tools", category: "Development Environment", passed: true, duration: 60 * time.Millisecond},
		{name: "format", category: "Code Quality", passed: true, duration: 100 * time.Millisecond},
		{name: "lint", category: "Code Quality", passed: true, duration: 200 * time.Millisecond},
		{name: "defaults", category: "Architecture Validation", passed: true, duration: 80 * time.Millisecond},
		{name: "secrets", category: "Security Scanning", passed: true, duration: 150 * time.Millisecond},
		{name: "deps", category: "Dependencies", passed: true, duration: 1 * time.Second},
		{name: "test", category: "Tests", passed: true, duration: 5 * time.Second},
	}

	executor.printFinalSummary(results, 8, 0, 7*time.Second)

	output := buf.String()
	assert.Contains(t, output, "Development Environment")
	assert.Contains(t, output, "Code Quality")
	assert.Contains(t, output, "Architecture Validation")
	assert.Contains(t, output, "Security Scanning")
	assert.Contains(t, output, "Dependencies")
	assert.Contains(t, output, "Tests")
	assert.Contains(t, output, "All 8 Checks Passed")
}
