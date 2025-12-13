// Package xdg provides XDG Base Directory Specification compliant paths.
//
// This package abstracts platform-specific conventions for storing application
// files, following the XDG Base Directory Specification on Linux/Unix and
// using native conventions on macOS and Windows.
//
// # Platform Behavior
//
// On Linux/Unix, it follows the XDG Base Directory Specification:
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
// # Usage
//
// The package uses a single source of truth pattern for the application name.
// Set the app name once at startup:
//
//	func init() {
//	    xdg.SetAppName("myapp")
//	}
//
// Then use the package functions throughout your application:
//
//	// Get config directory (creates if needed)
//	configDir, err := xdg.ConfigDir()
//	// Returns: ~/.config/myapp on Linux
//
//	// Get path to a config file
//	configPath, err := xdg.ConfigFile("config.yaml")
//	// Returns: ~/.config/myapp/config.yaml on Linux
//
//	// Get cache directory for temporary/regenerable data
//	cacheDir, err := xdg.CacheDir()
//	// Returns: ~/.cache/myapp on Linux
//
//	// Get state directory for logs, history, etc.
//	stateDir, err := xdg.StateDir()
//	// Returns: ~/.local/state/myapp on Linux
//
// # Directory Types
//
// Choose the appropriate directory type based on your data:
//
//   - Config: User configuration files (settings, preferences)
//   - Data: Persistent application data (databases, user files)
//   - Cache: Regenerable data (can be deleted without data loss)
//   - State: Runtime state (logs, history, recently-used)
//
// # Thread Safety
//
// All functions in this package are safe for concurrent use.
// SetAppName should ideally be called once during initialization.
package xdg
