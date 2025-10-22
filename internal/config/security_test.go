// internal/config/security_test.go

package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestValidateConfigFilePermissions(t *testing.T) {
	// Skip permission tests on Windows
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tests := []struct {
		name        string
		permissions os.FileMode
		wantErr     bool
		errContains string
	}{
		{
			name:        "Secure permissions (0600)",
			permissions: 0600,
			wantErr:     false,
		},
		{
			name:        "Secure permissions (0400)",
			permissions: 0400,
			wantErr:     false,
		},
		{
			name:        "Group-writable (0620) - warning only",
			permissions: 0620,
			wantErr:     false,
		},
		{
			name:        "World-writable (0666) - error",
			permissions: 0666,
			wantErr:     true,
			errContains: "world-writable",
		},
		{
			name:        "World-writable (0602) - error",
			permissions: 0602,
			wantErr:     true,
			errContains: "world-writable",
		},
		{
			name:        "Executable (0700) - ok but permissive",
			permissions: 0700,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(tmpFile, []byte("test: value\n"), tt.permissions); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Explicitly set permissions to overcome umask
			if err := os.Chmod(tmpFile, tt.permissions); err != nil {
				t.Fatalf("Failed to chmod test file: %v", err)
			}

			// Validate permissions
			err := ValidateConfigFilePermissions(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigFilePermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateConfigFilePermissions() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateConfigFilePermissions_NonexistentFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	err := ValidateConfigFilePermissions("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("ValidateConfigFilePermissions() should error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to stat") {
		t.Errorf("Error should mention stat failure, got: %v", err)
	}
}

func TestValidateConfigFilePermissions_Windows(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping Windows-specific test on non-Windows platform")
	}

	// On Windows, should always return nil
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(tmpFile, []byte("test: value\n"), 0666); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err := ValidateConfigFilePermissions(tmpFile)
	if err != nil {
		t.Errorf("ValidateConfigFilePermissions() on Windows should return nil, got: %v", err)
	}
}

func TestValidateConfigFileSize(t *testing.T) {
	tests := []struct {
		name        string
		fileSize    int
		maxSize     int64
		wantErr     bool
		errContains string
	}{
		{
			name:     "Small file within limit",
			fileSize: 100,
			maxSize:  1000,
			wantErr:  false,
		},
		{
			name:     "File at exact limit",
			fileSize: 1000,
			maxSize:  1000,
			wantErr:  false,
		},
		{
			name:        "File exceeds limit",
			fileSize:    2000,
			maxSize:     1000,
			wantErr:     true,
			errContains: "too large",
		},
		{
			name:     "Empty file",
			fileSize: 0,
			maxSize:  1000,
			wantErr:  false,
		},
		{
			name:        "Very large file",
			fileSize:    10000000,
			maxSize:     1048576, // 1MB
			wantErr:     true,
			errContains: "too large",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")

			// Create file of specific size
			content := make([]byte, tt.fileSize)
			if err := os.WriteFile(tmpFile, content, 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Validate size
			err := ValidateConfigFileSize(tmpFile, tt.maxSize)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfigFileSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateConfigFileSize() error = %v, should contain %q", err, tt.errContains)
				}
			}
		})
	}
}

func TestValidateConfigFileSize_NonexistentFile(t *testing.T) {
	err := ValidateConfigFileSize("/nonexistent/path/config.yaml", 1000)
	if err == nil {
		t.Error("ValidateConfigFileSize() should error for nonexistent file")
	}
	if !strings.Contains(err.Error(), "failed to stat") {
		t.Errorf("Error should mention stat failure, got: %v", err)
	}
}

func TestValidateConfigFileSize_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	// Create a 1-byte file
	if err := os.WriteFile(tmpFile, []byte("x"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test with various max sizes
	testCases := []struct {
		maxSize int64
		wantErr bool
	}{
		{0, true},  // Max size 0 should reject any file
		{1, false}, // Exact match
		{2, false}, // Above size
		{-1, true}, // Negative max size (file size can't be negative, so 1 > -1)
	}

	for _, tc := range testCases {
		err := ValidateConfigFileSize(tmpFile, tc.maxSize)
		if (err != nil) != tc.wantErr {
			t.Errorf("ValidateConfigFileSize(maxSize=%d) error = %v, wantErr %v", tc.maxSize, err, tc.wantErr)
		}
	}
}
