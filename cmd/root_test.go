// cmd/root_test.go

package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestInitConfig_Defaults(t *testing.T) {
	// Reset Viper to ensure a clean state
	viper.Reset()
	cfgFile = ""

	// Initialize configuration
	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() failed: %v", err)
	}

	// Assert default log level
	if viper.GetString("app.log_level") != "info" {
		t.Errorf("Expected 'info', got '%s'", viper.GetString("app.log_level"))
	}
}

func TestInitConfig_EnvironmentOverride(t *testing.T) {
	// Set environment variable to override log level
	os.Setenv("APP_LOG_LEVEL", "debug")
	defer os.Unsetenv("APP_LOG_LEVEL")

	viper.Reset()
	cfgFile = ""

	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() failed: %v", err)
	}

	// Assert that the environment variable took precedence
	if viper.GetString("app.log_level") != "debug" {
		t.Errorf("Expected log_level to be 'debug', got '%s'", viper.GetString("app.log_level"))
	}
}

func TestInitConfig_CustomConfigFile(t *testing.T) {
	// Get the path to the testdata directory
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current test file path")
	}
	projectRoot := filepath.Dir(filepath.Dir(filename))
	cfgFile = filepath.Join(projectRoot, "testdata", "config.yaml")

	viper.Reset()
	if err := initConfig(); err != nil {
		t.Fatalf("initConfig() failed: %v", err)
	}

	// Assert that the config file was used
	if viper.ConfigFileUsed() != cfgFile {
		t.Errorf("Expected config file to be '%s', got '%s'", cfgFile, viper.ConfigFileUsed())
	}

	// Assert values from the config file
	if viper.GetString("app.log_level") != "debug" {
		t.Errorf("Expected 'debug', got '%s'", viper.GetString("app.log_level"))
	}
	if viper.GetString("app.ping.output_message") != "Config Message" {
		t.Errorf("Expected 'Config Message', got '%s'", viper.GetString("app.ping.output_message"))
	}
}

func TestInitConfig_InvalidConfigFile(t *testing.T) {
	// Provide an invalid config file path
	cfgFile = "/invalid/path/to/config.yaml"

	viper.Reset()

	// Capture log output
	var buf bytes.Buffer
	log.Logger = log.Output(&buf)

	// Mock osExit to prevent the test from exiting
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }() // Restore osExit after the test

	// Call initConfig()
	err := initConfig()

	// Verify that an error was returned
	if err == nil {
		t.Errorf("Expected initConfig() to return an error for invalid config file")
	}

	// Verify the error message
	if !strings.Contains(err.Error(), "Failed to read config file") {
		t.Errorf("Expected error message to contain 'Failed to read config file', got '%v'", err)
	}

	// Verify that the appropriate log message is present
	if !strings.Contains(buf.String(), "Failed to read config file") {
		t.Errorf("Expected log output to contain 'Failed to read config file', got '%s'", buf.String())
	}

	// Verify that osExit was not called (exitCode should be zero)
	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}
}

func TestExecute(t *testing.T) {
	// Create a buffer to capture output
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)

	// Set arguments to display help message
	rootCmd.SetArgs([]string{"--help"})

	// Execute the command
	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("Execute() returned an error: %v", err)
	}

	// Assert that help output was generated
	output := buf.String()
	if output == "" {
		t.Errorf("Expected help output, got empty string")
	}
}

func TestExecute_CommandExecutionError(t *testing.T) {
	// Create a new root command for testing
	testCmd := &cobra.Command{
		Use: "testcmd",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("mock command execution error")
		},
		Args: cobra.NoArgs,
	}

	// Save the original rootCmd and defer restoration
	originalRootCmd := rootCmd
	defer func() { rootCmd = originalRootCmd }()

	// Assign the testCmd to rootCmd
	rootCmd = testCmd

	// Override PersistentPreRunE to prevent interference
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return nil
	}

	// Capture log output
	var buf bytes.Buffer
	log.Logger = log.Output(&buf)

	// Mock osExit to prevent the test from exiting
	var exitCode int
	osExit = func(code int) {
		exitCode = code
	}
	defer func() { osExit = os.Exit }() // Restore osExit after the test

	// Execute the root command
	Execute()

	// Assert that osExit was called with code 1
	if exitCode != 1 {
		t.Fatalf("Expected osExit(1) to be called, got exit code %d", exitCode)
	}

	// Assert that the error was logged
	if !strings.Contains(buf.String(), "Command execution failed") {
		t.Errorf("Expected 'Command execution failed' in log output, got '%s'", buf.String())
	}
}

func TestLoggerInitialization(t *testing.T) {
	buf := new(bytes.Buffer)
	viper.Set("app.log_level", "debug")

	// Initialize logger with the buffer as output
	if err := logger.Init(buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Log a debug message
	log.Debug().Msg("Test debug message")

	// Assert that the message was logged
	if !bytes.Contains(buf.Bytes(), []byte("Test debug message")) {
		t.Errorf("Expected 'Test debug message' in log output")
	}
}

func TestPersistentPreRunE_Error(t *testing.T) {
	// Save the original logger
	originalLogger := log.Logger
	defer func() { log.Logger = originalLogger }()

	// Mock the logger to discard output
	mockLog := log.Output(io.Discard).With().Str("mock", "true").Logger()
	log.Logger = mockLog

	// Save and defer restoration of the original PersistentPreRunE
	originalPersistentPreRunE := rootCmd.PersistentPreRunE
	defer func() { rootCmd.PersistentPreRunE = originalPersistentPreRunE }()

	// Override PersistentPreRunE to simulate an error
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("mock logger initialization error")
	}

	// Invoke the overridden PersistentPreRunE function
	err := rootCmd.PersistentPreRunE(nil, nil)
	if err == nil || err.Error() != "mock logger initialization error" {
		t.Fatalf("Expected logger initialization error, got %v", err)
	}
}
