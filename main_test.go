// main_test.go

package main

import (
	"fmt"
	"testing"

	"github.com/peiman/ckeletin-go/cmd"
	"github.com/spf13/cobra"
)

func TestMainFunction(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		cmd      string
		cmdError error
		wantCode int
	}{
		{
			name:     "Success scenario",
			cmd:      "success",
			cmdError: nil,
			wantCode: 0,
		},
		{
			name:     "Failure scenario",
			cmd:      "fail",
			cmdError: fmt.Errorf("simulated failure"),
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save the original RootCmd
			originalRoot := cmd.RootCmd
			// Create a test root command
			testRoot := &cobra.Command{Use: "test"}
			// Replace global RootCmd with our test root
			cmd.RootCmd = testRoot
			// Restore after the test
			defer func() { cmd.RootCmd = originalRoot }()

			// Add a dummy command with the specified behavior
			testRoot.AddCommand(&cobra.Command{
				Use: tt.cmd,
				RunE: func(cmd *cobra.Command, args []string) error {
					return tt.cmdError
				},
			})

			// Set command arguments
			testRoot.SetArgs([]string{tt.cmd})

			// EXECUTION PHASE
			code := run()

			// ASSERTION PHASE
			if code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, code)
			}
		})
	}
}
