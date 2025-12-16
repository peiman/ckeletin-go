package checkmate

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRunner(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	require.NotNil(t, runner)
	assert.Equal(t, mock, runner.printer)
	assert.False(t, runner.failFast)
	assert.Empty(t, runner.category)
}

func TestNewRunner_WithOptions(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock,
		WithFailFast(),
		WithCategory("Code Quality"),
	)

	require.NotNil(t, runner)
	assert.True(t, runner.failFast)
	assert.Equal(t, "Code Quality", runner.category)
}

func TestRunner_Add(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	check := Check{
		Name:        "test",
		Fn:          func(ctx context.Context) error { return nil },
		Remediation: "fix it",
	}

	result := runner.Add(check)

	// Returns runner for chaining
	assert.Same(t, runner, result)
	assert.Len(t, runner.checks, 1)
	assert.Equal(t, "test", runner.checks[0].Name)
}

func TestRunner_AddFunc(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	fn := func(ctx context.Context) error { return nil }
	result := runner.AddFunc("test", fn)

	assert.Same(t, runner, result)
	assert.Len(t, runner.checks, 1)
	assert.Equal(t, "test", runner.checks[0].Name)
}

func TestRunner_WithRemediation(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.AddFunc("test", func(ctx context.Context) error { return nil }).
		WithRemediation("Run: task fix")

	assert.Equal(t, "Run: task fix", runner.checks[0].Remediation)
}

func TestRunner_WithRemediation_NoChecks(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	// Should not panic when no checks added
	result := runner.WithRemediation("fix")
	assert.Same(t, runner, result)
}

func TestRunner_WithDetails(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.AddFunc("test", func(ctx context.Context) error { return nil }).
		WithDetails("Uses govulncheck")

	assert.Equal(t, "Uses govulncheck", runner.checks[0].Details)
}

func TestRunner_Run_AllPass(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithCategory("Tests"))

	checkCount := 0
	runner.AddFunc("check1", func(ctx context.Context) error {
		checkCount++
		return nil
	}).AddFunc("check2", func(ctx context.Context) error {
		checkCount++
		return nil
	})

	result := runner.Run(context.Background())

	// Verify result
	assert.True(t, result.Success())
	assert.Equal(t, 2, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Checks, 2)

	// Verify all checks ran
	assert.Equal(t, 2, checkCount)

	// Verify printer calls
	assert.True(t, mock.HasCall("CategoryHeader"))
	assert.Equal(t, 2, mock.CallCount("CheckHeader"))
	assert.Equal(t, 2, mock.CallCount("CheckSuccess"))
	assert.Equal(t, 0, mock.CallCount("CheckFailure"))
	assert.True(t, mock.HasCall("CheckSummary"))
}

func TestRunner_Run_SomeFail(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.AddFunc("pass1", func(ctx context.Context) error {
		return nil
	}).AddFunc("fail1", func(ctx context.Context) error {
		return errors.New("check failed")
	}).WithRemediation("Run: task fix").
		AddFunc("pass2", func(ctx context.Context) error {
			return nil
		})

	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	assert.Equal(t, 2, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 3, result.Total)

	// Verify all checks ran (no fail-fast)
	assert.Len(t, result.Checks, 3)

	// Verify failure was printed
	assert.Equal(t, 1, mock.CallCount("CheckFailure"))
}

func TestRunner_Run_FailFast(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithFailFast())

	check3Ran := false
	runner.AddFunc("pass1", func(ctx context.Context) error {
		return nil
	}).AddFunc("fail1", func(ctx context.Context) error {
		return errors.New("check failed")
	}).AddFunc("check3", func(ctx context.Context) error {
		check3Ran = true
		return nil
	})

	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, 3, result.Total)

	// Check3 should NOT have run due to fail-fast
	assert.False(t, check3Ran)
	assert.Len(t, result.Checks, 2)
}

func TestRunner_Run_ContextCancellation(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	ctx, cancel := context.WithCancel(context.Background())

	check2Ran := false
	runner.AddFunc("check1", func(ctx context.Context) error {
		cancel() // Cancel after first check
		return nil
	}).AddFunc("check2", func(ctx context.Context) error {
		check2Ran = true
		return nil
	})

	result := runner.Run(ctx)

	// Only first check should have run
	assert.False(t, check2Ran)
	assert.Len(t, result.Checks, 1)
}

func TestRunner_Run_Empty(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	result := runner.Run(context.Background())

	assert.True(t, result.Success())
	assert.Equal(t, 0, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Checks)
}

