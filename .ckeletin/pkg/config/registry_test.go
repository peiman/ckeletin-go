// internal/config/registry_test.go

package config_test

import (
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	_ "github.com/peiman/ckeletin-go/.ckeletin/pkg/config/commands" // Import to trigger init() registration
	"github.com/spf13/viper"
)

func TestRegistryHasExpectedKeys(t *testing.T) {
	// SETUP PHASE
	// Only test framework-level keys, not project-specific keys (like ping, check)
	requiredKeys := []string{
		"app.log_level",
		"app.docs.output_format",
	}

	// EXECUTION PHASE
	registry := config.Registry()

	// ASSERTION PHASE
	// Check that the registry has the expected minimum number of entries
	if len(registry) < 2 {
		t.Errorf("Registry() returned %d entries, expected at least 2", len(registry))
	}

	// Check for essential keys
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
	// SETUP PHASE
	// Reset viper before test
	viper.Reset()

	// EXECUTION PHASE
	// Apply defaults
	config.SetDefaults()

	// ASSERTION PHASE
	// Check that defaults were set
	registry := config.Registry()
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

// Test command-specific options are included in Registry
func TestRegistryIncludesCommandOptions(t *testing.T) {
	// SETUP PHASE
	// Only test framework command keys, not project-specific keys (like ping, check)
	docsKeys := map[string]bool{
		"app.docs.output_format": false,
		"app.docs.output_file":   false,
	}

	coreKeys := map[string]bool{
		"app.log_level": false,
	}

	// EXECUTION PHASE
	registry := config.Registry()

	// ASSERTION PHASE
	// Mark keys as found
	for _, opt := range registry {
		if _, ok := docsKeys[opt.Key]; ok {
			docsKeys[opt.Key] = true
		}
		if _, ok := coreKeys[opt.Key]; ok {
			coreKeys[opt.Key] = true
		}
	}

	// Check that all docs keys were found
	for key, found := range docsKeys {
		if !found {
			t.Errorf("Registry() missing docs key %q", key)
		}
	}

	// Check that all core keys were found
	for key, found := range coreKeys {
		if !found {
			t.Errorf("Registry() missing core key %q", key)
		}
	}
}
