// internal/logger/logger_bench_test.go

package logger

import (
	"fmt"
	"strings"
	"testing"
)

// Benchmark log sanitization functions that are called frequently

func BenchmarkSanitizeLogString(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"Short", "short string"},
		{"Medium", strings.Repeat("x", 500)},
		{"Long", strings.Repeat("x", 2000)},
		{"VeryLong", strings.Repeat("x", 10000)},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = SanitizeLogString(tt.input)
			}
		})
	}
}

func BenchmarkSanitizePath(b *testing.B) {
	tests := []struct {
		name string
		path string
	}{
		{"Short", "/usr/bin/app"},
		{"Medium", "/home/user/.config/myapp/config.yaml"},
		{"Long", "/home/user/very/deep/nested/directory/structure/with/many/components/file.yaml"},
		{"VeryLong", strings.Repeat("/dir", 50) + "/file.yaml"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = SanitizePath(tt.path)
			}
		})
	}
}

func BenchmarkSanitizeError(b *testing.B) {
	tests := []struct {
		name   string
		errMsg error
	}{
		{"Short", fmt.Errorf("file not found")},
		{"Medium", fmt.Errorf("failed to read configuration file: permission denied")},
		{"Long", fmt.Errorf("an error occurred while processing the request: %s", strings.Repeat("details ", 50))},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = SanitizeError(tt.errMsg)
			}
		})
	}
}

func BenchmarkSanitizeLogStringWithCustomMaxLength(b *testing.B) {
	// Test performance impact of different max lengths
	input := strings.Repeat("x", 10000)

	maxLengths := []int{100, 500, 1000, 5000}

	for _, maxLen := range maxLengths {
		b.Run(string(rune(maxLen)), func(b *testing.B) {
			SetMaxLogLength(maxLen)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = SanitizeLogString(input)
			}
		})
	}

	// Reset to default
	SetMaxLogLength(1000)
}
