// internal/config/security_test.go

package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigFilePermissions(t *testing.T) {
	t.Parallel()
	testutil.SkipOnWindowsWithReason(t, "permission tests require Unix file permissions")

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
			t.Parallel()
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")

			require.NoError(t,
				os.WriteFile(tmpFile, []byte("test: value\n"), tt.permissions),
				"Failed to create test file")

			// Explicitly set permissions to overcome umask
			require.NoError(t, os.Chmod(tmpFile, tt.permissions), "Failed to chmod test file")

			// Validate permissions
			err := ValidateConfigFilePermissions(tmpFile)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errContains),
						"ValidateConfigFilePermissions() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigFilePermissions_NonexistentFile(t *testing.T) {
	t.Parallel()
	testutil.SkipOnWindowsWithReason(t, "permission tests require Unix file permissions")

	err := ValidateConfigFilePermissions("/nonexistent/path/config.yaml")
	require.Error(t, err, "ValidateConfigFilePermissions() should error for nonexistent file")
	assert.True(t, strings.Contains(err.Error(), "failed to stat"),
		"Error should mention stat failure, got: %v", err)
}

func TestValidateConfigFilePermissions_Windows(t *testing.T) {
	testutil.SkipOnNonWindows(t)

	// On Windows, should always return nil
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	require.NoError(t,
		os.WriteFile(tmpFile, []byte("test: value\n"), 0666),
		"Failed to create test file")

	err := ValidateConfigFilePermissions(tmpFile)
	assert.NoError(t, err, "ValidateConfigFilePermissions() on Windows should return nil")
}

func TestValidateConfigFileSize(t *testing.T) {
	t.Parallel()
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
			t.Parallel()
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")

			// Create file of specific size
			content := make([]byte, tt.fileSize)
			require.NoError(t,
				os.WriteFile(tmpFile, content, 0600),
				"Failed to create test file")

			// Validate size
			err := ValidateConfigFileSize(tmpFile, tt.maxSize)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errContains),
						"ValidateConfigFileSize() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigFileSize_NonexistentFile(t *testing.T) {
	t.Parallel()
	err := ValidateConfigFileSize("/nonexistent/path/config.yaml", 1000)
	require.Error(t, err, "ValidateConfigFileSize() should error for nonexistent file")
	assert.True(t, strings.Contains(err.Error(), "failed to stat"),
		"Error should mention stat failure, got: %v", err)
}

func TestValidateConfigFileSize_EdgeCases(t *testing.T) {
	t.Parallel()
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "config.yaml")

	// Create a 1-byte file
	require.NoError(t,
		os.WriteFile(tmpFile, []byte("x"), 0600),
		"Failed to create test file")

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
		if tc.wantErr {
			assert.Error(t, err, "ValidateConfigFileSize(maxSize=%d) should error", tc.maxSize)
		} else {
			assert.NoError(t, err, "ValidateConfigFileSize(maxSize=%d) should not error", tc.maxSize)
		}
	}
}

func TestValidateConfigFileSecurity(t *testing.T) {
	testutil.SkipOnWindowsWithReason(t, "security validation requires Unix file permissions")

	tests := []struct {
		name        string
		fileSize    int64
		maxSize     int64
		permissions os.FileMode
		wantErr     bool
		errContains string
	}{
		{
			name:        "Valid file with secure permissions",
			fileSize:    100,
			maxSize:     1000,
			permissions: 0600,
			wantErr:     false,
		},
		{
			name:        "File too large",
			fileSize:    2000,
			maxSize:     1000,
			permissions: 0600,
			wantErr:     true,
			errContains: "too large",
		},
		{
			name:        "File with world-writable permissions",
			fileSize:    100,
			maxSize:     1000,
			permissions: 0666,
			wantErr:     true,
			errContains: "world-writable",
		},
		{
			name:        "Both size and permission violations - size checked first",
			fileSize:    2000,
			maxSize:     1000,
			permissions: 0666,
			wantErr:     true,
			errContains: "too large", // Size is checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary file with the specified size
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "config.yaml")

			// Write file with specified size
			content := make([]byte, tt.fileSize)
			for i := range content {
				content[i] = 'a'
			}
			require.NoError(t,
				os.WriteFile(tmpFile, content, 0600),
				"Failed to create test file")

			// Set the exact permissions we want (avoiding umask issues)
			require.NoError(t, os.Chmod(tmpFile, tt.permissions), "Failed to set file permissions")

			// Run validation
			err := ValidateConfigFileSecurity(tmpFile, tt.maxSize)

			// Check error expectation
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errContains),
						"ValidateConfigFileSecurity() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConfigFileSecurity_NonexistentFile(t *testing.T) {
	err := ValidateConfigFileSecurity("/nonexistent/path/config.yaml", 1000)
	require.Error(t, err, "ValidateConfigFileSecurity() should error for nonexistent file")
	assert.True(t, strings.Contains(err.Error(), "failed to stat"),
		"Expected error message to contain 'failed to stat', got: %v", err)
}

func TestValidateLogFilePath(t *testing.T) {
	t.Parallel()
	validate := ValidateLogFilePath()

	tests := []struct {
		name        string
		value       interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:    "Default relative path",
			value:   "./logs/ckeletin-go.log",
			wantErr: false,
		},
		{
			name:    "Absolute path",
			value:   "/var/log/ckeletin-go/app.log",
			wantErr: false,
		},
		{
			name:    "Relative path without leading dot",
			value:   "logs/app.log",
			wantErr: false,
		},
		{
			name:        "Parent directory traversal",
			value:       "../escape.log",
			wantErr:     true,
			errContains: "traversal",
		},
		{
			name:        "Nested traversal escaping the tree",
			value:       "logs/../../escape.log",
			wantErr:     true,
			errContains: "traversal",
		},
		{
			name:    "Inner traversal that stays inside cleans away",
			value:   "logs/../app.log",
			wantErr: false,
		},
		{
			name:    "Component with leading dots is not traversal",
			value:   "logs/..hidden/app.log",
			wantErr: false,
		},
		{
			name:        "Empty path",
			value:       "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:    "Nil value is skipped",
			value:   nil,
			wantErr: false,
		},
		{
			name:    "Non-string value is skipped",
			value:   42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validate(tt.value)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errContains),
						"ValidateLogFilePath() error = %v, should contain %q", err, tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateLogFilePath_ExistingFiles(t *testing.T) {
	t.Parallel()
	testutil.SkipOnWindowsWithReason(t, "symlink creation requires elevated privileges on Windows")

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "real.log")
	require.NoError(t,
		os.WriteFile(target, []byte("log line\n"), 0600),
		"Failed to create target file")

	link := filepath.Join(tmpDir, "link.log")
	require.NoError(t, os.Symlink(target, link), "Failed to create symlink")

	validate := ValidateLogFilePath()

	assert.NoError(t, validate(target),
		"existing regular file should be accepted")

	err := validate(link)
	require.Error(t, err, "existing symlink should be rejected")
	assert.True(t, strings.Contains(err.Error(), "symlink"),
		"ValidateLogFilePath() error = %v, should contain %q", err, "symlink")

	missing := filepath.Join(tmpDir, "missing", "app.log")
	assert.NoError(t, validate(missing),
		"missing file should be accepted (it is created on first write)")
}
