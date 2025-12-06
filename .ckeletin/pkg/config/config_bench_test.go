// internal/config/config_bench_test.go

package config

import (
	"testing"
)

// Benchmark config registry operations

func BenchmarkRegistry(b *testing.B) {
	// Benchmarks retrieving the full config registry
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Registry()
	}
}

func BenchmarkSetDefaults(b *testing.B) {
	// Benchmarks applying all default values to viper
	// Note: This is called once during initialization
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		SetDefaults()
	}
}

func BenchmarkValidateConfigValue(b *testing.B) {
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"String_short", "test.key", "short string"},
		{"String_medium", "test.key", "This is a medium length string that contains some text but is not too long"},
		{"String_array", "test.array", []string{"a", "b", "c", "d", "e"}},
		{"Int", "test.int", 42},
		{"Bool", "test.bool", true},
		{"Float", "test.float", 3.14},
		{"NestedMap", "test.nested", map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		}},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = ValidateConfigValue(tt.key, tt.value)
			}
		})
	}
}

func BenchmarkValidateAllConfigValues(b *testing.B) {
	// Benchmark validating a typical configuration structure
	testConfig := map[string]interface{}{
		"app": map[string]interface{}{
			"log_level": "info",
			"ping": map[string]interface{}{
				"output_message": "Pong!",
				"color":          "green",
				"ui":             false,
			},
			"docs": map[string]interface{}{
				"output_format": "markdown",
				"output_file":   "",
			},
		},
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ValidateAllConfigValues(testConfig)
	}
}

func BenchmarkConfigOptionEnvVarName(b *testing.B) {
	opt := ConfigOption{
		Key:          "app.log_level",
		DefaultValue: "info",
		Description:  "Log level",
		Type:         "string",
	}

	prefix := "MYAPP"

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = opt.EnvVarName(prefix)
	}
}

func BenchmarkConfigOptionDefaultValueString(b *testing.B) {
	tests := []struct {
		name string
		opt  ConfigOption
	}{
		{
			name: "String",
			opt: ConfigOption{
				Key:          "test.string",
				DefaultValue: "test value",
				Type:         "string",
			},
		},
		{
			name: "Int",
			opt: ConfigOption{
				Key:          "test.int",
				DefaultValue: 42,
				Type:         "int",
			},
		},
		{
			name: "Bool",
			opt: ConfigOption{
				Key:          "test.bool",
				DefaultValue: true,
				Type:         "bool",
			},
		},
		{
			name: "Nil",
			opt: ConfigOption{
				Key:          "test.nil",
				DefaultValue: nil,
				Type:         "string",
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tt.opt.DefaultValueString()
			}
		})
	}
}
