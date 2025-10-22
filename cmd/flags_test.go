// cmd/flags_test.go

package cmd

import (
	"testing"
)

func TestStringDefault(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  "",
		},
		{
			name:  "String value",
			input: "test",
			want:  "test",
		},
		{
			name:  "Empty string",
			input: "",
			want:  "",
		},
		{
			name:  "Integer conversion",
			input: 42,
			want:  "42",
		},
		{
			name:  "Float conversion",
			input: 3.14,
			want:  "3.14",
		},
		{
			name:  "Bool conversion",
			input: true,
			want:  "true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringDefault(tt.input)
			if got != tt.want {
				t.Errorf("stringDefault() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBoolDefault(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  false,
		},
		{
			name:  "Bool true",
			input: true,
			want:  true,
		},
		{
			name:  "Bool false",
			input: false,
			want:  false,
		},
		{
			name:  "String true",
			input: "true",
			want:  true,
		},
		{
			name:  "String false",
			input: "false",
			want:  false,
		},
		{
			name:  "String 1",
			input: "1",
			want:  true,
		},
		{
			name:  "String 0",
			input: "0",
			want:  false,
		},
		{
			name:  "Invalid string",
			input: "invalid",
			want:  false,
		},
		{
			name:  "Non-zero int",
			input: 42,
			want:  true,
		},
		{
			name:  "Zero int",
			input: 0,
			want:  false,
		},
		{
			name:  "Non-zero int64",
			input: int64(100),
			want:  true,
		},
		{
			name:  "Invalid type (float)",
			input: 3.14,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := boolDefault(tt.input)
			if got != tt.want {
				t.Errorf("boolDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIntDefault(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  0,
		},
		{
			name:  "Int value",
			input: 42,
			want:  42,
		},
		{
			name:  "Negative int",
			input: -10,
			want:  -10,
		},
		{
			name:  "Int64 value",
			input: int64(100),
			want:  100,
		},
		{
			name:  "Int32 value",
			input: int32(50),
			want:  50,
		},
		{
			name:  "Int16 value",
			input: int16(25),
			want:  25,
		},
		{
			name:  "Int8 value",
			input: int8(12),
			want:  12,
		},
		{
			name:  "Uint value",
			input: uint(33),
			want:  33,
		},
		{
			name:  "Float64 value",
			input: 3.14,
			want:  3,
		},
		{
			name:  "Float32 value",
			input: float32(2.71),
			want:  2,
		},
		{
			name:  "String valid number",
			input: "123",
			want:  123,
		},
		{
			name:  "String invalid",
			input: "not-a-number",
			want:  0,
		},
		{
			name:  "Invalid type",
			input: true,
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intDefault(tt.input)
			if got != tt.want {
				t.Errorf("intDefault() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestFloatDefault(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  0.0,
		},
		{
			name:  "Float64 value",
			input: 3.14,
			want:  3.14,
		},
		{
			name:  "Float32 value",
			input: float32(2.71),
			want:  float64(float32(2.71)), // Account for float32->float64 conversion
		},
		{
			name:  "Int value",
			input: 42,
			want:  42.0,
		},
		{
			name:  "Int64 value",
			input: int64(100),
			want:  100.0,
		},
		{
			name:  "String valid number",
			input: "3.14159",
			want:  3.14159,
		},
		{
			name:  "String invalid",
			input: "not-a-number",
			want:  0.0,
		},
		{
			name:  "Negative float",
			input: -2.5,
			want:  -2.5,
		},
		{
			name:  "Zero",
			input: 0.0,
			want:  0.0,
		},
		{
			name:  "Invalid type",
			input: true,
			want:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := floatDefault(tt.input)
			if got != tt.want {
				t.Errorf("floatDefault() = %f, want %f", got, tt.want)
			}
		})
	}
}

func TestStringSliceDefault(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  []string
	}{
		{
			name:  "Nil value",
			input: nil,
			want:  []string{},
		},
		{
			name:  "String slice",
			input: []string{"a", "b", "c"},
			want:  []string{"a", "b", "c"},
		},
		{
			name:  "Empty slice",
			input: []string{},
			want:  []string{},
		},
		{
			name:  "Interface slice with strings",
			input: []interface{}{"x", "y", "z"},
			want:  []string{"x", "y", "z"},
		},
		{
			name:  "Interface slice with mixed types",
			input: []interface{}{"a", 42, true},
			want:  []string{"a", "42", "true"},
		},
		{
			name:  "Single string (not slice)",
			input: "single",
			want:  []string{"single"},
		},
		{
			name:  "Invalid type",
			input: 123,
			want:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringSliceDefault(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("stringSliceDefault() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("stringSliceDefault()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestRegisterFlagsForPrefixWithOverrides(t *testing.T) {
	t.Skip("Skipping RegisterFlagsForPrefixWithOverrides test - requires complex setup with cobra and viper. " +
		"Function is indirectly tested via NewCommand tests and integration tests.")
}
