// internal/check/checks.go
//
// Individual check implementations for the check command.

package check

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

// checkFormat checks code formatting using goimports and gofmt
func (e *Executor) checkFormat(ctx context.Context) error {
	log.Debug().Msg("Running format check")

	// Check goimports
	cmd := exec.CommandContext(ctx, "goimports", "-l", ".")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("files need formatting:\n%s", strings.TrimSpace(string(output)))
	}

	// Check gofmt
	cmd = exec.CommandContext(ctx, "gofmt", "-l", ".")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}
	if len(strings.TrimSpace(string(output))) > 0 {
		return fmt.Errorf("files need formatting:\n%s", strings.TrimSpace(string(output)))
	}

	return nil
}

// checkLint runs go vet and golangci-lint
func (e *Executor) checkLint(ctx context.Context) error {
	log.Debug().Msg("Running lint check")

	// go vet
	cmd := exec.CommandContext(ctx, "go", "vet", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		filtered := filterLintOutput(string(output), "go vet")
		return fmt.Errorf("%s", filtered)
	}

	// golangci-lint
	cmd = exec.CommandContext(ctx, "golangci-lint", "run")
	if output, err := cmd.CombinedOutput(); err != nil {
		filtered := filterLintOutput(string(output), "golangci-lint")
		return fmt.Errorf("%s", filtered)
	}

	return nil
}

// filterLintOutput cleans up lint output to show only the issues
func filterLintOutput(output, tool string) string {
	lines := strings.Split(output, "\n")
	var issues []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip package headers and metadata
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Skip golangci-lint summary lines
		if strings.HasPrefix(trimmed, "level=") {
			continue
		}

		// Keep lines with file:line references (actual issues)
		if strings.Contains(trimmed, ".go:") {
			issues = append(issues, trimmed)
		}
	}

	if len(issues) == 0 {
		return tool + " found issues"
	}

	var sb strings.Builder
	sb.WriteString(tool + " found " + fmt.Sprintf("%d", len(issues)) + " issue(s):\n")
	for _, issue := range issues {
		sb.WriteString("  • " + issue + "\n")
	}
	return strings.TrimSpace(sb.String())
}

// checkTest runs tests with race detection and returns coverage
func (e *Executor) checkTest(ctx context.Context) error {
	log.Debug().Msg("Running test check")

	cmd := exec.CommandContext(ctx, "go", "test", "-race", "-cover", "./...")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Filter output to only show failures, not passing packages
		filtered := filterTestOutput(string(output))
		return fmt.Errorf("%s", filtered)
	}

	// Parse and store coverage
	coverage := parseCoverage(string(output))
	e.coverage = coverage

	// Call callback if set (for TUI mode)
	if e.onCoverage != nil {
		e.onCoverage(coverage)
	}

	return nil
}

// filterTestOutput extracts only the relevant failure information from go test output.
// Removes passing packages, cached results, and coverage info to show only errors.
func filterTestOutput(output string) string {
	lines := strings.Split(output, "\n")
	var result []string
	var failedPackages []string
	var failedTests []string
	inErrorBlock := false
	isCompileError := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines
		if trimmed == "" {
			inErrorBlock = false
			continue
		}

		// Skip passing packages (lines starting with "ok" or "?")
		if strings.HasPrefix(trimmed, "ok ") || strings.HasPrefix(trimmed, "? ") {
			inErrorBlock = false
			continue
		}

		// Skip lines that are just coverage info (not errors)
		if strings.Contains(trimmed, "coverage:") && !strings.Contains(trimmed, ".go:") {
			continue
		}

		// Track failed packages
		if strings.HasPrefix(trimmed, "FAIL") {
			// Extract package name from "FAIL package [duration]"
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				failedPackages = append(failedPackages, parts[1])
			}
			inErrorBlock = false
			continue
		}

		// Skip "exit status" lines
		if strings.HasPrefix(trimmed, "exit status") {
			continue
		}

		// Keep error lines (compilation errors, test failures, etc.)
		// These typically start with # (package header) or contain error info
		if strings.HasPrefix(trimmed, "#") {
			inErrorBlock = true
			isCompileError = true
			result = append(result, trimmed)
			continue
		}

		// Keep lines that look like errors (file:line: message)
		if inErrorBlock && strings.Contains(trimmed, ".go:") {
			result = append(result, trimmed)
			continue
		}

		// Track and keep --- FAIL lines for test failures
		if strings.HasPrefix(trimmed, "--- FAIL:") {
			// Extract test name from "--- FAIL: TestName (0.00s)"
			testName := extractTestName(trimmed)
			if testName != "" {
				failedTests = append(failedTests, testName)
			}
			result = append(result, trimmed)
			inErrorBlock = true
		}
	}

	// Build the final output
	var sb strings.Builder

	// Summary line
	if isCompileError {
		sb.WriteString(fmt.Sprintf("%d package(s) failed to compile\n\n", len(failedPackages)))
	} else if len(failedTests) > 0 {
		sb.WriteString(fmt.Sprintf("%d test(s) failed in %d package(s)\n\n", len(failedTests), len(failedPackages)))
	}

	if len(failedPackages) > 0 && !isCompileError {
		sb.WriteString("Failed packages:\n")
		for _, pkg := range failedPackages {
			sb.WriteString("  • " + pkg + "\n")
		}
		sb.WriteString("\n")
	}

	if len(failedTests) > 0 {
		sb.WriteString("Failed tests:\n")
		for _, test := range failedTests {
			sb.WriteString("  • " + test + "\n")
		}
		sb.WriteString("\n")
	}

	if len(result) > 0 {
		sb.WriteString("Details:\n")
		for _, line := range result {
			sb.WriteString("  " + line + "\n")
		}
	}

	if sb.Len() == 0 {
		return "tests failed (unknown error)"
	}

	return strings.TrimSpace(sb.String())
}

// extractTestName extracts the test name from a "--- FAIL: TestName (0.00s)" line
func extractTestName(line string) string {
	// Format: "--- FAIL: TestName (0.00s)"
	line = strings.TrimPrefix(line, "--- FAIL:")
	line = strings.TrimSpace(line)
	// Find the space before duration
	if idx := strings.LastIndex(line, " ("); idx > 0 {
		return line[:idx]
	}
	return line
}

// parseCoverage extracts average coverage from go test -cover output
func parseCoverage(output string) float64 {
	var total float64
	var count int

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		// Look for "coverage: XX.X% of statements"
		if idx := strings.Index(line, "coverage:"); idx != -1 {
			// Extract the percentage
			part := line[idx+len("coverage:"):]
			part = strings.TrimSpace(part)
			if pctIdx := strings.Index(part, "%"); pctIdx != -1 {
				pctStr := strings.TrimSpace(part[:pctIdx])
				var pct float64
				if _, err := fmt.Sscanf(pctStr, "%f", &pct); err == nil {
					total += pct
					count++
				}
			}
		}
	}

	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// checkDeps verifies dependency integrity
func (e *Executor) checkDeps(ctx context.Context) error {
	log.Debug().Msg("Running deps check")

	cmd := exec.CommandContext(ctx, "go", "mod", "verify")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dependency verification failed:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// checkVuln scans for vulnerabilities using govulncheck
func (e *Executor) checkVuln(ctx context.Context) error {
	log.Debug().Msg("Running vulnerability check")

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vulnerabilities found:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}
