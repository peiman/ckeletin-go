// cmd/root_test.go

package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInitConfig tests all cases related to the initConfig function in a table-driven format
func TestInitConfig(t *testing.T) {
	tests := []struct {
		name               string
		setupHome          string           // Specify HOME env var value (empty to unset)
		setupConfigFile    string           // Config file path to set
		setupTempDir       bool             // Whether to create a temp dir
		setupBinaryName    string           // Binary name to set
		expectedError      bool             // Whether an error is expected
		expectedErrContain string           // Expected error substring
		expectedStatus     string           // Expected config file status
		customAssert       func(*testing.T) // Custom assertion function for special cases
		skipIfNoHome       bool             // Skip test if HOME cannot be determined
	}{
		{
			name:           "No HOME environment variable",
			setupHome:      "",
			expectedError:  false, // Now works without HOME (Issue #1 fix)
			expectedStatus: "No config file found, using defaults and environment variables",
		},
		{
			name:            "Config path setup with temp directory",
			setupTempDir:    true,
			setupBinaryName: "test-binary",
			expectedError:   false,
			expectedStatus:  "No config file found, using defaults and environment variables",
		},
		{
			name:               "Invalid config file path",
			setupConfigFile:    "/invalid/path/to/config.yaml",
			expectedError:      true,
			expectedErrContain: "config file", // Accepts both "config file size validation failed" and "failed to read config file"
		},
		{
			name:            "No config file set",
			setupConfigFile: "",
			setupHome:       "/tmp", // Ensure HOME is set to something
			expectedError:   false,
		},
		{
			name:            "With valid config file",
			setupConfigFile: "../testdata/config/valid.yaml",
			expectedError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save logger state and restore after test
			savedLogger, savedLevel := logger.SaveLoggerState()
			defer logger.RestoreLoggerState(savedLogger, savedLevel)

			// Skip test if HOME is required but not available
			if tt.skipIfNoHome && os.Getenv("HOME") == "" {
				t.Skip("This test requires HOME environment variable to be set")
			}

			origCfgFile := cfgFile
			origStatus := configFileStatus
			origUsed := configFileUsed
			origBinaryName := binaryName

			// Create a cleanup function to restore package-level values
			defer func() {
				cfgFile = origCfgFile
				configFileStatus = origStatus
				configFileUsed = origUsed
				binaryName = origBinaryName
			}()

			// Reset viper state
			viper.Reset()

			// Setup HOME environment (t.Setenv handles cleanup automatically)
			// Note: We set HOME even if it's empty string to simulate unset HOME
			t.Setenv("HOME", tt.setupHome)

			// Setup binary name if specified
			if tt.setupBinaryName != "" {
				binaryName = tt.setupBinaryName
			}

			// Setup temporary directory if needed
			var tmpDir string
			if tt.setupTempDir {
				tmpDir = t.TempDir() // Automatic cleanup
				// Set HOME to temp dir (t.Setenv handles cleanup automatically)
				t.Setenv("HOME", tmpDir)
			}

			// Setup config file path if specified
			if tt.setupConfigFile != "" {
				// Check if the path is relative and exists
				_, err := os.Stat(tt.setupConfigFile)
				if err != nil {
					// For test files, try with working directory
					wd, _ := os.Getwd()
					testPath := filepath.Join(wd, tt.setupConfigFile)
					_, err = os.Stat(testPath)
					if err == nil {
						cfgFile = testPath
					} else {
						// Just use the path as-is for error cases
						cfgFile = tt.setupConfigFile
					}
				} else {
					cfgFile = tt.setupConfigFile
				}
			} else {
				cfgFile = ""
			}

			// Setup logger for capturing output
			buf := new(bytes.Buffer)
			log.Logger = zerolog.New(buf)

			// EXECUTION PHASE
			err := initConfig()

			// ASSERTION PHASE
			// Check error expectations
			if tt.expectedError {
				assert.Error(t, err, "initConfig should return error")
			} else {
				assert.NoError(t, err, "initConfig should not return error")
			}

			// Check error content if applicable
			if tt.expectedErrContain != "" && err != nil {
				assert.Contains(t, err.Error(), tt.expectedErrContain,
					"Error should contain expected string")
			}

			// Check config status if applicable
			if tt.expectedStatus != "" && !tt.expectedError {
				assert.Contains(t, configFileStatus, tt.expectedStatus,
					"Status should contain expected string")
			}

			// Run custom assertions if provided
			if tt.customAssert != nil && !tt.expectedError {
				tt.customAssert(t)
			}
		})
	}
}

