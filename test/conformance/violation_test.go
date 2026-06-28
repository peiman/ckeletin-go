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
// Violation: a run* function exceeds the 35-line hard limit (the failing
// gate; whole-file size above 80 lines is advisory-only and cannot prove
// enforcement). The injected file is named ping_violation.go so the
// metadata check resolves via the parent config (ping_config.go), and it
// carries MustNewCommand/MustAddToRoot wiring — every earlier check passes,
// isolating the line-count gate as the only possible error.
// ---------------------------------------------------------------------------

func TestViolation_ARCH006_OversizedCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	var b strings.Builder
	b.WriteString(`// cmd/ping_violation.go

package cmd

import (
	"github.com/spf13/cobra"
)

var pingViolationCmd = MustNewCommand(commands.PingMetadata, runPingViolation)

func init() {
	MustAddToRoot(pingViolationCmd)
}

func runPingViolation(cmd *cobra.Command, args []string) error {
`)
	// func line + 38 padding statements + return + closing brace = 41 lines,
	// past the 35-line hard limit. Plain assignments keep the business-logic
	// heuristic (check 6) quiet so the line-count gate is the sole finding.
	for i := 0; i < 38; i++ {
		b.WriteString("\t_ = \"padding statement\"\n")
	}
	b.WriteString("\treturn nil\n}\n")

	cleanup := writeViolationFile(t, "cmd/ping_violation.go", b.String())
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-command-patterns.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-command-patterns.sh must fail when a run* function exceeds the 35-line hard limit\nOutput: %s", output)
	assert.Contains(t, output,
		"ping_violation: runPingViolation() is 41 lines (target 30, hard limit 35) - move logic to internal/",
		"the failure must come from the run* line-count gate, not another check\nOutput: %s", output)
	assert.NotContains(t, output, "ping_violation: Missing metadata file",
		"the injected file must pass the metadata check so the line-count error is isolated\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-006: Entry point minimality (whitelist escape hatch)
// Enforcement: validate-command-patterns.sh (script level)
// Violation: a bare // ckeletin:allow-custom-command marker with no
// justification. The line above is a file-path header comment, which the
// script explicitly refuses to count as a reason, and the line below is the
// package clause — so the marker carries no justification anywhere the
// script looks. The marker makes the script skip every other check for the
// file, isolating the justification error.
// ---------------------------------------------------------------------------

func TestViolation_ARCH006_MarkerWithoutJustification(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	cleanup := writeViolationFile(t,
		"cmd/ckspec_violation.go",
		`// cmd/ckspec_violation.go
// ckeletin:allow-custom-command
package cmd
`)
	defer cleanup()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "validate-command-patterns.sh"))
	assert.NotEqual(t, 0, exitCode,
		"validate-command-patterns.sh must fail when the whitelist marker has no justification\nOutput: %s", output)
	assert.Contains(t, output,
		"ckspec_violation: ckeletin:allow-custom-command marker has no justification (add a short reason on or next to the marker line)",
		"the failure must come from the marker-justification check\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-TEST-003: Dependency injection over mocking
// Enforcement: grep for mock frameworks in go.mod (script level)
// Violation: go.mod references gomock
// ---------------------------------------------------------------------------

func TestViolation_TEST003_MockFrameworkImport(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// The semgrep rule ckeletin-no-mock-frameworks catches mock framework
	// imports in Go files. Create a file with a gomock import and verify
	// semgrep detects it.
	cleanup := writeViolationFile(t,
		"cmd/ckspec_violation.go",
		`package cmd

import "go.uber.org/mock/gomock"

// Violation: mock framework import.
// Semgrep rule ckeletin-no-mock-frameworks should catch this.
var _ = gomock.Controller{}
`)
	defer cleanup()

	// Run semgrep with local rules only (matches task check:sast)
	output, exitCode := runCheck(t, "semgrep", "scan", "--config", ".semgrep.yml",
		"--error", "--quiet", "cmd/ckspec_violation.go")
	assert.NotEqual(t, 0, exitCode,
		"semgrep should flag mock framework import\nOutput: %s", output)
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
// Enforcement: task conform validates every requirement ID from the spec's
// requirements.json is present in the mapping (count is machine-derived, not
// hand-maintained — see the ENF-008 drift guard)
// Violation: mapping file missing a requirement
// ---------------------------------------------------------------------------

func TestViolation_ENF005_IncompleteMapping(t *testing.T) {
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
// CKSPEC-ENF-006: Violation tests for enforcement claims
// Enforcement: conform.sh flags missing violation tests as feedback signals
// Violation: remove a violation test from mapping, verify generator flags it
// ---------------------------------------------------------------------------

func TestViolation_ENF006_MissingViolationTestFlagged(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	// Verify ENF-006 enforcement: the conform script contains logic that
	// flags requirements claiming enforcement above honor-system but
	// lacking proof. We verify this by checking that the script has the
	// detection logic AND that the current mapping carries proof for every
	// enforcement claim above honor-system.
	//
	// We can't run conform.sh here because it runs `go test -tags conformance`
	// which would recursively invoke this test. Instead, verify the mechanism
	// exists and the mapping is consistent.
	root := projectRoot(t)

	// 1. The conform script checks for missing violation tests
	script, err := os.ReadFile(filepath.Join(root, ".ckeletin", "scripts", "conform.sh"))
	require.NoError(t, err)
	assert.Contains(t, string(script), "violation_test",
		"conform.sh must check for missing violation tests")
	assert.Contains(t, string(script), "FEEDBACK_FILE",
		"conform.sh must collect feedback signals")

	// 2. Every enforcement claim above honor-system (linter, sast, script,
	// ci, test) must carry proof: a violation test or, per spec v0.4.0+, a
	// written violation_evidence analysis. This mirrors what conform.sh
	// enforces (lines under "ENF-006" there) so the two cannot drift apart.
	mapping, err := os.ReadFile(filepath.Join(root, "conformance-mapping.yaml"))
	require.NoError(t, err)
	content := string(mapping)

	needsProof := func(level string) bool {
		return level != "" && level != "honor-system"
	}

	lines := strings.Split(content, "\n")
	currentReq := ""
	currentLevel := ""
	hasViolationTest := false
	hasViolationEvidence := false
	mismatches := []string{}

	flush := func() {
		if currentReq != "" && needsProof(currentLevel) && !hasViolationTest && !hasViolationEvidence {
			mismatches = append(mismatches, currentReq+" ("+currentLevel+")")
		}
	}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "CKSPEC-") && strings.HasSuffix(trimmed, ":") {
			flush() // close out the previous requirement
			currentReq = strings.TrimSuffix(trimmed, ":")
			currentLevel = ""
			hasViolationTest = false
			hasViolationEvidence = false
		}
		if strings.HasPrefix(trimmed, "enforcement_level:") {
			currentLevel = strings.TrimSpace(strings.TrimPrefix(trimmed, "enforcement_level:"))
		}
		if strings.Contains(trimmed, "TestViolation_") {
			hasViolationTest = true
		}
		if strings.HasPrefix(trimmed, "violation_evidence:") {
			hasViolationEvidence = true
		}
	}
	flush() // close out the last requirement

	assert.Empty(t, mismatches,
		"Enforcement claims above honor-system must have a violation test or violation_evidence: %v", mismatches)
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

// ---------------------------------------------------------------------------
// Conformance tooling integrity: the mapping MUST be valid YAML.
// Enforcement: conform.sh parses conformance-mapping.yaml with yq and fails the
// build — and therefore the release gate (CKSPEC-ENF-009) — if it does not parse.
// This guards the conform.sh parser itself (not a spec requirement): before this
// gate existed, a lenient text scan silently tolerated an invalid-YAML mapping
// (e.g. a check string with an unescaped regex backslash), so a broken mapping
// could pass conformance. Violation: an unparseable mapping must be rejected.
// ---------------------------------------------------------------------------

func TestViolation_ConformMapping_InvalidYAMLRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	mappingPath := filepath.Join(root, "conformance-mapping.yaml")

	original, err := os.ReadFile(mappingPath)
	require.NoError(t, err)

	// Append a top-level scalar with an invalid YAML escape (\q). This is the
	// same defect class as a regex check written as a plain double-quoted string
	// ("...\.foo"): valid to a text scan, rejected by any real YAML parser.
	violated := string(original) + "\nbroken_invalid_yaml: \"x\\q\"\n"
	require.NoError(t, os.WriteFile(mappingPath, []byte(violated), 0644))
	defer func() {
		// Restore the original mapping regardless of assertion outcome.
		os.WriteFile(mappingPath, original, 0644)
	}()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "conform.sh"))

	assert.NotEqual(t, 0, exitCode,
		"conform.sh must fail when the mapping is not valid YAML\nOutput: %s", output)
	assert.Contains(t, output, "not valid YAML",
		"conform.sh must report the YAML parse failure\nOutput: %s", output)
	// The gate must run before the check loop so a broken mapping exits cheaply
	// and never re-invokes the conformance suite (recursion-free, like ENF-005).
	assert.NotContains(t, output, "Running checks",
		"the YAML gate must fail fast, before running any checks\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// Conformance tooling integrity: yq must be mikefarah/yq v4 (Go).
// Enforcement: conform.sh probes strenv() — a mikefarah-only function it relies
// on for safe field access — and fails with a clear message if the yq on PATH is
// the unrelated Python yq (kislyuk), instead of a cryptic mid-run strenv error.
// Violation: a non-mikefarah yq earlier on PATH must be rejected up front.
// ---------------------------------------------------------------------------

func TestViolation_ConformMapping_NonMikefarahYqRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests shell out to conform.sh")
	}

	root := projectRoot(t)

	// A stand-in "yq" that parses ('.' succeeds) but lacks strenv (any other
	// invocation fails) — the observable behaviour of the Python yq for our use.
	fakeBin := t.TempDir()
	fakeYq := "#!/bin/sh\n" +
		"if [ \"$1\" = \".\" ]; then exit 0; fi\n" +
		"echo 'yq (python stand-in): unknown function strenv' >&2\nexit 1\n"
	require.NoError(t, os.WriteFile(filepath.Join(fakeBin, "yq"), []byte(fakeYq), 0755))

	cmd := exec.Command("bash", scriptPath(t, "conform.sh"))
	cmd.Dir = root
	cmd.Env = append(os.Environ(), "PATH="+fakeBin+string(os.PathListSeparator)+os.Getenv("PATH"))
	out, err := cmd.CombinedOutput()
	output := string(out)

	require.Error(t, err, "conform.sh must fail when yq is not mikefarah/yq\nOutput: %s", output)
	assert.Contains(t, output, "not mikefarah/yq",
		"conform.sh must report the wrong yq variant clearly\nOutput: %s", output)
	assert.NotContains(t, output, "Running checks",
		"the yq-variant gate must fail fast, before running any checks\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// Conformance tooling integrity: the published report MUST match the mapping.
// Enforcement: conform.sh regenerates conformance-report.json from the mapping
// (gen-conformance-report.sh) and fails — and so does the release gate — if the
// committed report has drifted. This keeps the machine-readable report (which the
// spec repo aggregates instead of hand-authoring conformance/ckeletin-go.yaml)
// truthful. Violation: a stale committed report must be rejected.
// ---------------------------------------------------------------------------

func TestViolation_ConformReport_OutOfSyncRejected(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	reportPath := filepath.Join(root, "conformance-report.json")

	original, err := os.ReadFile(reportPath)
	require.NoError(t, err)

	// Corrupt the committed report so it no longer matches the mapping.
	require.NoError(t, os.WriteFile(reportPath, []byte("{\"summary\": {\"met\": 999}}\n"), 0644))
	defer func() {
		// Restore the original report regardless of assertion outcome.
		os.WriteFile(reportPath, original, 0644)
	}()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "conform.sh"))

	assert.NotEqual(t, 0, exitCode,
		"conform.sh must fail when conformance-report.json is out of sync\nOutput: %s", output)
	assert.Contains(t, output, "out of sync",
		"conform.sh must report the report drift\nOutput: %s", output)
	assert.NotContains(t, output, "Running checks",
		"the report-sync guard must fail fast, before running any checks\nOutput: %s", output)
}

