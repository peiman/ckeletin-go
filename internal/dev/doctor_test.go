package dev

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDoctor(t *testing.T) {
	// SETUP PHASE
	doctor := NewDoctor()

	// ASSERTION PHASE
	assert.NotNil(t, doctor, "Doctor should not be nil")
	assert.NotNil(t, doctor.checks, "Checks slice should be initialized")
	assert.Equal(t, 0, len(doctor.checks), "New doctor should have no checks yet")
}

func TestDoctorRunAllChecks(t *testing.T) {
	// SETUP PHASE
	doctor := NewDoctor()

	// EXECUTION PHASE
	doctor.RunAllChecks()

	// ASSERTION PHASE
	checks := doctor.GetResults()
	assert.NotNil(t, checks, "Results should not be nil")
	assert.Greater(t, len(checks), 0, "Should have run multiple checks")

	// Verify we have expected check categories
	checkNames := make(map[string]bool)
	for _, check := range checks {
		checkNames[check.Name] = true
	}

	// Should have tool checks
	assert.True(t, checkNames["Task runner"] || checkNames["Go compiler"],
		"Should have at least some tool checks")

	// Should have Go version check
	assert.True(t, checkNames["Go version"], "Should check Go version")

	// Should have project structure check
	assert.True(t, checkNames["Project structure"], "Should check project structure")

	// Should have git check
	assert.True(t, checkNames["Git repository"], "Should check git repository")

	// Should have dependencies check
	assert.True(t, checkNames["Dependencies"], "Should check dependencies")
}

func TestCheckStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   CheckStatus
		contains string
	}{
		{
			name:     "passed status",
			status:   CheckPassed,
			contains: "PASS",
		},
		{
			name:     "warning status",
			status:   CheckWarning,
			contains: "WARN",
		},
		{
			name:     "failed status",
			status:   CheckFailed,
			contains: "FAIL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// EXECUTION PHASE
			str := tt.status.String()

			// ASSERTION PHASE
			assert.Contains(t, str, tt.contains,
				"Status string should contain expected text")
		})
	}
}

func TestHasFailures(t *testing.T) {
	tests := []struct {
		name     string
		checks   []HealthCheck
		expected bool
	}{
		{
			name: "no failures",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckPassed},
			},
			expected: false,
		},
		{
			name: "with warnings only",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckWarning},
			},
			expected: false,
		},
		{
			name: "with failures",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckFailed},
			},
			expected: true,
		},
		{
			name: "multiple failures",
			checks: []HealthCheck{
				{Status: CheckFailed},
				{Status: CheckFailed},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			doctor := &Doctor{checks: tt.checks}

			// EXECUTION PHASE
			result := doctor.HasFailures()

			// ASSERTION PHASE
			assert.Equal(t, tt.expected, result,
				"HasFailures should return correct value")
		})
	}
}

func TestHasWarnings(t *testing.T) {
	tests := []struct {
		name     string
		checks   []HealthCheck
		expected bool
	}{
		{
			name: "no warnings",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckPassed},
			},
			expected: false,
		},
		{
			name: "with warnings",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckWarning},
			},
			expected: true,
		},
		{
			name: "with failures only",
			checks: []HealthCheck{
				{Status: CheckPassed},
				{Status: CheckFailed},
			},
			expected: false,
		},
		{
			name: "multiple warnings",
			checks: []HealthCheck{
				{Status: CheckWarning},
				{Status: CheckWarning},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			doctor := &Doctor{checks: tt.checks}

			// EXECUTION PHASE
			result := doctor.HasWarnings()

			// ASSERTION PHASE
			assert.Equal(t, tt.expected, result,
				"HasWarnings should return correct value")
		})
	}
}

func TestGetResults(t *testing.T) {
	// SETUP PHASE
	checks := []HealthCheck{
		{Name: "Check 1", Status: CheckPassed},
		{Name: "Check 2", Status: CheckFailed},
	}
	doctor := &Doctor{checks: checks}

	// EXECUTION PHASE
	results := doctor.GetResults()

	// ASSERTION PHASE
	assert.NotNil(t, results, "Results should not be nil")
	assert.Equal(t, len(checks), len(results), "Should return all checks")
	assert.Equal(t, checks[0].Name, results[0].Name, "Should preserve check order")
	assert.Equal(t, checks[1].Name, results[1].Name, "Should preserve check order")
}

