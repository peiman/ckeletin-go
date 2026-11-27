package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdoption_ReferenceRootExtensionPoints(t *testing.T) {
	t.Cleanup(func() {
		SetEnvPrefix("")
		SetExtendedRootFlags(true)
		applyRootOptions()
		SetConfigStatusLevel(zerolog.InfoLevel)
		config.SetStartupCheck(config.StartupStrict)
		viper.Set(config.KeyAppLogFileLevel, "debug")
		viper.Set(config.KeyAppLogColorEnabled, "auto")
	})

	t.Run("empty file level passes validation", func(t *testing.T) {
		config.SetDefaults()
		viper.Set(config.KeyAppLogFileLevel, "")
		for _, err := range config.ValidateRegisteredOptions() {
			assert.NotContains(t, err.Error(), "file_level")
		}
	})

	t.Run("warn does not fail an old value", func(t *testing.T) {
		config.SetStartupCheck(config.StartupWarn)
		config.SetDefaults()
		viper.Set(config.KeyAppLogColorEnabled, "nope")
		orig := cfgFile
		cfgFile = ""
		t.Cleanup(func() { cfgFile = orig })
		require.NoError(t, initConfig())
	})

	t.Run("prefix is explicit", func(t *testing.T) {
		SetEnvPrefix("daana")
		assert.Equal(t, "DAANA", EnvPrefix())
	})

	t.Run("status level is choosable", func(t *testing.T) {
		SetConfigStatusLevel(zerolog.DebugLevel)
		assert.Equal(t, zerolog.DebugLevel, configStatusLevel)
	})

	t.Run("extra flags can be hidden", func(t *testing.T) {
		SetExtendedRootFlags(false)
		applyRootOptions()
		flag := RootCmd.PersistentFlags().Lookup("log-file-path")
		require.NotNil(t, flag)
		assert.True(t, flag.Hidden)
		assert.False(t, RootCmd.PersistentFlags().Lookup("log-level").Hidden)
	})

	t.Run("help survives a broken config file", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "config.yaml")
		require.NoError(t, os.WriteFile(path, []byte(":\n: not yaml\n"), 0o644))
		orig := cfgFile
		cfgFile = path
		t.Cleanup(func() { cfgFile = orig })
		RootCmd.SetArgs([]string{"help"})
		require.NoError(t, RootCmd.Execute())
	})
}
