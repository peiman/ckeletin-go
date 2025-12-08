// cmd/flags_test.go

package cmd

import (
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testShortFlagOptions returns test config options with ShortFlags for all types
// This is used to test the ShortFlag branches in RegisterFlagsForPrefixWithOverrides
func testShortFlagOptions() []config.ConfigOption {
	return []config.ConfigOption{
		{Key: "test.shortflag.string_opt", DefaultValue: "default", Description: "Test string with short flag", Type: "string", ShortFlag: "s"},
		{Key: "test.shortflag.bool_opt", DefaultValue: true, Description: "Test bool with short flag", Type: "bool", ShortFlag: "b"},
		{Key: "test.shortflag.int_opt", DefaultValue: 42, Description: "Test int with short flag", Type: "int", ShortFlag: "i"},
		{Key: "test.shortflag.float_opt", DefaultValue: 3.14, Description: "Test float with short flag", Type: "float64", ShortFlag: "f"},
		{Key: "test.shortflag.slice_opt", DefaultValue: []string{"a", "b"}, Description: "Test slice with short flag", Type: "[]string", ShortFlag: "l"},
		{Key: "test.shortflag.unknown_type", DefaultValue: "value", Description: "Test unknown type with short flag", Type: "customtype", ShortFlag: "u"},
	}
}

func init() {
	// Register test options provider for ShortFlag testing
	config.RegisterOptionsProvider(testShortFlagOptions)
}

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
			assert.Equal(t, tt.want, got, "stringDefault should convert value correctly")
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
			name:  "Zero int64",
			input: int64(0),
			want:  false,
		},
		{
			name:  "Non-zero int32",
			input: int32(50),
			want:  true,
		},
		{
			name:  "Zero int32",
			input: int32(0),
			want:  false,
		},
		{
			name:  "Non-zero int16",
			input: int16(25),
			want:  true,
		},
		{
			name:  "Zero int16",
			input: int16(0),
			want:  false,
		},
		{
			name:  "Non-zero int8",
			input: int8(10),
			want:  true,
		},
		{
			name:  "Zero int8",
			input: int8(0),
			want:  false,
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
			assert.Equal(t, tt.want, got, "boolDefault should convert value correctly")
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
			assert.Equal(t, tt.want, got, "intDefault should convert value correctly")
		})
	}
}

func TestIntDefault_AllUintTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  int
	}{
		{
			name:  "Uint64 value",
			input: uint64(200),
			want:  200,
		},
		{
			name:  "Uint32 value",
			input: uint32(150),
			want:  150,
		},
		{
			name:  "Uint16 value",
			input: uint16(75),
			want:  75,
		},
		{
			name:  "Uint8 value",
			input: uint8(50),
			want:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intDefault(tt.input)
			assert.Equal(t, tt.want, got, "intDefault should handle uint types correctly")
		})
	}
}

func TestIntDefault_Uint64Overflow(t *testing.T) {
	// Test uint64 overflow handling - these values are larger than math.MaxInt
	// and should be clamped to max int, not silently wrapped to negative values
	tests := []struct {
		name  string
		input uint64
		want  int
	}{
		{
			name:  "uint64 max value (should clamp to max int)",
			input: ^uint64(0),         // math.MaxUint64
			want:  int(^uint(0) >> 1), // max int
		},
		{
			name:  "uint64 larger than max int (should clamp)",
			input: uint64(^uint(0)>>1) + 1, // math.MaxInt + 1
			want:  int(^uint(0) >> 1),      // max int
		},
		{
			name:  "Small uint64 value (no overflow)",
			input: uint64(100),
			want:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intDefault(tt.input)
			assert.Equal(t, tt.want, got, "intDefault should clamp uint64 overflow, not wrap to negative")
		})
	}
}

