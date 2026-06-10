// cmd/version_output_test.go
//
// CKSPEC-OUT-006: `--version` must surface four build-identity fields —
// semantic version, commit, build date, and working-tree state — and must
// degrade gracefully to "unknown" when ldflags are not injected.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setBuildIdentity overrides the ldflags-injected build identity vars for a
// test and restores the originals on cleanup.
func setBuildIdentity(t *testing.T, version, commit, date, dirty string) {
	t.Helper()
	origVersion, origCommit, origDate, origDirty := Version, Commit, Date, Dirty
	t.Cleanup(func() {
		Version, Commit, Date, Dirty = origVersion, origCommit, origDate, origDirty
	})
	Version, Commit, Date, Dirty = version, commit, date, dirty
}

// resetVersionFlagState clears cobra's --version flag after an Execute() call
// so later tests that run RootCmd do not inherit version=true.
func resetVersionFlagState(t *testing.T) {
	t.Helper()
	t.Cleanup(func() {
		if f := RootCmd.Flags().Lookup("version"); f != nil {
			f.Value.Set("false") //nolint:errcheck // resetting to known-good default
			f.Changed = false
		}
	})
}

func TestBuildIdentityDefaults_UninjectedBuild(t *testing.T) {
	// SETUP PHASE
	// Test binaries are compiled WITHOUT the Taskfile/GoReleaser ldflags, so
	// the package vars themselves carry the degradation defaults.

	// ASSERTION PHASE
	// CKSPEC-OUT-006 mandates the literal "unknown", not an empty string.
	assert.Equal(t, "unknown", Version, "uninjected Version must degrade to unknown")
	assert.Equal(t, "unknown", Commit, "uninjected Commit must degrade to unknown")
	assert.Equal(t, "unknown", Date, "uninjected Date must degrade to unknown")
	assert.Empty(t, Dirty, "Dirty has no default; treeState() derives unknown from Version")
}

func TestTreeState(t *testing.T) {
	tests := []struct {
		name    string
		version string
		dirty   string
		want    string
	}{
		{
			name:    "explicit dirty ldflag true (GoReleaser IsGitDirty)",
			version: "v1.2.3",
			dirty:   "true",
			want:    treeStateDirty,
		},
		{
			name:    "explicit dirty ldflag false wins over version suffix",
			version: "v1.2.3-dirty",
			dirty:   "false",
			want:    treeStateClean,
		},
		{
			name:    "explicit dirty ldflag is case- and space-insensitive",
			version: "v1.2.3",
			dirty:   " TRUE ",
			want:    treeStateDirty,
		},
		{
			name:    "git describe dirty suffix (Taskfile build)",
			version: "v1.2.3-dirty",
			dirty:   "",
			want:    treeStateDirty,
		},
		{
			name:    "clean injected version",
			version: "v1.2.3",
			dirty:   "",
			want:    treeStateClean,
		},
		{
			name:    "uninjected build degrades to unknown",
			version: "unknown",
			dirty:   "",
			want:    treeStateUnknown,
		},
		{
			name:    "empty version (pipeline injected empty ldflags) is unknown",
			version: "",
			dirty:   "",
			want:    treeStateUnknown,
		},
		{
			name:    "Taskfile dev fallback (git describe failed) is unknown",
			version: "dev",
			dirty:   "",
			want:    treeStateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			setBuildIdentity(t, tt.version, "abc1234", "2026-01-02_03:04:05", tt.dirty)

			// EXECUTION PHASE
			got := treeState()

			// ASSERTION PHASE
			assert.Equal(t, tt.want, got, "treeState() should resolve the working-tree state")
		})
	}
}

func TestVersionString_SurfacesAllFourFields(t *testing.T) {
	// SETUP PHASE
	setBuildIdentity(t, "v1.2.3", "abc1234", "2026-01-02_03:04:05", "")

	// EXECUTION PHASE
	got := versionString()

	// ASSERTION PHASE
	assert.Equal(t, "v1.2.3, commit abc1234, built at 2026-01-02_03:04:05, tree clean", got,
		"version string must carry semver, commit, date, and tree state")
}

// TestVersionString_EmptyLdflagsDegradeToUnknown guards the CKSPEC-OUT-006
// "never empty" backstop at runtime: a release pipeline that injects EMPTY
// strings via ldflags overrides the package-var "unknown" defaults, so each
// field needs its own normalization.
func TestVersionString_EmptyLdflagsDegradeToUnknown(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		date    string
		want    string
	}{
		{
			name:    "empty version",
			version: "",
			commit:  "abc1234",
			date:    "2026-01-02_03:04:05",
			want:    "unknown, commit abc1234, built at 2026-01-02_03:04:05, tree unknown",
		},
		{
			name:    "empty commit",
			version: "v1.2.3",
			commit:  "",
			date:    "2026-01-02_03:04:05",
			want:    "v1.2.3, commit unknown, built at 2026-01-02_03:04:05, tree clean",
		},
		{
			name:    "empty date",
			version: "v1.2.3",
			commit:  "abc1234",
			date:    "",
			want:    "v1.2.3, commit abc1234, built at unknown, tree clean",
		},
		{
			name:    "whitespace-only fields",
			version: "  ",
			commit:  "\t",
			date:    " ",
			want:    "unknown, commit unknown, built at unknown, tree unknown",
		},
		{
			name:    "all fields empty",
			version: "",
			commit:  "",
			date:    "",
			want:    "unknown, commit unknown, built at unknown, tree unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			setBuildIdentity(t, tt.version, tt.commit, tt.date, "")

			// EXECUTION PHASE
			got := versionString()

			// ASSERTION PHASE
			assert.Equal(t, tt.want, got,
				"empty ldflags injection must degrade to unknown, never empty (CKSPEC-OUT-006)")
		})
	}
}

