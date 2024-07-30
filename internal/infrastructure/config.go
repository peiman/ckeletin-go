package infrastructure

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config holds all configuration for our program
type Config struct {
	LogLevel string
	Server   ServerConfig
}

type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig holds all database-related configuration
type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	viper.SetDefault("LogLevel", "info")
	viper.SetDefault("Server.Port", 8080)
	viper.SetDefault("Server.Host", "localhost")

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}
