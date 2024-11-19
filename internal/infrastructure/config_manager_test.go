package infrastructure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigManager(t *testing.T) {
	cm := NewConfigManager("")
	assert.Equal(t, DefaultConfigFileName, cm.ConfigPath)

	customPath := "custom_config.json"
	cm = NewConfigManager(customPath)
	assert.Equal(t, customPath, cm.ConfigPath)
}

func TestEnsureConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test_config.json")
	cm := NewConfigManager(configPath)

	// Test creating a new config file
	err = cm.EnsureConfig()
	assert.NoError(t, err)
	assert.FileExists(t, configPath)

	// Test with existing config file
	err = cm.EnsureConfig()
	assert.NoError(t, err)
}

func TestCreateDefaultConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "config_test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test_config.json")
	cm := NewConfigManager(configPath)

	err = cm.CreateDefaultConfig()
	assert.NoError(t, err)
	assert.FileExists(t, configPath)

	// Verify content
	content, err := os.ReadFile(configPath)
	assert.NoError(t, err)
	assert.Contains(t, string(content), `"LogLevel": "info"`)
}

func TestCreateDefaultConfigWithError(t *testing.T) {
	// Test with invalid directory permissions
	if os.Geteuid() == 0 {
		t.Skip("Skipping test when running as root")
	}

	configPath := "/root/test_config.json" // This should fail for non-root users
	cm := NewConfigManager(configPath)

	err := cm.CreateDefaultConfig()
	assert.Error(t, err)
}
