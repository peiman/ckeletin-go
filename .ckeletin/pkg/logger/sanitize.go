// .ckeletin/pkg/logger/sanitize.go
//
// Sanitization functions for log output. Scope:
//   - SanitizeLogString strips control characters (log injection) and
//     truncates overly long strings (log flooding); truncated output exceeds
//     the configured byte limit by the appended truncation marker
//   - SanitizePath additionally masks the home directory in paths
//   - SanitizeError applies SanitizeLogString to an error's message
//
// These functions do NOT redact secret values (tokens, passwords, API keys).
// Callers are responsible for never passing secrets to the logger.

package logger

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"unicode/utf8"
)

// truncationMarker is appended to strings cut off by SanitizeLogString.
const truncationMarker = "...[truncated]"

var (
	// Remove control characters and newlines that could break log format or inject fake log entries
	controlCharsRegex = regexp.MustCompile(`[\x00-\x1F\x7F]+`)

	// Maximum length in bytes for logged strings to prevent log flooding.
	// Default is 1000, but can be overridden via LOG_TRUNCATE_LIMIT environment variable.
	// Atomic because SetMaxLogLength may run concurrently with SanitizeLogString.
	maxLogStringLength atomic.Int64
)

func init() {
	maxLogStringLength.Store(int64(initMaxLogLength()))
}

// initMaxLogLength initializes the max log length from environment variable or uses default
func initMaxLogLength() int {
	if envVal := os.Getenv("LOG_TRUNCATE_LIMIT"); envVal != "" {
		if limit, err := strconv.Atoi(envVal); err == nil && limit > 0 {
			return limit
		}
	}
	return 1000 // default value
}

// SanitizeLogString removes potentially dangerous characters from log output
// and truncates excessively long strings to prevent log flooding attacks.
// It does NOT redact secret values; never pass secrets to the logger.
func SanitizeLogString(s string) string {
	// Remove control characters (including newlines, tabs, etc.)
	// This prevents log injection where an attacker could:
	// 1. Insert fake log entries by injecting newlines
	// 2. Break log parsers with control characters
	// 3. Hide malicious activity by using ANSI escape codes
	s = controlCharsRegex.ReplaceAllString(s, "")

	// Truncate if too long (byte limit, cut on a rune boundary so valid
	// UTF-8 input never yields invalid UTF-8 output)
	if maxLen := int(maxLogStringLength.Load()); len(s) > maxLen {
		s = s[:truncationBoundary(s, maxLen)] + truncationMarker
	}

	return s
}

// truncationBoundary returns the largest index <= maxLen that starts a rune,
// so slicing s at the returned index never splits a multi-byte character.
// Callers must guarantee maxLen < len(s).
func truncationBoundary(s string, maxLen int) int {
	for maxLen > 0 && !utf8.RuneStart(s[maxLen]) {
		maxLen--
	}
	return maxLen
}

// SanitizePath removes sensitive information from file paths before logging.
// This prevents leakage of usernames and directory structures.
func SanitizePath(path string) string {
	// Handle Windows-style home paths first (takes precedence on Windows)
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		path = strings.ReplaceAll(path, userProfile, "~")
	}

	// Replace Unix-style home directory with ~ to avoid exposing usernames
	// Using ReplaceAll to handle paths that might contain home directory multiple times
	if home := os.Getenv("HOME"); home != "" {
		path = strings.ReplaceAll(path, home, "~")
	}

	// Still sanitize for control characters
	return SanitizeLogString(path)
}

// SanitizeError sanitizes error messages which may contain user input
func SanitizeError(err error) string {
	if err == nil {
		return ""
	}
	return SanitizeLogString(err.Error())
}

// SetMaxLogLength allows adjusting the maximum log string length.
// Useful for testing or specific security requirements.
// Safe for concurrent use with SanitizeLogString.
func SetMaxLogLength(length int) {
	if length > 0 {
		maxLogStringLength.Store(int64(length))
	}
}
