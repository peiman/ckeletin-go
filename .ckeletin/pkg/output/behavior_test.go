package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRendered_IsErrRenderedAndCarriesCode(t *testing.T) {
	err := Rendered(3)
	assert.ErrorIs(t, err, ErrRendered)
	assert.Equal(t, 3, ExitCodeOf(err))

	zero := Rendered(0)
	assert.ErrorIs(t, zero, ErrRendered)
	assert.Equal(t, 1, ExitCodeOf(zero), "exit 0 cannot accompany an error")
	assert.Equal(t, 1, ExitCodeOf(ErrRendered))
	assert.Equal(t, 0, ExitCodeOf(nil))
	assert.Equal(t, 1, ExitCodeOf(errors.New("plain")))
}

type codedError struct {
	code int
	msg  string
}

func (e codedError) Error() string { return e.msg }
func (e codedError) ExitCode() int { return e.code }

func TestStatusAgrees(t *testing.T) {
	assert.True(t, StatusAgrees("error", 1))
	assert.True(t, StatusAgrees("error", 2))
	assert.False(t, StatusAgrees("error", 0))
	assert.True(t, StatusAgrees("success", 0))
	assert.False(t, StatusAgrees("success", 1))
}

func TestCommandID_PathWithoutProgramName(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	check := &cobra.Command{Use: "check"}
	model := &cobra.Command{Use: "model"}
	version := &cobra.Command{Use: "version"}
	check.AddCommand(model)
	root.AddCommand(check, version)

	assert.Equal(t, "check model", CommandID(model))
	assert.Equal(t, "version", CommandID(version))
	assert.Empty(t, CommandID(root))

	SetCommandName("alias")
	t.Cleanup(func() { SetCommandName("") })
	SetCommandFrom(model)
	assert.Equal(t, "alias", CommandName())

	SetCommandName("")
	SetCommandFrom(model)
	assert.Equal(t, "check model", CommandName())
}

func TestExecute_NestedCommandEnvelope(t *testing.T) {
	SetOutputMode("json")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	var stdout bytes.Buffer
	root := &cobra.Command{Use: "app", SilenceErrors: true, SilenceUsage: true}
	check := &cobra.Command{Use: "check"}
	model := &cobra.Command{
		Use: "model",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
				Status:  "success",
				Command: CommandName(),
				Data:    map[string]string{"ok": "yes"},
			})
		},
	}
	check.AddCommand(model)
	root.AddCommand(check)
	RegisterOutputFlag(root, WithEnvPrefix("APP"))
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"check", "model"})

	code := Execute(root)

	require.Equal(t, 0, code)
	var envelope JSONEnvelope
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope))
	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "check model", envelope.Command)
	assert.True(t, StatusAgrees(envelope.Status, code))
}

func TestExecute_ErrorPaths(t *testing.T) {
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	t.Run("json RunE", func(t *testing.T) {
		SetOutputMode("json")
		stdout, code := runFailing(t, "json", func(cmd *cobra.Command) error {
			return errors.New("run failed")
		})
		assert.Equal(t, 1, code)
		assertOneErrorEnvelope(t, stdout, "run failed", "boom")
	})

	t.Run("json PreRun", func(t *testing.T) {
		root := newApp(t)
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				return errors.New("run should not be reached")
			},
		})
		root.PersistentPreRunE = func(*cobra.Command, []string) error {
			return errors.New("pre-run failed")
		}
		// Register wraps whatever pre-run is installed now.
		RegisterOutputFlag(root, WithEnvPrefix("APP"))
		root.SetArgs([]string{"boom", "--output", "json"})
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		code := Execute(root)
		assert.Equal(t, 1, code)
		assertOneErrorEnvelope(t, stdout.String(), "pre-run failed", "boom")
	})

	t.Run("text mode", func(t *testing.T) {
		SetOutputMode("text")
		root := newApp(t)
		root.SetArgs([]string{"boom"})
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				return errors.New("text failed")
			},
		})
		code := Execute(root)
		assert.Equal(t, 1, code)
		assert.Empty(t, stdout.String())
		assert.Equal(t, "Error: text failed\n", stderr.String())
	})

	t.Run("rendered code 2 is one envelope", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(cmd *cobra.Command, _ []string) error {
				require.NoError(t, RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
					Status:  "error",
					Command: "boom",
					Error:   &JSONError{Message: "usage"},
				}))
				return Rendered(2)
			},
		})
		code := Execute(root)
		assert.Equal(t, 2, code)
		assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`))
		assert.True(t, StatusAgrees("error", code))
	})

	t.Run("unrendered ExitCoder", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				return codedError{code: 3, msg: "needs a person"}
			},
		})
		code := Execute(root)
		assert.Equal(t, 3, code)
		assertOneErrorEnvelope(t, stdout.String(), "needs a person", "boom")
		assert.True(t, StatusAgrees("error", code))
	})

	t.Run("ExitCoder zero becomes 1", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				return codedError{code: 0, msg: "no"}
			},
		})
		code := Execute(root)
		assert.Equal(t, 1, code)
		assert.NotEqual(t, 0, code)
		assertOneErrorEnvelope(t, stdout.String(), "no", "boom")
	})

	t.Run("unresolved command is null", func(t *testing.T) {
		SetOutputMode("json")
		SetCommandName("")
		root := newApp(t)
		root.AddCommand(&cobra.Command{Use: "real", RunE: func(*cobra.Command, []string) error { return nil }})
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"nope"})
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		assert.Contains(t, stdout.String(), `"command": null`)
		assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`))
	})

	t.Run("deepest resolved path", func(t *testing.T) {
		SetOutputMode("json")
		SetCommandName("")
		root := newApp(t)
		check := &cobra.Command{
			Use:  "check",
			Args: cobra.NoArgs,
			RunE: func(*cobra.Command, []string) error { return nil },
		}
		root.AddCommand(check)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"check", "nope"})
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		var envelope JSONEnvelope
		require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope), stdout.String())
		assert.Equal(t, "error", envelope.Status)
		assert.Equal(t, "check", envelope.Command)
	})
}

