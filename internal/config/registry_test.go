// internal/config/registry_test.go

package config

import (
	"testing"

	"github.com/spf13/viper"
)

func TestConfigOptionEnvVarName(t *testing.T) {
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
			got := tt.opt.EnvVarName(tt.prefix)
			if got != tt.want {
				t.Errorf("EnvVarName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigOptionDefaultValueString(t *testing.T) {
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
			got := tt.opt.DefaultValueString()
			if got != tt.want {
				t.Errorf("DefaultValueString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigOptionExampleValueString(t *testing.T) {
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
			got := tt.opt.ExampleValueString()
			if got != tt.want {
				t.Errorf("ExampleValueString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegistryHasExpectedKeys(t *testing.T) {
	registry := Registry()

	// Check that the registry has the expected minimum number of entries
	if len(registry) < 4 {
		t.Errorf("Registry() returned %d entries, expected at least 4", len(registry))
	}

	// Check for essential keys
	requiredKeys := []string{
		"app.log_level",
		"app.ping.output_message",
		"app.ping.output_color",
		"app.ping.ui",
	}

	for _, key := range requiredKeys {
		found := false
		for _, opt := range registry {
			if opt.Key == key {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Registry() missing required key %q", key)
		}
	}
}

func TestSetDefaults(t *testing.T) {
	// Reset viper before test
	viper.Reset()

	// Apply defaults
	SetDefaults()

	// Check that defaults were set
	registry := Registry()
	for _, opt := range registry {
		// Skip nil defaults as they can't be reliably tested
		if opt.DefaultValue == nil {
			continue
		}

		// GetString works for all types in viper since everything is stored as strings internally
		got := viper.Get(opt.Key)
		if got != opt.DefaultValue {
			t.Errorf("Default for %q = %v, want %v", opt.Key, got, opt.DefaultValue)
		}
	}
}