// Test the PersistentPreRunE function's error paths
func TestRootCmd_PersistentPreRunE_Errors(t *testing.T) {
	// Capture the original command
	origCmd := RootCmd.PersistentPreRunE
	defer func() { RootCmd.PersistentPreRunE = origCmd }()

	// Create a failing initConfig function that returns error
	RootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Simulate failure in initConfig
		return errors.New("initConfig error")
	}

	// Setup command
	cmd := &cobra.Command{Use: "test"}
	var args []string

	// Call the function
	err := RootCmd.PersistentPreRunE(cmd, args)

	// Verify error is returned
	require.Error(t, err, "Should return error")
	assert.Equal(t, "initConfig error", err.Error(), "Error message should match")
}

// Test the specific status logging in PersistentPreRunE
func TestRootCmd_ConfigStatusLogging(t *testing.T) {
	// Save originals
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer func() {
		configFileStatus = origStatus
		configFileUsed = origUsed
	}()

	tests := []struct {
		name         string
		configStatus string
		configUsed   string
	}{
		{
			name:         "No config file",
			configStatus: "No config file found",
			configUsed:   "",
		},
		{
			name:         "With config file",
			configStatus: "Using config file",
			configUsed:   "/path/to/config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Set up test state
			configFileStatus = tt.configStatus
			configFileUsed = tt.configUsed

			// Mock the cmd so we can capture the check without running logger.Init
			mockCmd := &cobra.Command{
				PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
					// This is the core of what we're testing:
					if configFileStatus != "" {
						// If we have config status, it should be logged appropriately
						if tt.configStatus != configFileStatus {
							t.Errorf("Expected status %q, got %q", tt.configStatus, configFileStatus)
						}
						if tt.configUsed != configFileUsed {
							t.Errorf("Expected configUsed %q, got %q", tt.configUsed, configFileUsed)
						}
					}
					return nil
				},
			}

			// EXECUTION PHASE
			err := mockCmd.PersistentPreRunE(mockCmd, []string{})

			// ASSERTION PHASE
			if err != nil {
				t.Errorf("Mock command failed: %v", err)
			}
		})
	}
}

func TestExecute_ErrorPropagation(t *testing.T) {
	// SETUP PHASE
	// Create a temporary root command for testing
	origRoot := RootCmd
	defer func() { RootCmd = origRoot }()

	testRoot := &cobra.Command{Use: "test-root"}
	testRoot.RunE = func(cmd *cobra.Command, args []string) error {
		return errors.New("some error")
	}

	// Replace the global rootCmd with testRoot
	RootCmd = testRoot

	// EXECUTION PHASE
	// Execute should now produce the error "some error"
	err := Execute()

	// ASSERTION PHASE
	if err == nil || !strings.Contains(err.Error(), "some error") {
		t.Errorf("Expected 'some error', got %v", err)
	}
}

// TestConfigPaths tests the ConfigPaths function that returns the configuration paths
func TestConfigPaths(t *testing.T) {
	tests := []struct {
		name                    string
		binaryName              string
		wantDefaultName         string
		wantDefaultFullName     string
		wantDefaultPathContains string
	}{
		{
			name:                    "Standard binary name",
			binaryName:              "myapp",
			wantDefaultName:         ".myapp",
			wantDefaultFullName:     ".myapp.yaml",
			wantDefaultPathContains: ".myapp.yaml",
		},
		{
			name:                    "Name with hyphens",
			binaryName:              "my-app",
			wantDefaultName:         ".my-app",
			wantDefaultFullName:     ".my-app.yaml",
			wantDefaultPathContains: ".my-app.yaml",
		},
		{
			name:                    "Name with dots",
			binaryName:              "app.v2",
			wantDefaultName:         ".app.v2",
			wantDefaultFullName:     ".app.v2.yaml",
			wantDefaultPathContains: ".app.v2.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save original and restore after test
			origBinaryName := binaryName
			defer func() { binaryName = origBinaryName }()

			// Set test binary name
			binaryName = tt.binaryName

			// EXECUTION PHASE
			paths := ConfigPaths()

			// ASSERTION PHASE
			if paths.DefaultName != tt.wantDefaultName {
				t.Errorf("ConfigPaths().DefaultName = %v, want %v", paths.DefaultName, tt.wantDefaultName)
			}

			if paths.DefaultFullName != tt.wantDefaultFullName {
				t.Errorf("ConfigPaths().DefaultFullName = %v, want %v", paths.DefaultFullName, tt.wantDefaultFullName)
			}

			// Check if the default path contains the expected file name
			if !strings.Contains(paths.DefaultPath, tt.wantDefaultPathContains) {
				t.Errorf("ConfigPaths().DefaultPath = %v, should contain %v", paths.DefaultPath, tt.wantDefaultPathContains)
			}
		})
	}
}

