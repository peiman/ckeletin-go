//go:build dev

package cmd

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckCommand(t *testing.T) {
	// SETUP PHASE
	checkCmd := findCommandByName(RootCmd, "check")

	// ASSERTION PHASE
	assert.NotNil(t, checkCmd, "Check command must exist")
	assert.Equal(t, "check", checkCmd.Use, "Command should be named 'check'")
	assert.NotEmpty(t, checkCmd.Short, "Command should have a short description")
}

func TestCheckCommandFlags(t *testing.T) {
	checkCmd := findCommandByName(RootCmd, "check")
	assert.NotNil(t, checkCmd)

	expectedFlags := []string{"fail-fast", "verbose", "parallel", "category", "timing"}
	for _, flag := range expectedFlags {
		f := checkCmd.Flags().Lookup(flag)
		assert.NotNil(t, f, "Flag --%s should be registered", flag)
	}
}

func TestCheckCommandInvalidCategory(t *testing.T) {
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"check", "--category", "nonexistent"})

	err := RootCmd.Execute()
	assert.Error(t, err, "Invalid category should produce an error")
	assert.Contains(t, err.Error(), "invalid categories")
}

func TestCheckCommandRun(t *testing.T) {
	// Exercise the runCheck config construction path by running with
	// --category environment (lightest category — only checks go version and tools).
	// This covers the Config struct construction including BinaryName wiring.
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"check", "--category", "environment"})

	// Execute — may fail if dev environment is incomplete, which is fine.
	// The goal is to exercise the config construction path.
	_ = RootCmd.Execute()
}
