// internal/config/commands/docs_config_test.go

package commands

import (
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDocsConfigMetadata(t *testing.T) {
	t.Run("Required fields populated", func(t *testing.T) {
		assert.NotEmpty(t, DocsConfigMetadata.Use, "DocsConfigMetadata.Use is empty")
		assert.NotEmpty(t, DocsConfigMetadata.Short, "DocsConfigMetadata.Short is empty")
		assert.NotEmpty(t, DocsConfigMetadata.Long, "DocsConfigMetadata.Long is empty")
		assert.NotEmpty(t, DocsConfigMetadata.ConfigPrefix, "DocsConfigMetadata.ConfigPrefix is empty")
	})

	t.Run("Use command name matches convention", func(t *testing.T) {
		assert.Equal(t, "config", DocsConfigMetadata.Use)
	})

	t.Run("ConfigPrefix matches expected pattern", func(t *testing.T) {
		assert.Equal(t, "app.docs", DocsConfigMetadata.ConfigPrefix)
	})

	t.Run("Examples are valid", func(t *testing.T) {
		assert.NotEmpty(t, DocsConfigMetadata.Examples, "DocsConfigMetadata.Examples is empty")
		for i, example := range DocsConfigMetadata.Examples {
			assert.NotEmpty(t, example, "DocsConfigMetadata.Examples[%d] is empty", i)
			// Examples should start with "docs" command
			assert.True(t, strings.HasPrefix(example, "docs"),
				"DocsConfigMetadata.Examples[%d] = %q, should start with 'docs'", i, example)
		}
	})

	t.Run("FlagOverrides reference valid config keys", func(t *testing.T) {
		opts := DocsOptions()
		configKeys := make(map[string]bool)
		for _, opt := range opts {
			configKeys[opt.Key] = true
		}

		for configKey, flagName := range DocsConfigMetadata.FlagOverrides {
			assert.True(t, configKeys[configKey], "FlagOverride key %q not found in DocsOptions", configKey)
			assert.NotEmpty(t, flagName, "FlagOverride for %q has empty flag name", configKey)
			assert.False(t, strings.Contains(flagName, "_"),
				"Flag name %q should use kebab-case, not snake_case", flagName)
		}
	})
}

func TestDocsOptions(t *testing.T) {
	opts := DocsOptions()

	t.Run("Returns non-empty options", func(t *testing.T) {
		require.NotEmpty(t, opts, "DocsOptions() returned empty slice")
	})

	t.Run("All options have app.docs prefix", func(t *testing.T) {
		prefix := "app.docs."
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

	t.Run("Specific option: output_format", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.docs.output_format" {
				found = &opts[i]
				break
			}
		}

		require.NotNil(t, found, "app.docs.output_format not found in options")
		assert.Equal(t, "string", found.Type, "output_format.Type mismatch")
		assert.Equal(t, "markdown", found.DefaultValue, "output_format.DefaultValue mismatch")
		assert.False(t, found.Required, "output_format should not be required")
	})

	t.Run("Specific option: output_file", func(t *testing.T) {
		var found *config.ConfigOption
		for i := range opts {
			if opts[i].Key == "app.docs.output_file" {
				found = &opts[i]
				break
			}
		}

		require.NotNil(t, found, "app.docs.output_file not found in options")
		assert.Equal(t, "string", found.Type, "output_file.Type mismatch")
		assert.Equal(t, "", found.DefaultValue, "output_file.DefaultValue mismatch")
		assert.False(t, found.Required, "output_file should not be required")
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
			assert.True(t, found, "Config key %q not found in registry", key)
		}
	})
}
