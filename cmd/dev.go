//go:build dev

// ckeletin:allow-custom-command
// cmd/dev.go
//
// Development-only root command for dev tools.
// This command is only available when building with the 'dev' build tag.
//
// Build with dev tools:
//   go build -tags dev
//
// Build without dev tools (production):
//   go build

package cmd

import (
	"github.com/spf13/cobra"
)

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Development tools and utilities",
	Long: `Development tools for inspecting configuration, checking environment health,
and other development-focused utilities.

This command is only available in development builds (built with -tags dev).
Production builds do not include these commands.`,
}

func init() {
	// Add dev command to root command
	// This happens at package initialization time, so the command
	// is available whenever the cmd package is imported (with dev tag)
	RootCmd.AddCommand(devCmd)
}
