package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasFlag_LoggerFollowsResolvedMode(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	t.Cleanup(func() {
		logger.RestoreLoggerState(savedLogger, savedLevel)
		viper.Reset()
		SetOutputMode("")
		SetCommandName("")
	})
	t.Setenv("APP_APP_OUTPUT_FORMAT", "")

	t.Run("alias json silences console without config", func(t *testing.T) {
		viper.Reset()
		silenceLoggerNoise()
		_, child := newAliasCommand()
		require.NoError(t, child.ParseFlags([]string{"--format", "json"}))
		SetCommandFrom(child)
		require.NoError(t, UseOutputMode(PhaseResolved))
		require.True(t, IsJSONMode())
		require.Equal(t, "json", viper.GetString(config.KeyAppOutputFormat))

		buf := &bytes.Buffer{}
		require.NoError(t, logger.Init(buf))
		log.Warn().Msg("hidden-warn")
		logger.Cleanup()
		assert.Empty(t, buf.String(), "alias json is the logger's mode, not only the envelope's")
	})

	t.Run("alias text shows warnings over config json", func(t *testing.T) {
		viper.Reset()
		silenceLoggerNoise()
		viper.Set(config.KeyAppOutputFormat, "json")
		_, child := newAliasCommand()
		require.NoError(t, child.ParseFlags([]string{"--format", "text"}))
		SetCommandFrom(child)
		require.NoError(t, UseOutputMode(PhaseResolved))
		require.False(t, IsJSONMode())
		require.Equal(t, "text", viper.GetString(config.KeyAppOutputFormat))

		buf := &bytes.Buffer{}
		require.NoError(t, logger.Init(buf))
		log.Warn().Msg("shown-warn")
		logger.Cleanup()
		assert.Contains(t, buf.String(), "shown-warn")
	})
}

func TestExecute_RenderedStatusAgreesWithExit(t *testing.T) {
	SetCommandName("")
	t.Cleanup(func() {
		SetOutputMode("")
		SetCommandName("")
	})

	t.Run("rendered error and nil exit 1", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(cmd *cobra.Command, _ []string) error {
				return RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
					Status:  "error",
					Command: "boom",
					Error:   &JSONError{Message: "nope"},
				})
			},
		})
		code := Execute(root)
		assert.Equal(t, 1, code)
		assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
		assert.Contains(t, stdout.String(), "nope")
	})

	t.Run("rendered success and Rendered code exit 0", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&bytes.Buffer{})
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(cmd *cobra.Command, _ []string) error {
				if err := RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
					Status:  "success",
					Command: "boom",
					Data:    map[string]string{"ok": "yes"},
				}); err != nil {
					return err
				}
				return Rendered(2)
			},
		})
		code := Execute(root)
		assert.Equal(t, 0, code)
		assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
		assert.Contains(t, stdout.String(), `"success"`)
	})
}

func TestExecute_UsageErrorDoesNotLog(t *testing.T) {
	resetOutputState()
	t.Cleanup(resetOutputState)
	savedLogger, savedLevel := logger.SaveLoggerState()
	t.Cleanup(func() { logger.RestoreLoggerState(savedLogger, savedLevel) })

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	origStderr := os.Stderr
	os.Stderr = writer
	log.Logger = zerolog.New(os.Stderr).Level(zerolog.DebugLevel)
	t.Cleanup(func() { os.Stderr = origStderr })

	root := newApp(t)
	root.AddCommand(&cobra.Command{Use: "real", RunE: func(*cobra.Command, []string) error { return nil }})
	RegisterOutputFlag(root, WithEnvPrefix("APP"))
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"nosuch"})

	code := Execute(root)
	require.NoError(t, writer.Close())
	os.Stderr = origStderr
	captured, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.NotEqual(t, 0, code)
	assert.Empty(t, stdout.String())
	assert.Equal(t, "Error: unknown command \"nosuch\" for \"app\"\n", stderr.String())
	assert.Empty(t, string(captured), "no logger is installed, so stderr has no zerolog line")
}

func TestExecute_ParseErrorOutputFlagIsEnvelope(t *testing.T) {
	resetOutputState()
	viper.Reset()
	t.Cleanup(func() {
		resetOutputState()
		viper.Reset()
	})

	root := newApp(t)
	root.AddCommand(&cobra.Command{Use: "real", RunE: func(*cobra.Command, []string) error { return nil }})
	RegisterOutputFlag(root, WithEnvPrefix("APP"))
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--output", "json", "nosuch"})

	code := Execute(root)

	assert.NotEqual(t, 0, code)
	assert.Empty(t, stderr.String(), stderr.String())
	assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
	assert.Contains(t, stdout.String(), `"command": null`)
	assert.Contains(t, stdout.String(), "nosuch")
	var envelope JSONEnvelope
	require.NoError(t, json.Unmarshal(stdout.Bytes(), &envelope), stdout.String())
	assert.Equal(t, "error", envelope.Status)
	assert.Empty(t, envelope.Command)
}

