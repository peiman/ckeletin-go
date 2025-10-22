// internal/logger/sanitize.go
//
// Sanitization functions for log output to prevent log injection and information leakage

package logger

import (
	"os"
	"regexp"
	"strings"
)

var (
	// Remove control characters and newlines that could break log format or inject fake log entries
	controlCharsRegex = regexp.MustCompile(`[\x00-\x1F\x7F]+`)

	// Maximum length for logged strings to prevent log flooding
	maxLogStringLength = 1000
)

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
	// Replace home directory with ~ to avoid exposing usernames
	if home := os.Getenv("HOME"); home != "" {
		path = strings.Replace(path, home, "~", 1)
	}

	// Also handle Windows-style home paths
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		path = strings.Replace(path, userProfile, "~", 1)
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
