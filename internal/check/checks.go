// internal/check/checks.go
//
// Individual check implementations for the check command.
//
// Check Status Tracking:
// =====================
// Native Go (5):
//   ✅ format      - goimports + gofmt
//   ✅ lint        - go vet + golangci-lint
//   ✅ test        - go test -race -cover
//   ✅ deps        - go mod verify
//   ✅ vuln        - govulncheck
//
// Shell Delegators - Environment (2):
//   ✅ go-version  - check-go-version.sh
//   ✅ tools       - install_tools.sh --check
//
// Shell Delegators - Architecture Validation (10):
//   ✅ defaults           - check-defaults.sh (ADR-002)
//   ✅ commands           - validate-command-patterns.sh (ADR-001)
//   ✅ constants          - check-constants.sh (ADR-005)
//   ✅ task-naming        - validate-task-naming.sh (ADR-000)
//   ✅ architecture       - validate-architecture.sh (ADR-008)
//   ✅ layering           - validate-layering.sh (ADR-009)
//   ✅ package-org        - validate-package-organization.sh (ADR-010)
//   ✅ config-consumption - validate-config-consumption.sh (ADR-002)
//   ✅ output-patterns    - validate-output-patterns.sh (ADR-012)
//   ✅ security-patterns  - validate-security-patterns.sh (ADR-004)
//
// Shell Delegators - Security (2):
//   ✅ secrets     - check-secrets.sh
//   ✅ sast        - check-sast.sh
//
// Shell Delegators - Dependencies (3):
//   ✅ license-source - check-licenses-source.sh
//   ✅ license-binary - check-licenses-binary.sh
//   ✅ sbom-vulns     - check-sbom-vulns.sh
//
// Total: 22 checks (5 native Go + 17 shell delegators)

package check

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// scriptsDir is the base directory for check scripts
const scriptsDir = ".ckeletin/scripts"

// shellCheck creates a check function that delegates to a shell script.
// The script path is relative to the project root.
// Note: The script name must be a constant from this package's predefined scripts.
func (e *checkMethods) shellCheck(scriptName string, args ...string) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		scriptPath := filepath.Join(scriptsDir, scriptName)
		log.Debug().Str("script", scriptPath).Strs("args", args).Msg("Running shell check")

		cmdArgs := append([]string{scriptPath}, args...)
		// #nosec G204 -- scriptName is from predefined constants, not user input
		cmd := exec.CommandContext(ctx, "bash", cmdArgs...)

		output, err := cmd.CombinedOutput()
		if err != nil {
			// Extract meaningful error from output
			errMsg := extractShellError(string(output))
			if errMsg == "" {
				errMsg = fmt.Sprintf("script failed: %v", err)
			}
			return fmt.Errorf("%s", errMsg)
		}
		return nil
	}
}

// extractShellError extracts the most relevant error message from shell output.
// It looks for common error patterns and removes noise.
func extractShellError(output string) string {
	lines := strings.Split(output, "\n")
	var errorLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		// Skip success indicators
		if strings.HasPrefix(trimmed, "✅") || strings.HasPrefix(trimmed, "[0;32m✅") {
			continue
		}

		// Keep error indicators
		if strings.HasPrefix(trimmed, "❌") || strings.HasPrefix(trimmed, "✗") ||
			strings.Contains(trimmed, "error") || strings.Contains(trimmed, "Error") ||
			strings.Contains(trimmed, "failed") || strings.Contains(trimmed, "FAIL") {
			// Strip ANSI codes
			clean := stripANSI(trimmed)
			if clean != "" {
				errorLines = append(errorLines, clean)
			}
		}
	}

	if len(errorLines) == 0 {
		// Return last non-empty lines as fallback
		for i := len(lines) - 1; i >= 0 && len(errorLines) < 3; i-- {
			trimmed := strings.TrimSpace(lines[i])
			if trimmed != "" {
				errorLines = append([]string{stripANSI(trimmed)}, errorLines...)
			}
		}
	}

	return strings.Join(errorLines, "\n")
}

// stripANSI removes ANSI escape codes from a string
func stripANSI(s string) string {
	// Simple approach: remove escape sequences
	result := s
	for strings.Contains(result, "\033[") {
		start := strings.Index(result, "\033[")
		end := start + 2
		// Find the end of the escape sequence (letter)
		for end < len(result) && !isLetter(result[end]) {
			end++
		}
		if end < len(result) {
			end++ // Include the letter
		}
		result = result[:start] + result[end:]
	}
	return result
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

// checkFormat checks code formatting using goimports and gofmt
func (e *checkMethods) checkFormat(ctx context.Context) error {
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
func (e *checkMethods) checkLint(ctx context.Context) error {
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
func (e *checkMethods) checkTest(ctx context.Context) error {
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
func (e *checkMethods) checkDeps(ctx context.Context) error {
	log.Debug().Msg("Running deps check")

	cmd := exec.CommandContext(ctx, "go", "mod", "verify")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dependency verification failed:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}

// checkVuln scans for vulnerabilities using govulncheck
func (e *checkMethods) checkVuln(ctx context.Context) error {
	log.Debug().Msg("Running vulnerability check")

	cmd := exec.CommandContext(ctx, "govulncheck", "./...")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("vulnerabilities found:\n%s", strings.TrimSpace(string(output)))
	}
	return nil
}
