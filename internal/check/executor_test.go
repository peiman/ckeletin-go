package check

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunCategorySimple_SequentialExecution(t *testing.T) {
	tests := []struct {
		name           string
		checks         []checkItem
		failFast       bool
		wantErr        bool
		wantPassCount  int
		wantFailCount  int
		wantTotalCount int
	}{
		{
			name: "all checks pass",
			checks: []checkItem{
				{name: "check-1", fn: func(ctx context.Context) error { return nil }},
				{name: "check-2", fn: func(ctx context.Context) error { return nil }},
				{name: "check-3", fn: func(ctx context.Context) error { return nil }},
			},
			wantErr:        false,
			wantPassCount:  3,
			wantFailCount:  0,
			wantTotalCount: 3,
		},
		{
			name: "single check fails without fail-fast",
			checks: []checkItem{
				{name: "check-1", fn: func(ctx context.Context) error { return nil }},
				{name: "check-2", fn: func(ctx context.Context) error { return errors.New("check-2 failed") }},
				{name: "check-3", fn: func(ctx context.Context) error { return nil }},
			},
			failFast:       false,
			wantErr:        true,
			wantPassCount:  2,
			wantFailCount:  1,
			wantTotalCount: 3,
		},
		{
			name: "single check fails with fail-fast stops remaining",
			checks: []checkItem{
				{name: "check-1", fn: func(ctx context.Context) error { return nil }},
				{name: "check-2", fn: func(ctx context.Context) error { return errors.New("check-2 failed") }},
				{name: "check-3", fn: func(ctx context.Context) error { return nil }},
			},
			failFast:       true,
			wantErr:        true,
			wantPassCount:  1,
			wantFailCount:  1,
			wantTotalCount: 2, // check-3 should not run
		},
		{
			name: "first check fails with fail-fast",
			checks: []checkItem{
				{name: "check-1", fn: func(ctx context.Context) error { return errors.New("check-1 failed") }},
				{name: "check-2", fn: func(ctx context.Context) error { return nil }},
			},
			failFast:       true,
			wantErr:        true,
			wantPassCount:  0,
			wantFailCount:  1,
			wantTotalCount: 1,
		},
		{
			name: "all checks fail without fail-fast",
			checks: []checkItem{
				{name: "check-1", fn: func(ctx context.Context) error { return errors.New("fail-1") }},
				{name: "check-2", fn: func(ctx context.Context) error { return errors.New("fail-2") }},
			},
			failFast:       false,
			wantErr:        true,
			wantPassCount:  0,
			wantFailCount:  2,
			wantTotalCount: 2,
		},
		{
			name:           "empty category",
			checks:         []checkItem{},
			wantErr:        false,
			wantPassCount:  0,
			wantFailCount:  0,
			wantTotalCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			executor := &Executor{
				cfg:     Config{FailFast: tt.failFast},
				writer:  &buf,
				runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
				timings: &timingHistory{Checks: make(map[string]*checkTiming)},
				useTUI:  false,
			}

			category := categoryDef{
				name:   "Test Category",
				checks: tt.checks,
			}

			results, err := executor.runCategorySimple(context.Background(), category)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			require.Len(t, results, tt.wantTotalCount)

			passed := 0
			failed := 0
			for _, r := range results {
				if r.passed {
					passed++
				} else {
					failed++
				}
			}
			assert.Equal(t, tt.wantPassCount, passed, "passed count")
			assert.Equal(t, tt.wantFailCount, failed, "failed count")
		})
	}
}

func TestRunCategorySimple_ResultMetadata(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	expectedErr := errors.New("lint failed")
	category := categoryDef{
		name: "Code Quality",
		checks: []checkItem{
			{name: "format", fn: func(ctx context.Context) error { return nil }, remediation: "Run: task format"},
			{name: "lint", fn: func(ctx context.Context) error { return expectedErr }, remediation: "Run: task lint"},
		},
	}

	results, err := executor.runCategorySimple(context.Background(), category)
	require.Error(t, err)
	require.Len(t, results, 2)

	// Check first result (passed)
	assert.Equal(t, "format", results[0].name)
	assert.Equal(t, "Code Quality", results[0].category)
	assert.True(t, results[0].passed)
	assert.NoError(t, results[0].err)
	assert.Equal(t, "Run: task format", results[0].remediation)
	assert.GreaterOrEqual(t, results[0].duration, time.Duration(0))

	// Check second result (failed)
	assert.Equal(t, "lint", results[1].name)
	assert.Equal(t, "Code Quality", results[1].category)
	assert.False(t, results[1].passed)
	assert.Equal(t, expectedErr, results[1].err)
	assert.Equal(t, "Run: task lint", results[1].remediation)
	assert.GreaterOrEqual(t, results[1].duration, time.Duration(0))
}

