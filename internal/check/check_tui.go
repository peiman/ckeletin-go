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
	"github.com/charmbracelet/lipgloss"
	"github.com/peiman/ckeletin-go/internal/xdg"
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
	// Default estimates for first run
	defaults := map[string]time.Duration{
		// Environment
		"go-version": 100 * time.Millisecond,
		"tools":      100 * time.Millisecond,
		// Quality
		"format": 500 * time.Millisecond,
		"lint":   3 * time.Second,
		// Architecture
		"defaults":           100 * time.Millisecond,
		"commands":           200 * time.Millisecond,
		"constants":          500 * time.Millisecond,
		"task-naming":        200 * time.Millisecond,
		"architecture":       500 * time.Millisecond,
		"layering":           4 * time.Second,
		"package-org":        500 * time.Millisecond,
		"config-consumption": 100 * time.Millisecond,
		"output-patterns":    100 * time.Millisecond,
		"security-patterns":  100 * time.Millisecond,
		// Security
		"secrets": 200 * time.Millisecond,
		"sast":    4 * time.Second,
		// Dependencies
		"deps":           1 * time.Second,
		"vuln":           2 * time.Second,
		"outdated":       2 * time.Second,
		"license-source": 1 * time.Second,
		"license-binary": 1 * time.Second,
		"sbom-vulns":     5 * time.Second,
		// Tests
		"test": 10 * time.Second,
	}
	if d, ok := defaults[name]; ok {
		return d
	}
	return 3 * time.Second // Generic default for unknown checks
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

		// Choose execution mode based on environment
		var results []allCheckResult
		var err error
		if e.useTUI {
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

// printFinalSummary prints the final summary box with all check results.
// Uses styled output for TTY, plain ASCII for CI/pipes.
func (e *Executor) printFinalSummary(results []allCheckResult, passed, failed int, totalDuration time.Duration) {
	// Only clear screen in TUI mode
	if e.useTUI {
		_, _ = fmt.Fprint(e.writer, "\033[2J\033[H")
	} else {
		_, _ = fmt.Fprintln(e.writer) // Just add a blank line in CI mode
	}

	allPassed := failed == 0

	// Group results by category
	categoryOrder := []string{
		"Development Environment",
		"Code Quality",
		"Architecture Validation",
		"Security Scanning",
		"Dependencies",
		"Tests",
	}

	resultsByCategory := make(map[string][]allCheckResult)
	for _, r := range results {
		resultsByCategory[r.category] = append(resultsByCategory[r.category], r)
	}

	// Box characters - use ASCII for CI mode
	var topLeft, topRight, bottomLeft, bottomRight, horizontal, vertical string
	var catSeparator, treeConnector, treeLastConnector string
	var successIcon, failIcon string

	if e.useTUI {
		topLeft = "╭"
		topRight = "╮"
		bottomLeft = "╰"
		bottomRight = "╯"
		horizontal = "─"
		vertical = "│"
		catSeparator = "───"
		treeConnector = "├──"
		treeLastConnector = "└──"
		successIcon = "✓"
		failIcon = "✗"
	} else {
		topLeft = "+"
		topRight = "+"
		bottomLeft = "+"
		bottomRight = "+"
		horizontal = "-"
		vertical = "|"
		catSeparator = "---"
		treeConnector = "|--"
		treeLastConnector = "`--"
		successIcon = "[OK]"
		failIcon = "[FAIL]"
	}

	boxWidth := 60
	contentWidth := boxWidth - 2

	var sb strings.Builder

	// Define styles - only use colors in TUI mode
	var borderStyle, dimStyle, boldStyle, successStyle, failStyle, titleStyle lipgloss.Style

	if e.useTUI {
		accentColor := lipgloss.Color("#78B0E7")
		failColor := lipgloss.Color("#FF5555")
		successColor := lipgloss.Color("#50FA7B")
		dimColor := lipgloss.Color("#6272A4")

		var borderColor lipgloss.Color
		if allPassed {
			borderColor = accentColor
			titleStyle = lipgloss.NewStyle().
				Bold(true).
				Background(accentColor).
				Foreground(lipgloss.Color("#000000"))
		} else {
			borderColor = failColor
			titleStyle = lipgloss.NewStyle().
				Bold(true).
				Background(failColor).
				Foreground(lipgloss.Color("#000000"))
		}

		borderStyle = lipgloss.NewStyle().Foreground(borderColor)
		dimStyle = lipgloss.NewStyle().Foreground(dimColor)
		boldStyle = lipgloss.NewStyle().Bold(true)
		successStyle = lipgloss.NewStyle().Foreground(successColor)
		failStyle = lipgloss.NewStyle().Foreground(failColor)
	} else {
		// No styling in CI mode
		borderStyle = lipgloss.NewStyle()
		dimStyle = lipgloss.NewStyle()
		boldStyle = lipgloss.NewStyle()
		successStyle = lipgloss.NewStyle()
		failStyle = lipgloss.NewStyle()
		titleStyle = lipgloss.NewStyle()
	}

	// Top border
	sb.WriteString(borderStyle.Render(topLeft+strings.Repeat(horizontal, boxWidth-2)+topRight) + "\n")

	// Empty line
	sb.WriteString(borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth) + borderStyle.Render(vertical) + "\n")

	// Title
	var titleText string
	if allPassed {
		titleText = fmt.Sprintf(" %s All %d Checks Passed ", successIcon, passed)
	} else {
		titleText = fmt.Sprintf(" %s %d/%d Checks Failed ", failIcon, failed, passed+failed)
	}
	titleRendered := titleStyle.Render(titleText)
	// Calculate padding (lipgloss.Width handles unicode properly)
	titleWidth := lipgloss.Width(titleRendered)
	leftPad := (contentWidth - titleWidth) / 2
	rightPad := contentWidth - leftPad - titleWidth
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}
	sb.WriteString(borderStyle.Render(vertical))
	sb.WriteString(strings.Repeat(" ", leftPad))
	sb.WriteString(titleRendered)
	sb.WriteString(strings.Repeat(" ", rightPad))
	sb.WriteString(borderStyle.Render(vertical) + "\n")

	// Empty line
	sb.WriteString(borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth) + borderStyle.Render(vertical) + "\n")

	// Results grouped by category
	for _, catName := range categoryOrder {
		catResults, ok := resultsByCategory[catName]
		if !ok || len(catResults) == 0 {
			continue
		}

		// Category header
		catHeader := "  " + dimStyle.Render(catSeparator+" "+catName)
		catHeaderWidth := 2 + len(catSeparator) + 1 + len(catName) // "  " + separator + " " + name
		padding := contentWidth - catHeaderWidth
		sb.WriteString(borderStyle.Render(vertical) + catHeader)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(borderStyle.Render(vertical) + "\n")

		// Check results
		for i, r := range catResults {
			var iconStyle lipgloss.Style
			var icon string
			if r.passed {
				icon = successIcon
				iconStyle = successStyle
			} else {
				icon = failIcon
				iconStyle = failStyle
			}

			// Tree connector
			connector := treeConnector
			if i == len(catResults)-1 {
				connector = treeLastConnector
			}

			// Format duration
			durStr := ""
			durLen := 0
			if r.duration > 0 {
				durText := fmt.Sprintf("(%s)", r.duration.Round(time.Millisecond))
				durStr = dimStyle.Render(durText)
				durLen = len(durText)
			}

			// Build line: "  ├── ✓ name              (duration)"
			line := "  " + dimStyle.Render(connector) + " " + iconStyle.Render(icon) + " " + fmt.Sprintf("%-18s", r.name) + " " + durStr
			visibleLen := 2 + len(connector) + 1 + len(icon) + 1 + 18 + 1 + durLen
			padding := contentWidth - visibleLen
			sb.WriteString(borderStyle.Render(vertical) + line)
			if padding > 0 {
				sb.WriteString(strings.Repeat(" ", padding))
			}
			sb.WriteString(borderStyle.Render(vertical) + "\n")
		}

		// Empty line after category
		sb.WriteString(borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth) + borderStyle.Render(vertical) + "\n")
	}

	// Coverage
	if e.coverage > 0 {
		covText := fmt.Sprintf("%.1f%%", e.coverage)
		covLine := "  " + boldStyle.Render("Coverage:") + " " + covText
		covVisibleLen := 2 + 9 + 1 + len(covText)
		padding := contentWidth - covVisibleLen
		sb.WriteString(borderStyle.Render(vertical) + covLine)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(borderStyle.Render(vertical) + "\n")
	}

	// Duration
	durText := totalDuration.Round(time.Millisecond).String()
	durLine := "  " + boldStyle.Render("Duration:") + " " + durText
	durVisibleLen := 2 + 9 + 1 + len(durText)
	padding := contentWidth - durVisibleLen
	sb.WriteString(borderStyle.Render(vertical) + durLine)
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(borderStyle.Render(vertical) + "\n")

	// Empty line
	sb.WriteString(borderStyle.Render(vertical) + strings.Repeat(" ", contentWidth) + borderStyle.Render(vertical) + "\n")

	// Bottom border
	sb.WriteString(borderStyle.Render(bottomLeft+strings.Repeat(horizontal, boxWidth-2)+bottomRight) + "\n")

	_, _ = fmt.Fprint(e.writer, sb.String())

	// Print errors below the box if any
	if !allPassed {
		_, _ = fmt.Fprint(e.writer, "\n")
		var printer *checkmate.Printer
		if e.useTUI {
			printer = checkmate.New(checkmate.WithWriter(e.writer))
		} else {
			printer = checkmate.New(checkmate.WithWriter(e.writer), checkmate.WithTheme(checkmate.CITheme()))
		}
		for _, r := range results {
			if !r.passed && r.err != nil {
				printer.CheckFailure(r.name, r.err.Error(), r.remediation)
			}
		}
	}
}