func TestExecute_PanicBecomesJSONError(t *testing.T) {
	resetOutputState()
	t.Cleanup(resetOutputState)

	t.Run("json", func(t *testing.T) {
		SetOutputMode("json")
		root := newApp(t)
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				panic("kaboom")
			},
		})
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		assert.Empty(t, stderr.String(), stderr.String())
		assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
		assert.Contains(t, stdout.String(), "panic: kaboom")
	})

	t.Run("text", func(t *testing.T) {
		SetOutputMode("text")
		root := newApp(t)
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"boom"})
		root.AddCommand(&cobra.Command{
			Use: "boom",
			RunE: func(*cobra.Command, []string) error {
				panic("kaboom")
			},
		})
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		assert.Empty(t, stdout.String())
		assert.Equal(t, "Error: panic: kaboom\n", stderr.String())
	})
}

func TestSetup_AfterLoggerRunsOnConfigErrorAndInitError(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	t.Cleanup(func() {
		logger.RestoreLoggerState(savedLogger, savedLevel)
		viper.Reset()
		resetOutputState()
	})

	t.Run("config error still resolves json", func(t *testing.T) {
		viper.Reset()
		resetOutputState()
		silenceLoggerNoise()
		viper.Set(config.KeyAppOutputFormat, "json")
		var got error
		root := &cobra.Command{
			Use: "app",
			PersistentPreRunE: func(*cobra.Command, []string) error {
				return errors.New("load failed")
			},
		}
		root.AddCommand(&cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }})
		Setup(root, WithEnvPrefix("APP"), WithAfterLogger(func(err error) { got = err }))
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"sub"})
		code := Execute(root)
		require.Error(t, got)
		assert.Contains(t, got.Error(), "load failed")
		assert.NotEqual(t, 0, code)
		assert.True(t, IsJSONMode())
		assert.Contains(t, stdout.String(), `"status": "error"`)
		assert.Empty(t, stderr.String())
	})

	t.Run("unopenable audit file", func(t *testing.T) {
		viper.Reset()
		resetOutputState()
		silenceLoggerNoise()
		blocker := filepath.Join(t.TempDir(), "blocker")
		require.NoError(t, os.WriteFile(blocker, []byte("x"), 0o600))
		viper.Set(config.KeyAppLogFileEnabled, true)
		viper.Set(config.KeyAppLogFilePath, filepath.Join(blocker, "sub", "audit.log"))
		var got error
		root := &cobra.Command{Use: "app"}
		root.AddCommand(&cobra.Command{Use: "sub", RunE: func(*cobra.Command, []string) error { return nil }})
		Setup(root, WithEnvPrefix("APP"), WithAfterLogger(func(err error) { got = err }))
		var stdout, stderr bytes.Buffer
		root.SetOut(&stdout)
		root.SetErr(&stderr)
		root.SetArgs([]string{"--output", "json", "sub"})
		code := Execute(root)
		require.Error(t, got)
		assert.Contains(t, got.Error(), "failed to initialize logger")
		assert.NotEqual(t, 0, code)
		assert.Contains(t, stdout.String(), "log file", stdout.String())
		assert.Empty(t, stderr.String())
	})
}

func TestCoded_NilErrUsesCodeAsMessage(t *testing.T) {
	err := Coded("E_BARE", nil)
	assert.Equal(t, "E_BARE", err.Error())
	require.NotNil(t, errorClass(err))
	assert.Equal(t, "E_BARE", *errorClass(err))
}

func TestCommandArgs_IgnoresGoTestArguments(t *testing.T) {
	root := &cobra.Command{Use: "app"}
	assert.Empty(t, commandArgs(root))
	assert.Nil(t, commandArgs(nil))
}

func TestCoded_SetsEnvelopeCode(t *testing.T) {
	resetOutputState()
	t.Cleanup(resetOutputState)
	SetOutputMode("json")

	root := newApp(t)
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"boom"})
	root.AddCommand(&cobra.Command{
		Use: "boom",
		RunE: func(*cobra.Command, []string) error {
			return Coded("E_MISSING", errors.New("need a name"))
		},
	})
	code := Execute(root)
	assert.Equal(t, 1, code)
	assert.Contains(t, stdout.String(), `"code": "E_MISSING"`)
	assert.Contains(t, stdout.String(), "need a name")

	plain := Coded("", errors.New("x"))
	assert.Nil(t, errorClass(plain))
}

