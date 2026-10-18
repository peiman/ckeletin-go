package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restoreRootFlags(t *testing.T) {
	t.Helper()
	type snap struct {
		flag    *pflag.Flag
		value   string
		def     string
		changed bool
	}
	var saved []snap
	RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		saved = append(saved, snap{flag, flag.Value.String(), flag.DefValue, flag.Changed})
	})
	t.Cleanup(func() {
		for _, item := range saved {
			item.flag.Value.Set(item.value) //nolint:errcheck // restoring saved state
			item.flag.DefValue = item.def
			item.flag.Changed = item.changed
		}
	})
}

func TestRootFlags_ProjectDefaultIsFlagDefault(t *testing.T) {
	restoreRootFlags(t)

	err := applyRootFlagDefaults([]byte(`
[defaults]
"app.log_level" = "warn"
"app.log.file_enabled" = true
"app.log.file_max_size" = 7
"app.output_format" = "json"
`))
	require.NoError(t, err)

	logLevel := RootCmd.PersistentFlags().Lookup("log-level")
	require.NotNil(t, logLevel)
	assert.Equal(t, "warn", logLevel.DefValue)
	assert.Equal(t, "warn", logLevel.Value.String())
	assert.False(t, logLevel.Changed)
	assert.Equal(t, "Set the log level (trace, debug, info, warn, error, fatal, panic)", logLevel.Usage)

	enabled := RootCmd.PersistentFlags().Lookup("log-file-enabled")
	require.NotNil(t, enabled)
	assert.Equal(t, "true", enabled.DefValue)
	assert.False(t, enabled.Changed)

	maxSize := RootCmd.PersistentFlags().Lookup("log-file-max-size")
	require.NotNil(t, maxSize)
	assert.Equal(t, "7", maxSize.DefValue)
	assert.Equal(t, "7", maxSize.Value.String())

	outputFlag := RootCmd.PersistentFlags().Lookup("output")
	require.NotNil(t, outputFlag)
	assert.Equal(t, "json", outputFlag.DefValue)
	assert.False(t, outputFlag.Changed, "installing a default is not an explicit --output")

	assert.Equal(t, "", RootCmd.PersistentFlags().Lookup("config").DefValue)
	assert.Equal(t, ConfigPathModeXDG, RootCmd.PersistentFlags().Lookup("config-path-mode").DefValue)

	var published string
	for _, flag := range buildCatalog(RootCmd).GlobalFlags {
		if flag.Long == "log-level" {
			published = flag.Default
		}
	}
	assert.Equal(t, "warn", published, "catalog publishes the same default --help does")
}

func TestRootFlags_ProjectDefaultDoesNotOverrideExplicitViperValue(t *testing.T) {
	restoreRootFlags(t)
	viper.Reset()
	t.Cleanup(viper.Reset)

	require.NoError(t, applyRootFlagDefaults([]byte(`
[defaults]
"app.log_level" = "warn"
`)))
	require.NoError(t, bindFlags(RootCmd))
	viper.Set(config.KeyAppLogLevel, "error")

	assert.Equal(t, "error", viper.GetString(config.KeyAppLogLevel))
	assert.False(t, RootCmd.PersistentFlags().Lookup("log-level").Changed)
}

func TestRootFlags_UnknownProjectDefaultLeavesFlagsUntouched(t *testing.T) {
	before := RootCmd.PersistentFlags().Lookup("log-level").DefValue

	err := applyRootFlagDefaults([]byte(`
[defaults]
"app.not_a_key" = "x"
`))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown option")
	assert.Equal(t, before, RootCmd.PersistentFlags().Lookup("log-level").DefValue)
}
