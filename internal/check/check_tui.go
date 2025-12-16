package check

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/peiman/ckeletin-go/internal/xdg"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

// checkTiming stores historical timing data for a check
type checkTiming struct {
	LastDuration time.Duration `json:"last_duration"`
	AvgDuration  time.Duration `json:"avg_duration"`
	RunCount     int           `json:"run_count"`
}

// timingHistory stores timing data for all checks
type timingHistory struct {
	Checks map[string]*checkTiming `json:"checks"`
	mu     sync.RWMutex
}

// TUIExecutor runs checks with a Bubble Tea progress UI.
type TUIExecutor struct {
	cfg      Config
	writer   io.Writer
	checks   []checkItem
	program  *tea.Program   // Reference to the Bubble Tea program for sending messages
	timings  *timingHistory // Historical timing data for progress estimation
	coverage float64        // Code coverage percentage from test run
}

type checkItem struct {
	name        string
	fn          func(ctx context.Context) error
	remediation string
}

// timingFilePath returns the path to the timing history file.
// Uses XDG cache directory since timing data is ephemeral/regenerable.
func timingFilePath() string {
	path, err := xdg.CacheFile("check-timings.json")
	if err != nil {
		// Fallback to temp dir if XDG not configured
		return filepath.Join(os.TempDir(), "ckeletin-go-check-timings.json")
	}
	return path
}

// loadTimingHistory loads timing data from disk
func loadTimingHistory() *timingHistory {
	th := &timingHistory{Checks: make(map[string]*checkTiming)}

	data, err := os.ReadFile(timingFilePath())
	if err != nil {
		return th // Return empty history if file doesn't exist
	}

	// Ignore JSON errors, just use empty history
	_ = json.Unmarshal(data, th)
	if th.Checks == nil {
		th.Checks = make(map[string]*checkTiming)
	}
	return th
}

// save persists timing data to disk
func (th *timingHistory) save() {
	th.mu.RLock()
	defer th.mu.RUnlock()

	data, err := json.MarshalIndent(th, "", "  ")
	if err != nil {
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(timingFilePath())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return
	}

	// Write atomically with secure permissions
	_ = os.WriteFile(timingFilePath(), data, 0o600)
}

// getExpectedDuration returns the expected duration for a check
func (th *timingHistory) getExpectedDuration(name string) time.Duration {
	th.mu.RLock()
	defer th.mu.RUnlock()

	if t, ok := th.Checks[name]; ok && t.AvgDuration > 0 {
		return t.AvgDuration
	}
	// Default estimates for first run (in seconds)
	defaults := map[string]time.Duration{
		"format": 2 * time.Second,
		"lint":   5 * time.Second,
		"test":   10 * time.Second,
		"deps":   1 * time.Second,
		"vuln":   3 * time.Second,
	}
	if d, ok := defaults[name]; ok {
		return d
	}
	return 5 * time.Second // Generic default
}

// recordDuration updates timing data after a check completes
func (th *timingHistory) recordDuration(name string, duration time.Duration) {
	th.mu.Lock()
	defer th.mu.Unlock()

	t, ok := th.Checks[name]
	if !ok {
		t = &checkTiming{}
		th.Checks[name] = t
	}

	t.LastDuration = duration
	t.RunCount++

	// Update rolling average (exponential moving average with alpha=0.3)
	// This gives more weight to recent runs while considering history
	if t.AvgDuration == 0 {
		t.AvgDuration = duration
	} else {
		alpha := 0.3
		t.AvgDuration = time.Duration(alpha*float64(duration) + (1-alpha)*float64(t.AvgDuration))
	}
}

// categoryDef defines a category with its checks
type categoryDef struct {
	name   string
	checks []checkItem
}

// NewTUIExecutor creates a new TUI-based executor.
func NewTUIExecutor(cfg Config, writer io.Writer) *TUIExecutor {
	e := &TUIExecutor{
		cfg:     cfg,
		writer:  writer,
		timings: loadTimingHistory(),
	}

	// Build the full check list from all categories
	executor := &Executor{cfg: cfg}
	categories := e.buildCategories(executor)

	// Flatten for backwards compatibility (single category mode)
	for _, cat := range categories {
		e.checks = append(e.checks, cat.checks...)
	}

	return e
}

