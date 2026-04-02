package check

import (
	"context"
	"strings"
	"sync"
	"time"
)

// RunOptions configures how the Runner executes checks.
type RunOptions struct {
	FailFast bool
	Parallel bool
}

// Result holds the outcome of a single check execution.
type Result struct {
	Name        string
	Category    string
	Passed      bool
	Duration    time.Duration
	Err         error
	Remediation string
}

// OnCheckDone is called after each check completes, allowing the caller
// to update a TUI, log progress, etc. The index is the position within
// the current category's check list.
type OnCheckDone func(index int, result Result)

// Runner handles check orchestration: filtering, ordering, execution,
// and result collection. It contains no TUI or rendering logic.
type Runner struct {
	timings *timingHistory
}

// NewRunner creates a Runner with the given timing history.
func NewRunner(timings *timingHistory) *Runner {
	return &Runner{timings: timings}
}

// FilterCategories returns only the categories whose display name matches
// one of the requested filter names. If filters is empty, all categories
// are returned.
func (r *Runner) FilterCategories(categories []categoryDef, filters []string) []categoryDef {
	if len(filters) == 0 {
		return categories
	}

	// Map display names to filter names
	categoryMap := map[string]string{
		"Development Environment": CategoryEnvironment,
		"Code Quality":            CategoryQuality,
		"Architecture Validation": CategoryArchitecture,
		"Security Scanning":       CategorySecurity,
		"Dependencies":            CategoryDependencies,
		"Tests":                   CategoryTests,
	}

	var result []categoryDef
	for _, cat := range categories {
		if len(cat.checks) == 0 {
			continue
		}

		filterName, ok := categoryMap[cat.name]
		if !ok {
			// Unknown category, include it
			result = append(result, cat)
			continue
		}

		for _, f := range filters {
			if strings.EqualFold(f, filterName) {
				result = append(result, cat)
				break
			}
		}
	}
	return result
}

// RunChecks executes checks sequentially or in parallel, collecting results.
// It calls onDone (if non-nil) after each check completes. Returns all
// results and an error if any check failed.
func (r *Runner) RunChecks(ctx context.Context, category categoryDef, opts RunOptions, onDone OnCheckDone) ([]Result, error) {
	if opts.Parallel {
		return r.runChecksParallel(ctx, category, opts, onDone)
	}
	return r.runChecksSequential(ctx, category, opts, onDone)
}

// runChecksSequential runs checks one at a time in order.
func (r *Runner) runChecksSequential(ctx context.Context, category categoryDef, opts RunOptions, onDone OnCheckDone) ([]Result, error) {
	var results []Result
	var categoryErr error

	for i, check := range category.checks {
		start := time.Now()
		checkErr := check.fn(ctx)
		duration := time.Since(start)

		r.timings.recordDuration(check.name, duration)

		result := Result{
			Name:        check.name,
			Category:    category.name,
			Duration:    duration,
			Remediation: check.remediation,
		}

		if checkErr != nil {
			result.Passed = false
			result.Err = checkErr
			categoryErr = checkErr
		} else {
			result.Passed = true
		}
		results = append(results, result)

		if onDone != nil {
			onDone(i, result)
		}

		if checkErr != nil && opts.FailFast {
			break
		}
	}

	return results, categoryErr
}

// runChecksParallel runs all checks concurrently and returns results in
// original order.
func (r *Runner) runChecksParallel(ctx context.Context, category categoryDef, opts RunOptions, onDone OnCheckDone) ([]Result, error) {
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

			if checkErr != nil && opts.FailFast {
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
	for cr := range resultCh {
		ordered[cr.index] = cr
	}

	var results []Result
	var categoryErr error

	for i, check := range category.checks {
		cr := ordered[i]

		r.timings.recordDuration(check.name, cr.duration)

		result := Result{
			Name:        check.name,
			Category:    category.name,
			Duration:    cr.duration,
			Remediation: check.remediation,
		}

		if cr.err != nil {
			result.Passed = false
			result.Err = cr.err
			if categoryErr == nil {
				categoryErr = cr.err
			}
		} else {
			result.Passed = true
		}
		results = append(results, result)

		if onDone != nil {
			onDone(i, result)
		}
	}

	return results, categoryErr
}

// RecordTiming exposes timing recording for callers that need to record
// durations outside of RunChecks (e.g., TUI mode with its own execution loop).
func (r *Runner) RecordTiming(name string, duration time.Duration) {
	r.timings.recordDuration(name, duration)
}

// SaveTimings persists timing history to disk.
func (r *Runner) SaveTimings() {
	r.timings.save()
}
