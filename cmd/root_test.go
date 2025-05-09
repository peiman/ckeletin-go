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
