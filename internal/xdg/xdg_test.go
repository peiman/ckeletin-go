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

func TestHomeDirUsesHOMEEnv(t *testing.T) {
	t.Setenv("HOME", "/custom/home")
	assert.Equal(t, "/custom/home", homeDir())
}

func TestHomeDirFallsBackWhenHOMEEmpty(t *testing.T) {
	// When HOME is empty, homeDir() tries os.UserHomeDir(), then falls back to "."
	t.Setenv("HOME", "")
	home := homeDir()
	// Should return something (either os.UserHomeDir result or "." fallback)
	assert.NotEmpty(t, home)
}

func TestConfigBase_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	base := configBase()
	assert.Contains(t, base, filepath.Join("Library", "Application Support"))
}

func TestDataBase_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	base := dataBase()
	assert.Contains(t, base, filepath.Join("Library", "Application Support"))
}

func TestCacheBase_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	base := cacheBase()
	assert.Contains(t, base, filepath.Join("Library", "Caches"))
}

func TestStateBase_Darwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin-specific test")
	}
	base := stateBase()
	assert.Contains(t, base, filepath.Join("Library", "Application Support"))
}

func TestBasePathsUseHomeDir(t *testing.T) {
	// Verify all base functions incorporate the home directory
	t.Setenv("HOME", "/test/home")

	config := configBase()
	data := dataBase()
	cache := cacheBase()
	state := stateBase()

	switch runtime.GOOS {
	case "darwin":
		assert.Equal(t, "/test/home/Library/Application Support", config)
		assert.Equal(t, "/test/home/Library/Application Support", data)
		assert.Equal(t, "/test/home/Library/Caches", cache)
		assert.Equal(t, "/test/home/Library/Application Support", state)
	case "windows":
		// Windows uses env vars (AppData/LocalAppData), HOME fallback only if those are unset
	default:
		// Linux with no XDG vars set
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("XDG_CACHE_HOME", "")
		t.Setenv("XDG_STATE_HOME", "")
		assert.Equal(t, "/test/home/.config", configBase())
		assert.Equal(t, "/test/home/.local/share", dataBase())
		assert.Equal(t, "/test/home/.cache", cacheBase())
		assert.Equal(t, "/test/home/.local/state", stateBase())
	}
}

func TestDirCreatesDirectory(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	SetAppName("testapp")

	// All Dir functions should create their directories
	tests := []struct {
		name string
		fn   func() (string, error)
	}{
		{"ConfigDir", ConfigDir},
		{"DataDir", DataDir},
		{"CacheDir", CacheDir},
		{"StateDir", StateDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, err := tt.fn()
			require.NoError(t, err)
			assert.DirExists(t, dir)
			assert.Contains(t, dir, "testapp")

			// Verify permissions on non-Windows
			if runtime.GOOS != "windows" {
				info, err := os.Stat(dir)
				require.NoError(t, err)
				assert.Equal(t, os.FileMode(0700), info.Mode().Perm())
			}
		})
	}
}

func TestBasePaths_Linux(t *testing.T) {
	origOS := osName
	osName = "linux"
	t.Cleanup(func() { osName = origOS })

	home := "/home/testuser"
	t.Setenv("HOME", home)

	t.Run("defaults without XDG vars", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")
		t.Setenv("XDG_DATA_HOME", "")
		t.Setenv("XDG_CACHE_HOME", "")
		t.Setenv("XDG_STATE_HOME", "")

		// Use filepath.Join for cross-platform path separators
		assert.Equal(t, filepath.Join(home, ".config"), configBase())
		assert.Equal(t, filepath.Join(home, ".local", "share"), dataBase())
		assert.Equal(t, filepath.Join(home, ".cache"), cacheBase())
		assert.Equal(t, filepath.Join(home, ".local", "state"), stateBase())
	})

	t.Run("XDG env var overrides", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")
		t.Setenv("XDG_DATA_HOME", "/custom/data")
		t.Setenv("XDG_CACHE_HOME", "/custom/cache")
		t.Setenv("XDG_STATE_HOME", "/custom/state")

		assert.Equal(t, "/custom/config", configBase())
		assert.Equal(t, "/custom/data", dataBase())
		assert.Equal(t, "/custom/cache", cacheBase())
		assert.Equal(t, "/custom/state", stateBase())
	})
}

func TestBasePaths_Windows(t *testing.T) {
	origOS := osName
	osName = "windows"
	t.Cleanup(func() { osName = origOS })

	t.Setenv("HOME", `C:\Users\testuser`)

	t.Run("uses AppData and LocalAppData env vars", func(t *testing.T) {
		t.Setenv("AppData", `C:\Users\testuser\AppData\Roaming`)
		t.Setenv("LocalAppData", `C:\Users\testuser\AppData\Local`)

		assert.Equal(t, `C:\Users\testuser\AppData\Roaming`, configBase())
		assert.Equal(t, `C:\Users\testuser\AppData\Roaming`, dataBase())
		assert.Equal(t, `C:\Users\testuser\AppData\Local`, cacheBase())
		assert.Equal(t, `C:\Users\testuser\AppData\Roaming`, stateBase())
	})

	t.Run("falls back to HOME when env vars empty", func(t *testing.T) {
		t.Setenv("AppData", "")
		t.Setenv("LocalAppData", "")

		assert.Contains(t, configBase(), "AppData")
		assert.Contains(t, dataBase(), "AppData")
		assert.Contains(t, cacheBase(), "AppData")
		assert.Contains(t, stateBase(), "AppData")
	})
}

func TestBasePaths_Darwin(t *testing.T) {
	origOS := osName
	osName = "darwin"
	t.Cleanup(func() { osName = origOS })

	home := "/Users/testuser"
	t.Setenv("HOME", home)

	// Use filepath.Join for cross-platform path separators
	appSupport := filepath.Join(home, "Library", "Application Support")
	assert.Equal(t, appSupport, configBase())
	assert.Equal(t, appSupport, dataBase())
	assert.Equal(t, filepath.Join(home, "Library", "Caches"), cacheBase())
	assert.Equal(t, appSupport, stateBase())
}

func TestFileFunctions_ReturnAbsolutePaths(t *testing.T) {
	setupTest(t)
	tempDir := t.TempDir()
	t.Setenv("HOME", tempDir)
	SetAppName("testapp")

	tests := []struct {
		name     string
		fn       func(string) (string, error)
		filename string
	}{
		{"ConfigFile", ConfigFile, "config.yaml"},
		{"DataFile", DataFile, "data.db"},
		{"CacheFile", CacheFile, "cache.json"},
		{"StateFile", StateFile, "state.log"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := tt.fn(tt.filename)
			require.NoError(t, err)
			assert.True(t, filepath.IsAbs(file))
			assert.Equal(t, tt.filename, filepath.Base(file))
			assert.Contains(t, file, "testapp")
		})
	}
}