// TestViolation_AGENT006_CatalogCommandRemoved verifies CKSPEC-AGENT-006's
// enforcement: removing the `catalog` command (the CLI's machine-readable
// command surface) must fail the conformance check. The command file and its
// test are removed together so the cmd package still compiles and only the
// AGENT-006 `test -f cmd/catalog.go` check fires.
func TestViolation_AGENT006_CatalogCommandRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}

	root := projectRoot(t)
	files := []string{
		filepath.Join(root, "cmd", "catalog.go"),
		filepath.Join(root, "cmd", "catalog_test.go"),
	}
	saved := map[string][]byte{}
	for _, f := range files {
		b, err := os.ReadFile(f)
		require.NoError(t, err)
		saved[f] = b
		require.NoError(t, os.Remove(f))
	}
	defer func() {
		for f, b := range saved {
			os.WriteFile(f, b, 0644)
		}
	}()

	output, exitCode := runCheck(t, "bash", scriptPath(t, "conform.sh"))

	assert.NotEqual(t, 0, exitCode,
		"conform.sh must fail when the catalog command (CKSPEC-AGENT-006) is removed\nOutput: %s", output)
	assert.Contains(t, output, "CKSPEC-AGENT-006",
		"conform.sh must name the failed requirement\nOutput: %s", output)
}

