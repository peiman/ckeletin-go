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
	parallel bool
	workers  int // 0 means unlimited (all checks run concurrently)
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

// WithParallel enables parallel check execution.
// All checks run concurrently by default.
// Use WithWorkers to limit concurrency.
//
// Example:
//
//	runner := checkmate.NewRunner(printer, checkmate.WithParallel())
func WithParallel() RunnerOption {
	return func(r *Runner) { r.parallel = true }
}

// WithWorkers sets the number of concurrent workers for parallel execution.
// Implies WithParallel(). A value of 0 means unlimited (all checks run concurrently).
//
// Example:
//
//	// Run at most 3 checks concurrently
//	runner := checkmate.NewRunner(printer, checkmate.WithWorkers(3))
func WithWorkers(n int) RunnerOption {
	return func(r *Runner) {
		r.parallel = true
		r.workers = n
	}
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
// When WithParallel() is enabled, checks run concurrently.
// Use WithWorkers(n) to limit concurrency.
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
	parallel := r.parallel
	workers := r.workers
	printer := r.printer
	r.mu.Unlock()

	start := time.Now()

	// Print category header if set
	if category != "" {
		printer.CategoryHeader(category)
	}

	var result RunResult
	if parallel {
		result = r.runParallel(ctx, checks, printer, failFast, workers)
	} else {
		result = r.runSequential(ctx, checks, printer, failFast)
	}

	result.Duration = time.Since(start)
	r.printSummary(printer, result)

	return result
}

// runSequential executes checks one at a time in order.
func (r *Runner) runSequential(ctx context.Context, checks []Check, printer PrinterInterface, failFast bool) RunResult {
	result := RunResult{Total: len(checks)}

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
			printer.CheckSuccess(check.Name + " passed")
		}

		result.Checks = append(result.Checks, checkResult)
	}

	return result
}

// checkJob represents a check to be executed by a worker.
type checkJob struct {
	index int
	check Check
}

// checkJobResult represents the result of executing a check job.
type checkJobResult struct {
	index  int
	result CheckResult
	check  Check // Original check for remediation/details
}

// runParallel executes checks concurrently with optional worker limit.
func (r *Runner) runParallel(ctx context.Context, checks []Check, printer PrinterInterface, failFast bool, workers int) RunResult {
	result := RunResult{Total: len(checks)}

	if len(checks) == 0 {
		return result
	}

	// Create cancellable context for fail-fast
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Print all check headers upfront (shows what will run)
	for _, check := range checks {
		printer.CheckHeader(check.Name)
	}

	// Determine worker count
	workerCount := len(checks)
	if workers > 0 && workers < workerCount {
		workerCount = workers
	}

	// Create channels
	jobs := make(chan checkJob, len(checks))
	results := make(chan checkJobResult, len(checks))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go r.runWorker(ctx, &wg, jobs, results, failFast, cancel)
	}

	// Send all jobs
	for i, check := range checks {
		jobs <- checkJob{index: i, check: check}
	}
	close(jobs)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results (maintain order for consistent output)
	checkResults := make([]checkJobResult, len(checks))
	for jr := range results {
		checkResults[jr.index] = jr
	}

	// Process results in order and print outcomes
	for _, jr := range checkResults {
		result.Checks = append(result.Checks, jr.result)

		if jr.result.Status == StatusFailure {
			result.Failed++
			details := jr.check.Details
			if details == "" && jr.result.Error != nil {
				details = jr.result.Error.Error()
			}
			printer.CheckFailure(jr.check.Name+" failed", details, jr.check.Remediation)
		} else {
			result.Passed++
			printer.CheckSuccess(jr.check.Name + " passed")
		}
	}

	return result
}

// printSummary prints the final summary based on results.
func (r *Runner) printSummary(printer PrinterInterface, result RunResult) {
	if result.Success() {
		passedNames := make([]string, 0, result.Passed)
		for _, cr := range result.Checks {
			if cr.Status == StatusSuccess {
				passedNames = append(passedNames, cr.Name)
			}
		}
		printer.CheckSummary(StatusSuccess, "All checks passed", passedNames...)
	} else {
		failedNames := make([]string, 0, result.Failed)
		for _, cr := range result.Checks {
			if cr.Status == StatusFailure {
				failedNames = append(failedNames, cr.Name)
			}
		}
		printer.CheckSummary(StatusFailure,
			fmt.Sprintf("%d/%d checks failed", result.Failed, result.Total),
			failedNames...)
	}
}

// runWorker processes jobs from the jobs channel and sends results to results channel.
func (r *Runner) runWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan checkJob, results chan<- checkJobResult, failFast bool, cancel context.CancelFunc) {
	defer wg.Done()
	for job := range jobs {
		// Check if context is cancelled
		if ctx.Err() != nil {
			results <- checkJobResult{
				index: job.index,
				result: CheckResult{
					Name:   job.check.Name,
					Status: StatusFailure,
					Error:  ctx.Err(),
				},
				check: job.check,
			}
			continue
		}

		// Run the check
		checkStart := time.Now()
		err := runCheckSafe(ctx, job.check.Fn)
		checkDuration := time.Since(checkStart)

		jr := checkJobResult{
			index: job.index,
			result: CheckResult{
				Name:     job.check.Name,
				Duration: checkDuration,
			},
			check: job.check,
		}

		if err != nil {
			jr.result.Status = StatusFailure
			jr.result.Error = err
			if failFast {
				cancel() // Cancel remaining checks
			}
		} else {
			jr.result.Status = StatusSuccess
		}

		results <- jr
	}
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