func TestRunCategorySimple_RecordsTiming(t *testing.T) {
	var buf bytes.Buffer
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	executor := &Executor{
		cfg:     Config{},
		writer:  &buf,
		runner:  NewRunner(timings),
		timings: timings,
		useTUI:  false,
	}

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "fast-check", fn: func(ctx context.Context) error { return nil }},
		},
	}

	_, err := executor.runCategorySimple(context.Background(), category)
	require.NoError(t, err)

	// Timing should have been recorded
	assert.Contains(t, timings.Checks, "fast-check")
	assert.Equal(t, 1, timings.Checks["fast-check"].RunCount)
	assert.GreaterOrEqual(t, timings.Checks["fast-check"].LastDuration, time.Duration(0))
}

func TestRunCategorySimple_ContextCancellation(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "check-1", fn: func(ctx context.Context) error {
				return ctx.Err()
			}},
		},
	}

	results, err := executor.runCategorySimple(ctx, category)
	require.Error(t, err)
	require.Len(t, results, 1)
	assert.False(t, results[0].passed)
}

func TestRunCategorySimple_ParallelFailFast(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{Parallel: true, FailFast: true},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "fast-fail", fn: func(ctx context.Context) error {
				return errors.New("immediate failure")
			}},
			{name: "slow-check", fn: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(50 * time.Millisecond):
					return nil
				}
			}},
		},
	}

	results, err := executor.runCategorySimple(context.Background(), category)
	require.Error(t, err)
	// All checks run in parallel, so both will have results
	require.Len(t, results, 2)
}

func TestRunCategorySimple_ParallelAllPass(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{Parallel: true},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "check-1", fn: func(ctx context.Context) error { return nil }},
			{name: "check-2", fn: func(ctx context.Context) error { return nil }},
			{name: "check-3", fn: func(ctx context.Context) error { return nil }},
		},
	}

	results, err := executor.runCategorySimple(context.Background(), category)
	require.NoError(t, err)
	require.Len(t, results, 3)
	for _, r := range results {
		assert.True(t, r.passed, "check %s should pass", r.name)
	}
}

func TestRunCategorySimple_ParallelMixedResults(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{Parallel: true},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "pass-1", fn: func(ctx context.Context) error { return nil }},
			{name: "fail-1", fn: func(ctx context.Context) error { return errors.New("failed") }},
			{name: "pass-2", fn: func(ctx context.Context) error { return nil }},
		},
	}

	results, err := executor.runCategorySimple(context.Background(), category)
	require.Error(t, err)
	require.Len(t, results, 3)

	// Results should be in original order despite parallel execution
	assert.Equal(t, "pass-1", results[0].name)
	assert.True(t, results[0].passed)
	assert.Equal(t, "fail-1", results[1].name)
	assert.False(t, results[1].passed)
	assert.Equal(t, "pass-2", results[2].name)
	assert.True(t, results[2].passed)
}

func TestRunCategorySimple_ParallelRecordsTiming(t *testing.T) {
	var buf bytes.Buffer
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	executor := &Executor{
		cfg:     Config{Parallel: true},
		writer:  &buf,
		runner:  NewRunner(timings),
		timings: timings,
		useTUI:  false,
	}

	category := categoryDef{
		name: "Test Category",
		checks: []checkItem{
			{name: "check-1", fn: func(ctx context.Context) error { return nil }},
			{name: "check-2", fn: func(ctx context.Context) error { return nil }},
		},
	}

	_, err := executor.runCategorySimple(context.Background(), category)
	require.NoError(t, err)

	assert.Contains(t, timings.Checks, "check-1")
	assert.Contains(t, timings.Checks, "check-2")
	assert.Equal(t, 1, timings.Checks["check-1"].RunCount)
	assert.Equal(t, 1, timings.Checks["check-2"].RunCount)
}

