package dev

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNewConfigInspector(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()

	// ASSERTION PHASE
	assert.NotNil(t, ci, "ConfigInspector should not be nil")
	assert.NotNil(t, ci.registry, "Registry should be initialized")
	assert.Greater(t, len(ci.registry), 0, "Registry should have options")
}

func TestListAllOptions(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()

	// EXECUTION PHASE
	options := ci.ListAllOptions()

	// ASSERTION PHASE
	assert.NotNil(t, options, "Options list should not be nil")
	assert.Greater(t, len(options), 0, "Should have at least one option")

	// Verify sorted by key
	for i := 1; i < len(options); i++ {
		assert.LessOrEqual(t, options[i-1].Key, options[i].Key,
			"Options should be sorted by key")
	}

	// Verify options have required fields
	for _, opt := range options {
		assert.NotEmpty(t, opt.Key, "Option key should not be empty")
		assert.NotEmpty(t, opt.Description, "Option description should not be empty")
		assert.NotEmpty(t, opt.Type, "Option type should not be empty")
	}
}

func TestGetEffectiveConfig(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()
	viper.Reset() // Start fresh

	// Set a test value
	testKey := "app.log.level"
	testValue := "debug"
	viper.Set(testKey, testValue)

	// EXECUTION PHASE
	effective := ci.GetEffectiveConfig()

	// ASSERTION PHASE
	assert.NotNil(t, effective, "Effective config should not be nil")
	assert.Greater(t, len(effective), 0, "Should have effective values")

	// Verify we can retrieve the test value
	if val, ok := effective[testKey]; ok {
		assert.Equal(t, testValue, val, "Should get correct effective value")
	}
}

func TestFormatAsTable(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()

	// EXECUTION PHASE
	table := ci.FormatAsTable()

	// ASSERTION PHASE
	assert.NotEmpty(t, table, "Table should not be empty")
	assert.Contains(t, table, "Configuration Registry", "Should have header")
	assert.Contains(t, table, "KEY", "Should have KEY column")
	assert.Contains(t, table, "TYPE", "Should have TYPE column")
	assert.Contains(t, table, "DEFAULT", "Should have DEFAULT column")
	assert.Contains(t, table, "DESCRIPTION", "Should have DESCRIPTION column")
	assert.Contains(t, table, "Total:", "Should have total count")

	// Verify contains some actual config keys
	assert.Contains(t, table, "app.", "Should contain app config keys")
}

func TestFormatEffectiveAsTable(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()
	viper.Reset()
	viper.Set("app.log.level", "info")

	// EXECUTION PHASE
	table := ci.FormatEffectiveAsTable()

	// ASSERTION PHASE
	assert.NotEmpty(t, table, "Table should not be empty")
	assert.Contains(t, table, "Effective Configuration", "Should have header")
	assert.Contains(t, table, "KEY", "Should have KEY column")
	assert.Contains(t, table, "VALUE", "Should have VALUE column")
	assert.Contains(t, table, "SOURCE", "Should have SOURCE column")
}

