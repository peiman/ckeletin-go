// Package testutil provides common testing utilities and helpers
// used across the test suite.
package testutil

import (
	"runtime"
	"testing"
)

// SkipOnWindows skips the current test if running on Windows.
// This is useful for tests that rely on Unix-specific features
// like file permissions.
func SkipOnWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Skipping test on Windows")
	}
}

// SkipOnWindowsWithReason skips the current test if running on Windows,
// with a custom reason message.
func SkipOnWindowsWithReason(t *testing.T, reason string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skipf("Skipping test on Windows: %s", reason)
	}
}

// SkipOnNonWindows skips the current test if not running on Windows.
// This is useful for Windows-specific tests.
func SkipOnNonWindows(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}
}

// SkipOnPlatform skips the current test if running on the specified platform.
func SkipOnPlatform(t *testing.T, platform string) {
	t.Helper()
	if runtime.GOOS == platform {
		t.Skipf("Skipping test on %s", platform)
	}
}
