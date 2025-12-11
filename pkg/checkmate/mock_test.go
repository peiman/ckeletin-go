package checkmate

import (
	"sync"
	"testing"

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
