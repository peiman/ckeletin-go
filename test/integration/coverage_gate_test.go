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
