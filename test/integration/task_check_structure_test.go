package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Section represents an expected section in the task check output.
type Section struct {
	Name   string
	Marker string
}

// runTaskCheckForStructure executes "task check" and returns output for structure validation.
// Unlike golden file tests, this runs the full integration to verify structure.
func runTaskCheckForStructure(t *testing.T) string {
	t.Helper()

	// Get the project root directory
	projectRoot, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Change to project root if we're in test/integration
	if err := os.Chdir("../.."); err == nil {
		defer func() {
			_ = os.Chdir(projectRoot)
		}()
	}

	// Execute task check with environment variable to prevent recursion
	cmd := exec.Command("task", "check")
	cmd.Env = append(os.Environ(), "INSIDE_TASK_CHECK=1")
	output, err := cmd.CombinedOutput()

	// For structure validation, we want successful runs
	require.NoError(t, err, "task check should succeed for structure validation.\nOutput:\n%s", string(output))

	return string(output)
}

// TestTaskCheckOutputStructure verifies the overall structure of task check output.
// It validates that all expected sections appear in the correct order.
func TestTaskCheckOutputStructure(t *testing.T) {
	// Skip in short mode since this runs full task check
	if testing.Short() {
		t.Skip("Skipping structure validation test in short mode")
	}

	// Skip by default to prevent recursion/hanging during standard test runs.
	// Only run this if explicitly requested (e.g., CI or manual verification).
	if os.Getenv("RUN_TASK_CHECK_TESTS") != "1" {
		t.Skip("Skipping structure validation test. Set RUN_TASK_CHECK_TESTS=1 to enable.")
	}

	// Skip if we're already running inside task check to prevent recursion
	if os.Getenv("INSIDE_TASK_CHECK") == "1" {
		t.Skip("Skipping structure validation to prevent recursion (running inside task check)")
	}

	output := runTaskCheckForStructure(t)

	// Define expected sections in order
	// Note: Category headers use styled background text (not ─── separators)
	// since the Bubble Tea TUI refactor in Dec 2025
	expectedSections := []Section{
		{Name: "Development Environment", Marker: "Development Environment"},
		{Name: "Code Quality", Marker: "Code Quality"},
		{Name: "Architecture Validation", Marker: "Architecture Validation"},
		{Name: "Dependencies", Marker: "Dependencies"},
		{Name: "Tests", Marker: "Tests"},
	}

	// Validate all sections are present
	for _, section := range expectedSections {
		assert.Contains(t, output, section.Marker,
			"Output should contain section: %s", section.Name)
	}

	// Validate sections appear in the correct order
	ValidateSectionOrder(t, output, expectedSections)

	// Validate summary appears at the end
	// Current format: "[OK] All 23 Checks Passed"
	assert.Contains(t, output, "Checks Passed",
		"Output should contain success summary")
}

// TestTaskCheckCategoryHeaders verifies that all category headers are present.
func TestTaskCheckCategoryHeaders(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping category header test in short mode")
	}

	// Skip by default to prevent recursion/hanging during standard test runs.
	if os.Getenv("RUN_TASK_CHECK_TESTS") != "1" {
		t.Skip("Skipping category header test. Set RUN_TASK_CHECK_TESTS=1 to enable.")
	}

	// Skip if we're already running inside task check to prevent recursion
	if os.Getenv("INSIDE_TASK_CHECK") == "1" {
		t.Skip("Skipping structure validation to prevent recursion (running inside task check)")
	}

	output := runTaskCheckForStructure(t)

	// Verify each category header exists
	headers := []string{
		"Development Environment",
		"Code Quality",
		"Architecture Validation",
		"Dependencies",
		"Tests",
	}

	for _, header := range headers {
		assert.Contains(t, output, header,
			"Output should contain category header: %s", header)
	}
}

// TestTaskCheckSuccessIndicators verifies success indicators are present.
func TestTaskCheckSuccessIndicators(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping success indicators test in short mode")
	}

	// Skip by default to prevent recursion/hanging during standard test runs.
	if os.Getenv("RUN_TASK_CHECK_TESTS") != "1" {
		t.Skip("Skipping success indicators test. Set RUN_TASK_CHECK_TESTS=1 to enable.")
	}

	// Skip if we're already running inside task check to prevent recursion
	if os.Getenv("INSIDE_TASK_CHECK") == "1" {
		t.Skip("Skipping structure validation to prevent recursion (running inside task check)")
	}

	output := runTaskCheckForStructure(t)

	// Verify success indicators
	// Current format uses [OK] markers and box-drawing summary
	assert.Contains(t, output, "[OK]", "Output should contain success markers")
	assert.Contains(t, output, "Checks Passed", "Output should contain summary")
}

// ValidateSectionOrder verifies that sections appear in the expected order.
func ValidateSectionOrder(t *testing.T, output string, sections []Section) {
	t.Helper()

	positions := make([]int, len(sections))

	// Find position of each section
	for i, section := range sections {
		pos := strings.Index(output, section.Marker)
		require.NotEqual(t, -1, pos,
			"Section '%s' not found in output", section.Name)
		positions[i] = pos
	}

	// Verify positions are in ascending order
	for i := 0; i < len(positions)-1; i++ {
		assert.Less(t, positions[i], positions[i+1],
			"Section '%s' should appear before '%s'",
			sections[i].Name, sections[i+1].Name)
	}
}
