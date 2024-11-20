package infrastructure

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/peiman/ckeletin-go/internal/errors"
)

// ConfigManager handles operations related to the configuration file.
type ConfigManager struct {
	ConfigPath string
}

// NewConfigManager creates a new ConfigManager.
func NewConfigManager(configPath string) *ConfigManager {
	if configPath == "" {
		configPath = DefaultConfigFileName
	}
	return &ConfigManager{
		ConfigPath: configPath,
	}
}

// EnsureConfig makes sure a config file exists, creating a default one if it doesn't.
func (cm *ConfigManager) EnsureConfig() error {
	if _, err := os.Stat(cm.ConfigPath); os.IsNotExist(err) {
		return cm.CreateDefaultConfig()
	}
	return nil
}

// CreateDefaultConfig creates a default configuration file.
func (cm *ConfigManager) CreateDefaultConfig() error {
	// Use a separate struct for JSON to get human-readable log levels
	type jsonConfig struct {
		LogLevel string     `json:"logLevel"`
		Ping     PingConfig `json:"ping"`
	}

	defaultConfig := jsonConfig{
		LogLevel: DefaultLogLevel.String(), // Convert to string for readability
		Ping: PingConfig{
			DefaultCount:  DefaultPingCount,
			OutputMessage: DefaultPingMessage,
			ColoredOutput: false,
		},
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return errors.NewAppError(errors.ErrInvalidConfig, "Failed to marshal default config", err)
	}

	dir := filepath.Dir(cm.ConfigPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return errors.NewAppError(errors.ErrInvalidConfig, "Config directory does not exist", err)
	}

	if err := os.WriteFile(cm.ConfigPath, data, FilePerms); err != nil {
		return errors.NewAppError(errors.ErrInvalidConfig, "Failed to write default config file", err)
	}

	return nil
}