// ---------------------------------------------------------------------------
// CKSPEC-AGENT-002: No provider-specific content in universal guide
// Enforcement: grep-level conform check (script level) — provider names in
// AGENTS.md are rejected unless the line is a provider-file reference or
// cross-reference (CLAUDE.md, .cursorrules, copilot-instructions.md, ...).
// Violation: a provider-specific instruction line must be caught.
// The check runs against a doctored COPY of AGENTS.md in a temp dir, so this
// test neither mutates the real guide nor re-runs conform.sh (no recursion).
// ---------------------------------------------------------------------------

func TestViolation_AGENT002_ProviderInstructionCaught(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests shell out to the conform check")
	}

	root := projectRoot(t)

	// Pull the AGENT-002 check from the mapping (SSOT) so this test exercises
	// the exact command conform.sh runs, not a re-implementation of it.
	rawCheck, exitCode := runCheck(t, "yq",
		`.requirements["CKSPEC-AGENT-002"].checks[0] // ""`,
		filepath.Join(root, "conformance-mapping.yaml"))
	require.Equal(t, 0, exitCode, "yq must read the mapping")
	checkCmd := strings.TrimSpace(rawCheck)
	require.NotEmpty(t, checkCmd,
		"CKSPEC-AGENT-002 must declare a grep-level check in the mapping")

	agents, err := os.ReadFile(filepath.Join(root, "AGENTS.md"))
	require.NoError(t, err)

	// runAgainst runs the mapping's check in a temp dir holding the given
	// AGENTS.md content, returning the check's exit code.
	// `bash -c` is deliberate and safe here: checkCmd comes from the repo's
	// own conformance-mapping.yaml (developer-controlled, allowlisted), and
	// conform.sh executes the very same string the very same way.
	runAgainst := func(t *testing.T, content []byte) int {
		t.Helper()
		dir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "AGENTS.md"), content, 0644))
		cmd := exec.Command("bash", "-c", checkCmd) //nolint:gosec // repo-controlled check from the mapping, mirrors conform.sh
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		if err != nil {
			exitErr, ok := err.(*exec.ExitError)
			require.True(t, ok, "failed to run AGENT-002 check: %v (output: %s)", err, out)
			return exitErr.ExitCode()
		}
		return 0
	}

	assert.Equal(t, 0, runAgainst(t, agents),
		"the AGENT-002 check must pass on the current AGENTS.md")

	violated := append(append([]byte{}, agents...),
		[]byte("\nClaude should always run `task check` before committing.\n")...)
	assert.NotEqual(t, 0, runAgainst(t, violated),
		"the AGENT-002 check must catch a provider-specific instruction line")
}

