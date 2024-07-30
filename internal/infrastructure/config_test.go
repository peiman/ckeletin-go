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
		assert.Equal(t, 8080, config.Server.Port)
		assert.Equal(t, "localhost", config.Server.Host)
	})

	// Test with environment variables
	t.Run("Environment Variables", func(t *testing.T) {
		viper.Reset()
		os.Setenv("CKELETIN_LOGLEVEL", "debug")
		os.Setenv("CKELETIN_SERVER_PORT", "9090")
		os.Setenv("CKELETIN_SERVER_HOST", "127.0.0.1")
		defer os.Unsetenv("CKELETIN_LOGLEVEL")
		defer os.Unsetenv("CKELETIN_SERVER_PORT")
		defer os.Unsetenv("CKELETIN_SERVER_HOST")

		viper.AutomaticEnv()
		viper.SetEnvPrefix("CKELETIN")
		err := viper.BindEnv("LogLevel", "CKELETIN_LOGLEVEL")
		assert.NoError(t, err)
		err = viper.BindEnv("Server.Port", "CKELETIN_SERVER_PORT")
		assert.NoError(t, err)
		err = viper.BindEnv("Server.Host", "CKELETIN_SERVER_HOST")
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, 9090, config.Server.Port)
		assert.Equal(t, "127.0.0.1", config.Server.Host)
	})

	// Test with config file
	t.Run("Config File", func(t *testing.T) {
		viper.Reset()
		viper.SetConfigType("json")
		var jsonConfig = []byte(`{
					"LogLevel": "warn",
					"Server": {
							"Port": 3000,
							"Host": "127.0.0.1"
					}
			}`)
		err := viper.ReadConfig(bytes.NewBuffer(jsonConfig))
		assert.NoError(t, err)

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "warn", config.LogLevel)
		assert.Equal(t, 3000, config.Server.Port)
		assert.Equal(t, "127.0.0.1", config.Server.Host)
	})

	// Test error case
	t.Run("Unmarshal Error", func(t *testing.T) {
		viper.Reset()
		viper.Set("Server", "invalid")

		_, err := LoadConfig()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to decode into struct")
	})
}