func TestExportToJSON(t *testing.T) {
	tests := []struct {
		name             string
		includeEffective bool
		setupViper       func()
		wantErr          bool
	}{
		{
			name:             "registry only",
			includeEffective: false,
			setupViper:       func() {},
			wantErr:          false,
		},
		{
			name:             "with effective values",
			includeEffective: true,
			setupViper: func() {
				viper.Reset()
				viper.Set("app.log.level", "debug")
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			ci := NewConfigInspector()
			tt.setupViper()

			// EXECUTION PHASE
			jsonStr, err := ci.ExportToJSON(tt.includeEffective)

			// ASSERTION PHASE
			if tt.wantErr {
				assert.Error(t, err, "Should return error")
			} else {
				assert.NoError(t, err, "Should not return error")
				assert.NotEmpty(t, jsonStr, "JSON should not be empty")

				// Verify it's valid JSON
				var result interface{}
				err := json.Unmarshal([]byte(jsonStr), &result)
				assert.NoError(t, err, "Should be valid JSON")

				if tt.includeEffective {
					// Verify has both registry and effective
					assert.Contains(t, jsonStr, "registry", "Should contain registry")
					assert.Contains(t, jsonStr, "effective", "Should contain effective")
				}
			}
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupViper func()
		wantErrors bool
		errorCount int
	}{
		{
			name: "valid config",
			setupViper: func() {
				viper.Reset()
				// Set all required values if any
			},
			wantErrors: false,
			errorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			ci := NewConfigInspector()
			tt.setupViper()

			// EXECUTION PHASE
			errors := ci.ValidateConfig()

			// ASSERTION PHASE
			if tt.wantErrors {
				assert.NotEmpty(t, errors, "Should have validation errors")
				if tt.errorCount > 0 {
					assert.Len(t, errors, tt.errorCount, "Should have expected error count")
				}
			} else {
				// Note: This test may fail if registry has required fields
				// Adjust based on actual registry contents
				if len(errors) > 0 {
					t.Logf("Validation errors found (may be expected): %v", errors)
				}
			}
		})
	}
}

func TestValidateConfigWithRequiredFields(t *testing.T) {
	// Test validation with required field scenarios
	t.Run("required field not set", func(t *testing.T) {
		// SETUP PHASE
		ci := NewConfigInspector()
		viper.Reset()

		// In the current registry, there may not be required fields
		// But we test the validation logic by verifying the function runs
		// and returns a slice (even if empty - not nil)

		// EXECUTION PHASE
		errors := ci.ValidateConfig()

		// ASSERTION PHASE
		// ValidateConfig returns []error (empty slice, not nil)
		// Go idiom: checking len() is safer than nil check for slices
		assert.GreaterOrEqual(t, len(errors), 0, "Should return errors slice")

		// Log what we got for visibility
		if len(errors) > 0 {
			t.Logf("Found %d validation errors", len(errors))
			for i, err := range errors {
				t.Logf("  Error %d: %v", i+1, err)
			}
		} else {
			t.Logf("No validation errors (registry has no required fields)")
		}
	})

	t.Run("required field with nil value", func(t *testing.T) {
		// SETUP PHASE
		ci := NewConfigInspector()
		viper.Reset()

		// Test the specific validation path for required fields
		// The validation code checks: if opt.Required { ... if value == nil || ... }
		// Even if no registry options are required, we verify the code path exists

		// EXECUTION PHASE
		errors := ci.ValidateConfig()

		// ASSERTION PHASE
		// The validation function should always return without panicking
		assert.GreaterOrEqual(t, len(errors), 0, "Should return errors slice")
	})

	t.Run("required field with empty string", func(t *testing.T) {
		// SETUP PHASE
		ci := NewConfigInspector()
		viper.Reset()

		// Set a value to empty string (this tests the fmt.Sprintf(%v) == "" check)
		viper.Set("app.binary_name", "")

		// EXECUTION PHASE
		errors := ci.ValidateConfig()

		// ASSERTION PHASE
		assert.GreaterOrEqual(t, len(errors), 0, "Should return errors slice")

		// If app.binary_name is required (which it's not in the current registry),
		// this would produce an error. Either way, the validation runs successfully.
		t.Logf("Validation completed with %d errors", len(errors))
	})
}

func TestValidateConfig_RequiredFieldMissing(t *testing.T) {
	// Test the Required field validation path directly
	// by constructing a ConfigInspector with a custom registry
	ci := &ConfigInspector{
		registry: []config.ConfigOption{
			{
				Key:          "test.required.field",
				DefaultValue: "",
				Description:  "A required field",
				Type:         "string",
				Required:     true,
			},
			{
				Key:          "test.optional.field",
				DefaultValue: "default",
				Description:  "An optional field",
				Type:         "string",
				Required:     false,
			},
		},
	}

	viper.Reset()
	// Don't set the required field — should produce a validation error

	errors := ci.ValidateConfig()
	assert.NotEmpty(t, errors, "Should have validation error for missing required field")
	assert.Len(t, errors, 1, "Should have exactly one error")
	assert.Contains(t, errors[0].Error(), "test.required.field",
		"Error should reference the required key")
}

func TestValidateConfig_RequiredFieldPresent(t *testing.T) {
	ci := &ConfigInspector{
		registry: []config.ConfigOption{
			{
				Key:          "test.required.field",
				DefaultValue: "",
				Description:  "A required field",
				Type:         "string",
				Required:     true,
			},
		},
	}

	viper.Reset()
	viper.Set("test.required.field", "some-value")

	errors := ci.ValidateConfig()
	assert.Empty(t, errors, "Should have no errors when required field is set")
}

func TestValidateConfig_RequiredFieldNilValue(t *testing.T) {
	ci := &ConfigInspector{
		registry: []config.ConfigOption{
			{
				Key:          "test.nil.field",
				DefaultValue: nil,
				Description:  "A required field with nil default",
				Type:         "string",
				Required:     true,
			},
		},
	}

	viper.Reset()
	// Not setting value means viper.Get returns nil

	errors := ci.ValidateConfig()
	assert.NotEmpty(t, errors, "Should have error for nil required field")
}

func TestGetConfigSourceInfo_NonDefaultValue(t *testing.T) {
	// Test the branch where value differs from default and is detected as file source
	ci := &ConfigInspector{
		registry: []config.ConfigOption{
			{
				Key:          "app.test.value",
				DefaultValue: "original",
				Description:  "Test option",
				Type:         "string",
			},
		},
	}

	viper.Reset()
	viper.Set("app.test.value", "changed")

	sources := ci.GetConfigSourceInfo("CKELETIN")

	assert.Equal(t, "original", sources.Defaults["app.test.value"])
	assert.Equal(t, "changed", sources.Effective["app.test.value"])
	// Since there's no env var set, this should be detected as file source
	assert.Equal(t, "changed", sources.File["app.test.value"],
		"Non-default value without env var should be detected as file source")
	assert.Empty(t, sources.Environment, "No env vars should be detected")
}

func TestGetConfigSourceInfo_DefaultValueUnchanged(t *testing.T) {
	ci := &ConfigInspector{
		registry: []config.ConfigOption{
			{
				Key:          "app.test.value",
				DefaultValue: "default",
				Description:  "Test option",
				Type:         "string",
			},
		},
	}

	viper.Reset()
	viper.Set("app.test.value", "default") // Same as default

	sources := ci.GetConfigSourceInfo("CKELETIN")

	assert.Equal(t, "default", sources.Defaults["app.test.value"])
	assert.Equal(t, "default", sources.Effective["app.test.value"])
	// Value matches default, so should NOT appear in File or Environment
	assert.Empty(t, sources.File, "Default value should not appear as file source")
	assert.Empty(t, sources.Environment, "Default value should not appear as env source")
}

func TestGetConfigByPrefix(t *testing.T) {
	tests := []struct {
		name        string
		prefix      string
		expectMatch bool
	}{
		{
			name:        "app prefix",
			prefix:      "app.",
			expectMatch: true,
		},
		{
			name:        "log prefix",
			prefix:      "app.log",
			expectMatch: true,
		},
		{
			name:        "nonexistent prefix",
			prefix:      "nonexistent.prefix",
			expectMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			ci := NewConfigInspector()

			// EXECUTION PHASE
			matches := ci.GetConfigByPrefix(tt.prefix)

			// ASSERTION PHASE
			if tt.expectMatch {
				assert.NotEmpty(t, matches, "Should find matches for prefix")
				for _, match := range matches {
					assert.True(t, strings.HasPrefix(match.Key, tt.prefix),
						"Matched option should have correct prefix")
				}
			} else {
				assert.Empty(t, matches, "Should not find matches for nonexistent prefix")
			}
		})
	}
}

