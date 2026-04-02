// .ckeletin/pkg/config/validators.go
//
// Reusable validation functions for ConfigOption.Validation fields.
//
// These validators are designed to be used as the Validation field on
// ConfigOption entries. Each returns a function with the signature
// func(value interface{}) error that validates the config value at load time.

package config

import (
	"fmt"
	"strings"
)

// ValidLogLevels defines the set of valid log level strings.
// This matches zerolog's accepted level names.
var ValidLogLevels = map[string]bool{
	"trace":    true,
	"debug":    true,
	"info":     true,
	"warn":     true,
	"error":    true,
	"fatal":    true,
	"panic":    true,
	"disabled": true,
}

// ValidateOneOf returns a validation function that checks whether the string
// value is one of the allowed values. Empty strings are allowed when
// allowEmpty is true (useful for optional fields that fall back to defaults).
func ValidateOneOf(allowed []string, allowEmpty bool) func(interface{}) error {
	set := make(map[string]bool, len(allowed))
	for _, v := range allowed {
		set[v] = true
	}

	return func(value interface{}) error {
		s, ok := value.(string)
		if !ok {
			// Non-string values are skipped (type validation is separate)
			return nil
		}

		s = strings.TrimSpace(strings.ToLower(s))
		if s == "" && allowEmpty {
			return nil
		}

		if !set[s] {
			return fmt.Errorf("invalid value %q (valid options: %s)",
				s, strings.Join(allowed, ", "))
		}

		return nil
	}
}

// ValidateLogLevel returns a validation function for log level strings.
// If allowEmpty is true, empty string is accepted (for optional log level
// fields that fall back to a default).
func ValidateLogLevel(allowEmpty bool) func(interface{}) error {
	levels := make([]string, 0, len(ValidLogLevels))
	for k := range ValidLogLevels {
		levels = append(levels, k)
	}
	return ValidateOneOf(levels, allowEmpty)
}

// ValidateColor returns a validation function that checks whether a color
// name is in the provided set of valid color names.
func ValidateColor(validColors []string) func(interface{}) error {
	return ValidateOneOf(validColors, false)
}
