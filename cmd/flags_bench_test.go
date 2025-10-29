// cmd/flags_bench_test.go

package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

// Benchmark type conversion functions that are called during flag registration

func BenchmarkStringDefault(b *testing.B) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"String", "test string"},
		{"Int", 42},
		{"Float", 3.14},
		{"Bool", true},
		{"Nil", nil},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = stringDefault(tt.input)
			}
		})
	}
}

func BenchmarkBoolDefault(b *testing.B) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"Bool_true", true},
		{"Bool_false", false},
		{"String_true", "true"},
		{"String_false", "false"},
		{"Int_nonzero", 42},
		{"Int_zero", 0},
		{"Int64_nonzero", int64(100)},
		{"Int64_zero", int64(0)},
		{"Nil", nil},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = boolDefault(tt.input)
			}
		})
	}
}

func BenchmarkIntDefault(b *testing.B) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"Int", 42},
		{"Int64", int64(100)},
		{"Int32", int32(50)},
		{"Float64", 3.14},
		{"String", "123"},
		{"Nil", nil},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = intDefault(tt.input)
			}
		})
	}
}

func BenchmarkFloatDefault(b *testing.B) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"Float64", 3.14159},
		{"Float32", float32(2.71)},
		{"Int", 42},
		{"Int64", int64(100)},
		{"String", "3.14"},
		{"Nil", nil},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = floatDefault(tt.input)
			}
		})
	}
}

func BenchmarkStringSliceDefault(b *testing.B) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{"StringSlice", []string{"a", "b", "c"}},
		{"InterfaceSlice", []interface{}{"x", "y", "z"}},
		{"SingleString", "single"},
		{"Nil", nil},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = stringSliceDefault(tt.input)
			}
		})
	}
}

// Benchmark flag registration process
func BenchmarkRegisterFlagsForPrefixWithOverrides(b *testing.B) {
	// This benchmarks the flag registration which happens during command initialization
	// Note: This creates new commands each iteration, so it measures the full cost

	tests := []struct {
		name    string
		prefix  string
	}{
		{"PingCommand", "app.ping."},
		{"DocsCommand", "app.docs."},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				cmd := &cobra.Command{Use: "test"}
				_ = RegisterFlagsForPrefixWithOverrides(cmd, tt.prefix, nil)
			}
		})
	}
}
