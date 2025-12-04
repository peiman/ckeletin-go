// cmd/root_bench_test.go

package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Benchmark configuration retrieval functions

func BenchmarkGetConfigValueWithFlags(b *testing.B) {
	// Setup
	viper.Reset()
	viper.Set(config.KeyAppPingOutputMessage, "test message")

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("message", "default", "message flag")

	tests := []struct {
		name string
		fn   func() interface{}
	}{
		{
			"String",
			func() interface{} {
				return getConfigValueWithFlags[string](cmd, "message", config.KeyAppPingOutputMessage)
			},
		},
		{
			"Bool",
			func() interface{} {
				return getConfigValueWithFlags[bool](cmd, "ui", "app.ping.ui")
			},
		},
		{
			"Int",
			func() interface{} {
				return getConfigValueWithFlags[int](cmd, "count", "app.count")
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tt.fn()
			}
		})
	}
}

func BenchmarkGetKeyValue(b *testing.B) {
	// Setup
	viper.Reset()
	viper.Set("test.string", "value")
	viper.Set("test.int", 42)
	viper.Set("test.bool", true)
	viper.Set("test.float", 3.14)

	tests := []struct {
		name string
		fn   func() interface{}
	}{
		{
			"String",
			func() interface{} {
				return getKeyValue[string]("test.string")
			},
		},
		{
			"Int",
			func() interface{} {
				return getKeyValue[int]("test.int")
			},
		},
		{
			"Bool",
			func() interface{} {
				return getKeyValue[bool]("test.bool")
			},
		},
		{
			"Float64",
			func() interface{} {
				return getKeyValue[float64]("test.float")
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = tt.fn()
			}
		})
	}
}

func BenchmarkEnvPrefix(b *testing.B) {
	// Benchmark the environment variable prefix generation
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = EnvPrefix()
	}
}

func BenchmarkConfigPaths(b *testing.B) {
	// Benchmark the configuration paths generation
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ConfigPaths()
	}
}

// Benchmark regex pattern matching for EnvPrefix
func BenchmarkEnvPrefixWithDifferentNames(b *testing.B) {
	testCases := []struct {
		name       string
		binaryName string
	}{
		{"Simple", "myapp"},
		{"WithHyphens", "my-app-cli"},
		{"WithUnderscores", "my_app_cli"},
		{"WithNumbers", "app123"},
		{"Complex", "my-complex_app-v2"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			// Save original
			originalBinaryName := binaryName
			defer func() { binaryName = originalBinaryName }()

			binaryName = tc.binaryName
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = EnvPrefix()
			}
		})
	}
}
