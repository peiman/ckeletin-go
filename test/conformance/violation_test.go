//go:build conformance

// test/conformance/violation_test.go
//
// Violation tests for CKSPEC enforcement claims.
//
// Each test introduces a known violation and verifies the enforcement
// mechanism catches it. Without these tests, enforcement claims in the
// conformance report are unverified hypotheses.
//
// Per CKSPEC-ENF-006: enforcement claims above honor-system MUST be
// accompanied by a violation test that demonstrates the check catches
// a known violation.
//
// These tests create temporary .go files (named *_violation.go to avoid
// matching *_test.go exclusion patterns in validators), run the
// enforcement check, and verify it fails.

package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// projectRoot returns the absolute path to the project root.
func projectRoot(t *testing.T) string {
	t.Helper()
	abs, err := filepath.Abs("../..")
	require.NoError(t, err)
	return abs
}

// scriptPath returns the absolute path to a validation script.
func scriptPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(projectRoot(t), ".ckeletin", "scripts", name)
}

// writeViolationFile creates a temporary Go file that violates a rule.
// Returns a cleanup function that removes the file.
func writeViolationFile(t *testing.T, relPath string, content string) func() {
	t.Helper()
	absPath := filepath.Join(projectRoot(t), relPath)

	dir := filepath.Dir(absPath)
	require.DirExists(t, dir, "violation file directory must exist: %s", dir)

	err := os.WriteFile(absPath, []byte(content), 0644)
	require.NoError(t, err, "failed to write violation file")

	return func() {
		os.Remove(absPath)
	}
}

// runCheck runs a command in the project root and returns output and exit code.
func runCheck(t *testing.T, name string, args ...string) (string, int) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = projectRoot(t)

	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("failed to run %s: %v", name, err)
		}
	}
	return string(out), exitCode
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-002: Directed dependencies
// Enforcement: go-arch-lint (linter level)
// Violation: business logic imports from the command layer
// ---------------------------------------------------------------------------

func TestViolation_ARCH002_ReverseDepBusinessImportsCmd(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"internal/ping/ckspec_violation.go",
		`package ping

// Violation: business logic importing from the command layer.
// go-arch-lint must catch this reverse dependency.
import _ "github.com/peiman/ckeletin-go/cmd"
`)
	defer cleanup()

	output, exitCode := runCheck(t, "go-arch-lint", "check")
	assert.NotEqual(t, 0, exitCode,
		"go-arch-lint should fail when business logic imports cmd/\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-003: CLI framework isolation
// Enforcement: go-arch-lint (linter level)
// Violation: business logic imports Cobra directly
// ---------------------------------------------------------------------------

func TestViolation_ARCH003_BusinessImportsCobra(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// This test was originally a KNOWN GAP — depOnAnyVendor:true bypassed
	// canUse rules. Fixed by setting depOnAnyVendor:false and registering
	// all legitimate vendors in .go-arch-lint.yml.

	cleanup := writeViolationFile(t,
		"internal/ping/ckspec_violation.go",
		`package ping

// Violation: business logic importing Cobra directly.
// Only cmd/ is allowed to import the CLI framework.
import _ "github.com/spf13/cobra"
`)
	defer cleanup()

	output, exitCode := runCheck(t, "go-arch-lint", "check")
	assert.NotEqual(t, 0, exitCode,
		"go-arch-lint should fail when business logic imports Cobra\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-004: Business logic isolation
// Enforcement: go-arch-lint (linter level)
// Violation: one business logic package imports another
// ---------------------------------------------------------------------------

func TestViolation_ARCH004_BusinessImportsBusiness(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"internal/ping/ckspec_violation.go",
		`package ping

// Violation: business logic importing another business logic package.
// Business packages must be isolated from each other.
import _ "github.com/peiman/ckeletin-go/internal/docs"
`)
	defer cleanup()

	output, exitCode := runCheck(t, "go-arch-lint", "check")
	assert.NotEqual(t, 0, exitCode,
		"go-arch-lint should fail when business packages import each other\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-005: Infrastructure independence
// Enforcement: go-arch-lint (linter level)
// Violation: infrastructure imports business logic
// ---------------------------------------------------------------------------

func TestViolation_ARCH005_InfraImportsBusiness(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"internal/ui/ckspec_violation.go",
		`package ui

// Violation: infrastructure importing business logic.
// Infrastructure must not depend on upper layers.
import _ "github.com/peiman/ckeletin-go/internal/ping"
`)
	defer cleanup()

	output, exitCode := runCheck(t, "go-arch-lint", "check")
	assert.NotEqual(t, 0, exitCode,
		"go-arch-lint should fail when infrastructure imports business logic\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-007: Package location enforcement
// Enforcement: validate-package-organization.sh (script level)
// Violation: Go source file in the project root (not main.go)
// ---------------------------------------------------------------------------

func TestViolation_ARCH007_GoFileInRoot(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"ckspec_violation.go",
		`package main

// Violation: Go source file in the project root.
// Only main.go and main_test.go should be at root.
func ckspecViolation() {}
`)
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-package-organization.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-package-organization.sh should fail when Go file is in root\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-OUT-005: Output isolation from business logic
// Enforcement: validate-output-patterns.sh (script level)
// Violation: business logic uses fmt.Println directly
// ---------------------------------------------------------------------------

func TestViolation_OUT005_BusinessUsesFmtPrint(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"internal/ping/ckspec_violation.go",
		`package ping

import "fmt"

// Violation: business logic printing directly to stdout.
// Must use internal/ui for output.
func ckspecViolation() {
	fmt.Println("direct stdout access")
}
`)
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-output-patterns.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-output-patterns.sh should fail when business logic uses fmt.Println\nOutput: %s", output)
}
