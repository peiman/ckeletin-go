package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

// shouldUseTUI determines whether to use the interactive TUI or simple output.
// Returns false for CI environments, piped output, or non-TTY contexts.
func shouldUseTUI(w io.Writer) bool {
	ciEnvVars := []string{"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL", "CIRCLECI", "TRAVIS", "BUILDKITE", "TF_BUILD"}
	for _, env := range ciEnvVars {
		if os.Getenv(env) != "" {
			return false
		}
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return checkmate.IsTerminal(w)
}

// Executor runs checks with a Bubble Tea progress UI or simple output.
type Executor struct {
	cfg      Config
	writer   io.Writer
	checks   []checkItem
	runner   *Runner
	program  *tea.Program
	timings  *timingHistory
	coverage float64
	useTUI   bool
}

type checkItem struct {
	name        string
	fn          func(ctx context.Context) error
	remediation string
}

type categoryDef struct {
	name   string
	checks []checkItem
}

// NewExecutor creates a new executor with TUI or simple output based on environment.
func NewExecutor(cfg Config, writer io.Writer) *Executor {
	timings := loadTimingHistory()
	e := &Executor{
		cfg:     cfg,
		writer:  writer,
		timings: timings,
		runner:  NewRunner(timings),
		useTUI:  shouldUseTUI(writer),
	}
	methods := &checkMethods{cfg: cfg}
	categories := e.buildCategories(methods)
	for _, cat := range categories {
		e.checks = append(e.checks, cat.checks...)
	}
	return e
}

func (e *Executor) buildCategories(methods *checkMethods) []categoryDef {
	return []categoryDef{
		{name: "Development Environment", checks: []checkItem{
			{"go-version", methods.shellCheck("check-go-version.sh"), "Ensure Go version matches .go-version"},
			{"tools", methods.shellCheck("install_tools.sh", "--check"), "Run: task setup"},
		}},
		{name: "Code Quality", checks: []checkItem{
			{"format", methods.checkFormat, "Run: task format"},
			{"lint", methods.checkLint, "Run: task lint"},
		}},
		{name: "Architecture Validation", checks: []checkItem{
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
		}},
		{name: "Security Scanning", checks: []checkItem{
			{"secrets", methods.shellCheck("check-secrets.sh"), "Remove hardcoded secrets"},
			{"sast", methods.shellCheck("check-sast.sh"), "Fix SAST issues or update .semgrep.yml"},
		}},
		{name: "Dependencies", checks: []checkItem{
			{"deps", methods.checkDeps, "Run: go mod tidy"},
			{"vuln", methods.checkVuln, "Update vulnerable dependencies"},
			{"outdated", methods.shellCheck("check-deps-outdated.sh"), "Run: go get -u"},
			{"license-source", methods.shellCheck("check-licenses-source.sh"), "Check dependency licenses"},
			{"license-binary", methods.shellCheck("check-licenses-binary.sh"), "Check binary licenses"},
			{"sbom-vulns", methods.shellCheck("check-sbom-vulns.sh"), "Fix SBOM vulnerabilities"},
		}},
		{name: "Tests", checks: []checkItem{
			{"test", e.checkTest, "Fix failing tests"},
		}},
	}
}

type allCheckResult struct {
	name        string
	category    string
	passed      bool
	duration    time.Duration
	err         error
	remediation string
}

// Execute runs all checks with TUI progress display or simple output.
func (e *Executor) Execute(ctx context.Context) error {
	methods := &checkMethods{cfg: e.cfg}
	allCategories := e.buildCategories(methods)
	categories := e.runner.FilterCategories(allCategories, e.cfg.Categories)

	var allResults []allCheckResult
	var totalPassed, totalFailed int
	startTime := time.Now()

	for _, category := range categories {
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

	e.runner.SaveTimings()
	e.printFinalSummary(allResults, totalPassed, totalFailed, time.Since(startTime))

	if totalFailed > 0 {
		return fmt.Errorf("%d checks failed", totalFailed)
	}
	return nil
}

func (e *Executor) runCategoryTUI(ctx context.Context, category categoryDef) ([]allCheckResult, error) {
	names := make([]string, len(category.checks))
	for i, c := range category.checks {
		names[i] = c.name
	}
	model := checkmate.NewProgressModel(category.name, names, checkmate.WithSkipSummary())
	p := tea.NewProgram(model, tea.WithOutput(e.writer), tea.WithInput(nil))
	e.program = p

	var results []allCheckResult
	var mu sync.Mutex
	var categoryErr error

	go func() {
		for i, check := range category.checks {
			done := make(chan struct{})
			go e.animateProgress(p, i, check.name, done)

			start := time.Now()
			checkErr := check.fn(ctx)
			duration := time.Since(start)
			close(done)

			e.runner.RecordTiming(check.name, duration)

			mu.Lock()
			result := allCheckResult{name: check.name, category: category.name, duration: duration, remediation: check.remediation}
			if checkErr != nil {
				result.passed = false
				result.err = checkErr
				categoryErr = checkErr
				p.Send(checkmate.CheckUpdateMsg{Index: i, Status: checkmate.CheckFailed, Progress: 1.0, Duration: duration, Error: checkErr, Remediation: check.remediation})
			} else {
				result.passed = true
				p.Send(checkmate.CheckUpdateMsg{Index: i, Status: checkmate.CheckPassed, Progress: 1.0, Duration: duration})
			}
			results = append(results, result)
			mu.Unlock()

			if checkErr != nil && e.cfg.FailFast {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		time.Sleep(100 * time.Millisecond)
		p.Send(checkmate.DoneMsg{})
	}()

	if _, runErr := p.Run(); runErr != nil {
		return results, fmt.Errorf("TUI error: %w", runErr)
	}
	return results, categoryErr
}

// runCategorySimple delegates orchestration to the Runner and handles only presentation.
func (e *Executor) runCategorySimple(ctx context.Context, category categoryDef) ([]allCheckResult, error) {
	printer := checkmate.New(checkmate.WithWriter(e.writer), checkmate.WithTheme(checkmate.CITheme()))
	printer.CategoryHeader(category.name)

	opts := RunOptions{FailFast: e.cfg.FailFast, Parallel: e.cfg.Parallel}
	onDone := func(index int, r Result) {
		if r.Passed {
			printer.CheckLine(r.Name, checkmate.StatusSuccess, r.Duration)
		} else {
			printer.CheckLine(r.Name, checkmate.StatusFailure, r.Duration)
		}
	}

	runnerResults, categoryErr := e.runner.RunChecks(ctx, category, opts, onDone)

	results := make([]allCheckResult, len(runnerResults))
	for i, r := range runnerResults {
		results[i] = allCheckResult{name: r.Name, category: r.Category, passed: r.Passed, duration: r.Duration, err: r.Err, remediation: r.Remediation}
	}
	return results, categoryErr
}

func (e *Executor) checkTest(ctx context.Context) error {
	methods := &checkMethods{
		cfg: e.cfg,
		onCoverage: func(coverage float64) {
			e.coverage = coverage
			if e.program != nil {
				e.program.Send(checkmate.CoverageMsg{Coverage: coverage})
			}
		},
	}
	return methods.checkTest(ctx)
}

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
			progress := float64(elapsed) / float64(expectedDuration)
			if progress > 0.95 {
				progress = 0.95
			}
			p.Send(checkmate.CheckUpdateMsg{Index: idx, Status: checkmate.CheckRunning, Progress: progress})
		}
	}
}
