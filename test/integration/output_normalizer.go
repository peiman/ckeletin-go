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
// Handles all duration patterns including: "Completed in 12.34s", "took 45.2s", etc.
// Example: "Completed in 12.34s" -> "Completed in X.XXs"
func NormalizeTimings(output string) string {
	// Match patterns like: 0.001s, 1.23s, 123.45s
	timingPattern := regexp.MustCompile(`\d+\.\d+s`)
	return timingPattern.ReplaceAllString(output, "X.XXs")
}

// NormalizeTempPaths replaces temporary directory paths with placeholders.
// This prevents golden file tests from failing due to random temp directory names.
// Supports macOS, Linux, and Windows temp directory patterns.
//
// Examples:
//   - macOS: /var/folders/.../TestScaffoldInit1234567890/001 -> /tmp/TEMP_DIR/001
//   - Linux: /tmp/TestScaffoldInit1234567890/001 -> /tmp/TEMP_DIR/001
//   - Windows: C:\Users\...\AppData\Local\Temp\TestScaffoldInit1234567890\001 -> /tmp/TEMP_DIR/001
func NormalizeTempPaths(output string) string {
	// Normalize macOS temp directories
	tempPattern := regexp.MustCompile(`/var/folders/[^/]+/[^/]+/T/Test[^/]+\d+/(\d+)`)
	normalized := tempPattern.ReplaceAllString(output, "/tmp/TEMP_DIR/$1")

	// Normalize Linux temp directories
	tempPattern2 := regexp.MustCompile(`/tmp/Test[^/]+\d+/(\d+)`)
	normalized = tempPattern2.ReplaceAllString(normalized, "/tmp/TEMP_DIR/$1")

	// Normalize Windows temp directories
	// Matches: C:\Users\username\AppData\Local\Temp\TestScaffoldInit1234567890\001
	// Also handles: C:/Users/username/AppData/Local/Temp/... (forward slashes)
	tempPattern3 := regexp.MustCompile(`[A-Za-z]:[/\\]Users[/\\][^/\\]+[/\\]AppData[/\\]Local[/\\]Temp[/\\]Test[^/\\]+\d+[/\\](\d+)`)
	normalized = tempPattern3.ReplaceAllString(normalized, "/tmp/TEMP_DIR/$1")

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
	normalized = NormalizeTempPaths(normalized)
	return normalized
}
