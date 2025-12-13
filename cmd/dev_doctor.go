//go:build dev

// ckeletin:allow-custom-command
// cmd/dev_doctor.go
//
// Environment health checker subcommand (dev-only).
// Verifies development environment is properly configured.

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/internal/dev"
	"github.com/spf13/cobra"
)

var devDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check development environment health",
	Long: `Run health checks on your development environment to ensure all tools
are installed and properly configured.

Checks include:
  - Required development tools (task, golangci-lint, gotestsum, etc.)
  - Go version compatibility
  - Project structure validation
  - Git repository status
  - Dependency synchronization

Example:
  dev doctor`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Create doctor and run all checks
		doctor := dev.NewDoctor()
		doctor.RunAllChecks()

		// Format and display results
		output := doctor.FormatResults()
		fmt.Fprint(cmd.OutOrStdout(), output)

		// Return error if there are failures
		if doctor.HasFailures() {
			return fmt.Errorf("environment has issues that need attention")
		}

		return nil
	},
}

func init() {
	// Add to dev command
	devCmd.AddCommand(devDoctorCmd)
}