func TestSetup_MinimalRootAuditsFailureAndAliasSilence(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	t.Cleanup(func() {
		logger.RestoreLoggerState(savedLogger, savedLevel)
		viper.Reset()
		resetOutputState()
	})

	path := filepath.Join(t.TempDir(), "audit.log")
	viper.Reset()
	viper.Set(config.KeyAppLogFileEnabled, true)
	viper.Set(config.KeyAppLogFilePath, path)
	viper.Set(config.KeyAppLogFileLevel, "debug")
	viper.Set(config.KeyAppLogConsoleLevel, "debug")
	viper.Set(config.KeyAppLogLevel, "debug")
	viper.Set(config.KeyAppLogColorEnabled, "false")
	viper.Set(config.KeyAppLogSamplingEnabled, false)
	t.Setenv("APP_APP_OUTPUT_FORMAT", "")

	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	origStderr := os.Stderr
	os.Stderr = writer
	t.Cleanup(func() { os.Stderr = origStderr })

	root := &cobra.Command{Use: "app", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().String("format", "text", "alias")
	root.AddCommand(&cobra.Command{
		Use: "fail",
		RunE: func(*cobra.Command, []string) error {
			log.Warn().Msg("secret-warn")
			return errors.New("boom")
		},
	})
	Setup(root, WithEnvPrefix("APP"), WithAliasFlag("format"))
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs([]string{"--format", "json", "fail"})

	code := Execute(root)
	require.NoError(t, writer.Close())
	os.Stderr = origStderr
	console, err := io.ReadAll(reader)
	require.NoError(t, err)

	assert.Equal(t, 1, code)
	assert.Equal(t, 1, strings.Count(stdout.String(), `"status"`), stdout.String())
	assert.Contains(t, stdout.String(), "boom")
	assert.NotContains(t, string(console), "secret-warn")
	assert.NotContains(t, stderr.String(), "secret-warn")

	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "command failed")
	assert.Contains(t, string(body), "boom")
	assert.Contains(t, string(body), "secret-warn")
}

func TestShadowLog_FileKeepsDebugConsoleDoesNotEcho(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	t.Cleanup(func() {
		logger.RestoreLoggerState(savedLogger, savedLevel)
		viper.Reset()
		SetOutputMode("")
		SetCommandName("")
	})

	path := filepath.Join(t.TempDir(), "audit.log")
	viper.Reset()
	viper.Set(config.KeyAppLogLevel, "debug")
	viper.Set(config.KeyAppLogConsoleLevel, "")
	viper.Set(config.KeyAppLogFileLevel, "")
	viper.Set(config.KeyAppLogFileEnabled, true)
	viper.Set(config.KeyAppLogFilePath, path)
	viper.Set(config.KeyAppLogColorEnabled, "false")
	viper.Set(config.KeyAppLogSamplingEnabled, false)
	viper.Set(config.KeyAppOutputFormat, "text")

	console := &bytes.Buffer{}
	require.NoError(t, logger.Init(console))
	SetOutputMode("text")
	cmd := &cobra.Command{Use: "app"}
	child := &cobra.Command{Use: "ping"}
	cmd.AddCommand(child)
	SetCommandFrom(child)
	var out, errBuf bytes.Buffer
	child.SetOut(&out)
	child.SetErr(&errBuf)
	require.NoError(t, Resultf(child, "hello\n"))
	require.NoError(t, Statusf(child, "wait\n"))
	assert.Equal(t, "wait\n", errBuf.String())
	logger.Cleanup()

	assert.Equal(t, "hello\n", out.String())
	assert.NotContains(t, console.String(), "user output")
	assert.NotContains(t, console.String(), "hello")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "user output")
	assert.Contains(t, string(body), "hello")
	assert.Contains(t, string(body), "wait")
}

func newAliasCommand() (*cobra.Command, *cobra.Command) {
	root := &cobra.Command{Use: "app"}
	child := &cobra.Command{Use: "work"}
	child.Flags().String("format", "text", "alias")
	root.AddCommand(child)
	RegisterOutputFlag(root, WithEnvPrefix("APP"), WithAliasFlag("format"))
	return root, child
}

func silenceLoggerNoise() {
	viper.Set(config.KeyAppLogConsoleLevel, "warn")
	viper.Set(config.KeyAppLogLevel, "warn")
	viper.Set(config.KeyAppLogFileEnabled, false)
	viper.Set(config.KeyAppLogColorEnabled, "false")
	viper.Set(config.KeyAppLogSamplingEnabled, false)
}

func resetOutputState() {
	currentOutputMode = ""
	modePinned = false
	resolvedPhase = phaseUnset
	rendered = false
	renderedStatus = ""
	renderedErrorMessage = ""
	commandNameSet = false
	currentCommandName = ""
	currentCmd = nil
	modeWroteKey = false
	modeWroteValue = ""
	SetFailureLog(nil)
	SetAfterRun(nil)
}