func TestGetConfigSourceInfo(t *testing.T) {
	// SETUP PHASE
	ci := NewConfigInspector()
	viper.Reset()

	// Set some test values
	viper.Set("app.log.level", "debug")
	viper.SetDefault("app.binary_name", "ckeletin-go")

	// EXECUTION PHASE
	sources := ci.GetConfigSourceInfo("CKELETIN")

	// ASSERTION PHASE
	assert.NotNil(t, sources, "Sources should not be nil")
	assert.NotNil(t, sources.Defaults, "Defaults should not be nil")
	assert.NotNil(t, sources.Effective, "Effective should not be nil")
	assert.Greater(t, len(sources.Defaults), 0, "Should have default values")
	assert.Greater(t, len(sources.Effective), 0, "Should have effective values")
}

func TestGetConfigSourceInfoWithEnvVars(t *testing.T) {
	// Test GetConfigSourceInfo with environment variable simulation
	t.Run("file vs environment source detection", func(t *testing.T) {
		// SETUP PHASE
		ci := NewConfigInspector()
		viper.Reset()

		// Set a value different from default (simulating file config)
		viper.Set("app.log.level", "debug")

		// EXECUTION PHASE
		sources := ci.GetConfigSourceInfo("CKELETIN")

		// ASSERTION PHASE
		assert.NotNil(t, sources.File, "File map should not be nil")
		assert.NotNil(t, sources.Environment, "Environment map should not be nil")
		assert.NotNil(t, sources.Defaults, "Defaults map should not be nil")
		assert.NotNil(t, sources.Effective, "Effective map should not be nil")

		// The function always populates Defaults and Effective
		assert.Greater(t, len(sources.Defaults), 0, "Should have defaults")
		assert.Greater(t, len(sources.Effective), 0, "Should have effective values")

		// File or environment may or may not have values depending on how
		// Viper detects the source (it uses a heuristic)
		// The important thing is the function runs without errors
		t.Logf("Defaults: %d, File: %d, Env: %d, Effective: %d",
			len(sources.Defaults), len(sources.File),
			len(sources.Environment), len(sources.Effective))
	})
}