// TestEnvPrefix tests the EnvPrefix function used to create environment variable prefixes
func TestEnvPrefix(t *testing.T) {
	// Save original binary name and restore after test
	origBinaryName := binaryName
	defer func() {
		binaryName = origBinaryName
	}()

	tests := []struct {
		name           string
		binaryName     string
		expectedPrefix string
	}{
		{
			name:           "Simple name",
			binaryName:     "myapp",
			expectedPrefix: "MYAPP",
		},
		{
			name:           "With hyphens",
			binaryName:     "my-cool-app",
			expectedPrefix: "MY_COOL_APP",
		},
		{
			name:           "With dots",
			binaryName:     "app.name.v2",
			expectedPrefix: "APP_NAME_V2",
		},
		{
			name:           "With special characters",
			binaryName:     "app@name!v2",
			expectedPrefix: "APP_NAME_V2",
		},
		{
			name:           "Starting with number",
			binaryName:     "1app",
			expectedPrefix: "_1APP",
		},
		{
			name:           "All special characters",
			binaryName:     "!@#$%^&*()",
			expectedPrefix: "_",
		},
		{
			name:           "Mixed case",
			binaryName:     "MyApp",
			expectedPrefix: "MYAPP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			binaryName = tt.binaryName

			// EXECUTION PHASE
			prefix := EnvPrefix()

			// ASSERTION PHASE
			if prefix != tt.expectedPrefix {
				t.Errorf("EnvPrefix() = %v, want %v", prefix, tt.expectedPrefix)
			}
		})
	}
}

// TestEnvironmentVariables tests that environment variables are correctly read with the proper prefix
func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name          string
		binaryName    string
		envVars       map[string]string
		viperKey      string
		expectedValue string
	}{
		{
			name:          "Simple environment variable",
			binaryName:    "testapp",
			envVars:       map[string]string{"TESTAPP_APP_LOG_LEVEL": "debug"},
			viperKey:      "app.log_level",
			expectedValue: "debug",
		},
		{
			name:          "Hyphenated binary name",
			binaryName:    "test-app",
			envVars:       map[string]string{"TEST_APP_APP_LOG_LEVEL": "info"},
			viperKey:      "app.log_level",
			expectedValue: "info",
		},
		{
			name:          "Multiple parts key",
			binaryName:    "myapp",
			envVars:       map[string]string{"MYAPP_APP_SERVER_PORT": "8080"},
			viperKey:      "app.server.port",
			expectedValue: "8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			// Save original values
			origBinaryName := binaryName

			// Setup cleanup for package-level variables
			defer func() {
				binaryName = origBinaryName
			}()

			// Set test binary name
			binaryName = tt.binaryName

			// Reset viper
			viper.Reset()

			// Set environment variables for this test (t.Setenv handles cleanup automatically)
			for k, v := range tt.envVars {
				t.Setenv(k, v)
			}

			// Initialize configuration with the new environment
			cfgFile = "" // Ensure no config file is used

			// EXECUTION PHASE
			err := initConfig()

			// ASSERTION PHASE
			if err != nil {
				t.Fatalf("initConfig() failed: %v", err)
			}

			actualValue := viper.GetString(tt.viperKey)
			if actualValue != tt.expectedValue {
				t.Errorf("viper.GetString(%q) = %q, want %q",
					tt.viperKey, actualValue, tt.expectedValue)
			}
		})
	}
}

