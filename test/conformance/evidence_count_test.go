package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hardcodedRequirementCount is the phrase the ENF-008 drift guard rejects.
// A handwritten total rots when a requirement is added. The completeness
// line computed from the mapping is the count.
var hardcodedRequirementCount = regexp.MustCompile(`all [0-9]+ requirement`)

func TestMappingEvidenceDoesNotHardcodeRequirementCount(t *testing.T) {
	root := conformanceRoot(t)
	data, err := os.ReadFile(filepath.Join(root, "conformance-mapping.yaml"))
	require.NoError(t, err)

	if loc := hardcodedRequirementCount.Find(data); loc != nil {
		t.Fatalf("conformance evidence hardcodes a requirement count %q; the mapped total is computed and must not be restated", loc)
	}
}

func TestEvidenceCountDriftRejectsMatchingCount(t *testing.T) {
	root := conformanceRoot(t)
	dir := t.TempDir()
	mapping := filepath.Join(dir, "mapping.yaml")
	content := []byte("requirements:\n  ONLY:\n    evidence: task conform validates all 1 requirement IDs\n")
	require.NoError(t, os.WriteFile(mapping, content, 0o644))

	out := evidenceCountDrift(t, root, mapping, "1")
	assert.Contains(t, out, "all 1 requirement")
	assert.Contains(t, out, "ENF-008")
}

func TestEvidenceCountDriftAllowsProseWithoutCount(t *testing.T) {
	root := conformanceRoot(t)
	dir := t.TempDir()
	mapping := filepath.Join(dir, "mapping.yaml")
	content := []byte("requirements:\n  ONLY:\n    evidence: task conform validates every requirement ID\n")
	require.NoError(t, os.WriteFile(mapping, content, 0o644))

	out := evidenceCountDrift(t, root, mapping, "1")
	assert.Empty(t, out)
}

func conformanceRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	require.NoError(t, err)
	return root
}

func evidenceCountDrift(t *testing.T, root, mapping, total string) string {
	t.Helper()
	script := filepath.Join(root, ".ckeletin", "scripts", "lib", "evidence-drift.sh")
	cmd := exec.Command("bash", "-c", "source \"$1\" && evidence_count_drift \"$2\" \"$3\"", "evidence-drift", script, mapping, total)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "evidence_count_drift failed: %s", out)
	return strings.TrimSpace(string(out))
}
