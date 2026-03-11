package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

// shouldUseTUI determines whether to use the interactive TUI or simple output.
// Returns false for CI environments, piped output, or non-TTY contexts.
func shouldUseTUI(w io.Writer) bool {
	// Check common CI environment variables
	ciEnvVars := []string{
		"CI",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"JENKINS_URL",
		"CIRCLECI",
		"TRAVIS",
		"BUILDKITE",
		"TF_BUILD", // Azure DevOps
	}
	for _, env := range ciEnvVars {
		if os.Getenv(env) != "" {
			return false
		}
	}

	// Check if NO_COLOR is set (standard for disabling colors)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check if TERM is dumb (minimal terminal)
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	// Check if output is a TTY
	return checkmate.IsTerminal(w)
}

// Executor runs checks with a Bubble Tea progress UI or simple output.
type Executor struct {
	cfg      Config
	writer   io.Writer
	checks   []checkItem
	program  *tea.Program   // Reference to the Bubble Tea program for sending messages
	timings  *timingHistory // Historical timing data for progress estimation
	coverage float64        // Code coverage percentage from test run
	useTUI   bool           // Whether to use interactive TUI or simple output
}

type checkItem struct {
	name        string
	fn          func(ctx context.Context) error
	remediation string
}

// categoryDef defines a category with its checks
type categoryDef struct {
	name   string
	checks []checkItem
}

// NewExecutor creates a new executor with TUI or simple output based on environment.
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	e := &Executor{
		cfg:     cfg,
		writer:  writer,
		timings: loadTimingHistory(),
		useTUI:  shouldUseTUI(writer),
	}

	// Build the full check list from all categories
	methods := &checkMethods{cfg: cfg}
	categories := e.buildCategories(methods)

	// Flatten for backwards compatibility (single category mode)
	for _, cat := range categories {
		e.checks = append(e.checks, cat.checks...)
	}

	return e
}

// buildCategories returns all check categories with their checks
func (e *Executor) buildCategories(methods *checkMethods) []categoryDef {
	return []categoryDef{
		{
			name: "Development Environment",
			checks: []checkItem{
				{"go-version", methods.shellCheck("check-go-version.sh"), "Ensure Go version matches .go-version"},
				{"tools", methods.shellCheck("install_tools.sh", "--check"), "Run: task setup"},
			},
		},
		{
			name: "Code Quality",
			checks: []checkItem{
				{"format", methods.checkFormat, "Run: task format"},
				{"lint", methods.checkLint, "Run: task lint"},
			},
		},
		{
			name: "Architecture Validation",
			checks: []checkItem{
				{"defaults", methods.shellCheck("check-defaults.sh"), "Use registry for SetDefault (ADR-002)"},
				{"commands", methods.shellCheck("validate-command-patterns.sh"), "Keep commands ultra-thin (ADR-001)"},
				{"constants", methods.shellCheck("check-constants.sh"), "Run: task generate:config:key-constants"},
				{"task-naming", methods.shellCheck("validate-task-naming.sh"), "Follow ADR-000 naming convention"},
				{"architecture", methods.shellCheck("validate-architecture.sh"), "Update ARCHITECTURE.md (ADR-008)"},
				{"layering", methods.shellCheck("validate-layering.sh"), "Fix layer dependencies (ADR-009)"},
				{"package-org", methods.shellCheck("validate-package-organization.sh"), "Follow package organization (ADR-010)"},
				{"config-consumption", methods.shellCheck("validate-config-consumption.sh"), "Use type-safe config (ADR-002)"},
				{"output-patterns", methods.shellCheck("validate-output-patterns.sh"), "Follow output patterns (ADR-012)"},
				{"security-patterns", methods.shellCheck("validate-security-patterns.sh"), "Implement security patterns (ADR-004)"},
			},
		},
		{
			name: "Security Scanning",
			checks: []checkItem{
				{"secrets", methods.shellCheck("check-secrets.sh"), "Remove hardcoded secrets"},
				{"sast", methods.shellCheck("check-sast.sh"), "Fix SAST issues or update .semgrep.yml"},
			},
		},
		{
			name: "Dependencies",
			checks: []checkItem{
				{"deps", methods.checkDeps, "Run: go mod tidy"},
				{"vuln", methods.checkVuln, "Update vulnerable dependencies"},
				{"outdated", methods.shellCheck("check-deps-outdated.sh"), "Run: go get -u"},
				{"license-source", methods.shellCheck("check-licenses-source.sh"), "Check dependency licenses"},
				{"license-binary", methods.shellCheck("check-licenses-binary.sh"), "Check binary licenses"},
				{"sbom-vulns", methods.shellCheck("check-sbom-vulns.sh"), "Fix SBOM vulnerabilities"},
			},
		},
		{
			name: "Tests",
			checks: []checkItem{
				{"test", e.checkTest, "Fix failing tests"},
			},
		},
	}
}

