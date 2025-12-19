package checkmate

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockPrinter_NewMockPrinter(t *testing.T) {
	mock := NewMockPrinter()
	require.NotNil(t, mock)
	assert.Empty(t, mock.Calls)
}

func TestMockPrinter_RecordsCalls(t *testing.T) {
	mock := NewMockPrinter()

	mock.CategoryHeader("Test Category")
	mock.CheckHeader("Testing")
	mock.CheckSuccess("Passed")
	mock.CheckFailure("Failed", "details", "fix it")
	mock.CheckSummary(StatusSuccess, "Summary", "item1", "item2")
	mock.CheckInfo("line1", "line2")
	mock.CheckNote("A note")

	assert.Len(t, mock.Calls, 7)
}

func TestMockPrinter_HasCall(t *testing.T) {
	mock := NewMockPrinter()

	mock.CheckSuccess("test")

	assert.True(t, mock.HasCall("CheckSuccess"))
	assert.False(t, mock.HasCall("CheckFailure"))
	assert.False(t, mock.HasCall("NonExistent"))
}

func TestMockPrinter_CallCount(t *testing.T) {
	mock := NewMockPrinter()

	mock.CheckSuccess("one")
	mock.CheckSuccess("two")
	mock.CheckSuccess("three")
	mock.CheckFailure("fail", "", "")

	assert.Equal(t, 3, mock.CallCount("CheckSuccess"))
	assert.Equal(t, 1, mock.CallCount("CheckFailure"))
	assert.Equal(t, 0, mock.CallCount("CheckHeader"))
}

func TestMockPrinter_GetCalls(t *testing.T) {
	mock := NewMockPrinter()

	mock.CheckSuccess("first")
	mock.CheckHeader("middle")
	mock.CheckSuccess("second")

	calls := mock.GetCalls("CheckSuccess")
	assert.Len(t, calls, 2)
	assert.Equal(t, "first", calls[0][0])
	assert.Equal(t, "second", calls[1][0])
}

func TestMockPrinter_Reset(t *testing.T) {
	mock := NewMockPrinter()

	mock.CheckSuccess("test")
	assert.Len(t, mock.Calls, 1)

	mock.Reset()
	assert.Empty(t, mock.Calls)
	assert.Empty(t, mock.Output())
}

func TestMockPrinter_Arguments(t *testing.T) {
	mock := NewMockPrinter()

	mock.CategoryHeader("Category Title")
	mock.CheckHeader("Header Message")
	mock.CheckSuccess("Success Message")
	mock.CheckFailure("Fail Title", "Fail Details", "Fail Fix")
	mock.CheckSummary(StatusFailure, "Summary Title", "item1", "item2")
	mock.CheckInfo("info1", "info2", "info3")
	mock.CheckNote("Note Message")

	// Verify arguments were captured correctly
	assert.Equal(t, "Category Title", mock.Calls[0].Args[0])
	assert.Equal(t, "Header Message", mock.Calls[1].Args[0])
	assert.Equal(t, "Success Message", mock.Calls[2].Args[0])

	// CheckFailure has 3 args
	assert.Equal(t, "Fail Title", mock.Calls[3].Args[0])
	assert.Equal(t, "Fail Details", mock.Calls[3].Args[1])
	assert.Equal(t, "Fail Fix", mock.Calls[3].Args[2])

	// CheckSummary has status, title, and items
	assert.Equal(t, StatusFailure, mock.Calls[4].Args[0])
	assert.Equal(t, "Summary Title", mock.Calls[4].Args[1])
	assert.Equal(t, "item1", mock.Calls[4].Args[2])
	assert.Equal(t, "item2", mock.Calls[4].Args[3])

	// CheckInfo has variadic args
	assert.Equal(t, "info1", mock.Calls[5].Args[0])
	assert.Equal(t, "info2", mock.Calls[5].Args[1])
	assert.Equal(t, "info3", mock.Calls[5].Args[2])

	// CheckNote
	assert.Equal(t, "Note Message", mock.Calls[6].Args[0])
}

func TestMockPrinter_ConcurrentAccess(t *testing.T) {
	mock := NewMockPrinter()

	var wg sync.WaitGroup
	iterations := 100

	// Concurrent writes
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mock.CheckSuccess("test")
		}()
	}

	// Concurrent reads
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mock.HasCall("CheckSuccess")
			_ = mock.CallCount("CheckSuccess")
		}()
	}

	wg.Wait()

	assert.Equal(t, iterations, mock.CallCount("CheckSuccess"))
}

