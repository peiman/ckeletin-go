// Package checkmate provides beautiful terminal output for developer tools.
//
// Checkmate is designed for CLI tools that run checks, validations, or build
// processes. It provides consistent, visually appealing output with automatic
// TTY detection for graceful degradation in CI environments.
//
// # Basic Usage
//
//	p := checkmate.New()
//	p.CategoryHeader("Code Quality")
//	p.CheckHeader("Checking formatting")
//	p.CheckSuccess("All files properly formatted")
//
// # Themes
//
// Checkmate includes two built-in themes:
//   - DefaultTheme(): Colorful output with emojis for interactive terminals
//   - MinimalTheme(): Plain ASCII for CI/CD pipelines and piped output
//
// TTY detection is automatic - when output is piped or redirected, checkmate
// automatically switches to MinimalTheme unless ForceColors is set.
//
//	// Force a specific theme
//	p := checkmate.New(checkmate.WithTheme(checkmate.MinimalTheme()))
//
// # Writing to Different Outputs
//
//	// Write to stderr
//	p := checkmate.New(checkmate.WithStderr())
//
//	// Write to a buffer (for testing)
//	var buf bytes.Buffer
//	p := checkmate.New(checkmate.WithWriter(&buf))
//
// # Testing with Mock
//
// Use MockPrinter for testing code that uses checkmate:
//
//	func TestMyChecker(t *testing.T) {
//	    mock := checkmate.NewMockPrinter()
//	    myChecker := NewMyChecker(mock) // Accepts PrinterInterface
//	    myChecker.Run()
//
//	    assert.True(t, mock.HasCall("CheckSuccess"))
//	    assert.Equal(t, 1, mock.CallCount("CheckHeader"))
//	}
//
// # Thread Safety
//
// All Printer methods are thread-safe and can be called concurrently.
// This allows multiple goroutines to report progress simultaneously.
//
// # Output Examples
//
// CategoryHeader("Code Quality"):
//
//	─── Code Quality ────────────────────────
//
// CheckHeader("Checking formatting"):
//
//	🔍 Checking formatting...
//
// CheckSuccess("All files formatted"):
//
//	✅ All files formatted
//
// CheckFailure with details:
//
//	❌ Format check failed
//
//	Details:
//	  main.go:10: line too long
//
//	How to fix:
//	  • Run: task format
//
// CheckSummary:
//
//	━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
//	✅ All checks passed (5/5)
//
//	• Formatting
//	• Linting
//	• Tests
//	━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
package checkmate