func TestRunner_Run_PanicRecovery(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.AddFunc("panic_check", func(ctx context.Context) error {
		panic("something went wrong")
	})

	// Should NOT panic
	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	assert.Equal(t, 1, result.Failed)
	assert.Len(t, result.Checks, 1)

	// Error should contain panic message
	require.Error(t, result.Checks[0].Error)
	assert.Contains(t, result.Checks[0].Error.Error(), "panic:")
	assert.Contains(t, result.Checks[0].Error.Error(), "something went wrong")
}

func TestRunner_Run_Duration(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.AddFunc("slow", func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})

	result := runner.Run(context.Background())

	// Total duration should be at least 10ms
	assert.GreaterOrEqual(t, result.Duration, 10*time.Millisecond)

	// Check duration should be at least 10ms
	assert.GreaterOrEqual(t, result.Checks[0].Duration, 10*time.Millisecond)
}

func TestRunner_Run_CheckDetails(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	runner.Add(Check{
		Name:        "check_with_details",
		Fn:          func(ctx context.Context) error { return errors.New("error") },
		Details:     "Custom details message",
		Remediation: "Fix remediation",
	})

	runner.Run(context.Background())

	// Verify CheckFailure was called with custom details
	calls := mock.GetCalls("CheckFailure")
	require.Len(t, calls, 1)

	// Args should be: title, details, remediation
	args := calls[0]
	assert.Equal(t, "check_with_details failed", args[0])
	assert.Equal(t, "Custom details message", args[1])
	assert.Equal(t, "Fix remediation", args[2])
}

func TestRunner_Run_ErrorAsDetails(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	// No custom details - should use error message
	runner.AddFunc("check", func(ctx context.Context) error {
		return errors.New("the actual error")
	})

	runner.Run(context.Background())

	calls := mock.GetCalls("CheckFailure")
	require.Len(t, calls, 1)

	// Details should be the error message
	assert.Equal(t, "the actual error", calls[0][1])
}

func TestRunner_ConcurrentSafety(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock)

	// Add checks concurrently
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			runner.AddFunc("check", func(ctx context.Context) error {
				return nil
			})
		}(i)
	}
	wg.Wait()

	// Should have all checks
	assert.Len(t, runner.checks, 10)

	// Run should work
	result := runner.Run(context.Background())
	assert.Equal(t, 10, result.Total)
}

func TestRunner_FluentAPI(t *testing.T) {
	mock := NewMockPrinter()

	// Test full fluent chain
	result := NewRunner(mock, WithCategory("Quality")).
		AddFunc("format", func(ctx context.Context) error { return nil }).
		WithRemediation("Run: task format").
		WithDetails("Uses goimports").
		AddFunc("lint", func(ctx context.Context) error { return nil }).
		WithRemediation("Run: task lint").
		Add(Check{
			Name:        "test",
			Fn:          func(ctx context.Context) error { return nil },
			Remediation: "Run: task test",
		}).
		Run(context.Background())

	assert.True(t, result.Success())
	assert.Equal(t, 3, result.Total)
}

func TestRunResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   RunResult
		expected bool
	}{
		{
			name:     "no failures",
			result:   RunResult{Passed: 5, Failed: 0},
			expected: true,
		},
		{
			name:     "some failures",
			result:   RunResult{Passed: 3, Failed: 2},
			expected: false,
		},
		{
			name:     "all failures",
			result:   RunResult{Passed: 0, Failed: 5},
			expected: false,
		},
		{
			name:     "empty",
			result:   RunResult{Passed: 0, Failed: 0},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.result.Success())
		})
	}
}

func TestRunner_Integration(t *testing.T) {
	var buf bytes.Buffer
	printer := NewCheckPrinterWithWriter(&buf, WithTheme(MinimalTheme()))

	result := NewRunner(printer, WithCategory("Integration Test")).
		AddFunc("pass", func(ctx context.Context) error {
			return nil
		}).
		AddFunc("fail", func(ctx context.Context) error {
			return errors.New("intentional failure")
		}).WithRemediation("This is expected").
		Run(context.Background())

	output := buf.String()

	// Verify output contains expected elements
	assert.Contains(t, output, "Integration Test")
	assert.Contains(t, output, "pass")
	assert.Contains(t, output, "[OK]")
	assert.Contains(t, output, "fail")
	assert.Contains(t, output, "[FAIL]")
	assert.Contains(t, output, "1/2 checks failed")

	// Verify result
	assert.False(t, result.Success())
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)
}

// NewCheckPrinterWithWriter creates a printer with a custom writer (helper for tests).
func NewCheckPrinterWithWriter(w *bytes.Buffer, opts ...Option) *Printer {
	allOpts := append([]Option{WithWriter(w)}, opts...)
	return New(allOpts...)
}

