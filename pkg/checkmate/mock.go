package checkmate

import (
	"bytes"
	"sync"
)

// MockCall represents a single method call to the printer.
type MockCall struct {
	Method string
	Args   []interface{}
}

// MockPrinter records all output for testing.
// It captures method calls and arguments for verification.
//
// Example:
//
//	mock := checkmate.NewMockPrinter()
//	myChecker := NewMyChecker(mock)
//	myChecker.Run()
//
//	assert.True(t, mock.HasCall("CheckSuccess"))
//	assert.Equal(t, 1, mock.CallCount("CheckHeader"))
type MockPrinter struct {
	// Buffer captures any output written (currently unused but available for future use)
	Buffer bytes.Buffer
	// Calls records all method invocations with their arguments
	Calls []MockCall
	mu    sync.Mutex
}

// NewMockPrinter creates a new MockPrinter for testing.
func NewMockPrinter() *MockPrinter {
	return &MockPrinter{}
}

// CategoryHeader records the call.
func (m *MockPrinter) CategoryHeader(title string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CategoryHeader", Args: []interface{}{title}})
}

// CheckHeader records the call.
func (m *MockPrinter) CheckHeader(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CheckHeader", Args: []interface{}{message}})
}

// CheckSuccess records the call.
func (m *MockPrinter) CheckSuccess(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CheckSuccess", Args: []interface{}{message}})
}

// CheckFailure records the call.
func (m *MockPrinter) CheckFailure(title, details, remediation string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CheckFailure", Args: []interface{}{title, details, remediation}})
}

// CheckSummary records the call.
func (m *MockPrinter) CheckSummary(status Status, title string, items ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := []interface{}{status, title}
	for _, item := range items {
		args = append(args, item)
	}
	m.Calls = append(m.Calls, MockCall{Method: "CheckSummary", Args: args})
}

// CheckInfo records the call.
func (m *MockPrinter) CheckInfo(lines ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	args := make([]interface{}, len(lines))
	for i, line := range lines {
		args[i] = line
	}
	m.Calls = append(m.Calls, MockCall{Method: "CheckInfo", Args: args})
}

// CheckNote records the call.
func (m *MockPrinter) CheckNote(message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CheckNote", Args: []interface{}{message}})
}

// Output returns all captured output as a string.
func (m *MockPrinter) Output() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Buffer.String()
}

// Reset clears all recorded calls and output.
func (m *MockPrinter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
	m.Buffer.Reset()
}

// HasCall checks if a method was called at least once.
func (m *MockPrinter) HasCall(method string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, call := range m.Calls {
		if call.Method == method {
			return true
		}
	}
	return false
}

// CallCount returns the number of times a method was called.
func (m *MockPrinter) CallCount(method string) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	count := 0
	for _, call := range m.Calls {
		if call.Method == method {
			count++
		}
	}
	return count
}

// GetCalls returns all calls to a specific method.
func (m *MockPrinter) GetCalls(method string) []MockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	var calls []MockCall
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call)
		}
	}
	return calls
}

// Ensure MockPrinter implements PrinterInterface.
var _ PrinterInterface = (*MockPrinter)(nil)
