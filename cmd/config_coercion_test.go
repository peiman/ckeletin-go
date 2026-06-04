// cmd/config_coercion_test.go

package cmd

import (
	"testing"

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