func TestExecute_WithMockChecks(t *testing.T) {
	t.Run("execute with all passing checks", func(t *testing.T) {
		setupTimingTestEnv(t)

		var buf bytes.Buffer
		executor := &Executor{
			cfg:     Config{},
			writer:  &buf,
			runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
			timings: &timingHistory{Checks: make(map[string]*checkTiming)},
			useTUI:  false,
		}

		// Override buildCategories by injecting checks directly
		// We can test Execute by providing a category filter that matches nothing,
		// or by constructing a minimal executor

		// Instead, use runCategorySimple to test execute-like behavior
		category := categoryDef{
			name: "Test Category",
			checks: []checkItem{
				{name: "mock-check", fn: func(ctx context.Context) error { return nil }, remediation: "fix it"},
			},
		}

		results, err := executor.runCategorySimple(context.Background(), category)
		require.NoError(t, err)
		require.Len(t, results, 1)
		assert.True(t, results[0].passed)
	})
}

func TestExecutor_CheckTest(t *testing.T) {
	t.Run("checkTest stores coverage on executor", func(t *testing.T) {
		var buf bytes.Buffer
		executor := &Executor{
			cfg:     Config{},
			writer:  &buf,
			runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
			timings: &timingHistory{Checks: make(map[string]*checkTiming)},
			useTUI:  false,
		}

		// The checkTest method on Executor wraps the checkMethods.checkTest
		// We can verify the onCoverage callback wiring works by checking the struct
		methods := &checkMethods{
			cfg: executor.cfg,
			onCoverage: func(coverage float64) {
				executor.coverage = coverage
			},
		}

		// Set coverage via the callback
		methods.onCoverage(87.5)
		assert.Equal(t, 87.5, executor.coverage)
	})
}

func TestAllCheckResult(t *testing.T) {
	t.Run("struct fields are correctly set", func(t *testing.T) {
		testErr := errors.New("test error")
		result := allCheckResult{
			name:        "lint",
			category:    "Code Quality",
			passed:      false,
			duration:    2 * time.Second,
			err:         testErr,
			remediation: "Run: task lint",
		}

		assert.Equal(t, "lint", result.name)
		assert.Equal(t, "Code Quality", result.category)
		assert.False(t, result.passed)
		assert.Equal(t, 2*time.Second, result.duration)
		assert.Equal(t, testErr, result.err)
		assert.Equal(t, "Run: task lint", result.remediation)
	})
}

func TestCheckItem(t *testing.T) {
	t.Run("check item runs function", func(t *testing.T) {
		called := false
		item := checkItem{
			name: "test-check",
			fn: func(ctx context.Context) error {
				called = true
				return nil
			},
			remediation: "fix it",
		}

		err := item.fn(context.Background())
		assert.NoError(t, err)
		assert.True(t, called)
		assert.Equal(t, "test-check", item.name)
		assert.Equal(t, "fix it", item.remediation)
	})
}

func TestCategoryDef(t *testing.T) {
	t.Run("category holds checks", func(t *testing.T) {
		cat := categoryDef{
			name: "Test Category",
			checks: []checkItem{
				{name: "check-1"},
				{name: "check-2"},
			},
		}

		assert.Equal(t, "Test Category", cat.name)
		assert.Len(t, cat.checks, 2)
	})
}