// TestViolation_ENF008_AnchorResolution proves the ENF-008 anchor-resolution
// gate flags a dangling violation_test anchor — the regression guard for #15
// (issue #127). It exercises the real shell helper
// .ckeletin/scripts/lib/anchor.sh::anchor_resolve against fixtures, so it tests
// the actual gate logic (not a copy) and never invokes conform.sh — no
// recursion. anchor_resolve appends to the same FAIL_FILE the "met but
// unanchored" gate uses, which conform turns into a non-zero exit.
func TestViolation_ENF008_AnchorResolution(t *testing.T) {
	lib := filepath.Join(projectRoot(t), ".ckeletin", "scripts", "lib", "anchor.sh")
	require.FileExists(t, lib)

	dir := t.TempDir()
	fixture := filepath.Join(dir, "fixture.go")
	require.NoError(t, os.WriteFile(fixture,
		[]byte("package x\nfunc TestReal(t *T) {}\nfunc (s *S) TestMethod(t *T) {}\n"), 0o644))

	// run sources anchor.sh, calls anchor_resolve on the anchor, and returns
	// what was recorded to the fail file (empty == resolved, no failure).
	// Values are passed as bash positional args ($1/$2/$3), never interpolated
	// into the script string — no shell-injection surface.
	const script = `source "$1"; anchor_resolve "$2" "$3" REQ`
	run := func(anchor string) string {
		ff := filepath.Join(dir, "fail")
		_ = os.Remove(ff)
		out, err := exec.Command("bash", "-c", script, "_", lib, anchor, ff).CombinedOutput()
		require.NoError(t, err, "anchor_resolve must not error: %s", out)
		require.Empty(t, strings.TrimSpace(string(out)),
			"anchor_resolve records to the fail file, not stdout")
		b, _ := os.ReadFile(ff)
		return string(b)
	}

	// Resolving anchors record nothing.
	assert.Empty(t, run(fixture+"::TestReal"), "existing free function resolves")
	assert.Empty(t, run(fixture+"::TestMethod"), "method-style test resolves")
	assert.Empty(t, run(fixture), "bare existing file (no symbol) resolves")

	// Dangling anchors are flagged as such — the regression the gate exists to catch.
	assert.Contains(t, run(filepath.Join(dir, "missing.go")+"::TestReal"), "file not found",
		"missing file must be flagged dangling")
	assert.Contains(t, run(fixture+"::TestGone"), "not found",
		"missing symbol must be flagged dangling")
	assert.Contains(t, run(fixture+"::"), "empty symbol",
		"empty symbol after :: must be flagged dangling")
}
