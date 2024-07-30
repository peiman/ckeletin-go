package infrastructure

import (
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
		assert.Equal(t, "localhost", config.Database.Host)
		assert.Equal(t, 5432, config.Database.Port)
		assert.Equal(t, "postgres", config.Database.User)
		assert.Equal(t, "ckeletin", config.Database.Name)
	})

	// Test with environment variables
	t.Run("Environment Variables", func(t *testing.T) {
		viper.Reset()
		os.Setenv("CKELETIN_LOGLEVEL", "debug")
		os.Setenv("CKELETIN_SERVER_PORT", "9090")
		os.Setenv("CKELETIN_DATABASE_USER", "testuser")
		defer os.Unsetenv("CKELETIN_LOGLEVEL")
		defer os.Unsetenv("CKELETIN_SERVER_PORT")
		defer os.Unsetenv("CKELETIN_DATABASE_USER")

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "debug", config.LogLevel)
		assert.Equal(t, 9090, config.Server.Port)
		assert.Equal(t, "testuser", config.Database.User)
	})

	// Test with config file
	t.Run("Config File", func(t *testing.T) {
		viper.Reset()
		configContent := []byte(`{
			"LogLevel": "warn",
			"Server": {
				"Port": 3000,
				"Host": "127.0.0.1"
			},
			"Database": {
				"Host": "db.example.com",
				"Port": 5433,
				"User": "admin",
				"Password": "secret",
				"Name": "testdb"
			}
		}`)
		err := os.WriteFile("ckeletin-go.json", configContent, 0644)
		assert.NoError(t, err)
		defer os.Remove("ckeletin-go.json")

		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "warn", config.LogLevel)
		assert.Equal(t, 3000, config.Server.Port)
		assert.Equal(t, "127.0.0.1", config.Server.Host)
		assert.Equal(t, "db.example.com", config.Database.Host)
		assert.Equal(t, 5433, config.Database.Port)
		assert.Equal(t, "admin", config.Database.User)
		assert.Equal(t, "secret", config.Database.Password)
		assert.Equal(t, "testdb", config.Database.Name)
	})
}
