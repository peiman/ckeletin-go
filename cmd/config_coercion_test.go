// cmd/config_coercion_test.go

package cmd

import (
	"bytes"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// Env vars always arrive through viper as strings. These tests pin that
// non-string config values provided as strings (the env-var case) are coerced
// to their declared type instead of being silently dropped to the zero value.

func TestGetConfigValueWithFlags_BoolFromEnvString(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.ping.ui", "true") // as an env var would arrive

	cmd := &cobra.Command{}
	cmd.Flags().Bool("ui", false, "")

	got := getConfigValueWithFlags[bool](cmd, "ui", "app.ping.ui")
	assert.True(t, got, "string env value \"true\" must coerce to bool true, not fall back to false")
}

func TestGetConfigValueWithFlags_IntFromEnvString(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.some.count", "42")

	cmd := &cobra.Command{}
	cmd.Flags().Int("count", 0, "")

	got := getConfigValueWithFlags[int](cmd, "count", "app.some.count")
	assert.Equal(t, 42, got, "string env value \"42\" must coerce to int 42")
}

func TestGetConfigValueWithFlags_StringStillWorks(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.ping.output_message", "hello")

	cmd := &cobra.Command{}
	cmd.Flags().String("message", "", "")

	got := getConfigValueWithFlags[string](cmd, "message", "app.ping.output_message")
	assert.Equal(t, "hello", got)
}

func TestGetConfigValueWithFlags_FlagOverridesEnvCoercion(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.ping.ui", "false")

	cmd := &cobra.Command{}
	cmd.Flags().Bool("ui", false, "")
	_ = cmd.Flags().Set("ui", "true") // explicit flag wins over the env/config value

	got := getConfigValueWithFlags[bool](cmd, "ui", "app.ping.ui")
	assert.True(t, got, "an explicitly-set flag must override the (coerced) viper value")
}

func TestGetKeyValue_BoolFromEnvString(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.feature.enabled", "true")

	assert.True(t, getKeyValue[bool]("app.feature.enabled"),
		"string env value \"true\" must coerce to bool true")
}

func TestGetKeyValue_Float64FromEnvString(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.some.ratio", "3.14")

	assert.Equal(t, 3.14, getKeyValue[float64]("app.some.ratio"),
		"string env value \"3.14\" must coerce to float64 3.14")
}

func TestGetKeyValue_StringSliceFromEnvString(t *testing.T) {
	viper.Reset()
	defer viper.Reset()
	viper.Set("app.some.list", "a b c") // env-var shape: a delimited string, not a native slice

	assert.Equal(t, []string{"a", "b", "c"}, getKeyValue[[]string]("app.some.list"),
		"a delimited string env value must coerce to []string")
}

// TestCoerceViperValue_LogsAndIgnoresUnparseable pins that a value which cannot be
// coerced (e.g. a typo'd bool env var) is NOT silently dropped to zero — it is
// logged at WARN, so the failure is debuggable rather than invisible.
func TestCoerceViperValue_LogsAndIgnoresUnparseable(t *testing.T) {
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	var buf bytes.Buffer
	log.Logger = zerolog.New(&buf)
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	viper.Reset()
	defer viper.Reset()
	viper.Set("app.feature.enabled", "yse") // typo'd bool — not coercible

	got := getKeyValue[bool]("app.feature.enabled")
	assert.False(t, got, "an unparseable value falls back to the zero value")
	assert.Contains(t, buf.String(), "could not be coerced",
		"a warning must be logged so the coercion failure is not silent")
	assert.Contains(t, buf.String(), "app.feature.enabled",
		"the warning must name the offending key")
}
