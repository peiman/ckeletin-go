// Package infrastructure handles all infrastructure-related operations.
package infrastructure

import (
	"fmt"

	"github.com/spf13/viper"
)

// Default configuration values.
const (
	DefaultLogLevel = "info"
)

// Config holds all configuration for our program.
type Config struct {
	LogLevel string
}

// LoadConfig reads configuration from file or environment variables.
func LoadConfig() (*Config, error) {
	viper.SetDefault("LogLevel", DefaultLogLevel)

	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	return &config, nil
}