// TestRenderVersion_JSONNeverEmitsEmptyFields covers the machine-readable
// --version path: empty ldflags injection must surface "unknown" in every
// envelope field, never "" (CKSPEC-OUT-006).
func TestRenderVersion_JSONNeverEmitsEmptyFields(t *testing.T) {
	// SETUP PHASE
	setBuildIdentity(t, "", "", "", "")

	outputFlag := RootCmd.PersistentFlags().Lookup("output")
	require.NotNil(t, outputFlag, "--output flag must exist on RootCmd")
	origValue, origChanged := outputFlag.Value.String(), outputFlag.Changed
	t.Cleanup(func() {
		outputFlag.Value.Set(origValue) //nolint:errcheck // restoring saved state
		outputFlag.Changed = origChanged
	})
	require.NoError(t, outputFlag.Value.Set("json"))

	// EXECUTION PHASE
	got := renderVersion(RootCmd)

	// ASSERTION PHASE
	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal([]byte(got), &envelope),
		"renderVersion should emit a JSON envelope, got: %s", got)

	data, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok, "envelope data should be an object, got: %T", envelope.Data)
	for _, field := range []string{"version", "commit", "date"} {
		assert.Equal(t, versionUnknown, data[field],
			"JSON %q must degrade to unknown, never empty (CKSPEC-OUT-006)", field)
	}
	assert.Equal(t, treeStateUnknown, data["tree"],
		"tree state must be unknown when no build identity was injected")
}

func TestVersionFlag_TextOutput(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	setBuildIdentity(t, "v1.2.3-dirty", "abc1234", "2026-01-02_03:04:05", "")
	resetVersionFlagState(t)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"--version"})

	// EXECUTION PHASE
	err := Execute()

	// ASSERTION PHASE
	require.NoError(t, err, "--version should succeed")
	out := stdout.String()
	assert.Contains(t, out, "v1.2.3-dirty", "text output must include the semantic version")
	assert.Contains(t, out, "commit abc1234", "text output must include the commit")
	assert.Contains(t, out, "built at 2026-01-02_03:04:05", "text output must include the build date")
	assert.Contains(t, out, "tree dirty", "text output must surface the dirty working-tree state")
}

func TestVersionFlag_JSONOutput(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	setBuildIdentity(t, "v1.2.3", "abc1234", "2026-01-02_03:04:05", "false")
	resetVersionFlagState(t)

	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer resetOutputJSONTestState(origCfgFile, origStatus, origUsed)

	var stdout bytes.Buffer
	RootCmd.SetOut(&stdout)
	RootCmd.SetErr(&bytes.Buffer{})
	RootCmd.SetArgs([]string{"--version", "--output", "json"})

	// EXECUTION PHASE
	err := Execute()

	// ASSERTION PHASE
	require.NoError(t, err, "--version --output json should succeed")

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope),
		"stdout should contain a single JSON envelope, got: %s", stdout.String())

	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, RootCmd.Name(), envelope.Command)
	assert.Nil(t, envelope.Error)

	data, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok, "envelope data should be an object, got: %T", envelope.Data)
	assert.Equal(t, "v1.2.3", data["version"])
	assert.Equal(t, "abc1234", data["commit"])
	assert.Equal(t, "2026-01-02_03:04:05", data["date"])
	assert.Equal(t, treeStateClean, data["tree"], "JSON output must surface the working-tree state")
}

// versionColdStartEnv guards the subprocess branch of the cold-start test.
const versionColdStartEnv = "CKELETIN_TEST_VERSION_COLD_START"

// TestVersionFlag_JSONOutput_ColdStart re-executes the test binary so RootCmd
// is pristine — no earlier Execute() call has registered the --version flag.
// Guards a real regression: before Execute() called InitDefaultVersionFlag(),
// cobra's stripFlags treated the unregistered --version as value-taking during
// command lookup, swallowed --output, and failed with `unknown command "json"`.
// In-process tests cannot catch this because the first --version Execute()
// registers the flag for every test that follows.
func TestVersionFlag_JSONOutput_ColdStart(t *testing.T) {
	if os.Getenv(versionColdStartEnv) == "1" {
		// SUBPROCESS BRANCH: run --version --output json on the pristine RootCmd
		// and exit before the testing framework prints PASS to stdout.
		RootCmd.SetArgs([]string{"--version", "--output", "json"})
		if err := Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "execute failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// SETUP PHASE
	// HOME/XDG isolation: without it the subprocess inherits the developer's
	// real HOME, so a local ~/.config/<app>/config.yaml could alter behavior.
	home := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=^TestVersionFlag_JSONOutput_ColdStart$")
	cmd.Dir = home
	cmd.Env = isolateSubprocessEnv(home, versionColdStartEnv+"=1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// EXECUTION PHASE
	err := cmd.Run()

	// ASSERTION PHASE
	require.NoError(t, err, "cold-start --version --output json must succeed\nstdout: %s\nstderr: %s",
		stdout.String(), stderr.String())

	var envelope output.JSONEnvelope
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope),
		"cold-start stdout should be a single JSON envelope, got: %s", stdout.String())
	assert.Equal(t, "success", envelope.Status)

	data, ok := envelope.Data.(map[string]interface{})
	require.True(t, ok, "envelope data should be an object, got: %T", envelope.Data)
	for _, field := range []string{"version", "commit", "date", "tree"} {
		assert.NotEmpty(t, data[field], "cold-start JSON output must carry the %q field", field)
	}
}
