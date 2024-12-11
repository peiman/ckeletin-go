// cmd/root_test.go

package cmd

import (
	"bytes"
	"errors"
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
