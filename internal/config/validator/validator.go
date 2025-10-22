// internal/config/validator/validator.go
//
// Configuration validation functionality

package validator

import (
	"fmt"
	"os"

	"github.com/peiman/ckeletin-go/internal/config"
	_ "github.com/peiman/ckeletin-go/internal/config/commands" // Import to trigger init() registration
	"github.com/spf13/viper"
)

// Result represents the outcome of a configuration validation
type Result struct {
	Valid      bool
	Errors     []error
	Warnings   []string
	ConfigFile string
}

// Validate performs comprehensive validation of a configuration file
func Validate(configPath string) (*Result, error) {
	result := &Result{
		Valid:      true,
		ConfigFile: configPath,
	}

	// 1. Check if file exists
	if _, err := os.Stat(configPath); err != nil {
		return nil, fmt.Errorf("config file not found: %w", err)
	}

	// 2. Validate file size
	if err := config.ValidateConfigFileSize(configPath, config.MaxConfigFileSize); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err)
	}

	// 3. Validate file permissions
	if err := config.ValidateConfigFilePermissions(configPath); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err)
	}

	// 4. Try to parse the config file
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse config: %w", err))
		return result, nil // Return partial results
	}

	// 5. Validate all configuration values
	allSettings := v.AllSettings()
	valueErrors := config.ValidateAllConfigValues(allSettings)
	if len(valueErrors) > 0 {
		result.Valid = false
		result.Errors = append(result.Errors, valueErrors...)
	}

	// 6. Check for unknown keys (keys not in registry)
	knownKeys := make(map[string]bool)
	for _, opt := range config.Registry() {
		knownKeys[opt.Key] = true
	}

	unknownKeys := findUnknownKeys(allSettings, "", knownKeys)
	if len(unknownKeys) > 0 {
		for _, key := range unknownKeys {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("Unknown configuration key: %s (will be ignored)", key))
		}
	}

	return result, nil
}

// findUnknownKeys recursively finds configuration keys that aren't in the registry
func findUnknownKeys(settings map[string]interface{}, prefix string, knownKeys map[string]bool) []string {
	var unknown []string

	for key, value := range settings {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		// Check if this key is known
		if !knownKeys[fullKey] {
			// Check if it's a nested map
			if nestedMap, ok := value.(map[string]interface{}); ok {
				// Recursively check nested keys
				unknown = append(unknown, findUnknownKeys(nestedMap, fullKey, knownKeys)...)
			} else {
				unknown = append(unknown, fullKey)
			}
		}
	}

	return unknown
}
