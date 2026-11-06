// test/integration/coverage_gate_test.go

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckTaskEnforcesCoverageGate guards that the mandatory `task check` path
// still wires in the 85% project-coverage gate (test:coverage:project). If a
// future .ckeletin/ sync, merge, or refactor drops that step, `task check` would
// silently stop enforcing the coverage floor (the binary/bash check only
// measures coverage, it does not gate on it). This test fails loudly instead.
func TestCheckTaskEnforcesCoverageGate(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(root, ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)

	block := topLevelTaskBlock(string(data), "check")
	require.NotEmpty(t, block, "could not locate the 'check:' task in .ckeletin/Taskfile.yml")
	assert.Contains(t, block, "test:coverage:project",
		"the 'check' task must invoke test:coverage:project so the 85% coverage gate is enforced in `task check`")
}

// TestFrameworkChangelogPrecedesTheSummary guards issue #182. The binary
// prints "All N Checks Passed" and then exits. A later failing task still
// leaves that green box on screen. The changelog gate has to run first.
func TestFrameworkChangelogPrecedesTheSummary(t *testing.T) {
	root, err := filepath.Abs("../..")
	require.NoError(t, err)

	data, err := os.ReadFile(filepath.Join(root, ".ckeletin", "Taskfile.yml"))
	require.NoError(t, err)
	content := string(data)

	cases := []struct {
		task          string
		summaryMarker string
	}{
		{task: "check", summaryMarker: "check --help"},
		{task: "check:common", summaryMarker: "task: check:summary"},
		{task: "check:fast", summaryMarker: "task: check:summary"},
	}
	for _, tc := range cases {
		t.Run(tc.task, func(t *testing.T) {
			block := topLevelTaskBlock(content, tc.task)
			require.NotEmpty(t, block, "could not locate the %s task", tc.task)

			changelog := strings.Index(block, "task: check:framework-changelog")
			summary := strings.Index(block, tc.summaryMarker)
			require.NotEqual(t, -1, changelog, "%s must run the framework changelog gate", tc.task)
			require.NotEqual(t, -1, summary, "%s must print its summary (%s)", tc.task, tc.summaryMarker)
			assert.Less(t, changelog, summary,
				"%s must run the changelog gate before the summary", tc.task)
		})
	}
}

// topLevelTaskBlock returns the body lines of the named 2-space-indented task,
// up to (but excluding) the next top-level task.
func topLevelTaskBlock(content, name string) string {
	isTopTask := func(line string) bool {
		return len(line) >= 3 && line[0] == ' ' && line[1] == ' ' &&
			line[2] != ' ' && line[2] != '-' && line[2] != '#'
	}
	var body []string
	in := false
	for _, line := range strings.Split(content, "\n") {
		if isTopTask(line) {
			if in {
				break // reached the next task
			}
			if strings.TrimSpace(line) == name+":" {
				in = true
			}
			continue
		}
		if in {
			body = append(body, line)
		}
	}
	return strings.Join(body, "\n")
}
