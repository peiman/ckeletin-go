//go:build dev

package cmd

import (
	"bytes"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestDevCommand(t *testing.T) {
	// SETUP PHASE
	// The dev command should be added to RootCmd in init() during package initialization

	// EXECUTION PHASE
	devCmd := findCommandByName(RootCmd, "dev")

	// ASSERTION PHASE
	assert.NotNil(t, devCmd, "Dev command should exist in dev builds")
	assert.Equal(t, "dev", devCmd.Use, "Command should be named 'dev'")
	assert.NotEmpty(t, devCmd.Short, "Command should have a short description")
	assert.NotEmpty(t, devCmd.Long, "Command should have a long description")
}

func TestDevCommandHasSubcommands(t *testing.T) {
	// SETUP PHASE
	devCmd := findCommandByName(RootCmd, "dev")
	assert.NotNil(t, devCmd, "Dev command must exist for this test")

	// EXECUTION PHASE
	subcommands := devCmd.Commands()

	// ASSERTION PHASE
	assert.Greater(t, len(subcommands), 0, "Dev command should have subcommands")

	// Check for expected subcommands
	subcommandNames := make(map[string]bool)
	for _, cmd := range subcommands {
		subcommandNames[cmd.Use] = true
	}

	assert.True(t, subcommandNames["config"], "Dev command should have 'config' subcommand")
	assert.True(t, subcommandNames["doctor"], "Dev command should have 'doctor' subcommand")
}

func TestDevCommandHelp(t *testing.T) {
	// SETUP PHASE
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "--help"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "Help command should not return error")
	output := buf.String()
	assert.Contains(t, output, "dev", "Help should mention dev command")
	assert.Contains(t, output, "config", "Help should list config subcommand")
	assert.Contains(t, output, "doctor", "Help should list doctor subcommand")

	// Reset
	RootCmd.SetArgs([]string{})
}

func TestDevCommandWithoutSubcommand(t *testing.T) {
	// SETUP PHASE
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	// Running dev without subcommand should show help (or error asking for subcommand)
	// Either is acceptable behavior
	output := buf.String()
	assert.NotEmpty(t, output, "Command should produce output when run without subcommand")
	assert.True(t,
		err != nil || output != "",
		"Should either error or show help when run without subcommand")

	// Reset
	RootCmd.SetArgs([]string{})
}

// Helper function to find a command by name in the command tree
func findCommandByName(root *cobra.Command, name string) *cobra.Command {
	for _, cmd := range root.Commands() {
		if cmd.Use == name || cmd.Name() == name {
			return cmd
		}
	}
	return nil
}