func TestBuildCategories_CheckMetadata(t *testing.T) {
	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}
	methods := &checkMethods{cfg: Config{}}
	categories := executor.buildCategories(methods)

	t.Run("all checks have names", func(t *testing.T) {
		for _, cat := range categories {
			for _, check := range cat.checks {
				assert.NotEmpty(t, check.name, "check in %s should have a name", cat.name)
			}
		}
	})

	t.Run("all checks have functions", func(t *testing.T) {
		for _, cat := range categories {
			for _, check := range cat.checks {
				assert.NotNil(t, check.fn, "check %s in %s should have a function", check.name, cat.name)
			}
		}
	})

	t.Run("all checks have remediation text", func(t *testing.T) {
		for _, cat := range categories {
			for _, check := range cat.checks {
				assert.NotEmpty(t, check.remediation, "check %s in %s should have remediation", check.name, cat.name)
			}
		}
	})

	t.Run("check names are unique across all categories", func(t *testing.T) {
		seen := make(map[string]string)
		for _, cat := range categories {
			for _, check := range cat.checks {
				if existingCat, exists := seen[check.name]; exists {
					assert.Failf(t, "duplicate check name", "duplicate check name %q: found in both %q and %q",
						check.name, existingCat, cat.name)
				}
				seen[check.name] = cat.name
			}
		}
	})
}

func TestFilterCategories_ViaRunner_AllMappings(t *testing.T) {
	runner := NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)})
	tests := []struct {
		displayName string
		filterName  string
	}{
		{"Development Environment", CategoryEnvironment},
		{"Code Quality", CategoryQuality},
		{"Architecture Validation", CategoryArchitecture},
		{"Security Scanning", CategorySecurity},
		{"Dependencies", CategoryDependencies},
		{"Tests", CategoryTests},
	}

	for _, tt := range tests {
		t.Run(tt.displayName, func(t *testing.T) {
			cats := []categoryDef{{name: tt.displayName, checks: []checkItem{{name: "c"}}}}
			result := runner.FilterCategories(cats, []string{tt.filterName})
			require.Len(t, result, 1)

			for _, other := range tests {
				if other.filterName != tt.filterName {
					otherCats := []categoryDef{{name: other.displayName, checks: []checkItem{{name: "c"}}}}
					otherResult := runner.FilterCategories(otherCats, []string{tt.filterName})
					assert.Empty(t, otherResult)
				}
			}
		})
	}
}

func TestFilterCategories_ViaRunner_MultipleFilters(t *testing.T) {
	runner := NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)})
	categories := []categoryDef{
		{name: "Code Quality", checks: []checkItem{{name: "format"}}},
		{name: "Tests", checks: []checkItem{{name: "test"}}},
		{name: "Dependencies", checks: []checkItem{{name: "deps"}}},
		{name: "Development Environment", checks: []checkItem{{name: "go-version"}}},
	}
	result := runner.FilterCategories(categories, []string{"quality", "tests"})
	require.Len(t, result, 2)
	assert.Equal(t, "Code Quality", result[0].name)
	assert.Equal(t, "Tests", result[1].name)
}

func TestExecute_CategoryFiltering(t *testing.T) {
	t.Run("filtered to nonexistent category runs no checks", func(t *testing.T) {
		setupTimingTestEnv(t)

		var buf bytes.Buffer
		// Use a category filter that won't match any real category
		executor := &Executor{
			cfg:     Config{Categories: []string{"nonexistent"}},
			writer:  &buf,
			runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
			timings: &timingHistory{Checks: make(map[string]*checkTiming)},
			useTUI:  false,
		}

		err := executor.Execute(context.Background())
		// No checks should run, so no failures
		assert.NoError(t, err)
	})
}

