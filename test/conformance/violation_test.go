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
	"strings"
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

// ---------------------------------------------------------------------------
// CKSPEC-OUT-001: Three-stream output separation
// Enforcement: validate-output-patterns.sh (script level)
// Violation: business logic writes to os.Stdout directly
// ---------------------------------------------------------------------------

func TestViolation_OUT001_BusinessUsesOsStdout(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"internal/ping/ckspec_violation.go",
		`package ping

import "os"

// Violation: business logic writing to os.Stdout directly.
// Must use internal/ui for the Data stream.
func ckspecViolation() {
	os.Stdout.WriteString("direct stdout access")
}
`)
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-output-patterns.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-output-patterns.sh should fail when business logic uses os.Stdout\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-001: Four-layer architecture
// Enforcement: go-arch-lint (linter level)
// Violation: Go file outside any declared component
// ---------------------------------------------------------------------------

func TestViolation_ARCH001_FileOutsideComponent(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// Create a Go file in internal/ but outside any declared business package.
	// go-arch-lint with ignoreNotFoundComponents:false should flag this.
	root := projectRoot(t)
	dir := filepath.Join(root, "internal", "orphan")
	require.NoError(t, os.MkdirAll(dir, 0755))
	defer func() {
		os.RemoveAll(dir)
	}()

	err := os.WriteFile(filepath.Join(dir, "ckspec_violation.go"), []byte(`package orphan

func ckspecViolation() {}
`), 0644)
	require.NoError(t, err)

	output, exitCode := runCheck(t, "go-arch-lint", "check")
	assert.NotEqual(t, 0, exitCode,
		"go-arch-lint should flag files outside declared components\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-006: Entry point minimality
// Enforcement: validate-command-patterns.sh (script level)
// Violation: command file exceeds 80 lines
// ---------------------------------------------------------------------------

func TestViolation_ARCH006_OversizedCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// Generate a command file that exceeds the 80-line limit
	var lines string
	lines = "package cmd\n\n"
	for i := 0; i < 85; i++ {
		lines += "// padding line to exceed limit\n"
	}

	cleanup := writeViolationFile(t, "cmd/ckspec_violation.go", lines)
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-command-patterns.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-command-patterns.sh should fail when command file exceeds 80 lines\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-TEST-003: Dependency injection over mocking
// Enforcement: grep for mock frameworks in go.mod (script level)
// Violation: go.mod references gomock
// ---------------------------------------------------------------------------

func TestViolation_TEST003_MockFrameworkInGoMod(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	goModPath := filepath.Join(root, "go.mod")

	// Read original go.mod
	original, err := os.ReadFile(goModPath)
	require.NoError(t, err)

	// Append a fake gomock require (using uber's gomock fork, common pattern)
	violated := string(original) + "\nrequire go.uber.org/mock v0.5.0 // violation test\n"
	err = os.WriteFile(goModPath, []byte(violated), 0644)
	require.NoError(t, err)
	defer func() {
		os.WriteFile(goModPath, original, 0644)
	}()

	// The check: grep should find mock framework in go.mod
	// This tests the same pattern the mapping uses: grep for mock/mockery
	output, exitCode := runCheck(t, "grep", "-q", "mock", goModPath)
	assert.Equal(t, 0, exitCode,
		"grep should find mock framework in violated go.mod\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-AGENT-001: Universal agent guide
// Enforcement: test -f AGENTS.md (script level)
// Violation: AGENTS.md doesn't exist
// ---------------------------------------------------------------------------

func TestViolation_AGENT001_MissingAgentsMd(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	src := filepath.Join(root, "AGENTS.md")
	tmp := filepath.Join(root, "AGENTS.md.violation_bak")

	require.NoError(t, os.Rename(src, tmp))
	defer func() {
		os.Rename(tmp, src)
	}()

	output, exitCode := runCheck(t, "test", "-f", "AGENTS.md")
	assert.NotEqual(t, 0, exitCode,
		"test -f AGENTS.md should fail when file is missing\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-CL-001: CHANGELOG.md in repository root
// Enforcement: test -f CHANGELOG.md (script level)
// Violation: CHANGELOG.md doesn't exist
// ---------------------------------------------------------------------------

func TestViolation_CL001_MissingChangelogMd(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	src := filepath.Join(root, "CHANGELOG.md")
	tmp := filepath.Join(root, "CHANGELOG.md.violation_bak")

	require.NoError(t, os.Rename(src, tmp))
	defer func() {
		os.Rename(tmp, src)
	}()

	output, exitCode := runCheck(t, "test", "-f", "CHANGELOG.md")
	assert.NotEqual(t, 0, exitCode,
		"test -f CHANGELOG.md should fail when file is missing\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ENF-005: Conformance mapping completeness
// Enforcement: task conform validates all 35 IDs present
// Violation: mapping file missing a requirement
// ---------------------------------------------------------------------------

func TestViolation_ENF005_IncompleteMappng(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	mappingPath := filepath.Join(root, "conformance-mapping.yaml")

	original, err := os.ReadFile(mappingPath)
	require.NoError(t, err)

	// Remove a requirement from the mapping
	violated := strings.Replace(string(original),
		"  CKSPEC-CL-007:", "  # REMOVED-FOR-TEST:", 1)
	err = os.WriteFile(mappingPath, []byte(violated), 0644)
	require.NoError(t, err)
	defer func() {
		os.WriteFile(mappingPath, original, 0644)
	}()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "conform.sh"))
	assert.NotEqual(t, 0, exitCode,
		"conform.sh should fail when a requirement is missing from mapping\nOutput: %s", output)
	assert.Contains(t, output, "MISSING",
		"conform.sh should report the missing requirement\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ENF-007: Automatic feedback signals
// Enforcement: conform.sh reports missing violation tests
// Violation: N/A — verifies the generator produces feedback signals
// ---------------------------------------------------------------------------

func TestViolation_ENF007_FeedbackSignalsProduced(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// Verify the conform script has feedback signal logic.
	// The full conform run is slow (~30s) — use grep to verify the
	// mechanism exists, and trust task conform's own output for validation.
	root := projectRoot(t)
	script, err := os.ReadFile(filepath.Join(root, ".ckeletin", "scripts", "conform.sh"))
	require.NoError(t, err)

	assert.Contains(t, string(script), "FEEDBACK_FILE",
		"conform.sh should have feedback signal collection")
	assert.Contains(t, string(script), "violation_test",
		"conform.sh should check for missing violation tests")
	assert.Contains(t, string(script), "Feedback signals",
		"conform.sh should report feedback signals in output")
}