func TestFormatResults(t *testing.T) {
	tests := []struct {
		name     string
		checks   []HealthCheck
		contains []string
	}{
		{
			name: "all passed",
			checks: []HealthCheck{
				{Name: "Test 1", Status: CheckPassed, Message: "OK"},
				{Name: "Test 2", Status: CheckPassed, Message: "OK"},
			},
			contains: []string{
				"Development Environment Health Check",
				"Test 1",
				"Test 2",
				"2 passed, 0 warnings, 0 failed",
				"Environment is healthy",
			},
		},
		{
			name: "with warnings",
			checks: []HealthCheck{
				{Name: "Test 1", Status: CheckPassed, Message: "OK"},
				{Name: "Test 2", Status: CheckWarning, Message: "Warning"},
			},
			contains: []string{
				"Development Environment Health Check",
				"Test 1",
				"Test 2",
				"1 passed, 1 warnings, 0 failed",
				"functional but has warnings",
			},
		},
		{
			name: "with failures",
			checks: []HealthCheck{
				{Name: "Test 1", Status: CheckPassed, Message: "OK"},
				{Name: "Test 2", Status: CheckFailed, Message: "Error"},
			},
			contains: []string{
				"Development Environment Health Check",
				"Test 1",
				"Test 2",
				"1 passed, 0 warnings, 1 failed",
				"issues that need attention",
			},
		},
		{
			name: "with details",
			checks: []HealthCheck{
				{
					Name:    "Test 1",
					Status:  CheckPassed,
					Message: "OK",
					Details: "Extra info here",
				},
			},
			contains: []string{
				"Test 1",
				"OK",
				"Extra info here",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			doctor := &Doctor{checks: tt.checks}

			// EXECUTION PHASE
			output := doctor.FormatResults()

			// ASSERTION PHASE
			assert.NotEmpty(t, output, "Output should not be empty")
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected,
					"Output should contain expected string: %s", expected)
			}
		})
	}
}

func TestGetToolVersion(t *testing.T) {
	// SETUP PHASE
	doctor := NewDoctor()

	// EXECUTION PHASE
	// Test with Go (should always be available in test environment)
	version := doctor.getToolVersion("go")

	// ASSERTION PHASE
	assert.NotEmpty(t, version, "Go version should not be empty")
	// Version should contain "go" or be "version unknown"
	assert.True(t,
		strings.Contains(version, "go") || version == "version unknown",
		"Version should be valid or unknown")
}

func TestCheckToolLogic(t *testing.T) {
	// Test the logic of checkTool by calling it directly
	doctor := NewDoctor()

	// EXECUTION PHASE
	// Check a tool that should exist (go)
	doctor.checkTool("go", "Go compiler")

	// ASSERTION PHASE
	results := doctor.GetResults()
	assert.Len(t, results, 1, "Should have one check result")
	assert.Equal(t, "Go compiler", results[0].Name, "Should have correct name")
	// Status depends on environment, but should be set
	assert.True(t,
		results[0].Status == CheckPassed || results[0].Status == CheckFailed,
		"Status should be either passed or failed")
}

func TestCheckToolNotFound(t *testing.T) {
	// Test checkTool with a tool that definitely doesn't exist
	t.Run("nonexistent tool", func(t *testing.T) {
		// SETUP PHASE
		doctor := NewDoctor()

		// EXECUTION PHASE
		// Use a tool name that definitely doesn't exist
		doctor.checkTool("this-tool-definitely-does-not-exist-12345", "Nonexistent Tool")

		// ASSERTION PHASE
		results := doctor.GetResults()
		assert.Len(t, results, 1, "Should have one check result")
		assert.Equal(t, CheckFailed, results[0].Status,
			"Should fail when tool not found")
		assert.Contains(t, results[0].Message, "not found in PATH",
			"Should report tool not found in PATH")
	})
}

func TestCheckGoVersionLogic(t *testing.T) {
	// Test the logic of checkGoVersion by calling it directly
	doctor := NewDoctor()

	// EXECUTION PHASE
	doctor.checkGoVersion()

	// ASSERTION PHASE
	results := doctor.GetResults()
	assert.Len(t, results, 1, "Should have one check result")
	assert.Equal(t, "Go version", results[0].Name, "Should check Go version")
	// Status depends on environment, but should be set
	assert.True(t,
		results[0].Status == CheckPassed ||
			results[0].Status == CheckWarning ||
			results[0].Status == CheckFailed,
		"Status should be set")
}

