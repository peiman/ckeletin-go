// internal/config/validator/validator_test.go

package validator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/peiman/ckeletin-go/internal/testutil"
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

			if tt.skipOnWindows {
				testutil.SkipOnWindowsWithReason(t, "permission test requires Unix file permissions")
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

// FuzzValidate performs fuzz testing on the end-to-end Validate function.
// This tests the complete validation workflow including YAML parsing, security checks,
// and value validation with malformed config file content.
func FuzzValidate(f *testing.F) {
	// Seed corpus with interesting YAML structures
	f.Add("app:\n  log_level: info\n")
	f.Add("app:\n  nested:\n    key: value\n")
	f.Add("malformed: [unclosed\n")
	f.Add("app:\n  very_long_key: " + "x")
	f.Add("app:\n  special_chars: \"!@#$%^&*()\"\n")

	f.Fuzz(func(t *testing.T, configContent string) {
		// Create temp file with fuzzed content
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		if err := os.WriteFile(configFile, []byte(configContent), 0600); err != nil {
			t.Skip("Failed to create test file")
		}

		// Validate should never panic, even with malformed input
		result, err := Validate(configFile)

		// If Validate returns an error, it's a system error (not validation error)
		// In that case, we can't make assertions about the result
		if err != nil {
			// System errors are acceptable (e.g., file I/O failures)
			return
		}

		// Result should always have a ConfigFile set
		if result.ConfigFile != configFile {
			t.Errorf("Validate() ConfigFile = %v, want %v", result.ConfigFile, configFile)
		}

		// If there are errors, Valid should be false
		if len(result.Errors) > 0 && result.Valid {
			t.Errorf("Validate() has errors but Valid=true: %v", result.Errors)
		}

		// If Valid is true, there should be no errors
		if result.Valid && len(result.Errors) > 0 {
			t.Errorf("Validate() Valid=true but has errors: %v", result.Errors)
		}
	})
}

// FuzzFindUnknownKeys performs fuzz testing on the recursive findUnknownKeys function.
// This tests the key traversal logic with deeply nested structures and special characters.
func FuzzFindUnknownKeys(f *testing.F) {
	// Seed corpus with interesting key patterns
	f.Add("prefix", "key")
	f.Add("app.nested", "value")
	f.Add("", "root")
	f.Add("deep.nest.level", "key")
	f.Add("special..chars", "test")

	f.Fuzz(func(t *testing.T, prefix string, key string) {
		// Skip empty keys
		if key == "" {
			t.Skip()
		}

		// Create a simple known keys map
		knownKeys := map[string]bool{
			"app.log_level":           true,
			"app.ping.output_message": true,
		}

		// Test with simple string value
		settings := map[string]interface{}{
			key: "value",
		}

		// findUnknownKeys should never panic
		unknown := findUnknownKeys(settings, prefix, knownKeys)

		// Result should be a slice (possibly empty)
		if unknown == nil {
			t.Error("findUnknownKeys() returned nil, expected empty slice")
		}

		// Test with nested structure
		nestedSettings := map[string]interface{}{
			key: map[string]interface{}{
				"nested": "value",
			},
		}

		// Should handle nested maps without panic
		unknown = findUnknownKeys(nestedSettings, prefix, knownKeys)
		if unknown == nil {
			t.Error("findUnknownKeys() returned nil for nested map, expected slice")
		}

		// Test with slice value
		sliceSettings := map[string]interface{}{
			key: []interface{}{"item1", "item2"},
		}

		// Should handle slices without panic
		unknown = findUnknownKeys(sliceSettings, prefix, knownKeys)
		if unknown == nil {
			t.Error("findUnknownKeys() returned nil for slice value, expected slice")
		}
	})
}