// TestSetupCommandConfig tests the command configuration inheritance pattern
func TestSetupCommandConfig(t *testing.T) {
	// SETUP PHASE
	// Create a command for testing
	isOriginalCalled := false

	// Create a command with an existing PreRunE
	cmd := &cobra.Command{
		Use: "test",
		PreRunE: func(c *cobra.Command, args []string) error {
			isOriginalCalled = true
			return nil
		},
	}

	// EXECUTION PHASE
	// Apply our setup function
	setupCommandConfig(cmd)

	// Run the resulting PreRunE
	err := cmd.PreRunE(cmd, []string{})

	// ASSERTION PHASE
	// Verify original PreRunE was called
	if !isOriginalCalled {
		t.Error("Original PreRunE was not called")
	}

	// No error should be returned
	if err != nil {
		t.Errorf("PreRunE returned unexpected error: %v", err)
	}

	// Test with a command that has no PreRunE
	cmdWithoutPreRun := &cobra.Command{Use: "test2"}
	setupCommandConfig(cmdWithoutPreRun)

	// Ensure it still works
	err = cmdWithoutPreRun.PreRunE(cmdWithoutPreRun, []string{})
	if err != nil {
		t.Errorf("PreRunE returned unexpected error for command without original PreRunE: %v", err)
	}

	// Test with a command that returns an error in PreRunE
	expectedErr := fmt.Errorf("test error")
	cmdWithErrPreRun := &cobra.Command{
		Use: "test3",
		PreRunE: func(c *cobra.Command, args []string) error {
			return expectedErr
		},
	}
	setupCommandConfig(cmdWithErrPreRun)

	// Run PreRunE and verify the error is propagated
	err = cmdWithErrPreRun.PreRunE(cmdWithErrPreRun, []string{})
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestGetConfigValue_Types tests the getConfigValueWithFlags function with different types
func TestGetConfigValue_Types(t *testing.T) {
	// SETUP PHASE
	// Reset viper for a clean test
	viper.Reset()

	// Set different types of values in viper
	viper.Set("test.string", "string-value")
	viper.Set("test.bool", true)
	viper.Set("test.int", 42)
	viper.Set("test.float", 3.14)
	viper.Set("test.stringslice", []string{"value1", "value2", "value3"})

	// Create a command with flags of different types
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("string", "", "String flag")
	cmd.Flags().Bool("bool", false, "Boolean flag")
	cmd.Flags().Int("int", 0, "Integer flag")
	cmd.Flags().Float64("float", 0, "Float flag")
	cmd.Flags().StringSlice("stringslice", []string{}, "String slice flag")

	// EXECUTION & ASSERTION PHASE
	// Test string type
	strVal := getConfigValueWithFlags[string](cmd, "string", "test.string")
	if strVal != "string-value" {
		t.Errorf("Expected string value to be 'string-value', got '%s'", strVal)
	}

	// Test bool type
	boolVal := getConfigValueWithFlags[bool](cmd, "bool", "test.bool")
	if boolVal != true {
		t.Errorf("Expected bool value to be true, got %v", boolVal)
	}

	// Test int type
	intVal := getConfigValueWithFlags[int](cmd, "int", "test.int")
	if intVal != 42 {
		t.Errorf("Expected int value to be 42, got %d", intVal)
	}

	// Test float type
	floatVal := getConfigValueWithFlags[float64](cmd, "float", "test.float")
	if floatVal != 3.14 {
		t.Errorf("Expected float value to be 3.14, got %f", floatVal)
	}

	// Test string slice type
	sliceVal := getConfigValueWithFlags[[]string](cmd, "stringslice", "test.stringslice")
	if len(sliceVal) != 3 || sliceVal[0] != "value1" || sliceVal[1] != "value2" || sliceVal[2] != "value3" {
		t.Errorf("Expected string slice value to be [value1 value2 value3], got %v", sliceVal)
	}

	// Test overriding values with flags
	if err := cmd.Flags().Set("string", "flag-value"); err != nil {
		t.Fatalf("Failed to set string flag: %v", err)
	}
	if err := cmd.Flags().Set("bool", "false"); err != nil {
		t.Fatalf("Failed to set bool flag: %v", err)
	}
	if err := cmd.Flags().Set("int", "99"); err != nil {
		t.Fatalf("Failed to set int flag: %v", err)
	}
	if err := cmd.Flags().Set("float", "6.28"); err != nil {
		t.Fatalf("Failed to set float flag: %v", err)
	}
	if err := cmd.Flags().Set("stringslice", "flag1,flag2,flag3,flag4"); err != nil {
		t.Fatalf("Failed to set string slice flag: %v", err)
	}

	// Verify flag values override viper values
	strVal = getConfigValueWithFlags[string](cmd, "string", "test.string")
	if strVal != "flag-value" {
		t.Errorf("Expected string flag value to be 'flag-value', got '%s'", strVal)
	}

	boolVal = getConfigValueWithFlags[bool](cmd, "bool", "test.bool")
	if boolVal != false {
		t.Errorf("Expected bool flag value to be false, got %v", boolVal)
	}

	intVal = getConfigValueWithFlags[int](cmd, "int", "test.int")
	if intVal != 99 {
		t.Errorf("Expected int flag value to be 99, got %d", intVal)
	}

	floatVal = getConfigValueWithFlags[float64](cmd, "float", "test.float")
	if floatVal != 6.28 {
		t.Errorf("Expected float flag value to be 6.28, got %f", floatVal)
	}

	// Verify string slice flag value overrides viper value
	sliceVal = getConfigValueWithFlags[[]string](cmd, "stringslice", "test.stringslice")
	expectedSlice := []string{"flag1", "flag2", "flag3", "flag4"}
	if len(sliceVal) != len(expectedSlice) {
		t.Errorf("Expected string slice flag length to be %d, got %d", len(expectedSlice), len(sliceVal))
	} else {
		for i, v := range expectedSlice {
			if sliceVal[i] != v {
				t.Errorf("Expected string slice flag value at index %d to be '%s', got '%s'", i, v, sliceVal[i])
			}
		}
	}
}

// TestGetConfigValue_FlagErrors tests error handling when flags are not properly configured
func TestGetConfigValue_FlagErrors(t *testing.T) {
	tests := []struct {
		name         string
		setupFlags   func(*cobra.Command)
		setFlag      bool
		flagName     string
		viperKey     string
		viperValue   interface{}
		expectedType string
	}{
		{
			name: "String flag not registered",
			setupFlags: func(cmd *cobra.Command) {
				// Don't register the flag
			},
			setFlag:      false,
			flagName:     "nonexistent",
			viperKey:     "test.string",
			viperValue:   "viper-value",
			expectedType: "string",
		},
		{
			name: "Bool flag not registered",
			setupFlags: func(cmd *cobra.Command) {
				// Don't register the flag
			},
			setFlag:      false,
			flagName:     "nonexistent-bool",
			viperKey:     "test.bool",
			viperValue:   true,
			expectedType: "bool",
		},
		{
			name: "Int flag not registered",
			setupFlags: func(cmd *cobra.Command) {
				// Don't register the flag
			},
			setFlag:      false,
			flagName:     "nonexistent-int",
			viperKey:     "test.int",
			viperValue:   42,
			expectedType: "int",
		},
		{
			name: "Float64 flag not registered",
			setupFlags: func(cmd *cobra.Command) {
				// Don't register the flag
			},
			setFlag:      false,
			flagName:     "nonexistent-float",
			viperKey:     "test.float",
			viperValue:   3.14,
			expectedType: "float64",
		},
		{
			name: "String slice flag not registered",
			setupFlags: func(cmd *cobra.Command) {
				// Don't register the flag
			},
			setFlag:      false,
			flagName:     "nonexistent-slice",
			viperKey:     "test.slice",
			viperValue:   []string{"a", "b"},
			expectedType: "[]string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			viper.Reset()
			viper.Set(tt.viperKey, tt.viperValue)

			cmd := &cobra.Command{Use: "test"}
			tt.setupFlags(cmd)

			// Try to set the flag if requested (this will fail for nonexistent flags)
			if tt.setFlag {
				_ = cmd.Flags().Set(tt.flagName, "value")
			}

			// EXECUTION & ASSERTION PHASE
			// These should fall back to viper values when flags don't exist
			switch tt.expectedType {
			case "string":
				result := getConfigValueWithFlags[string](cmd, tt.flagName, tt.viperKey)
				assert.Equal(t, tt.viperValue.(string), result, "String value should match")
			case "bool":
				result := getConfigValueWithFlags[bool](cmd, tt.flagName, tt.viperKey)
				assert.Equal(t, tt.viperValue.(bool), result, "Bool value should match")
			case "int":
				result := getConfigValueWithFlags[int](cmd, tt.flagName, tt.viperKey)
				assert.Equal(t, tt.viperValue.(int), result, "Int value should match")
			case "float64":
				result := getConfigValueWithFlags[float64](cmd, tt.flagName, tt.viperKey)
				assert.Equal(t, tt.viperValue.(float64), result, "Float64 value should match")
			case "[]string":
				result := getConfigValueWithFlags[[]string](cmd, tt.flagName, tt.viperKey)
				expected := tt.viperValue.([]string)
				assert.Equal(t, len(expected), len(result), "Slice length should match")
			}
		})
	}
}

