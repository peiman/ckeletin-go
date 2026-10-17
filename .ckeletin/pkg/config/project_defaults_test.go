package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const projectDefaultWarn = "[defaults]\n\"app.log_level\" = \"warn\"\n"

func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestApplyProjectDefaultsTOML_ReplacesFrameworkDefault(t *testing.T) {
	resetViper(t)
	config.SetDefaults()
	assert.Equal(t, "info", viper.GetString(config.KeyAppLogLevel))

	require.NoError(t, config.ApplyProjectDefaultsTOML([]byte(projectDefaultWarn)))

	assert.Equal(t, "warn", viper.GetString(config.KeyAppLogLevel),
		"a project default is what viper returns when nothing else sets the option")
}

func TestApplyProjectDefaultsTOML_ConfigFileWins(t *testing.T) {
	resetViper(t)
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte("app:\n  log_level: info\n"), 0o644))

	config.SetDefaults()
	require.NoError(t, config.ApplyProjectDefaultsTOML([]byte(projectDefaultWarn)))
	viper.SetConfigFile(path)
	require.NoError(t, viper.ReadInConfig())

	assert.Equal(t, "info", viper.GetString(config.KeyAppLogLevel),
		"a config file still wins over a project default")
}

func TestApplyProjectDefaultsTOML_EnvWins(t *testing.T) {
	resetViper(t)
	viper.SetEnvPrefix("CKELETIN_GO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	t.Setenv("CKELETIN_GO_APP_LOG_LEVEL", "error")

	config.SetDefaults()
	require.NoError(t, config.ApplyProjectDefaultsTOML([]byte(projectDefaultWarn)))

	assert.Equal(t, "error", viper.GetString(config.KeyAppLogLevel),
		"an environment variable still wins over a project default")
}

func TestApplyProjectDefaultsTOML_UnknownKey(t *testing.T) {
	resetViper(t)
	config.SetDefaults()

	err := config.ApplyProjectDefaultsTOML([]byte("[defaults]\n\"app.not_a_real_option\" = \"x\"\n"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "app.not_a_real_option")
	assert.Equal(t, "info", viper.GetString(config.KeyAppLogLevel),
		"a rejected file must not change defaults already applied")
}

func TestApplyProjectDefaultsTOML_EmptyTableKeepsFrameworkDefault(t *testing.T) {
	resetViper(t)
	config.SetDefaults()

	require.NoError(t, config.ApplyProjectDefaultsTOML([]byte("[defaults]\n")))
	require.NoError(t, config.ApplyProjectDefaultsTOML(nil))
	require.NoError(t, config.ApplyProjectDefaultsTOML([]byte("# no defaults\n")))

	assert.Equal(t, "info", viper.GetString(config.KeyAppLogLevel))
}

func TestApplyProjectDefaultsTOML_UnknownTable(t *testing.T) {
	resetViper(t)
	err := config.ApplyProjectDefaultsTOML([]byte("other = 1\n"))
	require.Error(t, err)
}

func TestApplyProjectDefaultsTOML_InvalidTOML(t *testing.T) {
	resetViper(t)
	err := config.ApplyProjectDefaultsTOML([]byte("[defaults\n"))
	require.Error(t, err)
}

func TestEffectiveRegistry_ReplacesDefaultOnACopy(t *testing.T) {
	resetViper(t)
	config.SetDefaults()

	opts, err := config.EffectiveRegistry([]byte(projectDefaultWarn))
	require.NoError(t, err)

	var effective interface{}
	for _, opt := range opts {
		if opt.Key == config.KeyAppLogLevel {
			effective = opt.DefaultValue
		}
	}
	assert.Equal(t, "warn", effective, "the overlay replaces the framework default on the returned copy")
	assert.Equal(t, "info", viper.GetString(config.KeyAppLogLevel),
		"reading the effective registry must not change the process defaults")

	var framework interface{}
	for _, opt := range config.Registry() {
		if opt.Key == config.KeyAppLogLevel {
			framework = opt.DefaultValue
		}
	}
	assert.Equal(t, "info", framework, "the registry itself keeps the framework default")
}

func TestEffectiveRegistry_UnknownKey(t *testing.T) {
	opts, err := config.EffectiveRegistry([]byte("[defaults]\n\"app.not_a_real_option\" = \"x\"\n"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "app.not_a_real_option")
	assert.Nil(t, opts)
}

func TestEffectiveRegistry_EmptyIsFrameworkDefault(t *testing.T) {
	opts, err := config.EffectiveRegistry([]byte("[defaults]\n"))
	require.NoError(t, err)
	var got interface{}
	for _, opt := range opts {
		if opt.Key == config.KeyAppLogLevel {
			got = opt.DefaultValue
		}
	}
	assert.Equal(t, "info", got)
}
