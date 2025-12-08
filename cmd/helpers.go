// cmd/helpers.go
//
// FRAMEWORK FILE - DO NOT EDIT unless modifying the framework itself
//
// This file provides helpers to create ultra-thin command files following
// the ckeletin-go pattern. All command files should use these helpers to
// maintain consistency and reduce boilerplate.

package cmd

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/cobra"
)

// NewCommand creates a Cobra command from metadata following ckeletin-go patterns.
//
// This helper enforces the ultra-thin command pattern by:
//  1. Creating the command from metadata (Use, Short, Long)
//  2. Auto-registering flags from the config registry
//  3. Applying custom flag overrides from metadata
//
// Returns an error if flag registration fails, allowing callers to handle errors gracefully.
//
// Usage:
//
//	cmd, err := NewCommand(config.MyMetadata, runMy)
//	if err != nil {
//	    return err
//	}
//
// For init() functions where you want to panic on error, use MustNewCommand instead.
//
// The runE function signature must be: func(*cobra.Command, []string) error
func NewCommand(meta config.CommandMetadata, runE func(*cobra.Command, []string) error) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:    meta.Use,
		Short:  meta.Short,
		Long:   meta.Long,
		RunE:   runE,
		Hidden: meta.Hidden,
	}

	// Auto-register flags from config registry based on ConfigPrefix
	// This reads all ConfigOptions with keys starting with meta.ConfigPrefix
	// and creates Cobra flags for them automatically
	if err := RegisterFlagsForPrefixWithOverrides(cmd, meta.ConfigPrefix+".", meta.FlagOverrides); err != nil {
		return nil, fmt.Errorf("failed to register flags for command %s: %w", meta.Use, err)
	}

	return cmd, nil
}

// MustNewCommand creates a Cobra command and panics on error.
//
// This is a convenience wrapper for NewCommand intended for use in init() functions
// where there's no way to handle errors gracefully. For testable code or runtime
// command creation, use NewCommand instead.
//
// Usage in init():
//
//	var myCmd = MustNewCommand(config.MyMetadata, runMy)
//
// The runE function signature must be: func(*cobra.Command, []string) error
func MustNewCommand(meta config.CommandMetadata, runE func(*cobra.Command, []string) error) *cobra.Command {
	cmd, err := NewCommand(meta, runE)
	if err != nil {
		panic(err)
	}
	return cmd
}

// MustAddToRoot adds a command to RootCmd and sets up configuration inheritance.
//
// This is a convenience wrapper that combines two common operations:
//  1. Adding the command to the root command
//  2. Setting up command configuration to inherit from parent
//
// Usage:
//
//	func init() {
//	    MustAddToRoot(myCmd)
//	}
//
// This should be called in the init() function of your command file.
func MustAddToRoot(cmd *cobra.Command) {
	RootCmd.AddCommand(cmd)
	setupCommandConfig(cmd)
}
