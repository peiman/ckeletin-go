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
//   - DefaultTheme(): Unicode icons (○ ✓ ✗ • !) and tree connectors with
//     colors for interactive terminals
//   - MinimalTheme(): Plain ASCII for CI/CD pipelines and piped output
//
// TTY detection is automatic - when no theme was chosen explicitly and
// output is piped or redirected, checkmate switches to MinimalTheme.
// A non-nil theme passed via WithTheme is always honored. Color emission
// is decided by the lipgloss renderer for the writer, so colorful themes
// still print without ANSI codes on writers that do not support them.
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
// The examples below show the default theme (colors omitted).
//
// CategoryHeader("Code Quality") prints the title surrounded by blank
// lines, with a colored background on terminals:
//
//	Code Quality
//
// CheckHeader("format") prints an in-progress line on terminals only,
// without a trailing newline so the final result overwrites it; piped
// or redirected output gets nothing:
//
//	├── ○ format
//
// CheckSuccess("All files formatted"):
//
//	├── ✓ All files formatted
//
// CheckFailure("Format check failed", "main.go:10: line too long", "Run: task format"):
//
//	├── ✗ Format check failed
//	│   Details:
//	│     main.go:10: line too long
//	│   How to fix:
//	│     • Run: task format
//	│
//
// CheckSummary(StatusSuccess, "All checks passed (5/5)", "Formatting", "Linting", "Tests"):
//
//	╭────────────────────────────────────────────────╮
//	│                                                │
//	│           ✓ All checks passed (5/5)            │
//	│                                                │
//	│  ├── ✓ Formatting                              │
//	│  ├── ✓ Linting                                 │
//	│  └── ✓ Tests                                   │
//	│                                                │
//	╰────────────────────────────────────────────────╯
//
// With MinimalTheme the same calls render plain ASCII:
//
//	|-- [OK] All files formatted
//
//	|-- [FAIL] Format check failed
//	|   Details:
//	|     main.go:10: line too long
//	|   How to fix:
//	|     * Run: task format
//	|
//
//	+===========================================+
//	|                                           |
//	|       [OK] All checks passed (5/5)        |
//	|                                           |
//	|  |-- [OK] Formatting                      |
//	|  |-- [OK] Linting                         |
//	|  `-- [OK] Tests                           |
//	|                                           |
//	+===========================================+
//
// # Check Runner
//
// Use Runner to orchestrate multiple checks with automatic output handling:
//
//	p := checkmate.New()
//	result := checkmate.NewRunner(p, checkmate.WithCategory("Code Quality")).
//	    AddFunc("format", checkFormat).WithRemediation("Run: task format").
//	    AddFunc("lint", checkLint).WithRemediation("Run: task lint").
//	    Run(context.Background())
//
//	if !result.Success() {
//	    os.Exit(1)
//	}
//
// The runner automatically:
//   - Displays category headers
//   - Shows check progress with CheckHeader
//   - Reports success/failure for each check
//   - Generates a summary at the end
//   - Recovers from panics (converts to failed checks)
//   - Respects context cancellation
//
// # Runner Options
//
//	// Stop on first failure
//	runner := checkmate.NewRunner(p, checkmate.WithFailFast())
//
//	// Set category header
//	runner := checkmate.NewRunner(p, checkmate.WithCategory("Tests"))
//
//	// Run checks in parallel (all at once)
//	runner := checkmate.NewRunner(p, checkmate.WithParallel())
//
//	// Run checks in parallel with limited concurrency
//	runner := checkmate.NewRunner(p, checkmate.WithWorkers(3))
//
// # Parallel Execution
//
// Use WithParallel() to run checks concurrently for faster execution:
//
//	result := checkmate.NewRunner(p, checkmate.WithParallel()).
//	    AddFunc("format", checkFormat).
//	    AddFunc("lint", checkLint).    // Runs concurrently with format
//	    AddFunc("test", checkTest).    // Runs concurrently with both
//	    Run(ctx)
//
// Parallel execution maintains result order - even though checks may complete
// in any order, the results and output are reported in the original order.
//
// Use WithWorkers(n) to limit concurrency:
//
//	// Run at most 2 checks at a time
//	runner := checkmate.NewRunner(p, checkmate.WithWorkers(2))
//
// Fail-fast works with parallel execution - when enabled, remaining checks
// are cancelled after the first failure:
//
// # Fluent API
//
// The runner supports a fluent API for easy check definition:
//
//	result := checkmate.NewRunner(p).
//	    AddFunc("check1", func(ctx context.Context) error {
//	        return nil // success
//	    }).WithRemediation("Fix instruction").
//	    Add(checkmate.Check{
//	        Name:        "check2",
//	        Fn:          check2Func,
//	        Remediation: "How to fix",
//	        Details:     "Additional context shown on failure",
//	    }).
//	    Run(ctx)
//
// # Testing Runners
//
// Use MockPrinter to test code that uses Runner:
//
//	func TestChecks(t *testing.T) {
//	    mock := checkmate.NewMockPrinter()
//	    result := checkmate.NewRunner(mock).
//	        AddFunc("test", func(ctx context.Context) error { return nil }).
//	        Run(context.Background())
//
//	    assert.True(t, result.Success())
//	    assert.True(t, mock.HasCall("CheckSuccess"))
//	}
package checkmate