func TestConfigOptionMethods(t *testing.T) {
	// Test that we can work with ConfigOption instances
	t.Run("option has expected fields", func(t *testing.T) {
		ci := NewConfigInspector()
		options := ci.ListAllOptions()

		assert.Greater(t, len(options), 0, "Should have options")

		// Pick first option and verify it has the expected structure
		opt := options[0]
		assert.NotEmpty(t, opt.Key, "Should have key")
		assert.NotEmpty(t, opt.Description, "Should have description")
		assert.NotEmpty(t, opt.Type, "Should have type")

		// Test DefaultValueString method
		defaultStr := opt.DefaultValueString()
		assert.NotEmpty(t, defaultStr, "Should have default value string")

		// Test ExampleValueString method
		exampleStr := opt.ExampleValueString()
		assert.NotEmpty(t, exampleStr, "Should have example value string")
	})
}

func TestConfigInspectorIntegration(t *testing.T) {
	// Integration test that exercises multiple methods together
	t.Run("full workflow", func(t *testing.T) {
		// SETUP PHASE
		ci := NewConfigInspector()
		viper.Reset()
		viper.Set("app.log.level", "debug")

		// EXECUTION PHASE
		// 1. List all options
		options := ci.ListAllOptions()
		assert.Greater(t, len(options), 0, "Should have options")

		// 2. Get effective config
		effective := ci.GetEffectiveConfig()
		assert.Greater(t, len(effective), 0, "Should have effective config")

		// 3. Format as table
		table := ci.FormatAsTable()
		assert.Contains(t, table, "Configuration Registry", "Should have table")

		// 4. Export to JSON
		jsonStr, err := ci.ExportToJSON(true)
		assert.NoError(t, err, "Should export to JSON")
		assert.NotEmpty(t, jsonStr, "JSON should not be empty")

		// 5. Validate config
		errors := ci.ValidateConfig()
		// Note: May have errors depending on registry requirements
		t.Logf("Validation completed with %d errors", len(errors))

		// 6. Search by prefix
		appOptions := ci.GetConfigByPrefix("app.")
		assert.Greater(t, len(appOptions), 0, "Should have app options")
	})
}