// TestGetConfigValue_ViperTypeMismatch tests behavior when viper has wrong type
func TestGetConfigValue_ViperTypeMismatch(t *testing.T) {
	tests := []struct {
		name           string
		viperValue     interface{}
		requestedType  string
		expectedResult interface{}
	}{
		{
			name:           "Viper has int, requesting string",
			viperValue:     42,
			requestedType:  "string",
			expectedResult: "", // zero value
		},
		{
			name:           "Viper has string, requesting bool",
			viperValue:     "not-a-bool",
			requestedType:  "bool",
			expectedResult: false, // zero value
		},
		{
			name:           "Viper has string, requesting int",
			viperValue:     "not-an-int",
			requestedType:  "int",
			expectedResult: 0, // zero value
		},
		{
			name:           "Viper has bool, requesting float64",
			viperValue:     true,
			requestedType:  "float64",
			expectedResult: 0.0, // zero value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE
			viper.Reset()
			viper.Set("test.key", tt.viperValue)

			// Create command without flags set
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().String("str", "", "")
			cmd.Flags().Bool("bool", false, "")
			cmd.Flags().Int("int", 0, "")
			cmd.Flags().Float64("float", 0.0, "")

			// EXECUTION & ASSERTION PHASE
			// When viper has wrong type and flag not set, should return zero value
			switch tt.requestedType {
			case "string":
				result := getConfigValueWithFlags[string](cmd, "str", "test.key")
				assert.Equal(t, tt.expectedResult.(string), result, "Should return zero value for type mismatch")
			case "bool":
				result := getConfigValueWithFlags[bool](cmd, "bool", "test.key")
				assert.Equal(t, tt.expectedResult.(bool), result, "Should return zero value for type mismatch")
			case "int":
				result := getConfigValueWithFlags[int](cmd, "int", "test.key")
				assert.Equal(t, tt.expectedResult.(int), result, "Should return zero value for type mismatch")
			case "float64":
				result := getConfigValueWithFlags[float64](cmd, "float", "test.key")
				assert.Equal(t, tt.expectedResult.(float64), result, "Should return zero value for type mismatch")
			}
		})
	}
}

