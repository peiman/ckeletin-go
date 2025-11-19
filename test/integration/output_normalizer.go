package integration

import (
	"regexp"
)

// NormalizePaths converts absolute paths to relative paths in the output.
// This ensures golden files work across different development environments.
// Example: /Users/peiman/dev/cli/ckeletin-go/cmd/root.go -> ./cmd/root.go
func NormalizePaths(output string) string {
	// Match any absolute path containing "ckeletin-go" followed by a path
	// Handles both Unix (/path/to/ckeletin-go/...) and potential variations
	pathPattern := regexp.MustCompile(`[^\s]*ckeletin-go/([^\s]+)`)
	return pathPattern.ReplaceAllString(output, "./$1")
}

// NormalizeTimings replaces timing values (e.g., "1.23s") with a placeholder.
// This prevents golden file tests from failing due to performance variations.
// Example: "Completed in 12.34s" -> "Completed in X.XXs"
func NormalizeTimings(output string) string {
	// Match patterns like: 0.001s, 1.23s, 123.45s
	timingPattern := regexp.MustCompile(`\d+\.\d+s`)
	return timingPattern.ReplaceAllString(output, "X.XXs")
}

// NormalizeDurations replaces duration-related timing values with placeholders.
// Similar to NormalizeTimings but catches timing patterns with context words.
// Example: "took 45.2s" -> "took X.XXs"
func NormalizeDurations(output string) string {
	// This is redundant with NormalizeTimings but kept for semantic clarity
	// Both functions normalize the same pattern
	durationPattern := regexp.MustCompile(`\d+\.\d+s`)
	return durationPattern.ReplaceAllString(output, "X.XXs")
}

// NormalizeTempPaths replaces temporary directory paths with placeholders.
// This prevents golden file tests from failing due to random temp directory names.
// Example: /var/folders/.../TestScaffoldInit1234567890/001 -> /tmp/TEMP_DIR/001
func NormalizeTempPaths(output string) string {
	// Normalize macOS temp directories
	tempPattern := regexp.MustCompile(`/var/folders/[^/]+/[^/]+/T/Test[^/]+\d+/(\d+)`)
	normalized := tempPattern.ReplaceAllString(output, "/tmp/TEMP_DIR/$1")

	// Normalize Linux temp directories
	tempPattern2 := regexp.MustCompile(`/tmp/Test[^/]+\d+/(\d+)`)
	normalized = tempPattern2.ReplaceAllString(normalized, "/tmp/TEMP_DIR/$1")

	return normalized
}

// NormalizeCheckOutput applies all normalization functions to the output.
// This is the main function used by golden file tests to ensure consistent,
// environment-independent output for comparison.
func NormalizeCheckOutput(output string) string {
	// Apply normalizations in sequence
	normalized := output
	normalized = NormalizePaths(normalized)
	normalized = NormalizeTimings(normalized)
	normalized = NormalizeDurations(normalized)
	normalized = NormalizeTempPaths(normalized)
	return normalized
}