func TestCheckProjectStructureLogic(t *testing.T) {
	// Test the logic of checkProjectStructure by calling it directly
	// This test assumes it's running from the project root
	doctor := NewDoctor()

	// EXECUTION PHASE
	doctor.checkProjectStructure()

	// ASSERTION PHASE
	results := doctor.GetResults()
	assert.Len(t, results, 1, "Should have one check result")
	assert.Equal(t, "Project structure", results[0].Name,
		"Should check project structure")
	// In the actual project, this should pass
	// (if running from project root)
}

func TestCheckGitStatusLogic(t *testing.T) {
	// Test the logic of checkGitStatus by calling it directly
	// This test assumes it's running in a git repository
	doctor := NewDoctor()

	// EXECUTION PHASE
	doctor.checkGitStatus()

	// ASSERTION PHASE
	results := doctor.GetResults()
	assert.Len(t, results, 1, "Should have one check result")
	assert.Equal(t, "Git repository", results[0].Name,
		"Should check git repository")
	// Status depends on environment (git repo or not)
}

func TestCheckDependenciesLogic(t *testing.T) {
	// Test the logic of checkDependencies by calling it directly
	// This test assumes go.mod exists
	doctor := NewDoctor()

	// EXECUTION PHASE
	doctor.checkDependencies()

	// ASSERTION PHASE
	results := doctor.GetResults()
	assert.Len(t, results, 1, "Should have one check result")
	assert.Equal(t, "Dependencies", results[0].Name,
		"Should check dependencies")
	// Status depends on go.mod state
}

func TestCheckDependenciesInNonGoDirectory(t *testing.T) {
	// Test checkDependencies when not in a Go module directory
	t.Run("missing go.mod", func(t *testing.T) {
		// SETUP PHASE
		doctor := NewDoctor()

		// Save current dir and change to temp dir without go.mod
		currentDir, err := os.Getwd()
		assert.NoError(t, err, "Should get current directory")
		defer os.Chdir(currentDir) // Restore after test

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		assert.NoError(t, err, "Should change to temp directory")

		// EXECUTION PHASE
		doctor.checkDependencies()

		// ASSERTION PHASE
		results := doctor.GetResults()
		assert.Len(t, results, 1, "Should have one check result")
		assert.Equal(t, CheckFailed, results[0].Status,
			"Should fail when go.mod missing")
		assert.Contains(t, results[0].Message, "go.mod not found",
			"Should report go.mod not found")
	})
}

func TestDoctorIntegration(t *testing.T) {
	// Integration test that exercises the full workflow
	t.Run("full workflow", func(t *testing.T) {
		// SETUP PHASE
		doctor := NewDoctor()

		// EXECUTION PHASE
		// 1. Run all checks
		doctor.RunAllChecks()

		// 2. Get results
		results := doctor.GetResults()
		assert.Greater(t, len(results), 0, "Should have results")

		// 3. Check for failures/warnings
		hasFailures := doctor.HasFailures()
		hasWarnings := doctor.HasWarnings()
		t.Logf("Has failures: %v, Has warnings: %v", hasFailures, hasWarnings)

		// 4. Format output
		output := doctor.FormatResults()
		assert.NotEmpty(t, output, "Should have formatted output")
		assert.Contains(t, output, "Development Environment Health Check",
			"Should have header")
		assert.Contains(t, output, "Summary:",
			"Should have summary")

		// Log the output for manual inspection
		t.Logf("Doctor output:\n%s", output)
	})
}

func TestHealthCheckStructure(t *testing.T) {
	// Test that HealthCheck structure works as expected
	t.Run("health check fields", func(t *testing.T) {
		// SETUP PHASE
		check := HealthCheck{
			Name:    "Test Check",
			Status:  CheckPassed,
			Message: "Test message",
			Details: "Test details",
		}

		// ASSERTION PHASE
		assert.Equal(t, "Test Check", check.Name, "Name should be set")
		assert.Equal(t, CheckPassed, check.Status, "Status should be set")
		assert.Equal(t, "Test message", check.Message, "Message should be set")
		assert.Equal(t, "Test details", check.Details, "Details should be set")
	})
}