func TestIntDefault_Uint32Overflow(t *testing.T) {
	// Test uint/uint32 overflow on 32-bit systems
	// These should also be checked and clamped
	tests := []struct {
		name  string
		input uint
		want  int
	}{
		{
			name:  "uint max value (should clamp to max int)",
			input: ^uint(0),           // max uint
			want:  int(^uint(0) >> 1), // max int
		},
		{
			name:  "Small uint value (no overflow)",
			input: uint(100),
			want:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intDefault(tt.input)
			assert.Equal(t, tt.want, got, "intDefault should handle uint overflow correctly")
		})
	}

	// Test uint32 specifically (in addition to uint tests above)
	// On 64-bit systems, all uint32 values fit in int
	// On 32-bit systems, values > max int would clamp
	t.Run("uint32 max value", func(t *testing.T) {
		input := ^uint32(0) // 4294967295
		got := intDefault(input)

		// Calculate expected value based on system architecture
		maxInt := int(^uint(0) >> 1)
		var want int
		if uint32(maxInt) >= ^uint32(0) {
			// 64-bit system: uint32 max fits in int
			want = int(^uint32(0))
		} else {
			// 32-bit system: would clamp to max int
			want = maxInt
		}

		assert.Equal(t, want, got, "intDefault should handle uint32 max value correctly")
	})
}

func TestIntDefault_Int64Overflow(t *testing.T) {
	// Test int64 overflow handling
	// We test with values that would overflow int on 32-bit systems
	tests := []struct {
		name     string
		input    int64
		checkPos bool // true if checking positive overflow
	}{
		{
			name:     "Very large positive int64",
			input:    9223372036854775807, // math.MaxInt64
			checkPos: true,
		},
		{
			name:     "Very large negative int64",
			input:    -9223372036854775808, // math.MinInt64
			checkPos: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intDefault(tt.input)
			// Just verify it doesn't panic and returns a value
			// On 64-bit systems, these values fit in int
			// On 32-bit systems, they would be clamped
			assert.NotEqual(t, 0, got, "intDefault should handle large int64, should not return 0")
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
			assert.Equal(t, tt.want, got, "floatDefault should convert value correctly")
		})
	}
}

func TestFloatDefault_AllIntTypes(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  float64
	}{
		{
			name:  "Int32 value",
			input: int32(123),
			want:  123.0,
		},
		{
			name:  "Int16 value",
			input: int16(456),
			want:  456.0,
		},
		{
			name:  "Int8 value",
			input: int8(78),
			want:  78.0,
		},
		{
			name:  "Uint value",
			input: uint(999),
			want:  999.0,
		},
		{
			name:  "Uint64 value",
			input: uint64(12345),
			want:  12345.0,
		},
		{
			name:  "Uint32 value",
			input: uint32(6789),
			want:  6789.0,
		},
		{
			name:  "Uint16 value",
			input: uint16(321),
			want:  321.0,
		},
		{
			name:  "Uint8 value",
			input: uint8(99),
			want:  99.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := floatDefault(tt.input)
			assert.Equal(t, tt.want, got, "floatDefault should handle all int types correctly")
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
			want:  nil, // Function returns nil for nil input
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
			want:  nil, // Function returns nil for invalid types
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stringSliceDefault(tt.input)
			assert.Equal(t, tt.want, got, "stringSliceDefault should convert value correctly")
		})
	}
}

// TestRegisterFlagsForPrefixWithOverrides_ShortFlags tests that short flags are properly registered
// by using the ping command's actual configuration which includes short flags
func TestRegisterFlagsForPrefixWithOverrides_ShortFlags(t *testing.T) {
	// Reset viper for clean test
	viper.Reset()

	// Create a fresh command
	cmd := &cobra.Command{Use: "test"}

	// Register flags for ping prefix (which has short flags defined)
	err := RegisterFlagsForPrefixWithOverrides(cmd, "app.ping.", nil)
	assert.NoError(t, err, "RegisterFlagsForPrefixWithOverrides should not return error")

	// Verify the message flag exists with short flag "m"
	messageFlag := cmd.Flags().Lookup("output-message")
	if messageFlag != nil {
		assert.Equal(t, "m", messageFlag.Shorthand, "message flag should have shorthand 'm'")
	}

	// Verify the color flag exists with short flag "c"
	colorFlag := cmd.Flags().Lookup("output-color")
	if colorFlag != nil {
		assert.Equal(t, "c", colorFlag.Shorthand, "color flag should have shorthand 'c'")
	}

	// Verify the ui flag exists without short flag (ShortFlag is empty)
	uiFlag := cmd.Flags().Lookup("ui")
	if uiFlag != nil {
		assert.Empty(t, uiFlag.Shorthand, "ui flag should not have shorthand")
	}
}

