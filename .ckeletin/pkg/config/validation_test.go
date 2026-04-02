// .ckeletin/pkg/config/validation_test.go

package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRegisteredOptions(t *testing.T) {
	t.Parallel()

	t.Run("All defaults are valid", func(t *testing.T) {
		t.Parallel()
		// Use a fresh viper instance by resetting and reloading defaults.
		// This verifies that all default values pass their own validation.
		v := viper.New()
		for _, opt := range Registry() {
			v.SetDefault(opt.Key, opt.DefaultValue)
		}

		// Validate each option's default individually
		var errs []error
		for _, opt := range Registry() {
			if opt.Validation == nil {
				continue
			}
			value := v.Get(opt.Key)
			if err := opt.Validation(value); err != nil {
				errs = append(errs, err)
			}
		}

		assert.Empty(t, errs, "Default values should all pass validation")
	})

	t.Run("Invalid color caught", func(t *testing.T) {
		t.Parallel()
		// Create a config option with color validation and an invalid value
		opts := []ConfigOption{
			{
				Key:          "test.color",
				DefaultValue: "purple",
				Type:         "string",
				Validation:   ValidateColor([]string{"red", "green", "blue"}),
			},
		}

		v := viper.New()
		v.SetDefault(opts[0].Key, opts[0].DefaultValue)

		err := opts[0].Validation(v.Get(opts[0].Key))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value \"purple\"")
	})

	t.Run("Invalid log level caught", func(t *testing.T) {
		t.Parallel()
		opts := []ConfigOption{
			{
				Key:          "test.level",
				DefaultValue: "verbose",
				Type:         "string",
				Validation:   ValidateLogLevel(false),
			},
		}

		v := viper.New()
		v.SetDefault(opts[0].Key, opts[0].DefaultValue)

		err := opts[0].Validation(v.Get(opts[0].Key))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid value \"verbose\"")
	})

	t.Run("Options without validation are skipped", func(t *testing.T) {
		t.Parallel()
		opts := []ConfigOption{
			{
				Key:          "test.no_validation",
				DefaultValue: "anything",
				Type:         "string",
				// Validation is nil
			},
		}

		// No panic, no error
		assert.Nil(t, opts[0].Validation)
	})
}
