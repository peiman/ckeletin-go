// internal/config/options_test.go

package config

import (
	"testing"
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
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.EnvVarName(tt.prefix)

			// ASSERTION PHASE
			if got != tt.want {
				t.Errorf("EnvVarName() = %v, want %v", got, tt.want)
			}
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
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.DefaultValueString()

			// ASSERTION PHASE
			if got != tt.want {
				t.Errorf("DefaultValueString() = %v, want %v", got, tt.want)
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
			// SETUP PHASE
			// No specific setup needed for this test

			// EXECUTION PHASE
			got := tt.opt.ExampleValueString()

			// ASSERTION PHASE
			if got != tt.want {
				t.Errorf("ExampleValueString() = %v, want %v", got, tt.want)
			}
		})
	}
}
