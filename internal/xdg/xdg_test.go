package xdg

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) {
	t.Helper()
	// Reset app name before each test
	mu.Lock()
	appName = ""
	mu.Unlock()
}

func TestSetAppName(t *testing.T) {
	setupTest(t)

	SetAppName("testapp")
	assert.Equal(t, "testapp", GetAppName())
}

func TestGetAppName_NotSet(t *testing.T) {
	setupTest(t)

	assert.Equal(t, "", GetAppName())
}

func TestConfigDir_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := ConfigDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestConfigDir(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	// Set XDG_CONFIG_HOME for Linux test
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_CONFIG_HOME", tempDir)
	}

	SetAppName("testapp")
	dir, err := ConfigDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.Contains(t, dir, "testapp")
}

func TestConfigFile(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_CONFIG_HOME", tempDir)
	}

	SetAppName("testapp")
	file, err := ConfigFile("config.yaml")
	require.NoError(t, err)
	assert.Contains(t, file, "testapp")
	assert.True(t, filepath.IsAbs(file))
	assert.Equal(t, "config.yaml", filepath.Base(file))
}

func TestCacheDir(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_CACHE_HOME", tempDir)
	}

	SetAppName("testapp")
	dir, err := CacheDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.Contains(t, dir, "testapp")
}

func TestCacheFile(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_CACHE_HOME", tempDir)
	}

	SetAppName("testapp")
	file, err := CacheFile("timings.json")
	require.NoError(t, err)
	assert.Contains(t, file, "testapp")
	assert.Equal(t, "timings.json", filepath.Base(file))
}

func TestDataDir(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_DATA_HOME", tempDir)
	}

	SetAppName("testapp")
	dir, err := DataDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.Contains(t, dir, "testapp")
}

func TestStateDir(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_STATE_HOME", tempDir)
	}

	SetAppName("testapp")
	dir, err := StateDir()
	require.NoError(t, err)
	assert.DirExists(t, dir)
	assert.Contains(t, dir, "testapp")
}

func TestStateFile(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_STATE_HOME", tempDir)
	}

	SetAppName("testapp")
	file, err := StateFile("app.log")
	require.NoError(t, err)
	assert.Contains(t, file, "testapp")
	assert.Equal(t, "app.log", filepath.Base(file))
}

func TestXDGEnvVarsFallback(t *testing.T) {
	setupTest(t)

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("XDG env vars only apply to Linux/Unix")
	}

	// Clear XDG env vars to test fallback
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("XDG_STATE_HOME", "")

	// Should use default paths
	assert.Contains(t, configBase(), ".config")
	assert.Contains(t, dataBase(), ".local/share")
	assert.Contains(t, cacheBase(), ".cache")
	assert.Contains(t, stateBase(), ".local/state")
}

func TestXDGEnvVarsOverride(t *testing.T) {
	setupTest(t)

	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("XDG env vars only apply to Linux/Unix")
	}

	customConfig := t.TempDir()
	customData := t.TempDir()
	customCache := t.TempDir()
	customState := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", customConfig)
	t.Setenv("XDG_DATA_HOME", customData)
	t.Setenv("XDG_CACHE_HOME", customCache)
	t.Setenv("XDG_STATE_HOME", customState)

	assert.Equal(t, customConfig, configBase())
	assert.Equal(t, customData, dataBase())
	assert.Equal(t, customCache, cacheBase())
	assert.Equal(t, customState, stateBase())
}

func TestDirectoryPermissions(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_CONFIG_HOME", tempDir)
	}

	SetAppName("testapp")
	dir, err := ConfigDir()
	require.NoError(t, err)

	// Check directory was created with secure permissions
	info, err := os.Stat(dir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())

	if runtime.GOOS != "windows" {
		// On Unix, check permissions are 0700
		assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
	}
}

func TestHomeDir(t *testing.T) {
	home := homeDir()
	assert.NotEmpty(t, home)
	assert.True(t, filepath.IsAbs(home))
}

func TestConcurrentAccess(t *testing.T) {
	setupTest(t)

	// Test thread-safety of SetAppName/GetAppName
	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			SetAppName("app1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			SetAppName("app2")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			_ = GetAppName()
		}
		done <- true
	}()

	<-done
	<-done
	<-done

	// Should complete without race conditions
	name := GetAppName()
	assert.True(t, name == "app1" || name == "app2")
}

func TestDataFile(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()

	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Setenv("XDG_DATA_HOME", tempDir)
	}

	SetAppName("testapp")
	file, err := DataFile("data.db")
	require.NoError(t, err)
	assert.Contains(t, file, "testapp")
	assert.Equal(t, "data.db", filepath.Base(file))
}

func TestDataFile_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := DataFile("data.db")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestConfigFile_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := ConfigFile("config.yaml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestCacheFile_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := CacheFile("cache.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestStateFile_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := StateFile("state.log")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestDataDir_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := DataDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestCacheDir_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := CacheDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestStateDir_AppNameNotSet(t *testing.T) {
	setupTest(t)

	_, err := StateDir()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "app name not set")
}

func TestHomeDirFallback(t *testing.T) {
	// Test that homeDir() returns something even with HOME unset
	// Note: Can't actually unset HOME in a running test safely,
	// but we can verify the function doesn't panic and returns valid path
	home := homeDir()
	assert.NotEmpty(t, home)
	assert.NotEqual(t, ".", home) // Should find actual home
}
