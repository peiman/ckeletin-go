package checkmate

import (
	"bytes"
	"context"
	"sync"
	"time"
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

// CheckLine records the call.
func (m *MockPrinter) CheckLine(name string, status Status, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockCall{Method: "CheckLine", Args: []interface{}{name, status, duration}})
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

// GetCalls returns all argument lists for calls to a specific method.
// Each element is a slice of arguments passed to that method call.
//
// Example:
//
//	calls := mock.GetCalls("CheckFailure")
//	assert.Equal(t, "title", calls[0][0])
//	assert.Equal(t, "details", calls[0][1])
func (m *MockPrinter) GetCalls(method string) [][]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	var calls [][]interface{}
	for _, call := range m.Calls {
		if call.Method == method {
			calls = append(calls, call.Args)
		}
	}
	return calls
}

// Ensure MockPrinter implements PrinterInterface.
var _ PrinterInterface = (*MockPrinter)(nil)

// MockRunner records all check registrations and run calls for testing.
// It allows you to verify that checks were registered correctly without
// actually running them.
//
// Example:
//
//	mock := checkmate.NewMockRunner()
//	registerChecks(mock)
//	assert.Equal(t, 3, mock.CheckCount())
//	assert.True(t, mock.HasCheck("format"))
type MockRunner struct {
	checks       []Check
	lastRunCalls int
	lastResult   RunResult
	mu           sync.Mutex
}

// NewMockRunner creates a new MockRunner for testing.
func NewMockRunner() *MockRunner {
	return &MockRunner{}
}

// Add records the check.
func (m *MockRunner) Add(check Check) *Runner {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checks = append(m.checks, check)
	// Return nil - this is a mock, chaining is not supported
	return nil
}

// AddFunc records the check.
func (m *MockRunner) AddFunc(name string, fn func(ctx context.Context) error) *Runner {
	m.Add(Check{Name: name, Fn: fn})
	return nil
}

// WithRemediation sets remediation for the last check.
func (m *MockRunner) WithRemediation(text string) *Runner {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.checks) > 0 {
		m.checks[len(m.checks)-1].Remediation = text
	}
	return nil
}

// WithDetails sets details for the last check.
func (m *MockRunner) WithDetails(text string) *Runner {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.checks) > 0 {
		m.checks[len(m.checks)-1].Details = text
	}
	return nil
}

// Run records the call and returns a configurable result.
func (m *MockRunner) Run(ctx context.Context) RunResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastRunCalls++
	return m.lastResult
}

// SetResult sets the result that Run() will return.
func (m *MockRunner) SetResult(result RunResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastResult = result
}

// CheckCount returns the number of checks registered.
func (m *MockRunner) CheckCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.checks)
}

// HasCheck checks if a check with the given name was registered.
func (m *MockRunner) HasCheck(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, check := range m.checks {
		if check.Name == name {
			return true
		}
	}
	return false
}

// GetCheck returns the check with the given name, or nil if not found.
func (m *MockRunner) GetCheck(name string) *Check {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, check := range m.checks {
		if check.Name == name {
			return &m.checks[i]
		}
	}
	return nil
}

// RunCalls returns the number of times Run was called.
func (m *MockRunner) RunCalls() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lastRunCalls
}

// Reset clears all recorded checks and run calls.
func (m *MockRunner) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checks = nil
	m.lastRunCalls = 0
	m.lastResult = RunResult{}
}
