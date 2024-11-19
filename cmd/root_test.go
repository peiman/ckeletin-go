package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/peiman/ckeletin-go/internal/infrastructure"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	assert.Equal(t, "ckeletin-go", rootCmd.Use)
	assert.Contains(t, rootCmd.Short, "brief description")
	assert.Contains(t, rootCmd.Long, "longer description")

	configFlag := rootCmd.PersistentFlags().Lookup("config")
	assert.NotNil(t, configFlag)
	assert.Equal(t, "config", configFlag.Name)

	logLevelFlag := rootCmd.PersistentFlags().Lookup("log-level")
	assert.NotNil(t, logLevelFlag)
	assert.Equal(t, "log-level", logLevelFlag.Name)

	toggleFlag := rootCmd.Flags().Lookup("toggle")
	assert.NotNil(t, toggleFlag)
	assert.Equal(t, "toggle", toggleFlag.Name)

	// Test the root command's Run function
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.Run(rootCmd, []string{})

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "Hello from ckeletin-go!")
}

func TestInitConfig(t *testing.T) {
	// Save the original stdout and restore it after the test
	oldStdout := os.Stdout
	defer func() {
		os.Stdout = oldStdout
		infrastructure.SetLogOutput(os.Stdout)
	}()

	t.Run("Invalid Log Level", func(t *testing.T) {
		// Save original values
		oldLogLevel := logLevel
		oldOsExit := osExit
		exitCode := 0

		// Restore original values after test
		defer func() {
			logLevel = oldLogLevel
			osExit = oldOsExit
		}()

		// Set test values
		logLevel = "invalid"
		osExit = func(code int) {
			exitCode = code
		}

		// Capture output
		r, w, _ := os.Pipe()
		os.Stdout = w
		infrastructure.SetLogOutput(w)

		// Run initConfig
		initConfig()

		// Close and restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Check results
		assert.Equal(t, 1, exitCode)

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Failed to initialize logger")
	})

	t.Run("Invalid Config Path", func(t *testing.T) {
		// Save original values
		oldCfgFile := cfgFile
		oldOsExit := osExit
		exitCode := 0

		// Restore original values after test
		defer func() {
			cfgFile = oldCfgFile
			osExit = oldOsExit
		}()

		// Set test values
		cfgFile = "/nonexistent/path/config.json"
		osExit = func(code int) {
			exitCode = code
		}

		// Reset viper
		viper.Reset()

		// Capture output
		r, w, _ := os.Pipe()
		os.Stdout = w
		infrastructure.SetLogOutput(w)

		// Run initConfig
		initConfig()

		// Close and restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Check results
		assert.Equal(t, 1, exitCode)

		var buf bytes.Buffer
		_, err := io.Copy(&buf, r)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Failed to ensure config file exists")
	})

	t.Run("Invalid Config Content", func(t *testing.T) {
		// Create a temporary invalid config file
		tempFile, err := os.CreateTemp("", "invalid_config*.json")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		_, err = tempFile.Write([]byte("{invalid json"))
		require.NoError(t, err)
		tempFile.Close()

		// Save original values
		oldCfgFile := cfgFile
		oldOsExit := osExit
		exitCode := 0

		// Restore original values after test
		defer func() {
			cfgFile = oldCfgFile
			osExit = oldOsExit
		}()

		// Set test values
		cfgFile = tempFile.Name()
		osExit = func(code int) {
			exitCode = code
		}

		// Reset viper
		viper.Reset()

		// Capture output
		r, w, _ := os.Pipe()
		os.Stdout = w
		infrastructure.SetLogOutput(w)

		// Run initConfig
		initConfig()

		// Close and restore stdout
		w.Close()
		os.Stdout = oldStdout

		// Check results
		assert.Equal(t, 1, exitCode)

		var buf bytes.Buffer
		_, err = io.Copy(&buf, r)
		require.NoError(t, err)
		assert.Contains(t, buf.String(), "Failed to read config file")
	})
}