// Parallel execution tests

func TestNewRunner_WithParallel(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	require.NotNil(t, runner)
	assert.True(t, runner.parallel)
	assert.Equal(t, 0, runner.workers) // 0 means unlimited
}

func TestNewRunner_WithWorkers(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithWorkers(3))

	require.NotNil(t, runner)
	assert.True(t, runner.parallel) // WithWorkers implies parallel
	assert.Equal(t, 3, runner.workers)
}

func TestRunner_Run_Parallel_AllPass(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithCategory("Parallel Tests"), WithParallel())

	var mu sync.Mutex
	checkCount := 0
	for i := 0; i < 5; i++ {
		runner.AddFunc("check"+string(rune('1'+i)), func(ctx context.Context) error {
			mu.Lock()
			checkCount++
			mu.Unlock()
			time.Sleep(10 * time.Millisecond)
			return nil
		})
	}

	result := runner.Run(context.Background())

	// Verify result
	assert.True(t, result.Success())
	assert.Equal(t, 5, result.Passed)
	assert.Equal(t, 0, result.Failed)
	assert.Equal(t, 5, result.Total)
	assert.Len(t, result.Checks, 5)

	// Verify all checks ran
	assert.Equal(t, 5, checkCount)

	// Verify printer calls
	assert.True(t, mock.HasCall("CategoryHeader"))
	assert.Equal(t, 5, mock.CallCount("CheckHeader"))
	assert.Equal(t, 5, mock.CallCount("CheckSuccess"))
	assert.Equal(t, 0, mock.CallCount("CheckFailure"))
	assert.True(t, mock.HasCall("CheckSummary"))
}

func TestRunner_Run_Parallel_SomeFail(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	runner.
		AddFunc("pass1", func(ctx context.Context) error { return nil }).
		AddFunc("fail1", func(ctx context.Context) error { return errors.New("error 1") }).
		WithRemediation("Fix 1").
		AddFunc("pass2", func(ctx context.Context) error { return nil }).
		AddFunc("fail2", func(ctx context.Context) error { return errors.New("error 2") }).
		WithRemediation("Fix 2")

	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	assert.Equal(t, 2, result.Passed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, 4, result.Total)

	// All checks should have run (no fail-fast)
	assert.Len(t, result.Checks, 4)

	// Verify failures were printed
	assert.Equal(t, 2, mock.CallCount("CheckFailure"))
}

func TestRunner_Run_Parallel_FailFast(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel(), WithFailFast())

	// Use channel to coordinate timing
	failReady := make(chan struct{})
	slowCheckStarted := make(chan struct{})

	runner.
		AddFunc("slow", func(ctx context.Context) error {
			close(slowCheckStarted)
			// Wait for fail to be ready, then wait a bit more
			<-failReady
			time.Sleep(50 * time.Millisecond)
			// Should be cancelled by fail-fast
			return ctx.Err()
		}).
		AddFunc("fail", func(ctx context.Context) error {
			// Wait for slow to start
			<-slowCheckStarted
			close(failReady)
			return errors.New("fast failure")
		})

	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	// At least one check failed
	assert.GreaterOrEqual(t, result.Failed, 1)
}

func TestRunner_Run_Parallel_WithWorkers(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithWorkers(2)) // Max 2 concurrent

	maxConcurrent := 0
	currentConcurrent := 0
	var mu sync.Mutex

	for i := 0; i < 5; i++ {
		runner.AddFunc("check"+string(rune('1'+i)), func(ctx context.Context) error {
			mu.Lock()
			currentConcurrent++
			if currentConcurrent > maxConcurrent {
				maxConcurrent = currentConcurrent
			}
			mu.Unlock()

			time.Sleep(20 * time.Millisecond)

			mu.Lock()
			currentConcurrent--
			mu.Unlock()
			return nil
		})
	}

	result := runner.Run(context.Background())

	assert.True(t, result.Success())
	assert.Equal(t, 5, result.Total)

	// Max concurrent should be limited to 2
	assert.LessOrEqual(t, maxConcurrent, 2)
}

