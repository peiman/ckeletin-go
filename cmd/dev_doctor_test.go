//go:build dev

package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDevDoctorCommand(t *testing.T) {
	// SETUP PHASE
	devCmd := findCommandByName(RootCmd, "dev")
	assert.NotNil(t, devCmd, "Dev command must exist")

	// EXECUTION PHASE
	doctorCmd := findCommandByName(devCmd, "doctor")

	// ASSERTION PHASE
	assert.NotNil(t, doctorCmd, "Doctor subcommand should exist")
	assert.Equal(t, "doctor", doctorCmd.Use, "Command should be named 'doctor'")
	assert.NotEmpty(t, doctorCmd.Short, "Command should have a short description")
}

func TestDevDoctorRun(t *testing.T) {
	// SETUP PHASE
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "doctor"})

	// EXECUTION PHASE
	_ = RootCmd.Execute() // May error if environment has issues, which is fine

	// ASSERTION PHASE
	output := buf.String()
	assert.NotEmpty(t, output, "Doctor should produce output")
	assert.Contains(t, output, "Development Environment Health Check",
		"Should show health check header")
	assert.Contains(t, output, "Summary:",
		"Should show summary")

	// Should check various tools
	assert.True(t,
		strings.Contains(output, "Task runner") ||
			strings.Contains(output, "Go compiler") ||
			strings.Contains(output, "Go version"),
		"Should check for development tools")

	// Reset
	RootCmd.SetArgs([]string{})
}

func TestDevDoctorHelp(t *testing.T) {
	// SETUP PHASE
	buf := new(bytes.Buffer)
	RootCmd.SetOut(buf)
	RootCmd.SetErr(buf)
	RootCmd.SetArgs([]string{"dev", "doctor", "--help"})

	// EXECUTION PHASE
	err := RootCmd.Execute()

	// ASSERTION PHASE
	assert.NoError(t, err, "Help should not error")
	output := buf.String()
	assert.Contains(t, output, "doctor", "Help should mention doctor")
	assert.Contains(t, output, "health", "Help should mention health checks")

	// Reset
	RootCmd.SetArgs([]string{})
}
