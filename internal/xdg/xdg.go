// Package xdg provides XDG Base Directory Specification compliant paths.
//
// On Linux/Unix, it follows the XDG spec:
//   - Config: $XDG_CONFIG_HOME or ~/.config
//   - Data:   $XDG_DATA_HOME or ~/.local/share
//   - Cache:  $XDG_CACHE_HOME or ~/.cache
//   - State:  $XDG_STATE_HOME or ~/.local/state
//
// On macOS, it uses Apple conventions:
//   - Config: ~/Library/Application Support
//   - Data:   ~/Library/Application Support
//   - Cache:  ~/Library/Caches
//   - State:  ~/Library/Application Support
//
// On Windows, it uses standard Windows paths:
//   - Config: %AppData%
//   - Data:   %AppData%
//   - Cache:  %LocalAppData%
//   - State:  %AppData%
//
// Usage:
//
//	// At startup, set the app name from your single source of truth
//	xdg.SetAppName(binaryName)
//
//	// Then use the package functions
//	configDir, _ := xdg.ConfigDir()
//	cacheFile, _ := xdg.CacheFile("timings.json")
package xdg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	appName string
	mu      sync.RWMutex
)

// SetAppName sets the application name used for XDG directories.
// This should be called once at startup from your single source of truth.
// The name should be lowercase and without spaces (e.g., "ckeletin-go").
func SetAppName(name string) {
	mu.Lock()
	defer mu.Unlock()
	appName = name
}

// GetAppName returns the configured application name.
func GetAppName() string {
	mu.RLock()
	defer mu.RUnlock()
	return appName
}

func getAppName() (string, error) {
	mu.RLock()
	defer mu.RUnlock()
	if appName == "" {
		return "", fmt.Errorf("xdg: app name not set, call xdg.SetAppName() first")
	}
	return appName, nil
}

// ConfigDir returns the directory for configuration files.
// Creates the directory if it doesn't exist.
func ConfigDir() (string, error) {
	name, err := getAppName()
	if err != nil {
		return "", err
	}
	base := configBase()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// ConfigFile returns the path to a config file.
func ConfigFile(filename string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// DataDir returns the directory for persistent data files.
// Creates the directory if it doesn't exist.
func DataDir() (string, error) {
	name, err := getAppName()
	if err != nil {
		return "", err
	}
	base := dataBase()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// DataFile returns the path to a data file.
func DataFile(filename string) (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// CacheDir returns the directory for cache files.
// Creates the directory if it doesn't exist.
func CacheDir() (string, error) {
	name, err := getAppName()
	if err != nil {
		return "", err
	}
	base := cacheBase()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// CacheFile returns the path to a cache file.
func CacheFile(filename string) (string, error) {
	dir, err := CacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// StateDir returns the directory for state files (logs, history, etc.).
// Creates the directory if it doesn't exist.
func StateDir() (string, error) {
	name, err := getAppName()
	if err != nil {
		return "", err
	}
	base := stateBase()
	dir := filepath.Join(base, name)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// StateFile returns the path to a state file.
func StateFile(filename string) (string, error) {
	dir, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, filename), nil
}

// configBase returns the base config directory.
func configBase() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Application Support")
	case "windows":
		if dir := os.Getenv("AppData"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), "AppData", "Roaming")
	default: // Linux and other Unix
		if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), ".config")
	}
}

// dataBase returns the base data directory.
func dataBase() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Application Support")
	case "windows":
		if dir := os.Getenv("AppData"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), "AppData", "Roaming")
	default: // Linux and other Unix
		if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), ".local", "share")
	}
}

// cacheBase returns the base cache directory.
func cacheBase() string {
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir(), "Library", "Caches")
	case "windows":
		if dir := os.Getenv("LocalAppData"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), "AppData", "Local")
	default: // Linux and other Unix
		if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), ".cache")
	}
}

// stateBase returns the base state directory.
// State is for data that should persist between restarts but isn't config (logs, history).
func stateBase() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS doesn't have a state concept, use Application Support
		return filepath.Join(homeDir(), "Library", "Application Support")
	case "windows":
		if dir := os.Getenv("AppData"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), "AppData", "Roaming")
	default: // Linux and other Unix
		if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
			return dir
		}
		return filepath.Join(homeDir(), ".local", "state")
	}
}

// homeDir returns the user's home directory.
func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return "."
}
