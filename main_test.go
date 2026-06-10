// main_test.go

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
	"github.com/peiman/ckeletin-go/cmd"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Env vars guarding the subprocess branches of the run()-level E2E tests.
// These live in package main (not cmd) because they assert on the error
// envelope that run() itself renders — the part no cmd-package test can see.
const (
	jsonInitErrorRunEnv  = "CKELETIN_TEST_JSON_INIT_ERROR_RUN"
	auditOpenFailRunEnv  = "CKELETIN_TEST_AUDIT_OPEN_FAIL_RUN"
	auditOpenFailPathEnv = "CKELETIN_TEST_AUDIT_OPEN_FAIL_PATH"
)

// isolateSubprocessEnv builds the environment for a re-exec subprocess with
// HOME and XDG_CONFIG_HOME pointing into dir, so a developer's real
// ~/.config/<app>/config.yaml can never leak into the test. os/exec keeps the
// LAST value for duplicate keys, so appending overrides the inherited ones.
func isolateSubprocessEnv(dir string, extra ...string) []string {
	env := append(os.Environ(),
		"HOME="+dir,
		"XDG_CONFIG_HOME="+filepath.Join(dir, ".config"),
	)
	return append(env, extra...)
}

// outputFormatEnvVarUnderTest derives the environment variable viper maps to
// config.KeyAppOutputFormat (SetEnvPrefix + the "." -> "_" key replacer).
func outputFormatEnvVarUnderTest() string {
	return cmd.EnvPrefix() + "_" + strings.ToUpper(strings.ReplaceAll(config.KeyAppOutputFormat, ".", "_"))
}

// TestRun_JSONModeInitError_EmitsErrorEnvelope guards the init-error contract
// for JSON mode that is NOT flag-driven only: when configuration loading
// fails, run() must already know it is in JSON mode so stdout carries exactly
// one error envelope and stderr stays silent — no raw pre-init logs, no plain
// "Error:" text. Pins the EARLY output-mode resolution site (flag OR env):
// remove it and main.go falls back to the text error path because the
// post-initConfig viper read never happens on this error path.
func TestRun_JSONModeInitError_EmitsErrorEnvelope(t *testing.T) {
	if mode := os.Getenv(jsonInitErrorRunEnv); mode != "" {
		// SUBPROCESS BRANCH: ping against an invalid config file; JSON mode
		// comes from the env var or the --output flag depending on the mode.
		args := []string{"ping"}
		if mode == "flag" {
			args = append(args, "--output", "json")
		}
		cmd.RootCmd.SetArgs(args)
		os.Exit(run())
	}

	tests := []struct {
		name     string
		mode     string
		extraEnv []string
	}{
		{
			name:     "env-driven JSON mode",
			mode:     "env",
			extraEnv: []string{outputFormatEnvVarUnderTest() + "=json"},
		},
		{
			name: "flag-driven JSON mode",
			mode: "flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			home := t.TempDir()
			configDir := filepath.Join(home, ".config", cmd.RootCmd.Name())
			require.NoError(t, os.MkdirAll(configDir, 0o700), "failed to create config dir")
			require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.yaml"),
				[]byte("app:\n  log_level: NOT_A_LEVEL\n"), 0o600),
				"failed to write invalid config file")

			proc := exec.Command(os.Args[0], "-test.run=^TestRun_JSONModeInitError_EmitsErrorEnvelope$")
			proc.Dir = home
			proc.Env = isolateSubprocessEnv(home,
				append([]string{jsonInitErrorRunEnv + "=" + tt.mode}, tt.extraEnv...)...)
			var stdout, stderr bytes.Buffer
			proc.Stdout = &stdout
			proc.Stderr = &stderr

			// EXECUTION PHASE
			err := proc.Run()

			// ASSERTION PHASE
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr,
				"invalid config must exit non-zero\nstdout: %s\nstderr: %s", stdout.String(), stderr.String())
			assert.Equal(t, 1, exitErr.ExitCode(), "init failure must exit 1")

			var envelope output.JSONEnvelope
			require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope),
				"stdout must carry exactly one JSON error envelope, got: %s", stdout.String())
			assert.Equal(t, "error", envelope.Status)
			require.NotNil(t, envelope.Error, "error envelope must carry the error")
			assert.Contains(t, envelope.Error.Message, "configuration validation failed",
				"envelope must surface the config validation failure")

			assert.Empty(t, stderr.String(),
				"stderr must stay silent in JSON mode even when init fails (no raw logs, no Error: text)")
		})
	}
}

