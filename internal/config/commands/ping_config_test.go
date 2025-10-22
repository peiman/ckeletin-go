// internal/config/commands/ping_config_test.go

package commands

import (
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
)

func TestPingMetadata(t *testing.T) {
	t.Run("Required fields populated", func(t *testing.T) {
		if PingMetadata.Use == "" {
			t.Error("PingMetadata.Use is empty")
		}
		if PingMetadata.Short == "" {
			t.Error("PingMetadata.Short is empty")
		}
		if PingMetadata.Long == "" {
			t.Error("PingMetadata.Long is empty")
		}
		if PingMetadata.ConfigPrefix == "" {
			t.Error("PingMetadata.ConfigPrefix is empty")
		}
	})

	t.Run("Use command name matches convention", func(t *testing.T) {
		expected := "ping"
		if PingMetadata.Use != expected {
			t.Errorf("PingMetadata.Use = %q, want %q", PingMetadata.Use, expected)
		}
	})

	t.Run("ConfigPrefix matches expected pattern", func(t *testing.T) {
		expected := "app.ping"
		if PingMetadata.ConfigPrefix != expected {
			t.Errorf("PingMetadata.ConfigPrefix = %q, want %q", PingMetadata.ConfigPrefix, expected)
		}
	})

	t.Run("Examples are valid", func(t *testing.T) {
		if len(PingMetadata.Examples) == 0 {
			t.Error("PingMetadata.Examples is empty")
		}
		for i, example := range PingMetadata.Examples {
			if example == "" {
				t.Errorf("PingMetadata.Examples[%d] is empty", i)
			}
			// Examples should start with command name
			if !strings.HasPrefix(example, "ping") {
				t.Errorf("PingMetadata.Examples[%d] = %q, should start with 'ping'", i, example)
			}
		}
	})

	t.Run("FlagOverrides reference valid config keys", func(t *testing.T) {
		opts := PingOptions()
		configKeys := make(map[string]bool)
		for _, opt := range opts {
			configKeys[opt.Key] = true
		}

		for configKey, flagName := range PingMetadata.FlagOverrides {
			// Verify the config key exists in PingOptions
			if !configKeys[configKey] {
				t.Errorf("FlagOverride key %q not found in PingOptions", configKey)
			}

			// Verify flag name is not empty
			if flagName == "" {
				t.Errorf("FlagOverride for %q has empty flag name", configKey)
			}

			// Verify flag name uses kebab-case convention
			if strings.Contains(flagName, "_") {
				t.Errorf("Flag name %q should use kebab-case, not snake_case", flagName)
			}
		}
	})
}

func TestPingOptions(t *testing.T) {
	opts := PingOptions()

	t.Run("Returns non-empty options", func(t *testing.T) {
		if len(opts) == 0 {
			t.Fatal("PingOptions() returned empty slice")
		}
	})

	t.Run("All options have app.ping prefix", func(t *testing.T) {
		prefix := "app.ping."
		for i, opt := range opts {
			if !strings.HasPrefix(opt.Key, prefix) {
				t.Errorf("Option[%d].Key = %q, should start with %q", i, opt.Key, prefix)
			}
		}
	})

	t.Run("All required fields populated", func(t *testing.T) {
		for i, opt := range opts {
			if opt.Key == "" {
				t.Errorf("Option[%d].Key is empty", i)
			}
			if opt.Description == "" {
				t.Errorf("Option[%d].Description is empty for key %q", i, opt.Key)
			}
			if opt.Type == "" {
				t.Errorf("Option[%d].Type is empty for key %q", i, opt.Key)
			}
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
			if !validTypes[opt.Type] {
				t.Errorf("Option[%d] (%s) has invalid type %q", i, opt.Key, opt.Type)
			}
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

		if found == nil {
			t.Fatal("app.ping.output_message not found in options")
		}

		if found.Type != "string" {
			t.Errorf("output_message.Type = %q, want %q", found.Type, "string")
		}
		if found.DefaultValue != "Pong" {
			t.Errorf("output_message.DefaultValue = %q, want %q", found.DefaultValue, "Pong")
		}
		if found.Required {
			t.Error("output_message should not be required")
		}
	})

	t.Run("Specific option: output_color", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.ping.output_color" {
				found = &opts[i]
				break
			}
		}

		if found == nil {
			t.Fatal("app.ping.output_color not found in options")
		}

		if found.Type != "string" {
			t.Errorf("output_color.Type = %q, want %q", found.Type, "string")
		}
		if found.DefaultValue != "white" {
			t.Errorf("output_color.DefaultValue = %q, want %q", found.DefaultValue, "white")
		}
	})

	t.Run("Specific option: ui", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.ping.ui" {
				found = &opts[i]
				break
			}
		}

		if found == nil {
			t.Fatal("app.ping.ui not found in options")
		}

		if found.Type != "bool" {
			t.Errorf("ui.Type = %q, want %q", found.Type, "bool")
		}
		if found.DefaultValue != false {
			t.Errorf("ui.DefaultValue = %v, want %v", found.DefaultValue, false)
		}
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
			if !found {
				t.Errorf("Config key %q not found in registry", key)
			}
		}
	})
}