func TestExecute_EmptySuccessEnvelope(t *testing.T) {
	SetOutputMode("json")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	root := newApp(t)
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"quiet"})
	root.AddCommand(&cobra.Command{
		Use:  "quiet",
		RunE: func(*cobra.Command, []string) error { return nil },
	})

	code := Execute(root)

	assert.Equal(t, 0, code)
	assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
	var envelope JSONEnvelope
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope), stdout.String())
	assert.Equal(t, "success", envelope.Status)
	assert.Equal(t, "quiet", envelope.Command)
	assert.Nil(t, envelope.Data)
	assert.Nil(t, envelope.Error)
	assert.True(t, StatusAgrees(envelope.Status, code))

	SetOutputMode("text")
	stdout.Reset()
	root.SetArgs([]string{"quiet"})
	code = Execute(root)
	assert.Equal(t, 0, code)
	assert.Empty(t, stdout.String(), "text mode does not invent an envelope")
}

func TestExecute_ExemptCommandPrintsNativeOutput(t *testing.T) {
	SetOutputMode("json")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	root := newApp(t)
	native := &cobra.Command{
		Use: "script",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), "native-script")
			return err
		},
	}
	ExemptFromEnvelope(native, "shell completion script")
	root.AddCommand(native)
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"script"})

	code := Execute(root)

	assert.Equal(t, 0, code)
	assert.Equal(t, "native-script\n", stdout.String())
	assert.NotContains(t, stdout.String(), `"status"`)

	found := EnvelopeExemptions(root)
	require.Len(t, found, 1)
	assert.Equal(t, "script", found[0].Command)
	assert.Equal(t, "shell completion script", found[0].Reason)

	ExemptFromEnvelope(&cobra.Command{Use: "blank"}, "  ")
	assert.Empty(t, EnvelopeExemptions(&cobra.Command{Use: "app"}))
}

func TestExecute_InvalidOutputExitCode(t *testing.T) {
	SetOutputMode("")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	for _, args := range [][]string{
		{"boom", "--output", "yaml"},
		{"--output", "yaml", "boom"},
	} {
		root := newApp(t)
		root.AddCommand(&cobra.Command{Use: "boom", RunE: func(*cobra.Command, []string) error { return nil }})
		RegisterOutputFlag(root, WithEnvPrefix("APP"))
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs(args)
		code := Execute(root)
		assert.Equal(t, ExitInvalidOutput, code, "args %v", args)
		assert.Empty(t, stdout.String(), "args %v", args)
		assert.Equal(t, 1, strings.Count(stderr.String(), "Error:"), "args %v\n%s", args, stderr.String())
		assert.Contains(t, stderr.String(), "yaml")
	}
}

