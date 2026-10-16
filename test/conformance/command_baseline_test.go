//go:build conformance

// test/conformance/command_baseline_test.go

package conformance

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// CKSPEC-ARCH-006: Entry point minimality (adoption ratchet)
// Enforcement: validate-command-patterns.sh with .ckeletin-command-baseline
//
// A project that adopts the 35-line run* limit with long commands already in
// place records them in .ckeletin-command-baseline. A recorded function at its
// recorded length warns. It fails when it grows, when its entry is stale
// (shorter, or back under the limit), and an unrecorded long function fails as
// before. So the list can only shrink.
// ---------------------------------------------------------------------------

const baselineFile = ".ckeletin-command-baseline"

// pingCommandWithRun returns a command file whose runPingViolation is exactly
// lines long, wired so every other check passes (see ARCH006_OversizedCommand).
func pingCommandWithRun(lines int) string {
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
	// func line + padding + return + closing brace = lines.
	for i := 0; i < lines-3; i++ {
		b.WriteString("\t_ = \"padding statement\"\n")
	}
	b.WriteString("\treturn nil\n}\n")
	return b.String()
}

func writeBaseline(t *testing.T, content string) func() {
	t.Helper()
	path := filepath.Join(projectRoot(t), baselineFile)
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "%s already exists in the project root; the test would overwrite it", baselineFile)
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))
	return func() { os.Remove(path) }
}

func runCommandPatterns(t *testing.T, args ...string) (string, int) {
	t.Helper()
	return runCheck(t, "bash", append([]string{scriptPath(t, "validate-command-patterns.sh")}, args...)...)
}

func TestBaseline_RecordedLongFunctionWarns(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "# long run* functions recorded at adoption\nping_violation runPingViolation 41\n")()

	output, exitCode := runCommandPatterns(t)
	assert.Equal(t, 0, exitCode, "a function at its recorded length must not fail\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is 41 lines, recorded in .ckeletin-command-baseline (hard limit 35) - shrink it and lower the entry",
		"a recorded function still shows as a warning\nOutput: %s", output)
}

func TestBaseline_GrowthPastTheRecordFails(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "ping_violation runPingViolation 39\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "a recorded function that grew must fail\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is 41 lines, over its recorded 39 in .ckeletin-command-baseline - move logic to internal/",
		"\nOutput: %s", output)
}

func TestBaseline_ShrunkFunctionNeedsALowerEntry(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "ping_violation runPingViolation 45\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "an entry above the current length must fail, or the function could grow back\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is 41 lines but recorded at 45 - lower its entry in .ckeletin-command-baseline to 41",
		"\nOutput: %s", output)
}

func TestBaseline_EntryForAShortFunctionMustBeRemoved(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(20))()
	defer writeBaseline(t, "ping_violation runPingViolation 41\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "a stale entry must fail, so the list shrinks\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is within the limit now - remove its entry from .ckeletin-command-baseline",
		"\nOutput: %s", output)
}

func TestBaseline_UnrecordedLongFunctionStillFails(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "some_other runSomething 50\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is 41 lines (target 30, hard limit 35) - move logic to internal/",
		"an unrecorded function fails exactly as without a baseline\nOutput: %s", output)
	assert.Contains(t, output, "some_other: runSomething() is within the limit now - remove its entry from .ckeletin-command-baseline",
		"an entry whose function no longer exists is stale too\nOutput: %s", output)
}

func TestBaseline_WriteRecordsTheCurrentLongFunctions(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	path := filepath.Join(projectRoot(t), baselineFile)
	_, err := os.Stat(path)
	require.True(t, os.IsNotExist(err), "%s already exists in the project root", baselineFile)
	defer os.Remove(path)

	output, exitCode := runCommandPatterns(t, "--write-baseline")
	require.Equal(t, 0, exitCode, "\nOutput: %s", output)
	written, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(written), "ping_violation runPingViolation 41\n")

	output, exitCode = runCommandPatterns(t)
	assert.Equal(t, 0, exitCode, "the written baseline must make the next run pass\nOutput: %s", output)
}

// A baseline saved with CRLF line endings must still compare as numbers. A
// trailing \r makes every numeric test error out, which used to fall through
// to "at its recorded length" and let a grown function pass.
func TestBaseline_CRLFEntriesStillCatchGrowth(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "# recorded\r\nping_violation runPingViolation 39\r\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "growth past a CRLF entry must fail\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is 41 lines, over its recorded 39 in .ckeletin-command-baseline", "\nOutput: %s", output)
}

func TestBaseline_DuplicateEntryFails(t *testing.T) {
	if testing.Short() {
		t.Skip("violation tests modify the source tree")
	}
	defer writeViolationFile(t, "cmd/ping_violation.go", pingCommandWithRun(41))()
	defer writeBaseline(t, "ping_violation runPingViolation 41\nping_violation runPingViolation 41\n")()

	output, exitCode := runCommandPatterns(t)
	assert.NotEqual(t, 0, exitCode, "\nOutput: %s", output)
	assert.Contains(t, output, "ping_violation: runPingViolation() is listed more than once in .ckeletin-command-baseline - keep one entry", "\nOutput: %s", output)
}
