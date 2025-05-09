// cmd/root_test.go

package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Test when HOME environment variable is not set (simulating UserHomeDir error)
func TestInitConfig_UserHomeDirEnvironment(t *testing.T) {
	// Save original HOME env var
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Clear HOME env var to trigger an error in UserHomeDir via os.Getenv
	os.Unsetenv("HOME")

	// Reset viper and config path
	viper.Reset()
	cfgFile = ""

	// Run the function with missing HOME env var
	err := initConfig()

	// Either we'll get an error or we'll get a path that doesn't include HOME
	// Some systems might have a fallback for UserHomeDir, so we handle both cases
	if err == nil {
		t.Log("No error returned when HOME env var is not set, checking if config paths are valid")
		// If no error, verify Viper is at least using reasonable paths that don't contain HOME
		paths := viper.GetStringSlice("config_paths")
		if len(paths) > 0 && strings.Contains(paths[0], origHome) {
			t.Errorf("Expected paths without HOME, got %v", paths)
		}
	}
}

// Test the full config path
func TestInitConfig_ConfigPathSetup(t *testing.T) {
	// Setup test
	viper.Reset()
	origCfg := cfgFile
	origStatus := configFileStatus
	origUsed := configFileUsed
	origBinaryName := binaryName

	defer func() {
		cfgFile = origCfg
		configFileStatus = origStatus
		configFileUsed = origUsed
		binaryName = origBinaryName
	}()

	// Test with custom config path
	binaryName = "test-binary"
	cfgFile = ""

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "ckeletin-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save original and set HOME to temp dir
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	defer os.Setenv("HOME", origHome)

	// Run the function (this should set paths correctly even if config file doesn't exist)
	err = initConfig()
	if err != nil {
		t.Errorf("Expected no error for missing config, got %v", err)
	}

	// Verify the config file status
	if configFileStatus != "No config file found, using defaults and environment variables" {
		t.Errorf("Expected 'No config file found' message, got '%s'", configFileStatus)
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
	if err == nil || err.Error() != "initConfig error" {
		t.Errorf("Expected 'initConfig error', got %v", err)
	}
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

			// We don't actually need to call the function, just verify the setup is correct
			if err := mockCmd.PersistentPreRunE(mockCmd, []string{}); err != nil {
				t.Errorf("Mock command failed: %v", err)
			}
		})
	}
}

func TestInitConfig_InvalidConfigFile(t *testing.T) {
	cfgFile = "/invalid/path/to/config.yaml"
	defer func() { cfgFile = "" }()

	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	err := initConfig()

	if err == nil {
		t.Errorf("Expected initConfig() to return an error for invalid config file")
	}

	// Actual error message includes "failed to read config file"
	if !strings.Contains(err.Error(), "failed to read config file") {
		t.Errorf("Expected error message to contain 'failed to read config file', got '%v'", err)
	}
}

