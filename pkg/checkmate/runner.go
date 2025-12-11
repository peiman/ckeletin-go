package checkmate

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Check represents a single check to run.
type Check struct {
	// Name identifies this check (e.g., "format", "lint")
	Name string

	// Fn runs the check. Return nil for success, error for failure.
	Fn func(ctx context.Context) error

	// Remediation is shown on failure (e.g., "Run: task format")
	Remediation string

	// Details is shown on failure (optional additional context)
	Details string
}

// CheckResult represents the outcome of a single check.
type CheckResult struct {
	Name     string
	Status   Status
	Error    error
	Duration time.Duration
}

// RunResult represents the outcome of running all checks.
type RunResult struct {
	Passed   int
	Failed   int
	Total    int
	Checks   []CheckResult
	Duration time.Duration
}

// Success returns true if all checks passed.
func (r RunResult) Success() bool { return r.Failed == 0 }

// Runner orchestrates running multiple checks with beautiful output.
// All methods are thread-safe for concurrent use.
type Runner struct {
	printer  PrinterInterface
	checks   []Check
	failFast bool
	category string
	mu       sync.Mutex
}

// RunnerOption configures a Runner.
type RunnerOption func(*Runner)

// WithFailFast stops execution on first failure.
// Default behavior is to run all checks regardless of failures.
func WithFailFast() RunnerOption {
	return func(r *Runner) { r.failFast = true }
}

// WithCategory sets category header displayed before checks.
// Example: WithCategory("Code Quality")
func WithCategory(name string) RunnerOption {
	return func(r *Runner) { r.category = name }
}

// NewRunner creates a runner that outputs to the given printer.
//
// Example:
//
//	runner := checkmate.NewRunner(printer, checkmate.WithCategory("Code Quality"))
//	result := runner.AddFunc("lint", lintFn).Run(ctx)
func NewRunner(printer PrinterInterface, opts ...RunnerOption) *Runner {
	r := &Runner{printer: printer}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Add adds a check to the runner. Returns the runner for chaining.
//
// Example:
//
//	runner.Add(checkmate.Check{
//	    Name:        "security",
//	    Fn:          checkSecurity,
//	    Remediation: "Run: task check:vuln",
//	})
func (r *Runner) Add(check Check) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.checks = append(r.checks, check)
	return r
}

// AddFunc is a convenience for adding simple checks.
// Returns the runner for chaining.
//
// Example:
//
//	runner.AddFunc("format", checkFormat).AddFunc("lint", checkLint)
func (r *Runner) AddFunc(name string, fn func(ctx context.Context) error) *Runner {
	return r.Add(Check{Name: name, Fn: fn})
}

// WithRemediation sets remediation text for the last added check.
// Returns the runner for chaining.
//
// Example:
//
//	runner.AddFunc("format", checkFormat).WithRemediation("Run: task format")
func (r *Runner) WithRemediation(text string) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.checks) > 0 {
		r.checks[len(r.checks)-1].Remediation = text
	}
	return r
}

// WithDetails sets details text for the last added check.
// Returns the runner for chaining.
//
// Example:
//
//	runner.AddFunc("security", checkSecurity).WithDetails("Uses govulncheck")
func (r *Runner) WithDetails(text string) *Runner {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.checks) > 0 {
		r.checks[len(r.checks)-1].Details = text
	}
	return r
}

// Run executes all checks and returns results.
// Automatically prints category header, check headers, success/failure, and summary.
// Respects context cancellation.
//
// Example:
//
//	result := runner.Run(ctx)
//	if !result.Success() {
//	    os.Exit(1)
//	}
func (r *Runner) Run(ctx context.Context) RunResult {
	r.mu.Lock()
	checks := make([]Check, len(r.checks))
	copy(checks, r.checks)
	category := r.category
	failFast := r.failFast
	printer := r.printer
	r.mu.Unlock()

	start := time.Now()
	result := RunResult{Total: len(checks)}

	// Print category header if set
	if category != "" {
		printer.CategoryHeader(category)
	}

	passedNames := []string{}

	for _, check := range checks {
		// Check context cancellation
		if ctx.Err() != nil {
			break
		}

		// Print check header
		printer.CheckHeader(check.Name)

		// Run check with panic recovery
		checkStart := time.Now()
		err := runCheckSafe(ctx, check.Fn)
		checkDuration := time.Since(checkStart)

		checkResult := CheckResult{
			Name:     check.Name,
			Duration: checkDuration,
		}

		if err != nil {
			checkResult.Status = StatusFailure
			checkResult.Error = err
			result.Failed++

			details := check.Details
			if details == "" {
				details = err.Error()
			}
			printer.CheckFailure(check.Name+" failed", details, check.Remediation)

			if failFast {
				result.Checks = append(result.Checks, checkResult)
				break
			}
		} else {
			checkResult.Status = StatusSuccess
			result.Passed++
			passedNames = append(passedNames, check.Name)
			printer.CheckSuccess(check.Name + " passed")
		}

		result.Checks = append(result.Checks, checkResult)
	}

	result.Duration = time.Since(start)

	// Print summary
	if result.Success() {
		printer.CheckSummary(StatusSuccess, "All checks passed", passedNames...)
	} else {
		failedNames := []string{}
		for _, cr := range result.Checks {
			if cr.Status == StatusFailure {
				failedNames = append(failedNames, cr.Name)
			}
		}
		printer.CheckSummary(StatusFailure,
			fmt.Sprintf("%d/%d checks failed", result.Failed, result.Total),
			failedNames...)
	}

	return result
}

// runCheckSafe runs a check function with panic recovery.
func runCheckSafe(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return fn(ctx)
}

// Ensure Runner implements RunnerInterface.
var _ RunnerInterface = (*Runner)(nil)
