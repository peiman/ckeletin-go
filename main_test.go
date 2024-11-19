// Package main_test contains tests for the main package.
package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/peiman/ckeletin-go/cmd"
	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock for cmd.Execute.
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute() error {
	args := m.Called()
	return args.Error(0)
}

func TestMain(m *testing.M) {
	// Initialize logger
	if err := infrastructure.InitLogger("debug"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	infrastructure.SetLogOutput(os.Stdout)

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

func TestMainFunction(t *testing.T) {
	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set up test arguments
	os.Args = []string{"ckeletin-go"}

	// Run main function
	main()

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Check if the output contains expected content
	output := buf.String()
	assert.Contains(t, output, "Hello from ckeletin-go!")
}

func TestMainWithError(t *testing.T) {
	// Save the original runFunc and exitFunc, and restore them after the test
	originalRun := runFunc
	originalExit := exitFunc
	defer func() {
		runFunc = originalRun
		exitFunc = originalExit
	}()

	// Mock the runFunc to return an error
	runFunc = func() error {
		return errors.New("test error")
	}

	// Mock the exitFunc to capture the exit code
	var exitCode int
	exitFunc = func(code int) {
		exitCode = code
	}

	// Redirect stderr to capture output
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	// Run main function
	main()

	// Restore stderr
	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	// Check if the output contains expected error message
	output := buf.String()
	assert.Contains(t, output, "Error: test error")
	assert.Equal(t, 1, exitCode)
}

func TestRun(t *testing.T) {
	// Test successful execution
	t.Run("Success", func(t *testing.T) {
		mockExecutor := new(MockExecutor)
		mockExecutor.On("Execute").Return(nil)

		oldExecute := cmd.Execute
		cmd.Execute = mockExecutor.Execute
		defer func() { cmd.Execute = oldExecute }()

		err := defaultRun()
		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
	})

	// Test execution with error
	t.Run("Error", func(t *testing.T) {
		mockExecutor := new(MockExecutor)
		mockExecutor.On("Execute").Return(assert.AnError)

		oldExecute := cmd.Execute
		cmd.Execute = mockExecutor.Execute
		defer func() { cmd.Execute = oldExecute }()

		err := defaultRun()
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
		mockExecutor.AssertExpectations(t)
	})
}
