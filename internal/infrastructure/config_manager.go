package infrastructure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/peiman/ckeletin-go/internal/errors"
)

// Default configuration constants.
const (
	DefaultConfigFileName = "ckeletin-go.json"
	DirPerms              = 0o755
	FilePerms             = 0o600
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
	defaultConfig := Config{
		LogLevel: DefaultLogLevel,
	}

	data, err := json.MarshalIndent(defaultConfig, "", "  ")
	if err != nil {
		return errors.NewAppError(errors.ErrInvalidConfig, "Failed to marshal default config", err)
	}

	dir := filepath.Dir(cm.ConfigPath)
	if err := os.MkdirAll(dir, DirPerms); err != nil {
		return errors.NewAppError(errors.ErrInvalidConfig, "Failed to create config directory", err)
	}

	if err := os.WriteFile(cm.ConfigPath, data, FilePerms); err != nil {
		return errors.NewAppError(errors.ErrInvalidConfig, "Failed to write default config file", err)
	}

	fmt.Printf("Created default configuration file: %s\n", cm.ConfigPath)
	return nil
}
