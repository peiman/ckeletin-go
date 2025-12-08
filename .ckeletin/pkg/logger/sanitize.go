// internal/logger/sanitize.go
//
// Sanitization functions for log output to prevent log injection and information leakage

package logger

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Remove control characters and newlines that could break log format or inject fake log entries
	controlCharsRegex = regexp.MustCompile(`[\x00-\x1F\x7F]+`)

	// Maximum length for logged strings to prevent log flooding
	// Default is 1000, but can be overridden via LOG_TRUNCATE_LIMIT environment variable
	maxLogStringLength = initMaxLogLength()
)

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
func SanitizeLogString(s string) string {
	// Remove control characters (including newlines, tabs, etc.)
	// This prevents log injection where an attacker could:
	// 1. Insert fake log entries by injecting newlines
	// 2. Break log parsers with control characters
	// 3. Hide malicious activity by using ANSI escape codes
	s = controlCharsRegex.ReplaceAllString(s, "")

	// Truncate if too long
	if len(s) > maxLogStringLength {
		s = s[:maxLogStringLength] + "...[truncated]"
	}

	return s
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
func SetMaxLogLength(length int) {
	if length > 0 {
		maxLogStringLength = length
	}
}
