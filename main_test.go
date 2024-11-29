// main_test.go
package main

import (
	"os"
	"testing"
)

func TestMainFunction(t *testing.T) {
	// Mock osExit to prevent the test from exiting
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }() // Restore osExit after the test

	// Call main
	main()

	// Check the exit code
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}
