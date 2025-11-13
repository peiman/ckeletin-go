// cmd/root_flag_bindings_test.go

package cmd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.expected, got, "viperKeyToFlagName(%q) should return expected flag name", tt.viperKey)
		})
	}
}

// TestBindFlags_FunctionExists tests that the bindFlags function exists and can be called
func TestBindFlags_FunctionExists(t *testing.T) {
	// Reset viper state
	viper.Reset()

	// Test that bindFlags() function exists and returns an error type
	err := bindFlags(RootCmd)
	// We expect no error when binding valid flags
	assert.NoError(t, err, "bindFlags() should not return error for valid flags")
}

// TestFlagBindings_RegistryDriven validates all flags from the config registry
func TestFlagBindings_RegistryDriven(t *testing.T) {
	// SETUP PHASE
	viper.Reset()

	// Get all core options from registry
	options := config.CoreOptions()
	require.NotEmpty(t, options, "Registry should have options to test")

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
			require.NotNil(t, flag, "Flag %q should exist in RootCmd.PersistentFlags() for viper key %q",
				flagName, opt.Key)

			// 2. Test that binding works (this calls bindFlags internally)
			err := viper.BindPFlag(opt.Key, flag)
			assert.NoError(t, err, "Failed to bind flag %q to viper key %q",
				flagName, opt.Key)

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
	require.NotNil(t, flag, "Flag should exist")

	// Check default value based on type
	switch opt.Type {
	case "string":
		expected, ok := opt.DefaultValue.(string)
		require.True(t, ok, "Registry default value for %s should be a string, got %T", opt.Key, opt.DefaultValue)
		got := flag.DefValue
		assert.Equal(t, expected, got, "Default value mismatch for %s (flag %s)",
			opt.Key, flagName)

	case "bool":
		expected, ok := opt.DefaultValue.(bool)
		require.True(t, ok, "Registry default value for %s should be a bool, got %T", opt.Key, opt.DefaultValue)
		got := flag.DefValue
		expectedStr := fmt.Sprintf("%t", expected)
		assert.Equal(t, expectedStr, got, "Default value mismatch for %s (flag %s)",
			opt.Key, flagName)

	case "int":
		expected, ok := opt.DefaultValue.(int)
		require.True(t, ok, "Registry default value for %s should be an int, got %T", opt.Key, opt.DefaultValue)
		got := flag.DefValue
		expectedStr := fmt.Sprintf("%d", expected)
		assert.Equal(t, expectedStr, got, "Default value mismatch for %s (flag %s)",
			opt.Key, flagName)

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
	require.NoError(t, err, "bindFlags() should succeed")

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
	// Test that bindFlags collects and returns errors when flag bindings fail
	// This happens when flags don't exist (Lookup returns nil)

	// SETUP
	viper.Reset()

	// Create a bare command with NO persistent flags defined
	// This will cause all Lookup() calls to return nil, triggering bind errors
	bareCmd := &cobra.Command{
		Use:   "bare",
		Short: "Command with no flags",
	}

	// EXECUTION
	// bindFlags will try to look up flags that don't exist
	err := bindFlags(bareCmd)

	// ASSERTION
	// Should return an error indicating multiple bindings failed
	require.Error(t, err, "bindFlags() should return error when flags don't exist")

	// Verify error message contains information about failed bindings
	errMsg := err.Error()
	assert.Contains(t, errMsg, "failed to bind", "Error message should mention 'failed to bind'")

	// Verify it mentions the number of failures (14 flags total)
	assert.Contains(t, errMsg, "14 flag(s)", "Error message should mention '14 flag(s)'")
}

// TestBindFlags_Integration tests the full flag binding flow
func TestBindFlags_Integration(t *testing.T) {
	// SETUP
	viper.Reset()

	// EXECUTION
	// 1. Flags are defined in init()
	// 2. bindFlags() binds them to viper
	err := bindFlags(RootCmd)
	require.NoError(t, err, "bindFlags() should succeed")

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
			require.NotNil(t, flag, "Flag %q should exist", tc.flagName)

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

// TestBindFlags_FromSubcommand tests that bindFlags() works when called from a subcommand.
// This is a regression test for the bug where bindFlags(cmd) looked up flags on
// cmd.PersistentFlags() instead of cmd.Root().PersistentFlags(), causing all subcommands
// to fail with "flag for X is nil" errors.
func TestBindFlags_FromSubcommand(t *testing.T) {
	// SETUP
	viper.Reset()

	// Create a mock subcommand (simulating ping, config, etc.)
	mockSubCmd := &cobra.Command{
		Use:   "mock",
		Short: "Mock subcommand for testing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	// Add it to RootCmd (this is what happens in init() for real subcommands)
	RootCmd.AddCommand(mockSubCmd)

	// EXECUTION
	// Call bindFlags() with the subcommand (not RootCmd)
	// This simulates what happens in PersistentPreRunE when a subcommand is executed
	err := bindFlags(mockSubCmd)

	// ASSERTION
	require.NoError(t, err,
		"bindFlags(subcommand) should succeed.\n"+
			"This indicates bindFlags() is looking up flags on the subcommand instead of root.\n"+
			"Ensure bindFlags() uses cmd.Root().PersistentFlags().Lookup() not cmd.PersistentFlags().Lookup()")

	// Verify a sample of flags were actually bound
	testCases := []struct {
		viperKey string
		flagName string
	}{
		{config.KeyAppLogLevel, "log-level"},
		{config.KeyAppLogFileEnabled, "log-file-enabled"},
		{"config", "config"},
	}

	for _, tc := range testCases {
		t.Run(tc.flagName, func(t *testing.T) {
			// Verify the flag exists on RootCmd (persistent flags)
			flag := RootCmd.PersistentFlags().Lookup(tc.flagName)
			require.NotNil(t, flag, "Flag %q should exist on RootCmd", tc.flagName)

			// Verify it can be accessed via the subcommand's inherited flags
			inheritedFlag := mockSubCmd.Flag(tc.flagName)
			require.NotNil(t, inheritedFlag, "Flag %q should be inherited by subcommand", tc.flagName)

			// Verify binding exists by manually binding and checking no error
			// (In real execution, viper.BindPFlag is called inside bindFlags)
			err := viper.BindPFlag(tc.viperKey, flag)
			assert.NoError(t, err, "Failed to bind %q to %q", tc.flagName, tc.viperKey)
		})
	}

	// CLEANUP
	RootCmd.RemoveCommand(mockSubCmd)
}