// buildCategories returns all check categories with their checks
func (e *TUIExecutor) buildCategories(executor *Executor) []categoryDef {
	return []categoryDef{
		{
			name: "Development Environment",
			checks: []checkItem{
				{"go-version", executor.shellCheck("check-go-version.sh"), "Ensure Go version matches .go-version"},
				{"tools", executor.shellCheck("install_tools.sh", "--check"), "Run: task setup"},
			},
		},
		{
			name: "Code Quality",
			checks: []checkItem{
				{"format", executor.checkFormat, "Run: task format"},
				{"lint", executor.checkLint, "Run: task lint"},
			},
		},
		{
			name: "Architecture Validation",
			checks: []checkItem{
				{"defaults", executor.shellCheck("check-defaults.sh"), "Use registry for SetDefault (ADR-002)"},
				{"commands", executor.shellCheck("validate-command-patterns.sh"), "Keep commands ultra-thin (ADR-001)"},
				{"constants", executor.shellCheck("check-constants.sh"), "Run: task generate:config:key-constants"},
				{"task-naming", executor.shellCheck("validate-task-naming.sh"), "Follow ADR-000 naming convention"},
				{"architecture", executor.shellCheck("validate-architecture.sh"), "Update ARCHITECTURE.md (ADR-008)"},
				{"layering", executor.shellCheck("validate-layering.sh"), "Fix layer dependencies (ADR-009)"},
				{"package-org", executor.shellCheck("validate-package-organization.sh"), "Follow package organization (ADR-010)"},
				{"config-consumption", executor.shellCheck("validate-config-consumption.sh"), "Use type-safe config (ADR-002)"},
				{"output-patterns", executor.shellCheck("validate-output-patterns.sh"), "Follow output patterns (ADR-012)"},
				{"security-patterns", executor.shellCheck("validate-security-patterns.sh"), "Implement security patterns (ADR-004)"},
			},
		},
		{
			name: "Security Scanning",
			checks: []checkItem{
				{"secrets", executor.shellCheck("check-secrets.sh"), "Remove hardcoded secrets"},
				{"sast", executor.shellCheck("check-sast.sh"), "Fix SAST issues or update .semgrep.yml"},
			},
		},
		{
			name: "Dependencies",
			checks: []checkItem{
				{"deps", executor.checkDeps, "Run: go mod tidy"},
				{"vuln", executor.checkVuln, "Update vulnerable dependencies"},
				{"license-source", executor.shellCheck("check-licenses-source.sh"), "Check dependency licenses"},
				{"license-binary", executor.shellCheck("check-licenses-binary.sh"), "Check binary licenses"},
				{"sbom-vulns", executor.shellCheck("check-sbom-vulns.sh"), "Fix SBOM vulnerabilities"},
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

// Execute runs all checks with the TUI progress display.
// Runs each category sequentially with its own progress display.
func (e *TUIExecutor) Execute(ctx context.Context) error {
	executor := &Executor{cfg: e.cfg}
	categories := e.buildCategories(executor)

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

		results, err := e.runCategoryTUI(ctx, category)
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
func (e *TUIExecutor) shouldRunCategory(categoryName string) bool {
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
func (e *TUIExecutor) runCategoryTUI(ctx context.Context, category categoryDef) ([]allCheckResult, error) {
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

// checkTest wraps the executor's checkTest to send coverage updates to the TUI.
func (e *TUIExecutor) checkTest(ctx context.Context) error {
	executor := &Executor{
		cfg: e.cfg,
		onCoverage: func(coverage float64) {
			e.coverage = coverage // Store for final summary
			if e.program != nil {
				e.program.Send(checkmate.CoverageMsg{Coverage: coverage})
			}
		},
	}
	return executor.checkTest(ctx)
}

// animateProgress sends progress updates based on historical timing data
func (e *TUIExecutor) animateProgress(p *tea.Program, idx int, checkName string, done <-chan struct{}) {
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

// printFinalSummary prints the final summary box with all check results
func (e *TUIExecutor) printFinalSummary(results []allCheckResult, passed, failed int, totalDuration time.Duration) {
	// Use checkmate printer for consistent styling
	printer := checkmate.New(checkmate.WithWriter(e.writer))

	// Determine status
	allPassed := failed == 0
	var status checkmate.Status
	var title string
	if allPassed {
		status = checkmate.StatusSuccess
		title = fmt.Sprintf("All %d Checks Passed", passed)
	} else {
		status = checkmate.StatusFailure
		title = fmt.Sprintf("%d/%d Checks Failed", failed, passed+failed)
	}

	// Build items list (just check names with timing)
	var items []string
	for _, r := range results {
		item := r.name
		if r.duration > 0 {
			item = fmt.Sprintf("%s (%s)", item, r.duration.Round(time.Millisecond))
		}
		items = append(items, item)
	}

	// Print summary
	printer.CheckSummary(status, title, items...)

	// Print errors if any
	for _, r := range results {
		if !r.passed && r.err != nil {
			printer.CheckFailure(r.name, r.err.Error(), r.remediation)
		}
	}

	// Print coverage if available
	if e.coverage > 0 {
		printer.CheckInfo(fmt.Sprintf("Coverage: %.1f%%", e.coverage))
	}

	// Print total duration
	printer.CheckInfo(fmt.Sprintf("Total Duration: %s", totalDuration.Round(time.Millisecond)))
}