// allCheckResult stores result info for final summary
type allCheckResult struct {
	name        string
	category    string
	passed      bool
	duration    time.Duration
	err         error
	remediation string
}

// Execute runs all checks with TUI progress display or simple output.
// Uses TUI for interactive terminals, simple output for CI/pipes.
func (e *Executor) Execute(ctx context.Context) error {
	methods := &checkMethods{cfg: e.cfg}
	categories := e.buildCategories(methods)

	var allResults []allCheckResult
	var totalPassed, totalFailed int
	startTime := time.Now()

	for _, category := range categories {
		// Skip empty categories or filtered out
		if len(category.checks) == 0 {
			continue
		}

		// Check category filter
		if len(e.cfg.Categories) > 0 && !e.shouldRunCategory(category.name) {
			continue
		}

		// Choose execution mode based on environment.
		// Parallel mode runs in simple mode to keep check execution concurrent
		// without conflicting with per-check TUI animation semantics.
		var results []allCheckResult
		var err error
		useTUI := e.useTUI && !e.cfg.Parallel
		if useTUI {
			results, err = e.runCategoryTUI(ctx, category)
		} else {
			results, err = e.runCategorySimple(ctx, category)
		}
		allResults = append(allResults, results...)

		for _, r := range results {
			if r.passed {
				totalPassed++
			} else {
				totalFailed++
			}
		}

		if err != nil && e.cfg.FailFast {
			break
		}
	}

	// Save timing history for next run
	e.timings.save()

	// Print final summary
	e.printFinalSummary(allResults, totalPassed, totalFailed, time.Since(startTime))

	if totalFailed > 0 {
		return fmt.Errorf("%d checks failed", totalFailed)
	}
	return nil
}

// shouldRunCategory checks if a category matches the filter
func (e *Executor) shouldRunCategory(categoryName string) bool {
	// Map display names to filter names
	categoryMap := map[string]string{
		"Development Environment": CategoryEnvironment,
		"Code Quality":            CategoryQuality,
		"Architecture Validation": CategoryArchitecture,
		"Security Scanning":       CategorySecurity,
		"Dependencies":            CategoryDependencies,
		"Tests":                   CategoryTests,
	}

	filterName, ok := categoryMap[categoryName]
	if !ok {
		return true // Unknown category, run it
	}

	for _, c := range e.cfg.Categories {
		if strings.EqualFold(c, filterName) {
			return true
		}
	}
	return false
}

// runCategoryTUI runs a single category with TUI progress display
func (e *Executor) runCategoryTUI(ctx context.Context, category categoryDef) ([]allCheckResult, error) {
	// Get check names for this category
	names := make([]string, len(category.checks))
	for i, c := range category.checks {
		names[i] = c.name
	}

	// Create the progress model for this category (skip per-category summary)
	model := checkmate.NewProgressModel(category.name, names, checkmate.WithSkipSummary())

	// Create the program with options for non-TTY compatibility
	p := tea.NewProgram(model,
		tea.WithOutput(e.writer),
		tea.WithInput(nil), // Don't try to open /dev/tty for input
	)
	e.program = p // Store reference for coverage callback

	var results []allCheckResult
	var mu sync.Mutex
	var categoryErr error

	// Run checks in a goroutine
	go func() {
		for i, check := range category.checks {
			// Start progress animation with expected duration from history
			done := make(chan struct{})
			go e.animateProgress(p, i, check.name, done)

			// Run the check
			start := time.Now()
			checkErr := check.fn(ctx)
			duration := time.Since(start)

			// Stop progress animation
			close(done)

			// Record timing for future runs
			e.timings.recordDuration(check.name, duration)

			// Store result
			mu.Lock()
			result := allCheckResult{
				name:        check.name,
				category:    category.name,
				duration:    duration,
				remediation: check.remediation,
			}

			if checkErr != nil {
				result.passed = false
				result.err = checkErr
				categoryErr = checkErr
				p.Send(checkmate.CheckUpdateMsg{
					Index:       i,
					Status:      checkmate.CheckFailed,
					Progress:    1.0,
					Duration:    duration,
					Error:       checkErr,
					Remediation: check.remediation,
				})
			} else {
				result.passed = true
				p.Send(checkmate.CheckUpdateMsg{
					Index:    i,
					Status:   checkmate.CheckPassed,
					Progress: 1.0,
					Duration: duration,
				})
			}
			results = append(results, result)
			mu.Unlock()

			if checkErr != nil && e.cfg.FailFast {
				break
			}

			// Small delay to see the progress
			time.Sleep(50 * time.Millisecond)
		}

		// Signal done
		time.Sleep(100 * time.Millisecond)
		p.Send(checkmate.DoneMsg{})
	}()

	// Run the TUI
	if _, runErr := p.Run(); runErr != nil {
		return results, fmt.Errorf("TUI error: %w", runErr)
	}

	return results, categoryErr
}

