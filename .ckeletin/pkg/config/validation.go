// .ckeletin/pkg/config/validation.go
//
// Config-time validation for registered options.
//
// This file provides ValidateRegisteredOptions(), which iterates over all
// registered ConfigOption entries that have a Validation function and runs
// them against the current Viper values. This catches invalid user-facing
// values (colors, log levels, etc.) at config load time rather than during
// command execution.

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ValidateRegisteredOptions validates all registered config options that have
// a Validation function. It returns a slice of errors for any values that fail
// validation. An empty slice means all validations passed.
func ValidateRegisteredOptions() []error {
	var errs []error

	for _, opt := range Registry() {
		if opt.Validation == nil {
			continue
		}

		value := viper.Get(opt.Key)
		if err := opt.Validation(value); err != nil {
			errs = append(errs, fmt.Errorf("config %q: %w", opt.Key, err))
		}
	}

	return errs
}
