package infrastructure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test_config.json")

	// Create config manager
	cm := NewConfigManager(configPath)

	// Create default config
	err := cm.CreateDefaultConfig()
	require.NoError(t, err)

	// Read the created config file
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	// Verify the content
	configStr := string(data)
	t.Logf("Config file content: %s", configStr) // Add this line for debugging

	assert.True(t, strings.Contains(configStr, `"logLevel": "info"`))
	assert.True(t, strings.Contains(configStr, `"defaultCount": 1`))
	assert.True(t, strings.Contains(configStr, `"outputMessage": "pong"`))
	assert.True(t, strings.Contains(configStr, `"coloredOutput": false`))

	// Verify JSON structure
	var config struct {
		LogLevel string `json:"logLevel"`
		Ping     struct {
			DefaultCount  int    `json:"defaultCount"`
			OutputMessage string `json:"outputMessage"`
			ColoredOutput bool   `json:"coloredOutput"`
		} `json:"ping"`
	}
	err = json.Unmarshal(data, &config)
	require.NoError(t, err)
	assert.Equal(t, "info", config.LogLevel)
}

func TestNewConfigManager(t *testing.T) {
	tests := []struct {
		name         string
		configPath   string
		expectedPath string
	}{
		{
			name:         "With custom path",
			configPath:   "/custom/path/config.json",
			expectedPath: "/custom/path/config.json",
		},
		{
			name:         "Empty path uses default",
			configPath:   "",
			expectedPath: DefaultConfigFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := NewConfigManager(tt.configPath)
			assert.Equal(t, tt.expectedPath, cm.ConfigPath)
		})
	}
}
