package cmd

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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
}

func TestExecute(t *testing.T) {
	// Save the original os.Args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Set up test cases
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "Default command",
			args:     []string{"ckeletin-go"},
			expected: "Hello from ckeletin-go!",
		},
		{
			name:     "Version command",
			args:     []string{"ckeletin-go", "version"},
			expected: "ckeletin-go v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the command line arguments
			os.Args = tt.args

			// Redirect stdout to capture output
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Execute the command
			Execute()

			// Restore stdout
			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)

			// Check if the output contains expected content
			output := buf.String()
			assert.Contains(t, output, tt.expected)
		})
	}
}
func TestInitConfig(t *testing.T) {
	// Create a temporary config file
	tempFile, err := os.CreateTemp("", "ckeletin-go*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	// Write test configuration
	_, err = tempFile.WriteString(`{
		"LogLevel": "debug",
		"Server": {
			"Port": 9090,
			"Host": "localhost"
		}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	tempFile.Close()

	// Set config file path
	oldCfgFile := cfgFile
	cfgFile = tempFile.Name()
	defer func() { cfgFile = oldCfgFile }()

	// Reset viper to ensure a clean state
	viper.Reset()

	// Redirect stdout to capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call initConfig
	initConfig()

	// Restore stdout
	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	// Check if the output contains expected content
	output := buf.String()
	assert.Contains(t, output, "Using config file:")
	assert.Contains(t, output, "debug")
	assert.Contains(t, output, "9090")
}