func TestUseOutputMode_Precedence(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	SetOutputMode("")
	t.Cleanup(func() { SetOutputMode("") })

	t.Setenv("APP_APP_OUTPUT_FORMAT", "json")
	root := &cobra.Command{Use: "app"}
	root.AddCommand(&cobra.Command{Use: "child"})
	RegisterOutputFlag(root, WithEnvPrefix("APP"), WithAliasFlag("format"))
	SetCommandFrom(root)

	require.NoError(t, UseOutputMode(PhaseEarly))
	assert.True(t, IsJSONMode(), "env turns JSON on")

	require.NoError(t, root.ParseFlags([]string{"--output", "text"}))
	SetCommandFrom(root)
	require.NoError(t, UseOutputMode(PhaseEarly))
	assert.False(t, IsJSONMode(), "--output text overrides env json")

	// Config file loses to the flag and to the env var.
	viper.SetConfigType("yaml")
	require.NoError(t, viper.ReadConfig(strings.NewReader("app:\n  output_format: json\n")))
	require.NoError(t, UseOutputMode(PhaseResolved))
	assert.False(t, IsJSONMode(), "flag text overrides config json")

	// The cases above publish the resolved mode into viper. Reset so the
	// alias is the only json source: a dead alias branch must fail here.
	viper.Reset()
	root = &cobra.Command{Use: "app"}
	child := &cobra.Command{Use: "child"}
	child.Flags().String("format", "text", "alias")
	root.AddCommand(child)
	RegisterOutputFlag(root, WithEnvPrefix("APP"), WithAliasFlag("format"))
	// Clear the env so the alias is the decider. Empty is not a value.
	t.Setenv("APP_APP_OUTPUT_FORMAT", "")
	require.NoError(t, child.ParseFlags([]string{"--format", "json"}))
	SetCommandFrom(child)
	require.NoError(t, UseOutputMode(PhaseResolved))
	assert.True(t, IsJSONMode(), "alias flag turns JSON on")
	assert.Equal(t, "json", viper.GetString(config.KeyAppOutputFormat),
		"resolved mode is what the logger reads")

	require.NoError(t, root.ParseFlags([]string{"--output", "text"}))
	SetCommandFrom(child)
	require.NoError(t, UseOutputMode(PhaseEarly))
	assert.False(t, IsJSONMode(), "--output wins over the alias")
}

func TestUseOutputMode_ConfigAndInvalid(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	viper.SetConfigType("yaml")
	require.NoError(t, viper.ReadConfig(strings.NewReader("app:\n  output_format: json\n")))
	root := &cobra.Command{Use: "app"}
	RegisterOutputFlag(root, WithEnvPrefix("APP"))
	SetCommandFrom(root)
	require.NoError(t, UseOutputMode(PhaseResolved))
	assert.True(t, IsJSONMode())

	viper.Reset()
	viper.Set(config.KeyAppOutputFormat, "yaml")
	err := UseOutputMode(PhaseResolved)
	require.Error(t, err)
	assert.Equal(t, ExitInvalidOutput, ExitCodeOf(err))
	assert.Contains(t, err.Error(), "yaml")
}

func TestWriters_StreamsShadowAndJSON(t *testing.T) {
	SetOutputMode("text")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
		log.Logger = zerolog.New(os.Stderr)
	})

	var audit bytes.Buffer
	log.Logger = zerolog.New(&audit)
	cmd := &cobra.Command{Use: "app"}
	child := &cobra.Command{Use: "ping"}
	cmd.AddCommand(child)
	SetCommandFrom(child)

	var out, errBuf bytes.Buffer
	child.SetOut(&out)
	child.SetErr(&errBuf)

	loggedDuringWrite := false
	order := &orderWriter{after: func() {
		if audit.Len() > 0 {
			loggedDuringWrite = true
		}
	}}
	child.SetOut(order)

	require.NoError(t, Resultf(child, "hello\n"))
	assert.Equal(t, "hello\n", order.buf.String())
	assert.Empty(t, errBuf.String())
	assert.False(t, loggedDuringWrite, "shadow log is written after the user-facing write")
	assert.Contains(t, audit.String(), `"stream":"stdout"`)
	assert.Contains(t, audit.String(), `"command":"ping"`)
	assert.Contains(t, audit.String(), "hello")

	audit.Reset()
	require.NoError(t, Statusf(child, "wait\n"))
	assert.Equal(t, "wait\n", errBuf.String())
	assert.Contains(t, audit.String(), `"stream":"stderr"`)

	audit.Reset()
	fail := &failWriter{}
	child.SetOut(fail)
	err := Resultf(child, "nope")
	require.Error(t, err)
	assert.NotContains(t, audit.String(), "nope")
	assert.NotContains(t, audit.String(), "user output")

	SetOutputMode("json")
	var jsonOut bytes.Buffer
	child.SetOut(&jsonOut)
	require.NoError(t, Resultf(child, "hidden\n"))
	require.NoError(t, Success("also hidden"))
	assert.Empty(t, jsonOut.String(), "JSON mode stdout stays free of result text")

	var jsonErr bytes.Buffer
	child.SetErr(&jsonErr)
	require.NoError(t, Statusf(child, "hidden status"))
	assert.Empty(t, jsonErr.String())
}

