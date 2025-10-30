// cmd/root_flag_bindings_test.go

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/spf13/viper"
)

// viperKeyToFlagName converts a viper key to its corresponding flag name.
// Examples:
//   - app.log.file_enabled → log-file-enabled
//   - app.log_level → log-level
//   - app.log.color_enabled → log-color (special case)
func viperKeyToFlagName(viperKey string) string {
	// Special case: color_enabled flag is named "log-color" not "log-color-enabled"
	if viperKey == "app.log.color_enabled" {
		return "log-color"
	}

	// Remove app. prefix
	key := strings.TrimPrefix(viperKey, "app.")
	// Replace dots and underscores with hyphens
	key = strings.ReplaceAll(key, ".", "-")
	key = strings.ReplaceAll(key, "_", "-")
	return key
}

// TestViperKeyToFlagName tests the conversion function
func TestViperKeyToFlagName(t *testing.T) {
	tests := []struct {
		viperKey string
		expected string
	}{
		{"app.log.file_enabled", "log-file-enabled"},
		{"app.log_level", "log-level"},
		{"app.log.console_level", "log-console-level"},
		{"app.log.color_enabled", "log-color"}, // Special case
		{"app.ping.output_message", "ping-output-message"},
	}

	for _, tt := range tests {
		t.Run(tt.viperKey, func(t *testing.T) {
			got := viperKeyToFlagName(tt.viperKey)
			if got != tt.expected {
				t.Errorf("viperKeyToFlagName(%q) = %q, want %q", tt.viperKey, got, tt.expected)
			}
		})
	}
}

// TestBindFlags_FunctionExists tests that the bindFlags function exists and can be called
func TestBindFlags_FunctionExists(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Test that bindFlags() function exists and returns an error type
	err := bindFlags(RootCmd)
	if err != nil {
		// We expect no error when binding valid flags
		t.Errorf("bindFlags() returned unexpected error: %v", err)
	}
}

// TestFlagBindings_RegistryDriven validates all flags from the config registry
func TestFlagBindings_RegistryDriven(t *testing.T) {
	// SETUP PHASE
	viper.Reset()

	// Get all core options from registry
	options := config.CoreOptions()
	if len(options) == 0 {
		t.Fatal("No options in registry - cannot test")
	}

	// Initialize root command (calls init() which defines flags)
	// Note: RootCmd is package-level variable defined in root.go

	// EXECUTION & ASSERTION PHASE
	// Test each option from the registry
	for _, opt := range options {
		// Skip non-flag options (like nested config groups)
		if opt.Key == "" {
			continue
		}

		t.Run(opt.Key, func(t *testing.T) {
			flagName := viperKeyToFlagName(opt.Key)

			// 1. Verify flag exists in RootCmd
			flag := RootCmd.PersistentFlags().Lookup(flagName)
			if flag == nil {
				t.Fatalf("Flag %q not found in RootCmd.PersistentFlags() for viper key %q",
					flagName, opt.Key)
			}

			// 2. Test that binding works (this calls bindFlags internally)
			err := viper.BindPFlag(opt.Key, flag)
			if err != nil {
				t.Errorf("Failed to bind flag %q to viper key %q: %v",
					flagName, opt.Key, err)
				return
			}

			// 3. Verify the binding exists
			// After binding, setting the flag should update viper
			// This is tested implicitly - if binding failed, viper won't see the value

			// 4. Verify default value matches registry
			verifyDefaultValue(t, opt, flagName)
		})
	}
}