// runCategorySimple runs a single category with simple text output (no TUI).
// Used in CI environments, piped output, or non-TTY contexts.
func (e *Executor) runCategorySimple(ctx context.Context, category categoryDef) ([]allCheckResult, error) {
	// Create a printer with CI-appropriate theme
	printer := checkmate.New(
		checkmate.WithWriter(e.writer),
		checkmate.WithTheme(checkmate.CITheme()),
	)

	// Print category header
	printer.CategoryHeader(category.name)

	var results []allCheckResult
	var categoryErr error

	// Sequential execution when parallel mode is disabled.
	if !e.cfg.Parallel {
		for _, check := range category.checks {
			// Run the check
			start := time.Now()
			checkErr := check.fn(ctx)
			duration := time.Since(start)

			// Record timing for future runs
			e.timings.recordDuration(check.name, duration)

			// Build result
			result := allCheckResult{
				name:        check.name,
				category:    category.name,
				duration:    duration,
				remediation: check.remediation,
			}

			// Print result line
			if checkErr != nil {
				result.passed = false
				result.err = checkErr
				categoryErr = checkErr
				printer.CheckLine(check.name, checkmate.StatusFailure, duration)
			} else {
				result.passed = true
				printer.CheckLine(check.name, checkmate.StatusSuccess, duration)
			}
			results = append(results, result)

			if checkErr != nil && e.cfg.FailFast {
				break
			}
		}

		return results, categoryErr
	}

	// Parallel execution path.
	type checkResult struct {
		index    int
		duration time.Duration
		err      error
	}

	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	resultCh := make(chan checkResult, len(category.checks))
	var wg sync.WaitGroup

	for i, check := range category.checks {
		wg.Add(1)
		go func(idx int, item checkItem) {
			defer wg.Done()
			start := time.Now()
			checkErr := item.fn(runCtx)
			duration := time.Since(start)

			if checkErr != nil && e.cfg.FailFast {
				cancel()
			}

			resultCh <- checkResult{
				index:    idx,
				duration: duration,
				err:      checkErr,
			}
		}(i, check)
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	ordered := make([]checkResult, len(category.checks))
	for r := range resultCh {
		ordered[r.index] = r
	}

	for i, check := range category.checks {
		r := ordered[i]

		// Record timing for future runs
		e.timings.recordDuration(check.name, r.duration)

		result := allCheckResult{
			name:        check.name,
			category:    category.name,
			duration:    r.duration,
			remediation: check.remediation,
		}

		if r.err != nil {
			result.passed = false
			result.err = r.err
			if categoryErr == nil {
				categoryErr = r.err
			}
			printer.CheckLine(check.name, checkmate.StatusFailure, r.duration)
		} else {
			result.passed = true
			printer.CheckLine(check.name, checkmate.StatusSuccess, r.duration)
		}
		results = append(results, result)
	}

	return results, categoryErr
}

// checkTest wraps the checkMethods' checkTest to send coverage updates to the TUI.
func (e *Executor) checkTest(ctx context.Context) error {
	methods := &checkMethods{
		cfg: e.cfg,
		onCoverage: func(coverage float64) {
			e.coverage = coverage // Store for final summary
			if e.program != nil {
				e.program.Send(checkmate.CoverageMsg{Coverage: coverage})
			}
		},
	}
	return methods.checkTest(ctx)
}

// animateProgress sends progress updates based on historical timing data
func (e *Executor) animateProgress(p *tea.Program, idx int, checkName string, done <-chan struct{}) {
	expectedDuration := e.timings.getExpectedDuration(checkName)
	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			// Calculate progress as percentage of expected duration
			// Cap at 95% to leave room for completion
			progress := float64(elapsed) / float64(expectedDuration)
			if progress > 0.95 {
				progress = 0.95
			}
			p.Send(checkmate.CheckUpdateMsg{
				Index:    idx,
				Status:   checkmate.CheckRunning,
				Progress: progress,
			})
		}
	}
}