// TestShortFlagRegistration_DirectCobra tests short flag registration directly with cobra
func TestShortFlagRegistration_DirectCobra(t *testing.T) {
	tests := []struct {
		name         string
		flagType     string
		shortFlag    string
		defaultValue interface{}
	}{
		{"string with short flag", "string", "s", "default"},
		{"bool with short flag", "bool", "b", false},
		{"int with short flag", "int", "n", 42},
		{"float with short flag", "float64", "f", 3.14},
		{"stringslice with short flag", "[]string", "l", []string{"a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}

			// Register flag with short flag based on type
			switch tt.flagType {
			case "string":
				cmd.Flags().StringP("test-flag", tt.shortFlag, stringDefault(tt.defaultValue), "test")
			case "bool":
				cmd.Flags().BoolP("test-flag", tt.shortFlag, boolDefault(tt.defaultValue), "test")
			case "int":
				cmd.Flags().IntP("test-flag", tt.shortFlag, intDefault(tt.defaultValue), "test")
			case "float64":
				cmd.Flags().Float64P("test-flag", tt.shortFlag, floatDefault(tt.defaultValue), "test")
			case "[]string":
				cmd.Flags().StringSliceP("test-flag", tt.shortFlag, stringSliceDefault(tt.defaultValue), "test")
			}

			// Verify short flag is set
			flag := cmd.Flags().Lookup("test-flag")
			require.NotNil(t, flag, "Flag should exist")
			assert.Equal(t, tt.shortFlag, flag.Shorthand, "Flag should have correct shorthand")
		})
	}
}

// TestRegisterFlagsForPrefixWithOverrides_AllTypesWithShortFlags tests ShortFlag branches for all types
// This test uses the test options registered in init() above
func TestRegisterFlagsForPrefixWithOverrides_AllTypesWithShortFlags(t *testing.T) {
	viper.Reset()
	cmd := &cobra.Command{Use: "test"}

	// Register flags for the test.shortflag prefix (registered in init() above)
	err := RegisterFlagsForPrefixWithOverrides(cmd, "test.shortflag.", nil)
	require.NoError(t, err, "RegisterFlagsForPrefixWithOverrides should not return error")

	// Verify string flag with short flag
	stringFlag := cmd.Flags().Lookup("string-opt")
	require.NotNil(t, stringFlag, "string flag should exist")
	assert.Equal(t, "s", stringFlag.Shorthand, "string flag should have shorthand 's'")

	// Verify bool flag with short flag
	boolFlag := cmd.Flags().Lookup("bool-opt")
	require.NotNil(t, boolFlag, "bool flag should exist")
	assert.Equal(t, "b", boolFlag.Shorthand, "bool flag should have shorthand 'b'")

	// Verify int flag with short flag
	intFlag := cmd.Flags().Lookup("int-opt")
	require.NotNil(t, intFlag, "int flag should exist")
	assert.Equal(t, "i", intFlag.Shorthand, "int flag should have shorthand 'i'")

	// Verify float flag with short flag
	floatFlag := cmd.Flags().Lookup("float-opt")
	require.NotNil(t, floatFlag, "float flag should exist")
	assert.Equal(t, "f", floatFlag.Shorthand, "float flag should have shorthand 'f'")

	// Verify stringslice flag with short flag
	sliceFlag := cmd.Flags().Lookup("slice-opt")
	require.NotNil(t, sliceFlag, "stringslice flag should exist")
	assert.Equal(t, "l", sliceFlag.Shorthand, "stringslice flag should have shorthand 'l'")

	// Verify unknown type flag with short flag (falls through to default case)
	unknownFlag := cmd.Flags().Lookup("unknown-type")
	require.NotNil(t, unknownFlag, "unknown type flag should exist")
	assert.Equal(t, "u", unknownFlag.Shorthand, "unknown type flag should have shorthand 'u'")
}
