// internal/config/limits_test.go

package config

import (
	"strings"
	"testing"
)

func TestValidateConfigValue_Strings(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		key         string
		value       string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Normal string",
			key:     "test.key",
			value:   "normal value",
			wantErr: false,
		},
		{
			name:    "Empty string",
			key:     "test.key",
			value:   "",
			wantErr: false,
		},
		{
			name:    "String at max length",
			key:     "test.key",
			value:   strings.Repeat("x", MaxStringValueLength),
			wantErr: false,
		},
		{
			name:        "String exceeds max length",
			key:         "test.key",
			value:       strings.Repeat("x", MaxStringValueLength+1),
			wantErr:     true,
			errContains: "exceeds max string length",
		},
		{
			name:        "Very large string",
			key:         "test.key",
			value:       strings.Repeat("x", MaxStringValueLength*2),
			wantErr:     true,
			errContains: "exceeds max string length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfigValue(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestValidateConfigValue_StringSlices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		key         string
		value       []string
		wantErr     bool
		errContains string
	}{
		{
			name:    "Normal slice",
			key:     "test.key",
			value:   []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name:    "Empty slice",
			key:     "test.key",
			value:   []string{},
			wantErr: false,
		},
		{
			name:    "Slice at max length",
			key:     "test.key",
			value:   make([]string, MaxSliceLength),
			wantErr: false,
		},
		{
			name:        "Slice exceeds max length",
			key:         "test.key",
			value:       make([]string, MaxSliceLength+1),
			wantErr:     true,
			errContains: "exceeds max length",
		},
		{
			name:        "Slice with oversized string",
			key:         "test.key",
			value:       []string{"normal", strings.Repeat("x", MaxStringValueLength+1)},
			wantErr:     true,
			errContains: "exceeds max string length",
		},
		{
			name: "Slice with all valid strings",
			key:  "test.key",
			value: []string{
				strings.Repeat("a", 100),
				strings.Repeat("b", 100),
				strings.Repeat("c", 100),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfigValue(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestValidateConfigValue_InterfaceSlices(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		key         string
		value       []interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:    "Normal interface slice",
			key:     "test.key",
			value:   []interface{}{"a", "b", 123},
			wantErr: false,
		},
		{
			name:        "Interface slice exceeds max length",
			key:         "test.key",
			value:       make([]interface{}, MaxSliceLength+1),
			wantErr:     true,
			errContains: "exceeds max length",
		},
		{
			name:        "Interface slice with oversized string",
			key:         "test.key",
			value:       []interface{}{"normal", strings.Repeat("x", MaxStringValueLength+1)},
			wantErr:     true,
			errContains: "exceeds max string length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfigValue(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestValidateConfigValue_NestedMaps(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		key         string
		value       map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name: "Normal nested map",
			key:  "test",
			value: map[string]interface{}{
				"nested": "value",
				"count":  123,
			},
			wantErr: false,
		},
		{
			name: "Nested map with oversized string",
			key:  "test",
			value: map[string]interface{}{
				"nested": strings.Repeat("x", MaxStringValueLength+1),
			},
			wantErr:     true,
			errContains: "exceeds max string length",
		},
		{
			name: "Deeply nested map",
			key:  "test",
			value: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"value": "ok",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfigValue(tt.key, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errContains, err)
				}
			}
		})
	}
}

func TestValidateConfigValue_NumericTypes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		key   string
		value interface{}
	}{
		{"int", "test.key", 123},
		{"int8", "test.key", int8(123)},
		{"int16", "test.key", int16(123)},
		{"int32", "test.key", int32(123)},
		{"int64", "test.key", int64(123)},
		{"uint", "test.key", uint(123)},
		{"uint8", "test.key", uint8(123)},
		{"uint16", "test.key", uint16(123)},
		{"uint32", "test.key", uint32(123)},
		{"uint64", "test.key", uint64(123)},
		{"float32", "test.key", float32(123.45)},
		{"float64", "test.key", float64(123.45)},
		{"bool", "test.key", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := ValidateConfigValue(tt.key, tt.value)
			if err != nil {
				t.Errorf("ValidateConfigValue() for numeric type should not error, got: %v", err)
			}
		})
	}
}

func TestValidateConfigValue_NilValue(t *testing.T) {
	t.Parallel()
	err := ValidateConfigValue("test.key", nil)
	if err != nil {
		t.Errorf("ValidateConfigValue() for nil value should not error, got: %v", err)
	}
}

func TestValidateAllConfigValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		values    map[string]interface{}
		wantCount int
	}{
		{
			name: "All valid values",
			values: map[string]interface{}{
				"string": "value",
				"int":    123,
				"slice":  []string{"a", "b"},
			},
			wantCount: 0,
		},
		{
			name: "One invalid value",
			values: map[string]interface{}{
				"good": "value",
				"bad":  strings.Repeat("x", MaxStringValueLength+1),
			},
			wantCount: 1,
		},
		{
			name: "Multiple invalid values",
			values: map[string]interface{}{
				"bad1": strings.Repeat("x", MaxStringValueLength+1),
				"good": "value",
				"bad2": make([]string, MaxSliceLength+1),
			},
			wantCount: 2,
		},
		{
			name:      "Empty map",
			values:    map[string]interface{}{},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errors := ValidateAllConfigValues(tt.values)

			if len(errors) != tt.wantCount {
				t.Errorf("ValidateAllConfigValues() returned %d errors, want %d", len(errors), tt.wantCount)
				for i, err := range errors {
					t.Logf("Error %d: %v", i, err)
				}
			}
		})
	}
}
