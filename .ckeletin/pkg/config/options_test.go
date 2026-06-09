// internal/config/options_test.go

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionEnvVarName(t *testing.T) {
	tests := []struct {
		name   string
		opt    ConfigOption
		prefix string
		want   string
	}{
		{
			name: "Simple key",
			opt: ConfigOption{
				Key: "simple",
			},
			prefix: "APP",
			want:   "APP_SIMPLE",
		},
		{
			name: "Nested key",
			opt: ConfigOption{
				Key: "app.service.option",
			},
			prefix: "MYAPP",
			want:   "MYAPP_APP_SERVICE_OPTION",
		},
		{
			name: "Empty prefix",
			opt: ConfigOption{
				Key: "app.key",
			},
			prefix: "",
			want:   "_APP_KEY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.EnvVarName(tt.prefix)

			// ASSERTION PHASE
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOptionDefaultValueString(t *testing.T) {
	tests := []struct {
		name string
		opt  ConfigOption
		want string
	}{
		{
			name: "String value",
			opt: ConfigOption{
				DefaultValue: "test",
			},
			want: "test",
		},
		{
			name: "Integer value",
			opt: ConfigOption{
				DefaultValue: 42,
			},
			want: "42",
		},
		{
			name: "Boolean value",
			opt: ConfigOption{
				DefaultValue: true,
			},
			want: "true",
		},
		{
			name: "Nil value",
			opt: ConfigOption{
				DefaultValue: nil,
			},
			want: "nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.DefaultValueString()

			// ASSERTION PHASE
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCoreOptionsValidators(t *testing.T) {
	// SETUP PHASE
	// Index core options by key for lookup
	byKey := make(map[string]ConfigOption)
	for _, opt := range CoreOptions() {
		byKey[opt.Key] = opt
	}

	tests := []struct {
		name          string
		key           string
		validValues   []interface{}
		invalidValues []interface{}
	}{
		{
			name:          "color_enabled accepts auto/true/false",
			key:           KeyAppLogColorEnabled,
			validValues:   []interface{}{"auto", "true", "false", ""},
			invalidValues: []interface{}{"maybe", "yes"},
		},
		{
			name:          "file_max_size must be positive",
			key:           KeyAppLogFileMaxSize,
			validValues:   []interface{}{1, 100},
			invalidValues: []interface{}{0, -1},
		},
		{
			name:          "file_max_backups must be non-negative",
			key:           KeyAppLogFileMaxBackups,
			validValues:   []interface{}{0, 3},
			invalidValues: []interface{}{-1},
		},
		{
			name:          "file_max_age must be non-negative",
			key:           KeyAppLogFileMaxAge,
			validValues:   []interface{}{0, 28},
			invalidValues: []interface{}{-3},
		},
		{
			name:          "sampling_initial must be positive",
			key:           KeyAppLogSamplingInitial,
			validValues:   []interface{}{1, 100},
			invalidValues: []interface{}{0, -1},
		},
		{
			name:          "sampling_thereafter must be positive",
			key:           KeyAppLogSamplingThereafter,
			validValues:   []interface{}{1, 100},
			invalidValues: []interface{}{0, -1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE
			opt, ok := byKey[tt.key]
			require.True(t, ok, "option %q not found in CoreOptions()", tt.key)
			require.NotNil(t, opt.Validation, "option %q has no Validation function", tt.key)

			// EXECUTION AND ASSERTION PHASE
			for _, v := range tt.validValues {
				assert.NoError(t, opt.Validation(v),
					"value %v should be valid for %q", v, tt.key)
			}
			for _, v := range tt.invalidValues {
				assert.Error(t, opt.Validation(v),
					"value %v should be invalid for %q", v, tt.key)
			}
		})
	}
}

func TestOptionExampleValueString(t *testing.T) {
	tests := []struct {
		name string
		opt  ConfigOption
		want string
	}{
		{
			name: "With example",
			opt: ConfigOption{
				DefaultValue: "default",
				Example:      "example",
			},
			want: "example",
		},
		{
			name: "Without example",
			opt: ConfigOption{
				DefaultValue: "default",
				Example:      "",
			},
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.ExampleValueString()

			// ASSERTION PHASE
			assert.Equal(t, tt.want, got)
		})
	}
}
