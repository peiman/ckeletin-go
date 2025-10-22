// internal/config/commands/docs_config_test.go

package commands

import (
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
)

func TestDocsConfigMetadata(t *testing.T) {
	t.Run("Required fields populated", func(t *testing.T) {
		if DocsConfigMetadata.Use == "" {
			t.Error("DocsConfigMetadata.Use is empty")
		}
		if DocsConfigMetadata.Short == "" {
			t.Error("DocsConfigMetadata.Short is empty")
		}
		if DocsConfigMetadata.Long == "" {
			t.Error("DocsConfigMetadata.Long is empty")
		}
		if DocsConfigMetadata.ConfigPrefix == "" {
			t.Error("DocsConfigMetadata.ConfigPrefix is empty")
		}
	})

	t.Run("Use command name matches convention", func(t *testing.T) {
		expected := "config"
		if DocsConfigMetadata.Use != expected {
			t.Errorf("DocsConfigMetadata.Use = %q, want %q", DocsConfigMetadata.Use, expected)
		}
	})

	t.Run("ConfigPrefix matches expected pattern", func(t *testing.T) {
		expected := "app.docs"
		if DocsConfigMetadata.ConfigPrefix != expected {
			t.Errorf("DocsConfigMetadata.ConfigPrefix = %q, want %q", DocsConfigMetadata.ConfigPrefix, expected)
		}
	})

	t.Run("Examples are valid", func(t *testing.T) {
		if len(DocsConfigMetadata.Examples) == 0 {
			t.Error("DocsConfigMetadata.Examples is empty")
		}
		for i, example := range DocsConfigMetadata.Examples {
			if example == "" {
				t.Errorf("DocsConfigMetadata.Examples[%d] is empty", i)
			}
			// Examples should start with "docs" command
			if !strings.HasPrefix(example, "docs") {
				t.Errorf("DocsConfigMetadata.Examples[%d] = %q, should start with 'docs'", i, example)
			}
		}
	})

	t.Run("FlagOverrides reference valid config keys", func(t *testing.T) {
		opts := DocsOptions()
		configKeys := make(map[string]bool)
		for _, opt := range opts {
			configKeys[opt.Key] = true
		}

		for configKey, flagName := range DocsConfigMetadata.FlagOverrides {
			// Verify the config key exists in DocsOptions
			if !configKeys[configKey] {
				t.Errorf("FlagOverride key %q not found in DocsOptions", configKey)
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

func TestDocsOptions(t *testing.T) {
	opts := DocsOptions()

	t.Run("Returns non-empty options", func(t *testing.T) {
		if len(opts) == 0 {
			t.Fatal("DocsOptions() returned empty slice")
		}
	})

	t.Run("All options have app.docs prefix", func(t *testing.T) {
		prefix := "app.docs."
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

	t.Run("Specific option: output_format", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.docs.output_format" {
				found = &opts[i]
				break
			}
		}

		if found == nil {
			t.Fatal("app.docs.output_format not found in options")
		}

		if found.Type != "string" {
			t.Errorf("output_format.Type = %q, want %q", found.Type, "string")
		}
		if found.DefaultValue != "markdown" {
			t.Errorf("output_format.DefaultValue = %q, want %q", found.DefaultValue, "markdown")
		}
		if found.Required {
			t.Error("output_format should not be required")
		}
	})

	t.Run("Specific option: output_file", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.docs.output_file" {
				found = &opts[i]
				break
			}
		}

		if found == nil {
			t.Fatal("app.docs.output_file not found in options")
		}

		if found.Type != "string" {
			t.Errorf("output_file.Type = %q, want %q", found.Type, "string")
		}
		if found.DefaultValue != "" {
			t.Errorf("output_file.DefaultValue = %q, want empty string", found.DefaultValue)
		}
		if found.Required {
			t.Error("output_file should not be required")
		}
	})
}

func TestDocsOptionsRegistered(t *testing.T) {
	t.Run("Options are registered in global registry", func(t *testing.T) {
		// Get all options from the registry
		allOpts := config.Registry()

		// Check that docs options are present
		docsKeys := map[string]bool{
			"app.docs.output_format": false,
			"app.docs.output_file":   false,
		}

		for _, opt := range allOpts {
			if _, exists := docsKeys[opt.Key]; exists {
				docsKeys[opt.Key] = true
			}
		}

		// Verify all docs keys were found
		for key, found := range docsKeys {
			if !found {
				t.Errorf("Config key %q not found in registry", key)
			}
		}
	})
}