func TestMockPrinter_ImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ PrinterInterface = (*MockPrinter)(nil)

	// Runtime check
	var p PrinterInterface = NewMockPrinter()
	require.NotNil(t, p)

	// Use as interface
	p.CheckSuccess("via interface")

	mock := p.(*MockPrinter)
	assert.True(t, mock.HasCall("CheckSuccess"))
}

func TestMockPrinter_CheckLine(t *testing.T) {
	mock := NewMockPrinter()
	mock.CheckLine("test", StatusSuccess, 100*time.Millisecond)

	assert.True(t, mock.HasCall("CheckLine"))
	calls := mock.GetCalls("CheckLine")
	require.Len(t, calls, 1)
	assert.Equal(t, "test", calls[0][0])
	assert.Equal(t, StatusSuccess, calls[0][1])
	assert.Equal(t, 100*time.Millisecond, calls[0][2])
}

func TestMockPrinter_Output(t *testing.T) {
	mock := NewMockPrinter()
	assert.Empty(t, mock.Output())
	mock.Buffer.WriteString("test output")
	assert.Equal(t, "test output", mock.Output())
}

func TestMockRunner_NewMockRunner(t *testing.T) {
	mock := NewMockRunner()
	require.NotNil(t, mock)
	assert.Equal(t, 0, mock.CheckCount())
}

func TestMockRunner_Add(t *testing.T) {
	mock := NewMockRunner()
	mock.Add(Check{Name: "test", Fn: func(_ context.Context) error { return nil }})

	assert.Equal(t, 1, mock.CheckCount())
	assert.True(t, mock.HasCheck("test"))
}

func TestMockRunner_AddFunc(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("format", func(_ context.Context) error { return nil })

	assert.True(t, mock.HasCheck("format"))
	assert.Equal(t, 1, mock.CheckCount())
}

func TestMockRunner_WithRemediation(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("test", nil)
	mock.WithRemediation("fix it")

	check := mock.GetCheck("test")
	require.NotNil(t, check)
	assert.Equal(t, "fix it", check.Remediation)
}

func TestMockRunner_WithDetails(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("test", nil)
	mock.WithDetails("some details")

	check := mock.GetCheck("test")
	require.NotNil(t, check)
	assert.Equal(t, "some details", check.Details)
}

func TestMockRunner_WithRemediationNoChecks(t *testing.T) {
	mock := NewMockRunner()
	// Should not panic when no checks exist
	mock.WithRemediation("fix it")
	mock.WithDetails("details")
	assert.Equal(t, 0, mock.CheckCount())
}

func TestMockRunner_Run(t *testing.T) {
	mock := NewMockRunner()
	mock.SetResult(RunResult{Passed: 5, Failed: 2, Total: 7})

	result := mock.Run(context.Background())
	assert.Equal(t, 5, result.Passed)
	assert.Equal(t, 2, result.Failed)
	assert.Equal(t, 7, result.Total)
	assert.Equal(t, 1, mock.RunCalls())
}

func TestMockRunner_HasCheck(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("format", nil)
	mock.AddFunc("lint", nil)

	assert.True(t, mock.HasCheck("format"))
	assert.True(t, mock.HasCheck("lint"))
	assert.False(t, mock.HasCheck("test"))
}

func TestMockRunner_GetCheck(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("format", nil)

	check := mock.GetCheck("format")
	require.NotNil(t, check)
	assert.Equal(t, "format", check.Name)

	noCheck := mock.GetCheck("nonexistent")
	assert.Nil(t, noCheck)
}

func TestMockRunner_Reset(t *testing.T) {
	mock := NewMockRunner()
	mock.AddFunc("test", nil)
	mock.Run(context.Background())
	mock.SetResult(RunResult{Passed: 1})

	mock.Reset()
	assert.Equal(t, 0, mock.CheckCount())
	assert.Equal(t, 0, mock.RunCalls())
}

func TestMockRunner_ConcurrentAccess(t *testing.T) {
	mock := NewMockRunner()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			mock.AddFunc(fmt.Sprintf("check%d", i), nil)
		}(i)
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mock.CheckCount()
			_ = mock.HasCheck("check0")
		}()
	}

	wg.Wait()
	assert.Equal(t, 50, mock.CheckCount())
}
