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

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/spf13/cobra"
)

// NewCommand creates a Cobra command from metadata following ckeletin-go patterns.
//
// This helper enforces the ultra-thin command pattern by:
//  1. Creating the command from metadata (Use, Short, Long)
//  2. Auto-registering flags from the config registry
//  3. Applying custom flag overrides from metadata
//
// Usage:
//
//	var myCmd = NewCommand(config.MyMetadata, runMy)
//
// The runE function signature must be: func(*cobra.Command, []string) error
//
// Note: This function will panic if flag registration fails, as it's called during
// initialization and there's no way to recover from invalid command configuration.
func NewCommand(meta config.CommandMetadata, runE func(*cobra.Command, []string) error) *cobra.Command {
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
		// Panic is acceptable here as this is called during init()
		// and indicates a programming error in command configuration
		panic(fmt.Sprintf("failed to register flags for command %s: %v", meta.Use, err))
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