// TestRun_AuditOpenFailure_FailsClosed guards the audit-trail contract end to
// end: when file logging is explicitly enabled but the log file cannot be
// opened, the command must FAIL — never exit 0 with a success envelope and a
// silently missing audit file. JSON mode gets a proper error envelope on
// stdout with a silent stderr; text mode fails with the error on stderr.
func TestRun_AuditOpenFailure_FailsClosed(t *testing.T) {
	if mode := os.Getenv(auditOpenFailRunEnv); mode != "" {
		// SUBPROCESS BRANCH: enable file logging against an unopenable path.
		args := []string{
			"ping",
			"--log-file-enabled",
			"--log-file-path", os.Getenv(auditOpenFailPathEnv),
		}
		if mode == "json" {
			args = append(args, "--output", "json")
		}
		cmd.RootCmd.SetArgs(args)
		os.Exit(run())
	}

	tests := []struct {
		name string
		mode string
	}{
		{name: "JSON mode emits error envelope", mode: "json"},
		{name: "text mode fails with error on stderr", mode: "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			home := t.TempDir()
			// A file where a directory is needed makes MkdirAll fail without
			// requiring root-only filesystem states.
			blocker := filepath.Join(t.TempDir(), "blocker")
			require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600),
				"failed to create blocker file")
			badPath := filepath.Join(blocker, "sub", "audit.log")

			proc := exec.Command(os.Args[0], "-test.run=^TestRun_AuditOpenFailure_FailsClosed$")
			proc.Dir = home
			proc.Env = isolateSubprocessEnv(home,
				auditOpenFailRunEnv+"="+tt.mode,
				auditOpenFailPathEnv+"="+badPath,
			)
			var stdout, stderr bytes.Buffer
			proc.Stdout = &stdout
			proc.Stderr = &stderr

			// EXECUTION PHASE
			err := proc.Run()

			// ASSERTION PHASE
			var exitErr *exec.ExitError
			require.ErrorAs(t, err, &exitErr,
				"unopenable audit log must exit non-zero, not silently succeed\nstdout: %s\nstderr: %s",
				stdout.String(), stderr.String())
			assert.Equal(t, 1, exitErr.ExitCode(), "audit open failure must exit 1")
			assert.NoFileExists(t, badPath, "the audit file genuinely cannot exist")

			if tt.mode == "json" {
				var envelope output.JSONEnvelope
				require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope),
					"stdout must carry exactly one JSON error envelope, got: %s", stdout.String())
				assert.Equal(t, "error", envelope.Status)
				require.NotNil(t, envelope.Error, "error envelope must carry the error")
				assert.Contains(t, envelope.Error.Message, "log file",
					"envelope must identify the log file as the failure")
				assert.Empty(t, stderr.String(), "stderr must stay silent in JSON mode")
			} else {
				assert.Empty(t, stdout.String(), "text mode must not print success output")
				assert.Contains(t, stderr.String(), "Error:",
					"text mode must report the failure on stderr")
				assert.Contains(t, stderr.String(), "log file",
					"text mode error must identify the log file as the failure")
			}
		})
	}
}

func TestMainFunction(t *testing.T) {
	// Define test cases
	tests := []struct {
		name     string
		cmd      string
		cmdError error
		wantCode int
	}{
		{
			name:     "Success scenario",
			cmd:      "success",
			cmdError: nil,
			wantCode: 0,
		},
		{
			name:     "Failure scenario",
			cmd:      "fail",
			cmdError: fmt.Errorf("simulated failure"),
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save the original RootCmd
			originalRoot := cmd.RootCmd
			// Create a test root command
			testRoot := &cobra.Command{Use: "test"}
			// Replace global RootCmd with our test root
			cmd.RootCmd = testRoot
			// Restore after the test
			defer func() { cmd.RootCmd = originalRoot }()

			// Add a dummy command with the specified behavior
			testRoot.AddCommand(&cobra.Command{
				Use: tt.cmd,
				RunE: func(cmd *cobra.Command, args []string) error {
					return tt.cmdError
				},
			})

			// Set command arguments
			testRoot.SetArgs([]string{tt.cmd})

			// EXECUTION PHASE
			code := run()

			// ASSERTION PHASE
			if code != tt.wantCode {
				t.Errorf("expected exit code %d, got %d", tt.wantCode, code)
			}
		})
	}
}

func TestRun_JSONMode_Error(t *testing.T) {
	originalRoot := cmd.RootCmd
	defer func() { cmd.RootCmd = originalRoot }()

	output.SetOutputMode("json")
	output.SetCommandName("fail")
	defer func() {
		output.SetOutputMode("")
		output.SetCommandName("")
	}()

	testRoot := &cobra.Command{Use: "test", SilenceErrors: true}
	testRoot.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(c *cobra.Command, args []string) error {
			return fmt.Errorf("simulated failure")
		},
	})
	testRoot.SetArgs([]string{"fail"})
	cmd.RootCmd = testRoot

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	code := run()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)

	assert.Equal(t, 1, code)

	var envelope output.JSONEnvelope
	err := json.Unmarshal(buf.Bytes(), &envelope)
	assert.NoError(t, err, "stdout should contain valid JSON, got: %s", buf.String())
	assert.Equal(t, "error", envelope.Status)
	assert.NotNil(t, envelope.Error)
	assert.Contains(t, envelope.Error.Message, "simulated failure")
}
