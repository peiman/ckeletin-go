package check

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRunner() *Runner {
	return NewRunner(&timingHistory{Checks: make(map[string]*checkTiming)})
}

func TestFilterCategories_EmptyFilters(t *testing.T) {
	runner := newTestRunner()
	categories := []categoryDef{
		{name: "Development Environment", checks: []checkItem{{name: "go-version"}}},
		{name: "Code Quality", checks: []checkItem{{name: "format"}}},
	}
	result := runner.FilterCategories(categories, nil)
	assert.Len(t, result, 2)
	result = runner.FilterCategories(categories, []string{})
	assert.Len(t, result, 2)
}

func TestFilterCategories_AllMappings(t *testing.T) {
	runner := newTestRunner()
	tests := []struct{ displayName, filterName string }{
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
		})
	}
}

func TestRunChecks_SequentialAllPass(t *testing.T) {
	runner := newTestRunner()
	cat := categoryDef{name: "T", checks: []checkItem{
		{name: "c1", fn: func(ctx context.Context) error { return nil }},
		{name: "c2", fn: func(ctx context.Context) error { return nil }},
	}}
	results, err := runner.RunChecks(context.Background(), cat, RunOptions{}, nil)
	assert.NoError(t, err)
	require.Len(t, results, 2)
	for _, r := range results {
		assert.True(t, r.Passed)
	}
}

func TestRunChecks_SequentialFailFast(t *testing.T) {
	runner := newTestRunner()
	cat := categoryDef{name: "T", checks: []checkItem{
		{name: "c1", fn: func(ctx context.Context) error { return nil }},
		{name: "c2", fn: func(ctx context.Context) error { return errors.New("fail") }},
		{name: "c3", fn: func(ctx context.Context) error { assert.Fail(t, "should not run"); return nil }},
	}}
	results, err := runner.RunChecks(context.Background(), cat, RunOptions{FailFast: true}, nil)
	assert.Error(t, err)
	require.Len(t, results, 2)
}

func TestRunChecks_ParallelMixedResults(t *testing.T) {
	runner := newTestRunner()
	cat := categoryDef{name: "T", checks: []checkItem{
		{name: "pass", fn: func(ctx context.Context) error { return nil }},
		{name: "fail", fn: func(ctx context.Context) error { return errors.New("f") }},
		{name: "pass2", fn: func(ctx context.Context) error { return nil }},
	}}
	results, err := runner.RunChecks(context.Background(), cat, RunOptions{Parallel: true}, nil)
	assert.Error(t, err)
	require.Len(t, results, 3)
	assert.True(t, results[0].Passed)
	assert.False(t, results[1].Passed)
	assert.True(t, results[2].Passed)
}

func TestRunChecks_ParallelActualConcurrency(t *testing.T) {
	runner := newTestRunner()
	started := make(chan struct{}, 2)
	release := make(chan struct{})
	go func() { <-started; <-started; close(release) }()
	cat := categoryDef{name: "P", checks: []checkItem{
		{name: "c1", fn: func(ctx context.Context) error {
			select {
			case started <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			select {
			case <-release:
				return nil
			case <-time.After(200 * time.Millisecond):
				return fmt.Errorf("timeout")
			}
		}},
		{name: "c2", fn: func(ctx context.Context) error {
			select {
			case started <- struct{}{}:
			case <-ctx.Done():
				return ctx.Err()
			}
			select {
			case <-release:
				return nil
			case <-time.After(200 * time.Millisecond):
				return fmt.Errorf("timeout")
			}
		}},
	}}
	results, err := runner.RunChecks(context.Background(), cat, RunOptions{Parallel: true}, nil)
	require.NoError(t, err)
	require.Len(t, results, 2)
}

func TestRunChecks_OnDoneCallback(t *testing.T) {
	runner := newTestRunner()
	cat := categoryDef{name: "T", checks: []checkItem{
		{name: "c1", fn: func(ctx context.Context) error { return nil }},
		{name: "c2", fn: func(ctx context.Context) error { return errors.New("f") }},
	}}
	var count int
	onDone := func(index int, r Result) { count++ }
	_, _ = runner.RunChecks(context.Background(), cat, RunOptions{}, onDone)
	assert.Equal(t, 2, count)
}

func TestRunChecks_RecordsTiming(t *testing.T) {
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	runner := NewRunner(timings)
	cat := categoryDef{name: "T", checks: []checkItem{
		{name: "c1", fn: func(ctx context.Context) error { return nil }},
	}}
	_, _ = runner.RunChecks(context.Background(), cat, RunOptions{}, nil)
	assert.Contains(t, timings.Checks, "c1")
}

func TestRunner_RecordTiming(t *testing.T) {
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	runner := NewRunner(timings)
	runner.RecordTiming("x", 2*time.Second)
	assert.Equal(t, 2*time.Second, timings.Checks["x"].LastDuration)
}

func TestRunner_SaveTimings(t *testing.T) {
	setupTimingTestEnv(t)
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	runner := NewRunner(timings)
	runner.RecordTiming("x", 3*time.Second)
	runner.SaveTimings()
	loaded := loadTimingHistory()
	assert.Contains(t, loaded.Checks, "x")
}

func TestNewRunner(t *testing.T) {
	timings := &timingHistory{Checks: make(map[string]*checkTiming)}
	runner := NewRunner(timings)
	require.NotNil(t, runner)
}
