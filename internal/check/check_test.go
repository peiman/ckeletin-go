package check

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewExecutor(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{FailFast: true, Verbose: true}

	executor := NewExecutor(cfg, &buf)

	require.NotNil(t, executor)
	assert.Equal(t, cfg, executor.cfg)
	assert.NotNil(t, executor.writer)
	assert.NotNil(t, executor.timings)
}

func TestConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfg      Config
		failFast bool
		verbose  bool
		parallel bool
	}{
		{
			name:     "default config",
			cfg:      Config{},
			failFast: false,
			verbose:  false,
			parallel: false,
		},
		{
			name:     "fail fast enabled",
			cfg:      Config{FailFast: true},
			failFast: true,
			verbose:  false,
			parallel: false,
		},
		{
			name:     "verbose enabled",
			cfg:      Config{Verbose: true},
			failFast: false,
			verbose:  true,
			parallel: false,
		},
		{
			name:     "parallel enabled",
			cfg:      Config{Parallel: true},
			failFast: false,
			verbose:  false,
			parallel: true,
		},
		{
			name:     "all enabled",
			cfg:      Config{FailFast: true, Verbose: true, Parallel: true},
			failFast: true,
			verbose:  true,
			parallel: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.failFast, tt.cfg.FailFast)
			assert.Equal(t, tt.verbose, tt.cfg.Verbose)
			assert.Equal(t, tt.parallel, tt.cfg.Parallel)
		})
	}
}

