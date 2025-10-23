// internal/config/security.go
//
// Security validation functions for configuration files

package config

import (
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog/log"
)

// ValidateConfigFilePermissions checks if a config file has secure permissions.
// On Unix-like systems, it ensures the file is not world-writable and warns if group-writable.
// On Windows, this check is skipped as Windows has a different permission model.
func ValidateConfigFilePermissions(path string) error {
	// Skip on Windows - different permission model
	if runtime.GOOS == "windows" {
		log.Debug().Str("path", path).Msg("Skipping permission check on Windows")
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	mode := info.Mode()
	perm := mode.Perm()

	// Check if world-writable (dangerous - anyone can modify config)
	if perm&0002 != 0 {
		return fmt.Errorf(`config file %s is world-writable (permissions: %04o)

Security Issue: Anyone on the system can modify this configuration file.

How to fix:
  chmod 600 %s

This will set owner-only read/write permissions.`, path, perm, path)
	}

	// Warn if group-writable (potentially dangerous depending on group membership)
	if perm&0020 != 0 {
		log.Warn().
			Str("path", path).
			Str("permissions", fmt.Sprintf("%04o", perm)).
			Msg("Config file is group-writable, consider restricting to 0600 or 0400")
	}

	// Recommend stricter permissions if too permissive
	if perm&0077 != 0 {
		log.Info().
			Str("path", path).
			Str("current", fmt.Sprintf("%04o", perm)).
			Str("recommended", "0600").
			Msg("Config file has permissive permissions, recommend tightening")
	}

	log.Debug().
		Str("path", path).
		Str("permissions", fmt.Sprintf("%04o", perm)).
		Msg("Config file permissions validated")

	return nil
}

// ValidateConfigFileSize checks if a config file size is within acceptable limits.
// This prevents DoS attacks via extremely large config files.
func ValidateConfigFileSize(path string, maxSize int64) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat config file: %w", err)
	}

	if info.Size() > maxSize {
		return fmt.Errorf(`config file %s is too large (%d bytes > %d bytes maximum)

Security Issue: Large config files can cause denial-of-service.

Suggestions:
  - Remove unnecessary configuration
  - Split into multiple smaller files
  - Check for accidental binary data in config file

Current size: %d bytes (%.2f MB)
Maximum allowed: %d bytes (%.2f MB)`,
			path, info.Size(), maxSize,
			info.Size(), float64(info.Size())/(1024*1024),
			maxSize, float64(maxSize)/(1024*1024))
	}

	log.Debug().
		Str("path", path).
		Int64("size", info.Size()).
		Int64("max_size", maxSize).
		Msg("Config file size validated")

	return nil
}

// ValidateConfigFileSecurity performs comprehensive security validation on a config file.
// It combines both size and permission checks in a single convenient function.
// This is the recommended function to use for config file security validation.
//
// It checks:
//   - File size is within acceptable limits (prevents DoS)
//   - File permissions are secure (prevents unauthorized modification)
//
// Returns the first validation error encountered, if any.
func ValidateConfigFileSecurity(path string, maxSize int64) error {
	// Validate file size first (cheaper check)
	if err := ValidateConfigFileSize(path, maxSize); err != nil {
		return err
	}

	// Validate file permissions
	if err := ValidateConfigFilePermissions(path); err != nil {
		return err
	}

	return nil
}