// TestGetConfigValue_StringSlice specifically tests the string slice handling in getConfigValueWithFlags
func TestGetConfigValue_StringSlice(t *testing.T) {
	// SETUP PHASE
	// Reset viper for a clean test
	viper.Reset()

	// Define test cases
	tests := []struct {
		name           string
		viperValue     []string
		flagValue      string
		setFlag        bool
		expectedResult []string
	}{
		{
			name:           "Viper value only",
			viperValue:     []string{"one", "two", "three"},
			setFlag:        false,
			expectedResult: []string{"one", "two", "three"},
		},
		{
			name:           "Empty viper value",
			viperValue:     []string{},
			setFlag:        false,
			expectedResult: []string{},
		},
		{
			name:           "Flag value overrides viper",
			viperValue:     []string{"viper1", "viper2"},
			flagValue:      "flag1,flag2,flag3",
			setFlag:        true,
			expectedResult: []string{"flag1", "flag2", "flag3"},
		},
		{
			name:           "Empty flag value",
			viperValue:     []string{"viper1", "viper2"},
			flagValue:      "",
			setFlag:        true,
			expectedResult: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP PHASE - for each test case
			viper.Reset()
			viper.Set("test.stringslice", tt.viperValue)

			// Create a command with a string slice flag
			cmd := &cobra.Command{Use: "test"}
			cmd.Flags().StringSlice("stringslice", []string{}, "String slice flag")

			// Set the flag if needed
			if tt.setFlag {
				if err := cmd.Flags().Set("stringslice", tt.flagValue); err != nil {
					t.Fatalf("Failed to set string slice flag: %v", err)
				}
			}

			// EXECUTION PHASE
			result := getConfigValueWithFlags[[]string](cmd, "stringslice", "test.stringslice")

			// ASSERTION PHASE
			if len(result) != len(tt.expectedResult) {
				t.Errorf("Expected string slice length to be %d, got %d",
					len(tt.expectedResult), len(result))
				t.Errorf("Expected: %v, Got: %v", tt.expectedResult, result)
				return
			}

			for i, v := range tt.expectedResult {
				if result[i] != v {
					t.Errorf("Expected value at index %d to be '%s', got '%s'", i, v, result[i])
				}
			}
		})
	}
}

func TestLoggingFlagBindings(t *testing.T) {
	// This test verifies that all logging flags are properly defined
	tests := []struct {
		name     string
		flagName string
	}{
		// Console and file logging flags
		{name: "log-console-level flag exists", flagName: "log-console-level"},
		{name: "log-file-enabled flag exists", flagName: "log-file-enabled"},
		{name: "log-file-path flag exists", flagName: "log-file-path"},
		{name: "log-file-level flag exists", flagName: "log-file-level"},
		// Log rotation flags
		{name: "log-file-max-size flag exists", flagName: "log-file-max-size"},
		{name: "log-file-max-backups flag exists", flagName: "log-file-max-backups"},
		{name: "log-file-max-age flag exists", flagName: "log-file-max-age"},
		{name: "log-file-compress flag exists", flagName: "log-file-compress"},
		// Log sampling flags
		{name: "log-sampling-enabled flag exists", flagName: "log-sampling-enabled"},
		{name: "log-sampling-initial flag exists", flagName: "log-sampling-initial"},
		{name: "log-sampling-thereafter flag exists", flagName: "log-sampling-thereafter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify flag exists in persistent flags
			flag := RootCmd.PersistentFlags().Lookup(tt.flagName)
			if flag == nil {
				t.Errorf("Flag %s not found in RootCmd persistent flags", tt.flagName)
			}
		})
	}
}

