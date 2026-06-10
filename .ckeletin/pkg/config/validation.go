// .ckeletin/pkg/config/validation.go
//
// Config-time validation for registered options.
//
// This file provides ValidateRegisteredOptions() and
// ValidateRegisteredOptionsWithViper(), which iterate over all registered
// ConfigOption entries that have a Validation function and run them against
// the Viper values (global instance or a provided one). This catches invalid
// user-facing values (colors, log levels, etc.) at config load time rather
// than during command execution.

package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ValidateRegisteredOptionsWithViper validates all registered config options
// that have a Validation function against the provided Viper instance. It
// returns a slice of errors for any values that fail validation. An empty
// slice means all validations passed.
func ValidateRegisteredOptionsWithViper(v *viper.Viper) []error {
	var errs []error

	for _, opt := range Registry() {
		if opt.Validation == nil {
			continue
		}

		value := v.Get(opt.Key)
		if err := opt.Validation(value); err != nil {
			errs = append(errs, fmt.Errorf("config %q: %w", opt.Key, err))
		}
	}

	return errs
}

// ValidateRegisteredOptions validates all registered config options that have
// a Validation function against the global Viper instance. It returns a slice
// of errors for any values that fail validation. An empty slice means all
// validations passed.
func ValidateRegisteredOptions() []error {
	return ValidateRegisteredOptionsWithViper(viper.GetViper())
}