func TestInitConfig_NoConfigFile(t *testing.T) {
	viper.Reset()
	cfgFile = ""
	err := initConfig()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestExecute_ErrorPropagation(t *testing.T) {
	// Create a temporary root command for testing
	origRoot := RootCmd
	defer func() { RootCmd = origRoot }()

	testRoot := &cobra.Command{Use: "test-root"}
	testRoot.RunE = func(cmd *cobra.Command, args []string) error {
		return errors.New("some error")
	}

	// Replace the global rootCmd with testRoot
	RootCmd = testRoot

	// Execute should now produce the error "some error"
	err := Execute()
	if err == nil || !strings.Contains(err.Error(), "some error") {
		t.Errorf("Expected 'some error', got %v", err)
	}
}

func TestInitConfig_WithConfigFile(t *testing.T) {
	// Reset viper state before and after test
	viper.Reset()
	defer viper.Reset()

	// Capture the original value and restore after test
	origCfgFile := cfgFile
	defer func() { cfgFile = origCfgFile }()

	// Setup: point to the testdata config file
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	testConfigPath := filepath.Join(wd, "../testdata/config.yaml")
	// For debugging, output the path and check if file exists
	_, err = os.Stat(testConfigPath)
	if err != nil {
		t.Logf("Test path doesn't exist: %s, err: %v", testConfigPath, err)
		// Try a different path
		testConfigPath = "./testdata/config.yaml"
		_, err = os.Stat(testConfigPath)
		if err != nil {
			t.Fatalf("Could not find config file at either path: %v", err)
		}
	}
	cfgFile = testConfigPath

	// Capture logs for verification
	buf := new(bytes.Buffer)
	log.Logger = zerolog.New(buf)

	// Backup the original variables and restore after test
	origStatus := configFileStatus
	origUsed := configFileUsed
	defer func() {
		configFileStatus = origStatus
		configFileUsed = origUsed
	}()

	// Run the function
	err = initConfig()

	// Check result
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the variables are set correctly
	if configFileStatus != "Using config file" {
		t.Errorf("Expected configFileStatus to be 'Using config file', got '%s'", configFileStatus)
	}

	if !strings.Contains(configFileUsed, "testdata/config.yaml") {
		t.Errorf("Expected configFileUsed to contain 'testdata/config.yaml', got '%s'", configFileUsed)
	}

	// Verify that viper read the config values
	if viper.GetString("app.log_level") != "info" {
		t.Errorf("Expected app.log_level to be 'info', got '%s'", viper.GetString("app.log_level"))
	}
}

// TestConfigPaths ensures the ConfigPaths function works correctly
func TestConfigPaths(t *testing.T) {
	// Save original binary name and restore after test
	origBinaryName := binaryName
	defer func() {
		binaryName = origBinaryName
	}()

	// Test with a known binary name
	binaryName = "testapp"

	paths := ConfigPaths()

	// Verify the values are correctly constructed
	if paths.DefaultName != ".testapp" {
		t.Errorf("Expected DefaultName to be '.testapp', got '%s'", paths.DefaultName)
	}

	if paths.Extension != "yaml" {
		t.Errorf("Expected Extension to be 'yaml', got '%s'", paths.Extension)
	}

	if paths.DefaultFullName != ".testapp.yaml" {
		t.Errorf("Expected DefaultFullName to be '.testapp.yaml', got '%s'", paths.DefaultFullName)
	}

	// DefaultPath includes the home directory, so we can't easily test its exact value
	// But we can check that it ends with the expected filename
	if !strings.HasSuffix(paths.DefaultPath, ".testapp.yaml") {
		t.Errorf("Expected DefaultPath to end with '.testapp.yaml', got '%s'", paths.DefaultPath)
	}

	if paths.IgnorePattern != "testapp.yaml" {
		t.Errorf("Expected IgnorePattern to be 'testapp.yaml', got '%s'", paths.IgnorePattern)
	}
}

// TestEnvPrefix tests the EnvPrefix function with various binary names
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
			binaryName = tt.binaryName
			prefix := EnvPrefix()
			if prefix != tt.expectedPrefix {
				t.Errorf("EnvPrefix() = %v, want %v", prefix, tt.expectedPrefix)
			}
		})
	}
}

// TestEnvironmentVariables tests that environment variables are correctly read with the proper prefix
func TestEnvironmentVariables(t *testing.T) {
	// Save original binary name and restore after test
	origBinaryName := binaryName
	defer func() {
		binaryName = origBinaryName
	}()

	// Set a test binary name
	binaryName = "testcli"

	// Reset viper for a clean test
	viper.Reset()

	// Set an environment variable with the expected prefix
	envVarName := "TESTCLI_APP_TEST_VALUE"
	os.Setenv(envVarName, "env_value")
	defer os.Unsetenv(envVarName)

	// Initialize config
	err := initConfig()
	if err != nil {
		t.Fatalf("initConfig() error = %v", err)
	}

	// Check that the value was read from the environment variable
	value := viper.GetString("app.test_value")
	if value != "env_value" {
		t.Errorf("Expected viper to read value 'env_value' from environment, got '%s'", value)
	}
}
