// internal/config/commands/ping_config_test.go

package commands

import (
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPingMetadata(t *testing.T) {
	t.Run("Required fields populated", func(t *testing.T) {
		assert.NotEmpty(t, PingMetadata.Use, "PingMetadata.Use is empty")
		assert.NotEmpty(t, PingMetadata.Short, "PingMetadata.Short is empty")
		assert.NotEmpty(t, PingMetadata.Long, "PingMetadata.Long is empty")
		assert.NotEmpty(t, PingMetadata.ConfigPrefix, "PingMetadata.ConfigPrefix is empty")
	})

	t.Run("Use command name matches convention", func(t *testing.T) {
		assert.Equal(t, "ping", PingMetadata.Use)
	})

	t.Run("ConfigPrefix matches expected pattern", func(t *testing.T) {
		assert.Equal(t, "app.ping", PingMetadata.ConfigPrefix)
	})

	t.Run("Examples are valid", func(t *testing.T) {
		assert.NotEmpty(t, PingMetadata.Examples, "PingMetadata.Examples is empty")
		for i, example := range PingMetadata.Examples {
			assert.NotEmpty(t, example, "PingMetadata.Examples[%d] is empty", i)
			assert.True(t, strings.HasPrefix(example, "ping"),
				"PingMetadata.Examples[%d] = %q, should start with 'ping'", i, example)
		}
	})

	t.Run("FlagOverrides reference valid config keys", func(t *testing.T) {
		opts := PingOptions()
		configKeys := make(map[string]bool)
		for _, opt := range opts {
			configKeys[opt.Key] = true
		}

		for configKey, flagName := range PingMetadata.FlagOverrides {
			assert.True(t, configKeys[configKey], "FlagOverride key %q not found in PingOptions", configKey)
			assert.NotEmpty(t, flagName, "FlagOverride for %q has empty flag name", configKey)
			assert.False(t, strings.Contains(flagName, "_"),
				"Flag name %q should use kebab-case, not snake_case", flagName)
		}
	})
}

func TestPingOptions(t *testing.T) {
	opts := PingOptions()

	t.Run("Returns non-empty options", func(t *testing.T) {
		require.NotEmpty(t, opts, "PingOptions() returned empty slice")
	})

	t.Run("All options have app.ping prefix", func(t *testing.T) {
		prefix := "app.ping."
		for i, opt := range opts {
			assert.True(t, strings.HasPrefix(opt.Key, prefix),
				"Option[%d].Key = %q, should start with %q", i, opt.Key, prefix)
		}
	})

	t.Run("All required fields populated", func(t *testing.T) {
		for i, opt := range opts {
			assert.NotEmpty(t, opt.Key, "Option[%d].Key is empty", i)
			assert.NotEmpty(t, opt.Description, "Option[%d].Description is empty for key %q", i, opt.Key)
			assert.NotEmpty(t, opt.Type, "Option[%d].Type is empty for key %q", i, opt.Key)
			// DefaultValue can be nil/empty, but let's check it exists
			// (even if it's the zero value)
		}
	})

	t.Run("Types are valid", func(t *testing.T) {
		validTypes := map[string]bool{
			"string":   true,
			"bool":     true,
			"int":      true,
			"float":    true,
			"[]string": true,
		}

		for i, opt := range opts {
			assert.True(t, validTypes[opt.Type],
				"Option[%d] (%s) has invalid type %q", i, opt.Key, opt.Type)
		}
	})

	t.Run("Specific option: output_message", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.ping.output_message" {
				found = &opts[i]
				break
			}
		}

		require.NotNil(t, found, "app.ping.output_message not found in options")
		assert.Equal(t, "string", found.Type, "output_message.Type mismatch")
		assert.Equal(t, "Pong", found.DefaultValue, "output_message.DefaultValue mismatch")
		assert.False(t, found.Required, "output_message should not be required")
	})

	t.Run("Specific option: output_color", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.ping.output_color" {
				found = &opts[i]
				break
			}
		}

		require.NotNil(t, found, "app.ping.output_color not found in options")
		assert.Equal(t, "string", found.Type, "output_color.Type mismatch")
		assert.Equal(t, "white", found.DefaultValue, "output_color.DefaultValue mismatch")
	})

	t.Run("Specific option: ui", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.ping.ui" {
				found = &opts[i]
				break
			}
		}

		require.NotNil(t, found, "app.ping.ui not found in options")
		assert.Equal(t, "bool", found.Type, "ui.Type mismatch")
		assert.Equal(t, false, found.DefaultValue, "ui.DefaultValue mismatch")
	})
}

func TestPingOptionsRegistered(t *testing.T) {
	t.Run("Options are registered in global registry", func(t *testing.T) {
		// Get all options from the registry
		allOpts := config.Registry()

		// Check that ping options are present
		pingKeys := map[string]bool{
			"app.ping.output_message": false,
			"app.ping.output_color":   false,
			"app.ping.ui":             false,
		}

		for _, opt := range allOpts {
			if _, exists := pingKeys[opt.Key]; exists {
				pingKeys[opt.Key] = true
			}
		}

		// Verify all ping keys were found
		for key, found := range pingKeys {
			assert.True(t, found, "Config key %q not found in registry", key)
		}
	})
}