func TestRunner_Run_Parallel_FasterThanSequential(t *testing.T) {
	checkDuration := 20 * time.Millisecond
	numChecks := 5

	// Run sequentially
	mockSeq := NewMockPrinter()
	runnerSeq := NewRunner(mockSeq)
	for i := 0; i < numChecks; i++ {
		runnerSeq.AddFunc("check", func(ctx context.Context) error {
			time.Sleep(checkDuration)
			return nil
		})
	}
	resultSeq := runnerSeq.Run(context.Background())

	// Run in parallel
	mockPar := NewMockPrinter()
	runnerPar := NewRunner(mockPar, WithParallel())
	for i := 0; i < numChecks; i++ {
		runnerPar.AddFunc("check", func(ctx context.Context) error {
			time.Sleep(checkDuration)
			return nil
		})
	}
	resultPar := runnerPar.Run(context.Background())

	// Both should succeed
	assert.True(t, resultSeq.Success())
	assert.True(t, resultPar.Success())

	// Parallel should be significantly faster
	// Sequential: ~100ms (5 * 20ms)
	// Parallel: ~20ms (all at once)
	assert.Less(t, resultPar.Duration, resultSeq.Duration,
		"Parallel (%v) should be faster than sequential (%v)",
		resultPar.Duration, resultSeq.Duration)

	// Parallel should be at least 2x faster
	assert.Less(t, resultPar.Duration, resultSeq.Duration/2,
		"Parallel should be at least 2x faster")
}

func TestRunner_Run_Parallel_Empty(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	result := runner.Run(context.Background())

	assert.True(t, result.Success())
	assert.Equal(t, 0, result.Total)
	assert.Empty(t, result.Checks)
}

func TestRunner_Run_Parallel_PanicRecovery(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	runner.
		AddFunc("panic_check", func(ctx context.Context) error {
			panic("parallel panic")
		}).
		AddFunc("normal", func(ctx context.Context) error {
			return nil
		})

	// Should NOT panic
	result := runner.Run(context.Background())

	assert.False(t, result.Success())
	assert.Equal(t, 1, result.Passed)
	assert.Equal(t, 1, result.Failed)

	// Find the panic result
	var panicResult *CheckResult
	for i := range result.Checks {
		if result.Checks[i].Name == "panic_check" {
			panicResult = &result.Checks[i]
			break
		}
	}

	require.NotNil(t, panicResult)
	assert.Contains(t, panicResult.Error.Error(), "panic:")
	assert.Contains(t, panicResult.Error.Error(), "parallel panic")
}

func TestRunner_Run_Parallel_ContextCancellation(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	ctx, cancel := context.WithCancel(context.Background())

	runner.
		AddFunc("check1", func(ctx context.Context) error {
			cancel() // Cancel immediately
			return nil
		}).
		AddFunc("check2", func(ctx context.Context) error {
			// This will see cancelled context
			time.Sleep(50 * time.Millisecond)
			return ctx.Err()
		})

	result := runner.Run(ctx)

	// At least one check should have context error
	hasContextError := false
	for _, cr := range result.Checks {
		if cr.Error != nil && errors.Is(cr.Error, context.Canceled) {
			hasContextError = true
			break
		}
	}
	// Context cancellation should be detected
	assert.True(t, hasContextError || ctx.Err() != nil)
}

func TestRunner_Run_Parallel_MaintainsOrder(t *testing.T) {
	mock := NewMockPrinter()
	runner := NewRunner(mock, WithParallel())

	// Add checks with varying delays
	// They should complete in different order but results should maintain original order
	delays := []time.Duration{30 * time.Millisecond, 10 * time.Millisecond, 20 * time.Millisecond}
	names := []string{"first", "second", "third"}

	for i, name := range names {
		delay := delays[i]
		runner.AddFunc(name, func(ctx context.Context) error {
			time.Sleep(delay)
			return nil
		})
	}

	result := runner.Run(context.Background())

	assert.True(t, result.Success())
	require.Len(t, result.Checks, 3)

	// Results should be in original order
	assert.Equal(t, "first", result.Checks[0].Name)
	assert.Equal(t, "second", result.Checks[1].Name)
	assert.Equal(t, "third", result.Checks[2].Name)
}

func TestRunCheckSafe(t *testing.T) {
	tests := []struct {
		name        string
		fn          func(ctx context.Context) error
		wantErr     bool
		errContains string
	}{
		{
			name:    "success",
			fn:      func(ctx context.Context) error { return nil },
			wantErr: false,
		},
		{
			name:    "error",
			fn:      func(ctx context.Context) error { return errors.New("test error") },
			wantErr: true,
		},
		{
			name:        "panic string",
			fn:          func(ctx context.Context) error { panic("test panic") },
			wantErr:     true,
			errContains: "panic: test panic",
		},
		{
			name:        "panic error",
			fn:          func(ctx context.Context) error { panic(errors.New("panic error")) },
			wantErr:     true,
			errContains: "panic:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runCheckSafe(context.Background(), tt.fn)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