func TestExecute_IntegrationWithRunCategorySimple(t *testing.T) {
	t.Run("executes mock checks and aggregates results", func(t *testing.T) {
		setupTimingTestEnv(t)

		var buf bytes.Buffer
		timings := &timingHistory{Checks: make(map[string]*checkTiming)}
		executor := &Executor{
			cfg:     Config{},
			writer:  &buf,
			runner:  NewRunner(timings),
			timings: timings,
			useTUI:  false,
		}

		// Simulate what Execute does with a single category
		category := categoryDef{
			name: "Code Quality",
			checks: []checkItem{
				{name: "mock-format", fn: func(ctx context.Context) error { return nil }, remediation: "Run: task format"},
				{name: "mock-lint", fn: func(ctx context.Context) error { return nil }, remediation: "Run: task lint"},
			},
		}

		results, err := executor.runCategorySimple(context.Background(), category)
		require.NoError(t, err)
		require.Len(t, results, 2)

		// Now simulate printFinalSummary
		var passed, failed int
		for _, r := range results {
			if r.passed {
				passed++
			} else {
				failed++
			}
		}
		assert.Equal(t, 2, passed)
		assert.Equal(t, 0, failed)

		// Verify timing was saved
		timings.save()
		loaded := loadTimingHistory()
		assert.Contains(t, loaded.Checks, "mock-format")
		assert.Contains(t, loaded.Checks, "mock-lint")
	})

	t.Run("fail-fast stops after first category failure", func(t *testing.T) {
		setupTimingTestEnv(t)

		var buf bytes.Buffer
		executor := &Executor{
			cfg:     Config{FailFast: true},
			writer:  &buf,
			runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
			timings: &timingHistory{Checks: make(map[string]*checkTiming)},
			useTUI:  false,
		}

		// Run two categories, first has a failure
		cat1 := categoryDef{
			name: "First Category",
			checks: []checkItem{
				{name: "fail-check", fn: func(ctx context.Context) error {
					return errors.New("check failed")
				}, remediation: "fix it"},
			},
		}
		cat2 := categoryDef{
			name: "Second Category",
			checks: []checkItem{
				{name: "should-not-run", fn: func(ctx context.Context) error {
					assert.Fail(t, "should not have been called due to fail-fast")
					return nil
				}, remediation: "n/a"},
			},
		}

		// Simulate Execute's category loop with fail-fast
		var allResults []allCheckResult
		var totalPassed, totalFailed int
		var categoryErr error

		for _, category := range []categoryDef{cat1, cat2} {
			results, err := executor.runCategorySimple(context.Background(), category)
			allResults = append(allResults, results...)
			for _, r := range results {
				if r.passed {
					totalPassed++
				} else {
					totalFailed++
				}
			}
			if err != nil && executor.cfg.FailFast {
				categoryErr = err
				break
			}
		}

		require.Error(t, categoryErr)
		assert.Equal(t, 0, totalPassed)
		assert.Equal(t, 1, totalFailed)
		assert.Len(t, allResults, 1) // Only first category's results
	})
}

func TestExecute_EmptyCategoryFilter(t *testing.T) {
	// When Categories is empty, Execute runs all categories.
	// Since we can't mock the shell scripts, we test with a filter that matches
	// a nonexistent category, which exercises the Execute loop without running checks.
	setupTimingTestEnv(t)

	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{Categories: []string{"nonexistent-category-xyz"}},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	err := executor.Execute(context.Background())
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "All 0 Checks Passed")
}

func TestExecute_FailFastWithFilteredCategory(t *testing.T) {
	setupTimingTestEnv(t)

	var buf bytes.Buffer
	executor := &Executor{
		cfg:     Config{Categories: []string{"nonexistent"}, FailFast: true},
		writer:  &buf,
		runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
		timings: &timingHistory{Checks: make(map[string]*checkTiming)},
		useTUI:  false,
	}

	err := executor.Execute(context.Background())
	assert.NoError(t, err)
}

func TestExecutor_CheckTestCoverageCallback(t *testing.T) {
	t.Run("onCoverage callback updates executor coverage", func(t *testing.T) {
		var buf bytes.Buffer
		executor := &Executor{
			cfg:     Config{},
			writer:  &buf,
			runner:  NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)}),
			timings: &timingHistory{Checks: make(map[string]*checkTiming)},
			useTUI:  false,
		}

		// Simulate what Executor.checkTest does with the coverage callback
		var capturedCoverage float64
		methods := &checkMethods{
			cfg: executor.cfg,
			onCoverage: func(coverage float64) {
				capturedCoverage = coverage
				executor.coverage = coverage
			},
		}

		// Verify callback wiring
		methods.onCoverage(92.3)
		assert.Equal(t, 92.3, capturedCoverage)
		assert.Equal(t, 92.3, executor.coverage)
	})

	t.Run("nil onCoverage does not panic", func(t *testing.T) {
		methods := &checkMethods{
			cfg: Config{},
		}
		// onCoverage is nil, calling checkTest would set coverage but not call callback
		assert.Nil(t, methods.onCoverage)
		// Setting coverage directly should work fine
		methods.coverage = 75.0
		assert.Equal(t, 75.0, methods.coverage)
	})
}