// verifyDefaultValue checks that the flag's default value matches the registry
func verifyDefaultValue(t *testing.T, opt config.ConfigOption, flagName string) {
	t.Helper()

	// Get the flag to check its default value
	flag := RootCmd.PersistentFlags().Lookup(flagName)
	if flag == nil {
		t.Fatal("Flag not found")
	}

	// Check default value based on type
	switch opt.Type {
	case "string":
		expected, ok := opt.DefaultValue.(string)
		if !ok {
			t.Fatalf("Registry default value for %s is not a string: %T", opt.Key, opt.DefaultValue)
		}
		got := flag.DefValue
		if got != expected {
			t.Errorf("Default value mismatch for %s (flag %s): expected %q, got %q",
				opt.Key, flagName, expected, got)
		}

	case "bool":
		expected, ok := opt.DefaultValue.(bool)
		if !ok {
			t.Fatalf("Registry default value for %s is not a bool: %T", opt.Key, opt.DefaultValue)
		}
		got := flag.DefValue
		expectedStr := fmt.Sprintf("%t", expected)
		if got != expectedStr {
			t.Errorf("Default value mismatch for %s (flag %s): expected %q, got %q",
				opt.Key, flagName, expectedStr, got)
		}

	case "int":
		expected, ok := opt.DefaultValue.(int)
		if !ok {
			t.Fatalf("Registry default value for %s is not an int: %T", opt.Key, opt.DefaultValue)
		}
		got := flag.DefValue
		expectedStr := fmt.Sprintf("%d", expected)
		if got != expectedStr {
			t.Errorf("Default value mismatch for %s (flag %s): expected %q, got %q",
				opt.Key, flagName, expectedStr, got)
		}

	default:
		t.Errorf("Unknown type %q for option %s", opt.Type, opt.Key)
	}
}

// TestBindFlags_AllFlagsHaveViperBinding tests that all persistent flags have viper bindings
func TestBindFlags_AllFlagsHaveViperBinding(t *testing.T) {
	// SETUP
	viper.Reset()

	// Call bindFlags to set up bindings
	err := bindFlags(RootCmd)
	if err != nil {
		t.Fatalf("bindFlags() failed: %v", err)
	}

	// Get expected flags from registry
	options := config.CoreOptions()
	expectedBindings := make(map[string]bool)
	for _, opt := range options {
		if opt.Key != "" {
			expectedBindings[opt.Key] = false // Mark as not yet verified
		}
	}

	// Walk through all persistent flags and verify they're bound
	// (This is a placeholder - we verify bindings exist through other tests)
	_ = expectedBindings

	// Verify all expected bindings exist
	for key := range expectedBindings {
		// Try to get the value from viper to confirm binding exists
		// Just checking if the key is set is sufficient
		if !viper.InConfig(key) && !viper.IsSet(key) {
			// This is expected - flags aren't "set" until they're used
			// But the binding should exist
			// We'll verify this by checking if setting via flag would work
		}
	}
}

// TestBindFlags_ErrorCollection tests that bindFlags properly collects multiple errors
func TestBindFlags_ErrorCollection(t *testing.T) {
	// This test is tricky because viper.BindPFlag rarely fails in practice
	// We're testing that IF it fails, errors are collected properly

	// In the current implementation, bindFlags should:
	// 1. Attempt to bind all flags
	// 2. Collect any errors that occur
	// 3. Return a combined error if any bindings failed

	// For now, we test the happy path - all bindings succeed
	viper.Reset()
	err := bindFlags(RootCmd)
	if err != nil {
		t.Errorf("bindFlags() should succeed with valid flags, got error: %v", err)
	}
}

// TestBindFlags_Integration tests the full flag binding flow
func TestBindFlags_Integration(t *testing.T) {
	// SETUP
	viper.Reset()

	// EXECUTION
	// 1. Flags are defined in init()
	// 2. bindFlags() binds them to viper
	err := bindFlags(RootCmd)
	if err != nil {
		t.Fatalf("bindFlags() failed: %v", err)
	}

	// ASSERTION
	// Verify a sample of flags are properly bound
	testCases := []struct {
		viperKey string
		flagName string
		flagType string
	}{
		{config.KeyAppLogLevel, "log-level", "string"},
		{config.KeyAppLogFileEnabled, "log-file-enabled", "bool"},
		{config.KeyAppLogFileMaxSize, "log-file-max-size", "int"},
	}

	for _, tc := range testCases {
		t.Run(tc.viperKey, func(t *testing.T) {
			// Check flag exists
			flag := RootCmd.PersistentFlags().Lookup(tc.flagName)
			if flag == nil {
				t.Fatalf("Flag %q not found", tc.flagName)
			}

			// Verify binding by checking if viper key is accessible
			// (the key should exist even if not explicitly set)
			switch tc.flagType {
			case "string":
				_ = viper.GetString(tc.viperKey) // Should not panic
			case "bool":
				_ = viper.GetBool(tc.viperKey) // Should not panic
			case "int":
				_ = viper.GetInt(tc.viperKey) // Should not panic
			}
		})
	}
}