func TestValidateCategories(t *testing.T) {
	tests := []struct {
		name       string
		categories []string
		wantErr    bool
	}{
		{
			name:       "valid single category",
			categories: []string{"environment"},
			wantErr:    false,
		},
		{
			name:       "valid multiple categories",
			categories: []string{"environment", "quality", "tests"},
			wantErr:    false,
		},
		{
			name:       "all valid categories",
			categories: AllCategories,
			wantErr:    false,
		},
		{
			name:       "invalid category",
			categories: []string{"invalid"},
			wantErr:    true,
		},
		{
			name:       "mixed valid and invalid",
			categories: []string{"environment", "invalid"},
			wantErr:    true,
		},
		{
			name:       "empty slice",
			categories: []string{},
			wantErr:    false,
		},
		{
			name:       "case insensitive",
			categories: []string{"ENVIRONMENT", "Quality", "TESTS"},
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCategories(tt.categories)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAllCategories(t *testing.T) {
	// Verify all expected categories exist
	expectedCategories := []string{
		CategoryEnvironment,
		CategoryQuality,
		CategoryArchitecture,
		CategorySecurity,
		CategoryDependencies,
		CategoryTests,
	}

	assert.Equal(t, expectedCategories, AllCategories)
	assert.Len(t, AllCategories, 6)
}

func TestCategoryConstants(t *testing.T) {
	// Verify category constants have expected values
	assert.Equal(t, "environment", CategoryEnvironment)
	assert.Equal(t, "quality", CategoryQuality)
	assert.Equal(t, "architecture", CategoryArchitecture)
	assert.Equal(t, "security", CategorySecurity)
	assert.Equal(t, "dependencies", CategoryDependencies)
	assert.Equal(t, "tests", CategoryTests)
}

func TestExecutor_BuildCategories(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{}

	executor := NewExecutor(cfg, &buf)
	methods := &checkMethods{cfg: cfg}
	categories := executor.buildCategories(methods)

	// Verify we have all 6 categories
	assert.Len(t, categories, 6)

	// Verify category names and check counts
	expectedCategories := map[string]int{
		"Development Environment": 2,
		"Code Quality":            2,
		"Architecture Validation": 10,
		"Security Scanning":       2,
		"Dependencies":            6,
		"Tests":                   1,
	}

	for _, cat := range categories {
		expectedCount, ok := expectedCategories[cat.name]
		require.True(t, ok, "unexpected category: %s", cat.name)
		assert.Len(t, cat.checks, expectedCount, "wrong check count for %s", cat.name)
	}

	// Verify total check count is 23
	total := 0
	for _, cat := range categories {
		total += len(cat.checks)
	}
	assert.Equal(t, 23, total, "should have 23 total checks")
}

func TestExecutor_ShouldRunCategory(t *testing.T) {
	var buf bytes.Buffer

	tests := []struct {
		name         string
		categories   []string
		categoryName string
		want         bool
	}{
		{
			// Note: When categories is nil, Execute() doesn't call shouldRunCategory
			// so this returns false (no match), but the caller handles this case
			name:         "no filter returns false (caller handles this)",
			categories:   nil,
			categoryName: "Development Environment",
			want:         false,
		},
		{
			name:         "filter matches environment",
			categories:   []string{"environment"},
			categoryName: "Development Environment",
			want:         true,
		},
		{
			name:         "filter doesn't match",
			categories:   []string{"security"},
			categoryName: "Development Environment",
			want:         false,
		},
		{
			name:         "filter matches quality",
			categories:   []string{"quality"},
			categoryName: "Code Quality",
			want:         true,
		},
		{
			name:         "filter matches architecture",
			categories:   []string{"architecture"},
			categoryName: "Architecture Validation",
			want:         true,
		},
		{
			name:         "case insensitive filter",
			categories:   []string{"SECURITY"},
			categoryName: "Security Scanning",
			want:         true,
		},
		{
			name:         "unknown category returns true",
			categories:   []string{"security"},
			categoryName: "Unknown Category",
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Categories: tt.categories}
			executor := NewExecutor(cfg, &buf)

			got := executor.shouldRunCategory(tt.categoryName)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldUseTUI(t *testing.T) {
	// Helper to save and restore environment variables
	saveEnv := func(keys []string) map[string]string {
		saved := make(map[string]string)
		for _, key := range keys {
			saved[key] = os.Getenv(key)
		}
		return saved
	}
	restoreEnv := func(saved map[string]string) {
		for key, val := range saved {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}

	// All CI-related environment variables to test
	ciEnvVars := []string{
		"CI", "GITHUB_ACTIONS", "GITLAB_CI", "JENKINS_URL",
		"CIRCLECI", "TRAVIS", "BUILDKITE", "TF_BUILD",
		"NO_COLOR", "TERM",
	}

	tests := []struct {
		name    string
		envVars map[string]string
		want    bool
	}{
		{
			name:    "CI environment variable set",
			envVars: map[string]string{"CI": "true"},
			want:    false,
		},
		{
			name:    "GITHUB_ACTIONS set",
			envVars: map[string]string{"GITHUB_ACTIONS": "true"},
			want:    false,
		},
		{
			name:    "GITLAB_CI set",
			envVars: map[string]string{"GITLAB_CI": "true"},
			want:    false,
		},
		{
			name:    "JENKINS_URL set",
			envVars: map[string]string{"JENKINS_URL": "http://jenkins"},
			want:    false,
		},
		{
			name:    "CIRCLECI set",
			envVars: map[string]string{"CIRCLECI": "true"},
			want:    false,
		},
		{
			name:    "TRAVIS set",
			envVars: map[string]string{"TRAVIS": "true"},
			want:    false,
		},
		{
			name:    "BUILDKITE set",
			envVars: map[string]string{"BUILDKITE": "true"},
			want:    false,
		},
		{
			name:    "TF_BUILD set (Azure DevOps)",
			envVars: map[string]string{"TF_BUILD": "True"},
			want:    false,
		},
		{
			name:    "NO_COLOR set",
			envVars: map[string]string{"NO_COLOR": "1"},
			want:    false,
		},
		{
			name:    "TERM is dumb",
			envVars: map[string]string{"TERM": "dumb"},
			want:    false,
		},
		{
			name:    "no CI environment (non-TTY buffer)",
			envVars: map[string]string{},
			want:    false, // bytes.Buffer is not a TTY
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current environment
			saved := saveEnv(ciEnvVars)
			defer restoreEnv(saved)

			// Clear all CI environment variables first
			for _, key := range ciEnvVars {
				os.Unsetenv(key)
			}

			// Set test-specific environment variables
			for key, val := range tt.envVars {
				os.Setenv(key, val)
			}

			// Test with a buffer (non-TTY)
			var buf bytes.Buffer
			got := shouldUseTUI(&buf)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShouldUseTUI_MultipleEnvVars(t *testing.T) {
	// Save and restore environment
	ciEnvVars := []string{"CI", "GITHUB_ACTIONS", "NO_COLOR"}
	saved := make(map[string]string)
	for _, key := range ciEnvVars {
		saved[key] = os.Getenv(key)
		os.Unsetenv(key)
	}
	defer func() {
		for key, val := range saved {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	// Test that first matching CI var causes TUI to be disabled
	os.Setenv("CI", "true")
	os.Setenv("GITHUB_ACTIONS", "true")

	var buf bytes.Buffer
	got := shouldUseTUI(&buf)
	assert.False(t, got, "should return false when multiple CI vars are set")
}

func TestExecutor_UseTUI(t *testing.T) {
	// Save and restore environment
	saved := os.Getenv("CI")
	defer func() {
		if saved == "" {
			os.Unsetenv("CI")
		} else {
			os.Setenv("CI", saved)
		}
	}()

	t.Run("useTUI is false in CI environment", func(t *testing.T) {
		os.Setenv("CI", "true")

		var buf bytes.Buffer
		executor := NewExecutor(Config{}, &buf)

		assert.False(t, executor.useTUI, "executor should have useTUI=false in CI")
	})

	t.Run("useTUI is false for non-TTY writer", func(t *testing.T) {
		os.Unsetenv("CI")

		var buf bytes.Buffer
		executor := NewExecutor(Config{}, &buf)

		// Buffer is not a TTY, so useTUI should be false
		assert.False(t, executor.useTUI, "executor should have useTUI=false for non-TTY")
	})
}
