// cmd/root_test.go
package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/internal/logger"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func TestInitConfig_Defaults(t *testing.T) {
	// Reset Viper before the test
	viper.Reset()

	// Test default values from initConfig
	initConfig()

	if viper.GetString("app.log_level") != "info" {
		t.Errorf("Expected default log_level to be 'info', got '%s'", viper.GetString("app.log_level"))
	}
}

func TestInitConfig_EnvironmentOverride(t *testing.T) {
	// Set an environment variable to override the default
	os.Setenv("APP_LOG_LEVEL", "debug")
	defer os.Unsetenv("APP_LOG_LEVEL")

	// Reset Viper and re-initialize configurations
	viper.Reset()
	// Ensure that the environment variables are correctly mapped
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	initConfig()

	if viper.GetString("app.log_level") != "debug" {
		t.Errorf("Expected log_level to be 'debug', got '%s'", viper.GetString("app.log_level"))
	}
}

func TestInitConfig_CustomConfigFile(t *testing.T) {
	// Get the directory of the current test file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("Failed to get current test file path")
	}
	// Build the path to testdata/config.yaml relative to the project root
	projectRoot := filepath.Dir(filepath.Dir(filename)) // Assuming cmd/ is one level under project root
	cfgFile = filepath.Join(projectRoot, "testdata", "config.yaml")

	viper.Reset()
	initConfig()

	if viper.ConfigFileUsed() != cfgFile {
		t.Errorf("Expected config file to be '%s', got '%s'", cfgFile, viper.ConfigFileUsed())
	}
}

func TestLoggerInitialization(t *testing.T) {
	buf := new(bytes.Buffer)
	viper.Set("app.log_level", "debug")

	if err := logger.Init(buf); err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	log.Debug().Msg("Test debug message")
	if !bytes.Contains(buf.Bytes(), []byte("Test debug message")) {
		t.Errorf("Expected 'Test debug message' in log output")
	}
}