func TestProgressf_StripsANSIWhenNotATTY(t *testing.T) {
	SetOutputMode("text")
	t.Cleanup(func() { SetOutputMode("") })

	cmd := &cobra.Command{Use: "app"}
	var errBuf bytes.Buffer
	cmd.SetErr(&errBuf)
	require.NoError(t, Progressf(cmd, "\x1b[31mworking\x1b[0m"))
	assert.Equal(t, "working\n", errBuf.String())
	assert.NotContains(t, errBuf.String(), "\x1b")
	assert.Equal(t, "working", stripANSI("\x1b[31mworking\x1b[0m"))
}

func TestExecute_LogsFailureBeforeCleanup(t *testing.T) {
	SetOutputMode("json")
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
		log.Logger = zerolog.New(os.Stderr)
		SetAfterRun(nil)
		SetFailureLog(nil)
	})

	var audit bytes.Buffer
	log.Logger = zerolog.New(&audit).Level(zerolog.DebugLevel)
	SetFailureLog(func(err error, command string, code int) {
		log.Debug().Err(err).Str("command", command).Int("exit_code", code).Msg("command failed")
	})
	closed := false
	SetAfterRun(func() {
		assert.Contains(t, audit.String(), "command failed")
		assert.Contains(t, audit.String(), "boom")
		closed = true
	})

	root := newApp(t)
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"boom"})
	root.AddCommand(&cobra.Command{
		Use: "boom",
		RunE: func(*cobra.Command, []string) error {
			return errors.New("boom")
		},
	})

	code := Execute(root)
	assert.Equal(t, 1, code)
	assert.True(t, closed, "cleanup runs after the error is logged")
	assert.Contains(t, stdout.String(), "boom")
}

func TestEnvelope_UnresolvedCommandIsNull(t *testing.T) {
	var buf bytes.Buffer
	err := RenderJSON(&buf, JSONEnvelope{Status: "error", Error: &JSONError{Message: "x"}}.WithUnresolvedCommand())
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"command": null`)

	buf.Reset()
	err = RenderJSON(&buf, JSONEnvelope{Status: "success", Command: "ping"})
	require.NoError(t, err)
	assert.Contains(t, buf.String(), `"command": "ping"`)
	assert.Contains(t, buf.String(), "\n")
}

type orderWriter struct {
	buf   bytes.Buffer
	after func()
}

func (w *orderWriter) Write(p []byte) (int, error) {
	n, err := w.buf.Write(p)
	if w.after != nil {
		w.after()
	}
	return n, err
}

type failWriter struct{}

func (failWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }

func newApp(t *testing.T) *cobra.Command {
	t.Helper()
	return &cobra.Command{Use: "app", SilenceErrors: true, SilenceUsage: true}
}

func runFailing(t *testing.T, mode string, run func(*cobra.Command) error) (string, int) {
	t.Helper()
	SetOutputMode(mode)
	root := newApp(t)
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"boom"})
	root.AddCommand(&cobra.Command{Use: "boom", RunE: func(cmd *cobra.Command, _ []string) error {
		return run(cmd)
	}})
	code := Execute(root)
	return stdout.String(), code
}

func assertOneErrorEnvelope(t *testing.T, stdout, message, command string) {
	t.Helper()
	assert.Equal(t, 1, strings.Count(stdout, `"status"`), stdout)
	var envelope JSONEnvelope
	require.NoError(t, json.Unmarshal([]byte(stdout), &envelope), stdout)
	assert.Equal(t, "error", envelope.Status)
	assert.Equal(t, command, envelope.Command)
	require.NotNil(t, envelope.Error)
	assert.Contains(t, envelope.Error.Message, message)
	assert.Nil(t, envelope.Data)
}
