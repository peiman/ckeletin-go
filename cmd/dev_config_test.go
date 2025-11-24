//go:build dev

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevConfigCommand(t *testing.T) {
	// SETUP PHASE
	devCmd := findCommandByName(RootCmd, "dev")
	assert.NotNil(t, devCmd, "Dev command must exist")

	// EXECUTION PHASE
	configCmd := findCommandByName(devCmd, "config")

	// ASSERTION PHASE
	assert.NotNil(t, configCmd, "Config subcommand should exist")
	assert.Equal(t, "config", configCmd.Use, "Command should be named 'config'")
	assert.NotEmpty(t, configCmd.Short, "Command should have a short description")
}

func TestDevConfigList(t *testing.T) {
	// SETUP PHASE
	// Reset flags to avoid test pollution
	configList = false
	configShow = false
	configValidate = false
	configExport = ""
	configPrefix = ""

	// Set up output capture on root command
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "config", "--list"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "List command should not error")
	output := buf.String()
	assert.NotEmpty(t, output, "List should produce output")
	assert.Contains(t, output, "Configuration Registry", "Should show registry header")
	assert.Contains(t, output, "KEY", "Should have KEY column")
	assert.Contains(t, output, "app.", "Should list app config keys")

	// Reset root command for next test
	RootCmd.SetArgs([]string{})
}

func TestDevConfigShow(t *testing.T) {
	// SETUP PHASE
	// Reset flags
	configList = false
	configShow = false
	configValidate = false
	configExport = ""
	configPrefix = ""

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "config", "--show"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "Show command should not error")
	output := buf.String()
	assert.NotEmpty(t, output, "Show should produce output")
	assert.Contains(t, output, "Effective Configuration", "Should show effective config header")

	// Reset
	RootCmd.SetArgs([]string{})
}

func TestDevConfigExportJSON(t *testing.T) {
	// SETUP PHASE
	// Reset flags
	configList = false
	configShow = false
	configValidate = false
	configExport = ""
	configPrefix = ""

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "config", "--export", "json"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "Export JSON should not error")
	output := buf.String()
	assert.NotEmpty(t, output, "Export should produce output")
	// Should be valid JSON
	assert.True(t,
		strings.HasPrefix(strings.TrimSpace(output), "[") ||
			strings.HasPrefix(strings.TrimSpace(output), "{"),
		"Output should be JSON (starts with [ or {)")

	// Reset
	RootCmd.SetArgs([]string{})
}

func TestDevConfigValidate(t *testing.T) {
	// SETUP PHASE
	// Reset flags
	configList = false
	configShow = false
	configValidate = false
	configExport = ""
	configPrefix = ""

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "config", "--validate"})

	// EXECUTION PHASE
	_ = RootCmd.Execute() // Error status depends on validation result

	// ASSERTION PHASE
	// Validation might pass or fail depending on config state,
	// but it should execute without panic
	output := buf.String()
	assert.NotEmpty(t, output, "Validate should produce output")
	assert.True(t,
		strings.Contains(output, "valid") ||
			strings.Contains(output, "error") ||
			strings.Contains(output, "passed"),
		"Output should indicate validation result")

	// Reset
	RootCmd.SetArgs([]string{})
}

func TestDevConfigHelp(t *testing.T) {
	// SETUP PHASE
	// Reset flags
	configList = false
	configShow = false
	configValidate = false
	configExport = ""
	configPrefix = ""

	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "config", "--help"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "Help should not error")
	output := buf.String()
	assert.Contains(t, output, "config", "Help should mention config")
	assert.Contains(t, output, "--list", "Help should mention list flag")
	assert.Contains(t, output, "--show", "Help should mention show flag")
	assert.Contains(t, output, "--export", "Help should mention export flag")
	assert.Contains(t, output, "--validate", "Help should mention validate flag")

	// Reset
	RootCmd.SetArgs([]string{})
}
