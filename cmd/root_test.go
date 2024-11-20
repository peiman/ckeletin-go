package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTest creates a test environment and returns a cleanup function
func setupTest(t *testing.T) (func(), *cobra.Command) {
	t.Helper()
	// Save original values
	origStdout := os.Stdout
	origStderr := os.Stderr
	origExit := osExit
	origCfgFile := cfgFile
	origLogLevel := logLevel

	// Create a fresh command for testing
	cmd := &cobra.Command{
		Use:   rootCmd.Use,
		Short: rootCmd.Short,
		Long:  rootCmd.Long,
		Run:   rootCmd.Run,
	}
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./ckeletin-go.json)")
	cmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set the logging level (debug, info, warn, error)")
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cleanup := func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
		osExit = origExit
		cfgFile = origCfgFile
		logLevel = origLogLevel
		viper.Reset()
	}

	return cleanup, cmd
}

// captureOutput captures stdout and stderr during a test
func captureOutput(f func()) string {
	old := os.Stdout
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	re, we, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = we

	outC := make(chan string)
	// Copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	errC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, re)
		errC <- buf.String()
	}()

	f()

	// Reset stdout and stderr
	w.Close()
	we.Close()
	os.Stdout = old
	os.Stderr = oldErr

	return fmt.Sprintf("%s%s", <-outC, <-errC)
}

func TestRootCommand(t *testing.T) {
	cleanup, cmd := setupTest(t)
	defer cleanup()

	t.Run("command configuration", func(t *testing.T) {
		t.Run("has correct name", func(t *testing.T) {
			assert.Equal(t, "ckeletin-go", cmd.Use)
		})

		t.Run("has description", func(t *testing.T) {
			assert.Contains(t, cmd.Short, "brief description")
			assert.Contains(t, cmd.Long, "longer description")
		})

		t.Run("has required flags", func(t *testing.T) {
			flags := []struct {
				name string
				flag string
			}{
				{"config flag", "config"},
				{"log-level flag", "log-level"},
				{"toggle flag", "toggle"},
			}

			for _, f := range flags {
				t.Run(f.name, func(t *testing.T) {
					flag := cmd.Flags().Lookup(f.flag)
					if flag == nil {
						flag = cmd.PersistentFlags().Lookup(f.flag)
					}
					require.NotNil(t, flag, "flag %s should exist", f.name)
					assert.Equal(t, f.flag, flag.Name)
				})
			}
		})
	})

	t.Run("command execution", func(t *testing.T) {
		output := captureOutput(func() {
			cmd.Run(cmd, []string{})
		})
		assert.Contains(t, output, "Hello from ckeletin-go!")
	})
}

func TestInitConfig(t *testing.T) {
	cleanup, _ := setupTest(t)
	defer cleanup()

	t.Run("logging configuration", func(t *testing.T) {
		t.Run("fails with invalid log level", func(t *testing.T) {
			logLevel = "invalid"
			exitCode := 0
			osExit = func(code int) {
				exitCode = code
			}

			output := captureOutput(func() {
				initConfig()
			})

			assert.Equal(t, 1, exitCode, "should exit with code 1")
			assert.Contains(t, output, "Failed to initialize logger",
				"should contain error about logger initialization")
		})
	})

	t.Run("configuration file handling", func(t *testing.T) {
		t.Run("fails with invalid config path", func(t *testing.T) {
			logLevel = "info" // Reset to valid log level
			nonExistentDir := filepath.Join(t.TempDir(), "nonexistent")
			cfgFile = filepath.Join(nonExistentDir, "config.json")

			exitCode := 0
			osExit = func(code int) {
				exitCode = code
			}

			output := captureOutput(func() {
				initConfig()
			})

			assert.Equal(t, 1, exitCode, "should exit with code 1")
			assert.Contains(t, output, "Failed to ensure config file exists",
				"should include error about config file")
		})

		t.Run("fails with invalid config content", func(t *testing.T) {
			logLevel = "info" // Reset to valid log level
			tempFile := filepath.Join(t.TempDir(), "invalid_config.json")
			require.NoError(t, os.WriteFile(tempFile, []byte("{invalid json"), 0o600))

			cfgFile = tempFile
			exitCode := 0
			osExit = func(code int) {
				exitCode = code
			}

			output := captureOutput(func() {
				initConfig()
			})

			assert.Equal(t, 1, exitCode, "should exit with code 1")
			assert.Contains(t, output, "Failed to read config file",
				"should include error about invalid config content")
		})
	})
}
