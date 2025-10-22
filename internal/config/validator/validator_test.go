// internal/config/validator/validator_test.go

package validator

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		configContent  string
		permissions    os.FileMode
		wantValid      bool
		wantErrorCount int
		wantWarnings   int
		skipOnWindows  bool
	}{
		{
			name: "Valid config file",
			configContent: `app:
  log_level: debug
  ping:
    output_message: "Test Message"
    output_color: green
    ui: false
`,
			permissions:    0600,
			wantValid:      true,
			wantErrorCount: 0,
			wantWarnings:   0,
			skipOnWindows:  false,
		},
		{
			name: "Config with unknown keys",
			configContent: `app:
  log_level: info
  unknown_key: "value"
  nested:
    also_unknown: "value"
`,
			permissions:    0600,
			wantValid:      true,
			wantErrorCount: 0,
			wantWarnings:   2, // Two unknown keys
			skipOnWindows:  false,
		},
		{
			name: "Invalid YAML syntax",
			configContent: `app:
  log_level: debug
  invalid_yaml: [unclosed
`,
			permissions:    0600,
			wantValid:      false,
			wantErrorCount: 1, // Parse error
			skipOnWindows:  false,
		},
		{
			name: "World-writable file",
			configContent: `app:
  log_level: info
`,
			permissions:    0666,
			wantValid:      false,
			wantErrorCount: 1,    // Permission error
			skipOnWindows:  true, // Skip on Windows
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.skipOnWindows && runtime.GOOS == "windows" {
				t.Skip("Skipping permission test on Windows")
			}

			// Create temp file
			tmpDir := t.TempDir()
			configFile := filepath.Join(tmpDir, "config.yaml")

			if err := os.WriteFile(configFile, []byte(tt.configContent), tt.permissions); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Set permissions explicitly (overcomes umask)
			if err := os.Chmod(configFile, tt.permissions); err != nil {
				t.Fatalf("Failed to chmod test file: %v", err)
			}

			// Validate
			result, err := Validate(configFile)
			if err != nil {
				t.Fatalf("Validate() unexpected error: %v", err)
			}

			if result.Valid != tt.wantValid {
				t.Errorf("Validate() valid = %v, want %v", result.Valid, tt.wantValid)
			}

			if len(result.Errors) != tt.wantErrorCount {
				t.Errorf("Validate() error count = %d, want %d. Errors: %v",
					len(result.Errors), tt.wantErrorCount, result.Errors)
			}

			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("Validate() warning count = %d, want %d. Warnings: %v",
					len(result.Warnings), tt.wantWarnings, result.Warnings)
			}

			if result.ConfigFile != configFile {
				t.Errorf("Validate() config file = %v, want %v", result.ConfigFile, configFile)
			}
		})
	}
}

func TestValidate_NonexistentFile(t *testing.T) {
	t.Parallel()

	_, err := Validate("/nonexistent/config.yaml")
	if err == nil {
		t.Error("Validate() should error for nonexistent file")
	}
}

func TestFindUnknownKeys(t *testing.T) {
	t.Parallel()

	knownKeys := map[string]bool{
		"app.log_level":           true,
		"app.ping.output_message": true,
	}

	tests := []struct {
		name      string
		settings  map[string]interface{}
		prefix    string
		wantCount int
	}{
		{
			name: "No unknown keys",
			settings: map[string]interface{}{
				"app": map[string]interface{}{
					"log_level": "info",
				},
			},
			prefix:    "",
			wantCount: 0,
		},
		{
			name: "One unknown key",
			settings: map[string]interface{}{
				"app": map[string]interface{}{
					"unknown_key": "value",
				},
			},
			prefix:    "",
			wantCount: 1,
		},
		{
			name: "Nested unknown keys",
			settings: map[string]interface{}{
				"app": map[string]interface{}{
					"nested": map[string]interface{}{
						"unknown": "value",
					},
				},
			},
			prefix:    "",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			unknown := findUnknownKeys(tt.settings, tt.prefix, knownKeys)
			if len(unknown) != tt.wantCount {
				t.Errorf("findUnknownKeys() found %d unknown keys, want %d. Keys: %v",
					len(unknown), tt.wantCount, unknown)
			}
		})
	}
}
