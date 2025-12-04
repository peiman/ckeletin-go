// internal/ui/mock.go

package ui

// MockUIRunner is a mock implementation of the UIRunner interface for testing
type MockUIRunner struct {
	CalledWithMessage string
	CalledWithColor   string
	ReturnError       error
}

func (m *MockUIRunner) RunUI(message, col string) error {
	m.CalledWithMessage = message
	m.CalledWithColor = col
	return m.ReturnError
}
