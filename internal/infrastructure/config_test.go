package infrastructure

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Test with default values
	t.Run("Default Values", func(t *testing.T) {
		viper.Reset()
		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "info", config.LogLevel)
	})

	// Test with environment variables
	t.Run("Environment Variables", func(t *testing.T) {
		viper.Reset()
		os.Setenv("CKELETIN_LOGLEVEL", "debug")
		defer os.Unsetenv("CKELETIN_LOGLEVEL")

		viper.AutomaticEnv()
		viper.SetEnvPrefix("CKELETIN")
		err := viper.BindEnv("LogLevel", "CKELETIN_LOGLEVEL")
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "debug", config.LogLevel)
	})

	// Test with config file
	t.Run("Config File", func(t *testing.T) {
		viper.Reset()
		viper.SetConfigType("json")
		jsonConfig := []byte(`{
			"LogLevel": "warn"
		}`)
		err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "warn", config.LogLevel)
	})

	// Test error case
	t.Run("Unmarshal Error", func(t *testing.T) {
		viper.Reset()
		viper.Set("LogLevel", make(chan int)) // This will cause an unmarshal error

		_, err := LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to decode into struct")
	})
}
