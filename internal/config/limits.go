// internal/config/limits.go
//
// Configuration value size limits to prevent DoS attacks

package config

import (
	"fmt"
)

const (
	// MaxStringValueLength is the maximum length for string config values (10 KB)
	MaxStringValueLength = 10 * 1024

	// MaxSliceLength is the maximum number of elements in array config values
	MaxSliceLength = 1000

	// MaxConfigFileSize is the maximum size for config files (1 MB)
	MaxConfigFileSize = 1 * 1024 * 1024
)

// ValidateConfigValue checks if a config value is within acceptable limits.
// This prevents DoS attacks via excessively large configuration values.
func ValidateConfigValue(key string, value interface{}) error {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string:
		if len(v) > MaxStringValueLength {
			return fmt.Errorf("config value for %s exceeds max string length (%d > %d bytes)",
				key, len(v), MaxStringValueLength)
		}

	case []string:
		if len(v) > MaxSliceLength {
			return fmt.Errorf("config array for %s exceeds max length (%d > %d elements)",
				key, len(v), MaxSliceLength)
		}
		// Also validate each string in the slice
		for i, s := range v {
			if len(s) > MaxStringValueLength {
				return fmt.Errorf("config array element %d for %s exceeds max string length (%d > %d bytes)",
					i, key, len(s), MaxStringValueLength)
			}
		}

	case []interface{}:
		if len(v) > MaxSliceLength {
			return fmt.Errorf("config array for %s exceeds max length (%d > %d elements)",
				key, len(v), MaxSliceLength)
		}
		// Validate each element
		for i, item := range v {
			if s, ok := item.(string); ok {
				if len(s) > MaxStringValueLength {
					return fmt.Errorf("config array element %d for %s exceeds max string length (%d > %d bytes)",
						i, key, len(s), MaxStringValueLength)
				}
			}
		}

	case map[string]interface{}:
		// For nested maps, recursively validate
		for nestedKey, nestedValue := range v {
			fullKey := fmt.Sprintf("%s.%s", key, nestedKey)
			if err := ValidateConfigValue(fullKey, nestedValue); err != nil {
				return err
			}
		}

	// Numeric types don't need size validation
	case int, int8, int16, int32, int64:
		return nil
	case uint, uint8, uint16, uint32, uint64:
		return nil
	case float32, float64:
		return nil
	case bool:
		return nil

	default:
		// Unknown types are allowed but logged elsewhere
		return nil
	}

	return nil
}

// ValidateAllConfigValues validates all configuration values in the registry.
// Returns a slice of errors for any values that exceed limits.
func ValidateAllConfigValues(values map[string]interface{}) []error {
	var errors []error

	for key, value := range values {
		if err := ValidateConfigValue(key, value); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
