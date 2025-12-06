package testutil

import (
	"runtime"
	"testing"
)

// TestSkipOnWindows verifies SkipOnWindows behavior.
func TestSkipOnWindows(t *testing.T) {
	t.Run("does not skip on non-Windows", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Test runs only on non-Windows")
		}

		// On non-Windows, calling SkipOnWindows should not skip
		// If we reach the end of this test, it worked correctly
		SkipOnWindows(t)

		// If we got here, the skip didn't happen (correct for non-Windows)
	})
}

// TestSkipOnWindowsWithReason verifies SkipOnWindowsWithReason behavior.
func TestSkipOnWindowsWithReason(t *testing.T) {
	t.Run("does not skip on non-Windows", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Test runs only on non-Windows")
		}

		// On non-Windows, should not skip
		SkipOnWindowsWithReason(t, "test reason")

		// If we got here, it worked correctly
	})
}

// TestSkipOnNonWindows verifies SkipOnNonWindows behavior.
func TestSkipOnNonWindows(t *testing.T) {
	t.Run("does not skip on Windows", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Test runs only on Windows")
		}

		// On Windows, calling SkipOnNonWindows should not skip
		SkipOnNonWindows(t)

		// If we got here, it worked correctly
	})
}

// TestSkipOnPlatform verifies SkipOnPlatform behavior.
func TestSkipOnPlatform(t *testing.T) {
	t.Run("does not skip when platform doesn't match", func(t *testing.T) {
		// Use a platform that doesn't match current
		fakePlatform := "nonexistent-os-xyz"

		// Should not skip since we're not on "nonexistent-os-xyz"
		SkipOnPlatform(t, fakePlatform)

		// If we got here, it correctly didn't skip
	})

	t.Run("skips when platform matches current", func(t *testing.T) {
		// Use current platform to test the skip path
		SkipOnPlatform(t, runtime.GOOS)

		// This line should not be reached
		t.Error("Should have skipped on current platform")
	})
}

// TestSkipOnNonWindows_ActualSkip tests the actual skip on non-Windows.
func TestSkipOnNonWindows_ActualSkip(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("This test verifies non-Windows skip behavior")
	}

	// On non-Windows, this should skip
	SkipOnNonWindows(t)

	// This line should not be reached on non-Windows
	t.Error("Should have skipped on non-Windows")
}
