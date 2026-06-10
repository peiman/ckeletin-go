// .ckeletin/pkg/config/validator/validator.go
//
// Configuration validation functionality

package validator

import (
	"fmt"
	"os"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	_ "github.com/peiman/ckeletin-go/.ckeletin/pkg/config/commands" // Import to trigger init() registration
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

	// 2. Validate file security (size and permissions)
	if err := config.ValidateConfigFileSecurity(configPath, config.MaxConfigFileSize); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err)
	}

	// 3. Try to parse the config file
	v := viper.New()
	v.SetConfigFile(configPath)
	if err := v.ReadInConfig(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Errorf("failed to parse config: %w", err))
		return result, nil // Return partial results
	}

	// Capture file-only settings before seeding defaults so steps 4-6 see
	// exactly what the file contains.
	allSettings := v.AllSettings()

	// 4. Validate registered option values (log levels, colors, formats).
	// Seed registry defaults so options absent from the file are validated
	// against their defaults rather than nil (some validators reject nil).
	for _, opt := range config.Registry() {
		v.SetDefault(opt.Key, opt.DefaultValue)
	}
	optionErrors := config.ValidateRegisteredOptionsWithViper(v)
	if len(optionErrors) > 0 {
		result.Valid = false
		for _, optErr := range optionErrors {
			result.Errors = append(result.Errors, attributeOptionError(optErr, allSettings))
		}
	}

	// 5. Validate all configuration values
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

// attributeOptionError re-attributes a registered-option validation error
// when the offending key is absent from the file: the failing value then came
// from the seeded registry default, and the plain `config "app.x": invalid
// value` message would send users hunting for a key they never set. The key
// is recovered from the stable `config "<key>":` prefix that
// ValidateRegisteredOptionsWithViper puts on every error it returns.
func attributeOptionError(err error, fileSettings map[string]interface{}) error {
	for _, opt := range config.Registry() {
		if opt.Validation == nil {
			continue
		}
		if !strings.HasPrefix(err.Error(), fmt.Sprintf("config %q:", opt.Key)) {
			continue
		}
		if settingsContainKey(fileSettings, opt.Key) {
			return err
		}
		return fmt.Errorf("(registry default, not set in file) %w", err)
	}
	return err
}

// settingsContainKey reports whether the dotted key is present in the nested
// settings map parsed from the config file.
func settingsContainKey(settings map[string]interface{}, key string) bool {
	parts := strings.Split(key, ".")
	current := settings
	for i, part := range parts {
		value, ok := current[part]
		if !ok {
			return false
		}
		if i == len(parts)-1 {
			return true
		}
		nested, ok := value.(map[string]interface{})
		if !ok {
			return false
		}
		current = nested
	}
	return false
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
