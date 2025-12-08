package integration

import (
	"os"
	"os/exec"
	"testing"

	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
)

// runCheckSummary executes the check summary script and returns its output.
// This helper function is used by golden file tests to capture the
// summary output that was restructured to be "neat and nice to read".
func runCheckSummary(t *testing.T) string {
	t.Helper()

	// Get the project root directory (two levels up from test/integration)
	projectRoot, err := os.Getwd()
	require.NoError(t, err, "Failed to get current working directory")

	// Change to project root if we're in test/integration
	// This ensures script paths are correct
	if err := os.Chdir("../.."); err == nil {
		// Store original directory to restore later
		defer func() {
			_ = os.Chdir(projectRoot)
		}()
	}

	// Execute the check summary script (check framework location first)
	scriptPath := ".ckeletin/scripts/check-summary.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		scriptPath = "scripts/check-summary.sh" // Fallback to old location
	}
	cmd := exec.Command("bash", scriptPath)
	output, err := cmd.CombinedOutput()

	// For golden file testing, we want to capture successful runs
	require.NoError(t, err, "check-summary.sh should succeed for golden file test.\nOutput:\n%s", string(output))

	return string(output)
}

// TestCheckSummaryGolden is the golden file test for the check summary output.
// It tests the "neat and nice" restructured summary from scripts/check-summary.sh.
//
// To update the golden file:
//
//	GOLDEN_UPDATE=1 go test ./test/integration -run Golden
//
// IMPORTANT: Always manually review golden file changes before committing!
//
//	git diff testdata/check-summary.golden
func TestCheckSummaryGolden(t *testing.T) {
	// Skip in short mode
	if testing.Short() {
		t.Skip("Skipping golden file test in short mode")
	}

	// Run check summary script and capture output
	output := runCheckSummary(t)

	// The summary script output is already clean and deterministic,
	// but we still normalize in case there are any timing references
	normalized := NormalizeCheckOutput(output)

	// Create golden file helper
	g := goldie.New(t,
		goldie.WithFixtureDir("testdata"),
		goldie.WithNameSuffix(".golden"),
	)

	// Check if we should update golden files
	// Support both -update flag and GOLDEN_UPDATE env var
	if os.Getenv("GOLDEN_UPDATE") != "" {
		// Update mode: write the normalized output to golden file
		g.Update(t, "check-summary", []byte(normalized))
	} else {
		// Normal mode: compare against golden file
		g.Assert(t, "check-summary", []byte(normalized))
	}
}

// TestGoldenFilesExist verifies that golden files are present and readable.
// This test helps catch accidental deletion of golden files.
func TestGoldenFilesExist(t *testing.T) {
	goldenFiles := []string{
		"testdata/check-summary.golden",
	}

	for _, file := range goldenFiles {
		t.Run(file, func(t *testing.T) {
			_, err := os.Stat(file)
			require.NoError(t, err, "Golden file should exist: %s\nRun: GOLDEN_UPDATE=1 go test ./test/integration -run Golden", file)
		})
	}
}
