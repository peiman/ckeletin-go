// .ckeletin/pkg/config/validators_test.go

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateOneOf(t *testing.T) {
	t.Parallel()

	allowed := []string{"red", "green", "blue"}

	tests := []struct {
		name       string
		value      interface{}
		allowEmpty bool
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "Valid value",
			value:      "red",
			allowEmpty: false,
			wantErr:    false,
		},
		{
			name:       "Valid value with different case",
			value:      "RED",
			allowEmpty: false,
			wantErr:    false,
		},
		{
			name:       "Valid value with whitespace",
			value:      "  green  ",
			allowEmpty: false,
			wantErr:    false,
		},
		{
			name:       "Invalid value",
			value:      "purple",
			allowEmpty: false,
			wantErr:    true,
			errMsg:     "invalid value \"purple\"",
		},
		{
			name:       "Empty string when not allowed",
			value:      "",
			allowEmpty: false,
			wantErr:    true,
			errMsg:     "invalid value",
		},
		{
			name:       "Empty string when allowed",
			value:      "",
			allowEmpty: true,
			wantErr:    false,
		},
		{
			name:       "Non-string value is skipped",
			value:      42,
			allowEmpty: false,
			wantErr:    false,
		},
		{
			name:       "Nil value is skipped",
			value:      nil,
			allowEmpty: false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validate := ValidateOneOf(allowed, tt.allowEmpty)
			err := validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		value      interface{}
		allowEmpty bool
		wantErr    bool
	}{
		{name: "trace", value: "trace", wantErr: false},
		{name: "debug", value: "debug", wantErr: false},
		{name: "info", value: "info", wantErr: false},
		{name: "warn", value: "warn", wantErr: false},
		{name: "error", value: "error", wantErr: false},
		{name: "fatal", value: "fatal", wantErr: false},
		{name: "panic", value: "panic", wantErr: false},
		{name: "disabled", value: "disabled", wantErr: false},
		{name: "invalid level", value: "verbose", wantErr: true},
		{name: "empty not allowed", value: "", allowEmpty: false, wantErr: true},
		{name: "empty allowed", value: "", allowEmpty: true, wantErr: false},
		{name: "uppercase", value: "DEBUG", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validate := ValidateLogLevel(tt.allowEmpty)
			err := validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateColor(t *testing.T) {
	t.Parallel()

	validColors := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}

	tests := []struct {
		name    string
		value   interface{}
		wantErr bool
	}{
		{name: "valid white", value: "white", wantErr: false},
		{name: "valid red", value: "red", wantErr: false},
		{name: "valid green", value: "green", wantErr: false},
		{name: "valid cyan", value: "cyan", wantErr: false},
		{name: "invalid color", value: "purple", wantErr: true},
		{name: "invalid color orange", value: "orange", wantErr: true},
		{name: "empty string", value: "", wantErr: true},
		{name: "uppercase", value: "WHITE", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			validate := ValidateColor(validColors)
			err := validate(tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid value")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
