// main_test.go

package main

import (
	"fmt"
	"testing"

	"github.com/peiman/ckeletin-go/cmd"
	"github.com/spf13/cobra"
)

func TestMainFunction(t *testing.T) {
	// Save the original RootCmd
	originalRoot := cmd.RootCmd
	// Create a test root command
	testRoot := &cobra.Command{Use: "test"}
	// Replace global RootCmd with our test root
	cmd.RootCmd = testRoot
	// Restore after the test
	defer func() { cmd.RootCmd = originalRoot }()

	// Add a dummy success command
	testRoot.AddCommand(&cobra.Command{
		Use: "success",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	})

	// Add a dummy fail command
	testRoot.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("simulated failure")
		},
	})

	// Test success scenario
	testRoot.SetArgs([]string{"success"})
	if code := run(); code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}

	// Test failure scenario
	testRoot.SetArgs([]string{"fail"})
	if code := run(); code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}