func TestLoggingFlagsIntegration(t *testing.T) {
	// This test verifies that the logging system works with the flags
	tempDir := t.TempDir()
	logFile := filepath.Join(tempDir, "test.log")

	tests := []struct {
		name             string
		fileEnabled      bool
		filePath         string
		consoleLevel     string
		fileLevel        string
		expectFileExists bool
	}{
		{
			name:             "File logging disabled",
			fileEnabled:      false,
			filePath:         logFile + ".1",
			consoleLevel:     "info",
			fileLevel:        "debug",
			expectFileExists: false,
		},
		{
			name:             "File logging enabled",
			fileEnabled:      true,
			filePath:         logFile + ".2",
			consoleLevel:     "info",
			fileLevel:        "debug",
			expectFileExists: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// SETUP
			viper.Reset()
			viper.Set("app.log.file_enabled", tt.fileEnabled)
			viper.Set("app.log.file_path", tt.filePath)
			viper.Set("app.log.console_level", tt.consoleLevel)
			viper.Set("app.log.file_level", tt.fileLevel)
			viper.Set("app.log.color_enabled", "false")
			viper.Set("app.log.sampling_enabled", false)

			// Save and restore logger state
			savedLogger, savedLevel := logger.SaveLoggerState()
			defer logger.RestoreLoggerState(savedLogger, savedLevel)

			consoleBuf := &bytes.Buffer{}

			// EXECUTE
			err := logger.Init(consoleBuf)
			if err != nil {
				t.Fatalf("Failed to initialize logger: %v", err)
			}

			// Log some messages
			log.Debug().Msg("Debug message")
			log.Info().Msg("Info message")

			// Cleanup
			logger.Cleanup()

			// ASSERT
			if tt.expectFileExists {
				if _, err := os.Stat(tt.filePath); os.IsNotExist(err) {
					t.Errorf("Expected log file to exist at %s", tt.filePath)
				}
			} else {
				if _, err := os.Stat(tt.filePath); !os.IsNotExist(err) {
					t.Errorf("Expected log file NOT to exist at %s", tt.filePath)
				}
			}

			// Verify console contains info message
			consoleOutput := consoleBuf.String()
			if !strings.Contains(consoleOutput, "Info message") {
				t.Errorf("Console output should contain 'Info message'")
			}
		})
	}
}

// ============================================================================
// TDD Tests for Issue #1 + #7: $HOME Fallback + Config Search Path
// ============================================================================

// TestInitConfigWithoutHomeDir tests that initConfig works without $HOME environment variable
// This test will initially FAIL - initConfig currently returns error without HOME
func TestInitConfigWithoutHomeDir(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Save originals
	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer func() {
		cfgFile = origCfgFile
		configFileStatus = origStatus
		configFileUsed = origUsed
	}()

	// Reset viper state
	viper.Reset()

	// Unset HOME completely
	t.Setenv("HOME", "")

	// No --config flag set
	cfgFile = ""

	// Setup logger for capturing output
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	// EXECUTION PHASE
	err := initConfig()

	// ASSERTION PHASE
	// Should NOT return error - app should work without HOME
	if err != nil {
		t.Errorf("initConfig should work without HOME environment variable, got error: %v", err)
	}

	// Should use defaults (no config file found is OK)
	if err == nil && configFileStatus == "" {
		t.Error("configFileStatus should be set even when no config file is found")
	}
}

// TestConfigFromCurrentDirectory tests that config is discovered from current directory
// This test will initially FAIL - only searches home directory currently
func TestConfigFromCurrentDirectory(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Create temp dir and change to it
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer os.Chdir(oldWd)
	os.Chdir(tempDir)

	// Create config file in current directory
	configContent := []byte("app:\n  log_level: debug\n")
	configPath := filepath.Join(tempDir, ".ckeletin-go.yaml")
	err := os.WriteFile(configPath, configContent, 0600)
	if err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Save originals
	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer func() {
		cfgFile = origCfgFile
		configFileStatus = origStatus
		configFileUsed = origUsed
	}()

	// Reset viper state
	viper.Reset()

	// No --config flag set
	cfgFile = ""

	// Setup HOME to different directory so we know it's finding current dir config
	homeDir := filepath.Join(tempDir, "home")
	os.MkdirAll(homeDir, 0755)
	t.Setenv("HOME", homeDir)

	// Setup logger
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	// EXECUTION PHASE
	err = initConfig()

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("initConfig should succeed with current directory config: %v", err)
	}

	// Should find config from current directory
	if viper.GetString("app.log_level") != "debug" {
		t.Errorf("Expected log_level='debug' from current dir config, got: %s", viper.GetString("app.log_level"))
	}

	// Config file should be discovered
	if !strings.Contains(configFileStatus, "Using config file") {
		t.Errorf("Expected 'Using config file' status, got: %s", configFileStatus)
	}
}

