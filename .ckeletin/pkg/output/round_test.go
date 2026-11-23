package output

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasBelowRoot_KeepsTheRealUsageError(t *testing.T) {
	resetRound(t)

	t.Run("group alias json names the bogus flag", func(t *testing.T) {
		root := aliasTree(true)
		root.SetArgs([]string{"check", "model", "x", "--bogus", "--format", "json"})
		out := &bytes.Buffer{}
		root.SetOut(out)
		root.SetErr(out)
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		assert.NotContains(t, out.String(), "unknown flag: --format")
		env := decodeEnvelope(t, out.String())
		assert.Equal(t, "error", env["status"])
		assert.Contains(t, out.String(), "--bogus")
	})

	t.Run("leaf alias json names the unknown command", func(t *testing.T) {
		root := aliasTree(false)
		root.SetArgs([]string{"nosuch", "--format", "json"})
		out := &bytes.Buffer{}
		root.SetOut(out)
		root.SetErr(out)
		code := Execute(root)
		assert.NotEqual(t, 0, code)
		assert.NotContains(t, out.String(), "unknown flag: --format")
		env := decodeEnvelope(t, out.String())
		assert.Equal(t, "error", env["status"])
		assert.Nil(t, env["command"])
		assert.Contains(t, out.String(), "unknown command")
	})

	t.Run("text mode keeps the bogus flag", func(t *testing.T) {
		root := aliasTree(true)
		root.SetArgs([]string{"check", "model", "x", "--bogus", "--format", "text"})
		out := &bytes.Buffer{}
		root.SetOut(out)
		root.SetErr(out)
		_ = Execute(root)
		assert.Contains(t, out.String(), "unknown flag: --bogus")
		assert.NotContains(t, out.String(), "unknown flag: --format")
	})
}

func TestEnvJSON_UnknownCommandIsAnEnvelope(t *testing.T) {
	resetRound(t)
	t.Setenv("APP_APP_OUTPUT_FORMAT", "json")
	root := &cobra.Command{Use: "dapp"}
	root.AddCommand(&cobra.Command{Use: "real", RunE: func(*cobra.Command, []string) error { return nil }})
	RegisterOutputFlag(root, WithEnvPrefix("APP"))
	root.SetArgs([]string{"nosuch"})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	code := Execute(root)
	assert.NotEqual(t, 0, code)
	env := decodeEnvelope(t, out.String())
	assert.Equal(t, "error", env["status"])
	assert.Nil(t, env["command"])
}

func TestSuccessEnvelope_DoesNotSwallowTheError(t *testing.T) {
	resetRound(t)
	SetOutputMode("json")
	var audited error
	SetFailureLog(func(err error, _ string, code int) {
		audited = err
		assert.Equal(t, 0, code)
	})
	t.Cleanup(func() { SetFailureLog(nil) })

	root := &cobra.Command{Use: "dapp"}
	root.AddCommand(&cobra.Command{
		Use: "ok",
		RunE: func(cmd *cobra.Command, _ []string) error {
			require.NoError(t, RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
				Status: "success", Command: "ok", Data: nil,
			}))
			return errors.New("after success")
		},
	})
	root.SetArgs([]string{"ok"})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	assert.Equal(t, 0, Execute(root))
	require.Error(t, audited)
	assert.Contains(t, audited.Error(), "after success")
	assert.Equal(t, 1, stringsCount(out.String(), `"status"`))
}

func TestRenderedError_ReturnErrorDoesNotPrintTwice(t *testing.T) {
	resetRound(t)
	SetOutputMode("json")
	var audited bool
	SetFailureLog(func(error, string, int) { audited = true })
	t.Cleanup(func() { SetFailureLog(nil) })

	root := &cobra.Command{Use: "dapp"}
	root.AddCommand(&cobra.Command{
		Use: "bad",
		RunE: func(cmd *cobra.Command, _ []string) error {
			require.NoError(t, RenderJSON(cmd.OutOrStdout(), JSONEnvelope{
				Status: "error", Command: "bad",
				Error: &JSONError{Message: "first"},
			}))
			return errors.New("second")
		},
	})
	root.SetArgs([]string{"bad"})
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	assert.NotEqual(t, 0, Execute(root))
	assert.True(t, audited)
	assert.Equal(t, 1, stringsCount(out.String(), `"status"`))
	assert.NotContains(t, out.String(), "second")
}

func TestSetupClosesTheAuditFile(t *testing.T) {
	resetRound(t)
	dir := t.TempDir()
	viper.Set(config.KeyAppLogFileEnabled, true)
	viper.Set(config.KeyAppLogFilePath, filepath.Join(dir, "app.log"))
	viper.Set(config.KeyAppLogFileLevel, "debug")
	viper.Set(config.KeyAppLogColorEnabled, "false")
	root := &cobra.Command{Use: "dapp", RunE: func(*cobra.Command, []string) error { return nil }}
	Setup(root, WithEnvPrefix("APP"))
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	assert.Equal(t, 0, Execute(root))
	assert.False(t, logger.AuditOpen())
}

func aliasTree(aliasOnGroup bool) *cobra.Command {
	root := &cobra.Command{Use: "dapp"}
	group := &cobra.Command{Use: "check"}
	leaf := &cobra.Command{
		Use:  "model",
		Args: cobra.ExactArgs(1),
		RunE: func(*cobra.Command, []string) error { return nil },
	}
	if aliasOnGroup {
		group.PersistentFlags().String("format", "text", "format")
	} else {
		leaf.Flags().String("format", "text", "format")
	}
	group.AddCommand(leaf)
	root.AddCommand(group)
	RegisterOutputFlag(root, WithAliasFlag("format"), WithEnvPrefix("APP"))
	return root
}

func resetRound(t *testing.T) {
	t.Helper()
	resetOutputState()
	viper.Reset()
	t.Cleanup(func() {
		resetOutputState()
		viper.Reset()
		logger.Cleanup()
	})
}

func decodeEnvelope(t *testing.T, text string) map[string]any {
	t.Helper()
	var env map[string]any
	require.NoError(t, json.Unmarshal([]byte(text), &env), text)
	return env
}

func stringsCount(text, part string) int {
	n := 0
	for i := 0; i+len(part) <= len(text); i++ {
		if text[i:i+len(part)] == part {
			n++
		}
	}
	return n
}
