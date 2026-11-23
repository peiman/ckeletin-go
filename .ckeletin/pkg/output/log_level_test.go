package output

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogLevelFlagBeatsEnvBeatsConfigBeatsDefault(t *testing.T) {
	viper.Reset()
	t.Cleanup(viper.Reset)
	config.SetDefaults()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	viper.SetConfigType("yaml")
	require.NoError(t, viper.MergeConfig(strings.NewReader("app:\n  log_level: error\n  log:\n    console_level: info\n    file_level: trace\n")))

	t.Run("flag", func(t *testing.T) {
		t.Setenv("APP_APP_LOG_LEVEL", "panic")
		t.Setenv("APP_APP_LOG_CONSOLE_LEVEL", "fatal")
		t.Setenv("APP_APP_LOG_FILE_LEVEL", "info")
		executeLogLevels(t, "--log-level", "debug", "--log-level-console", "warn", "--log-level-file", "error")
		assert.Equal(t, "debug", viper.GetString(config.KeyAppLogLevel))
		assert.Equal(t, "warn", viper.GetString(config.KeyAppLogConsoleLevel))
		assert.Equal(t, "error", viper.GetString(config.KeyAppLogFileLevel))
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("APP_APP_LOG_LEVEL", "warn")
		t.Setenv("APP_APP_LOG_CONSOLE_LEVEL", "debug")
		t.Setenv("APP_APP_LOG_FILE_LEVEL", "info")
		executeLogLevels(t)
		assert.Equal(t, "warn", viper.GetString(config.KeyAppLogLevel))
		assert.Equal(t, "debug", viper.GetString(config.KeyAppLogConsoleLevel))
		assert.Equal(t, "info", viper.GetString(config.KeyAppLogFileLevel))
	})

	t.Run("config", func(t *testing.T) {
		executeLogLevels(t)
		assert.Equal(t, "error", viper.GetString(config.KeyAppLogLevel))
		assert.Equal(t, "info", viper.GetString(config.KeyAppLogConsoleLevel))
		assert.Equal(t, "trace", viper.GetString(config.KeyAppLogFileLevel))
	})

	t.Run("default", func(t *testing.T) {
		viper.Reset()
		t.Cleanup(viper.Reset)
		config.SetDefaults()
		executeLogLevels(t)
		assert.Equal(t, "warn", viper.GetString(config.KeyAppLogLevel))
		assert.Equal(t, "", viper.GetString(config.KeyAppLogConsoleLevel))
		assert.Equal(t, "debug", viper.GetString(config.KeyAppLogFileLevel))
	})
}

func TestLogLevelFlagsApplyBeforeFirstLogLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	viper.Reset()
	t.Cleanup(func() {
		logger.Cleanup()
		viper.Reset()
	})
	viper.Set(config.KeyAppLogFileEnabled, true)
	viper.Set(config.KeyAppLogFilePath, path)
	viper.Set(config.KeyAppLogColorEnabled, "false")
	viper.Set(config.KeyAppLogSamplingEnabled, false)

	console := &bytes.Buffer{}
	root := &cobra.Command{
		Use: "app",
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return logger.Init(console)
		},
		RunE: func(*cobra.Command, []string) error {
			log.Debug().Msg("first-line")
			return nil
		},
	}
	RegisterLogLevelFlags(root)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs([]string{"--log-level-console", "debug", "--log-level-file", "debug"})

	require.NoError(t, root.Execute())

	assert.Contains(t, console.String(), "first-line")
	logger.Cleanup()
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(body), "first-line")
}

func executeLogLevels(t *testing.T, args ...string) {
	t.Helper()
	root := &cobra.Command{Use: "app", RunE: func(*cobra.Command, []string) error { return nil }}
	RegisterLogLevelFlags(root)
	root.SetOut(io.Discard)
	root.SetErr(io.Discard)
	root.SetArgs(args)
	require.NoError(t, root.Execute())
}