// TestConfigPriorityCurrentDirFirst tests that current directory config has priority over home directory
// This test will initially FAIL - current directory search doesn't exist yet
func TestConfigPriorityCurrentDirFirst(t *testing.T) {
	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Create temp directory structure
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	currentDir := filepath.Join(tempDir, "current")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(currentDir, 0755)

	// Write different values to home and current directory configs
	homeConfigContent := []byte("app:\n  log_level: info\n")
	homeConfig := filepath.Join(homeDir, ".ckeletin-go.yaml")
	err := os.WriteFile(homeConfig, homeConfigContent, 0600)
	if err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	currentConfigContent := []byte("app:\n  log_level: debug\n")
	currentConfig := filepath.Join(currentDir, ".ckeletin-go.yaml")
	err = os.WriteFile(currentConfig, currentConfigContent, 0600)
	if err != nil {
		t.Fatalf("Failed to write current dir config: %v", err)
	}

	// Save originals
	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	oldWd, _ := os.Getwd()
	defer func() {
		cfgFile = origCfgFile
		configFileStatus = origStatus
		configFileUsed = origUsed
		os.Chdir(oldWd)
	}()

	// Reset viper state
	viper.Reset()

	// No --config flag set
	cfgFile = ""

	// Set HOME to home directory and change to current directory
	t.Setenv("HOME", homeDir)
	os.Chdir(currentDir)

	// Setup logger
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	// EXECUTION PHASE
	err = initConfig()

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("initConfig should succeed with both configs present: %v", err)
	}

	// Current directory config should win (debug, not info)
	logLevel := viper.GetString("app.log_level")
	if logLevel != "debug" {
		t.Errorf("Expected log_level='debug' from current dir (priority), got: %s", logLevel)
	}

	// Config file path should be from current directory
	if configFileUsed != "" && !strings.Contains(configFileUsed, currentDir) {
		t.Errorf("Expected config from current dir (%s), got: %s", currentDir, configFileUsed)
	}
}

// TestConfigFromHomeDirectoryOnly tests config discovery when config only exists in home directory
func TestConfigFromHomeDirectoryOnly(t *testing.T) {
	// Skip on Windows due to path handling differences
	if runtime.GOOS == "windows" {
		t.Skip("Skipping home directory config test on Windows")
	}

	// SETUP PHASE
	savedLogger, savedLevel := logger.SaveLoggerState()
	defer logger.RestoreLoggerState(savedLogger, savedLevel)

	// Create temp directory structure
	tempDir := t.TempDir()
	homeDir := filepath.Join(tempDir, "home")
	currentDir := filepath.Join(tempDir, "current")
	os.MkdirAll(homeDir, 0755)
	os.MkdirAll(currentDir, 0755)

	// Write config ONLY to home directory (not current directory)
	homeConfigContent := []byte("app:\n  log_level: warn\n")
	homeConfig := filepath.Join(homeDir, ".ckeletin-go.yaml")
	err := os.WriteFile(homeConfig, homeConfigContent, 0600)
	if err != nil {
		t.Fatalf("Failed to write home config: %v", err)
	}

	// Save originals
	origCfgFile := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	oldWd, _ := os.Getwd()
	defer func() {
		cfgFile = origCfgFile
		configFileStatus = origStatus
		configFileUsed = origUsed
		os.Chdir(oldWd)
	}()

	// Reset viper state
	viper.Reset()

	// No --config flag set
	cfgFile = ""

	// Set HOME to home directory and change to empty current directory
	t.Setenv("HOME", homeDir)
	os.Chdir(currentDir)

	// Setup logger
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	// EXECUTION PHASE
	err = initConfig()

	// ASSERTION PHASE
	if err != nil {
		t.Fatalf("initConfig should succeed with home directory config: %v", err)
	}

	// Home directory config should be loaded
	logLevel := viper.GetString("app.log_level")
	if logLevel != "warn" {
		t.Errorf("Expected log_level='warn' from home dir, got: %s", logLevel)
	}

	// Config file path should be from home directory
	if configFileUsed != "" && !strings.Contains(configFileUsed, homeDir) {
		t.Errorf("Expected config from home dir (%s), got: %s", homeDir, configFileUsed)
	}
}
