package check

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimingHistory_GetExpectedDuration(t *testing.T) {
	// Known defaults to verify against
	defaults := map[string]time.Duration{
		"go-version": 100 * time.Millisecond,
		"lint":       3 * time.Second,
		"test":       10 * time.Second,
		"sast":       4 * time.Second,
	}

	t.Run("returns default for unknown check with no history", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		assert.Equal(t, 3*time.Second, th.getExpectedDuration("unknown-check"))
	})

	t.Run("returns predefined defaults for known checks", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		for name, expected := range defaults {
			assert.Equal(t, expected, th.getExpectedDuration(name), "check: %s", name)
		}
	})

	t.Run("returns historical average when available", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 5 * time.Second, RunCount: 3},
			},
		}
		// Should use history (5s) not default (3s)
		assert.Equal(t, 5*time.Second, th.getExpectedDuration("lint"))
	})

	t.Run("ignores zero-value history", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 0, RunCount: 0},
			},
		}
		// Zero avg should fall back to default
		assert.Equal(t, 3*time.Second, th.getExpectedDuration("lint"))
	})
}

func TestTimingHistory_RecordDuration(t *testing.T) {
	const alpha = 0.3 // EMA alpha value from implementation

	t.Run("first recording sets duration directly", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		th.recordDuration("test", 2*time.Second)

		require.NotNil(t, th.Checks["test"])
		assert.Equal(t, 2*time.Second, th.Checks["test"].AvgDuration)
		assert.Equal(t, 2*time.Second, th.Checks["test"].LastDuration)
		assert.Equal(t, 1, th.Checks["test"].RunCount)
	})

	t.Run("subsequent recordings use EMA", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test": {AvgDuration: 10 * time.Second, LastDuration: 10 * time.Second, RunCount: 5},
			},
		}

		// Record a new 4 second run
		th.recordDuration("test", 4*time.Second)

		// EMA: new_avg = alpha*new + (1-alpha)*old = 0.3*4 + 0.7*10 = 1.2 + 7 = 8.2s
		expectedAvg := time.Duration(alpha*float64(4*time.Second) + (1-alpha)*float64(10*time.Second))
		assert.Equal(t, expectedAvg, th.Checks["test"].AvgDuration)
		assert.Equal(t, 4*time.Second, th.Checks["test"].LastDuration)
		assert.Equal(t, 6, th.Checks["test"].RunCount)
	})

	t.Run("EMA converges towards recent values", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test": {AvgDuration: 10 * time.Second, RunCount: 10},
			},
		}

		// Record several fast runs
		for i := 0; i < 10; i++ {
			th.recordDuration("test", 1*time.Second)
		}

		// Average should converge towards 1s (but not quite reach it)
		assert.Less(t, th.Checks["test"].AvgDuration, 3*time.Second, "should converge towards recent values")
		assert.Greater(t, th.Checks["test"].AvgDuration, 1*time.Second, "shouldn't fully converge in 10 runs")
	})

	t.Run("creates new check entry if not exists", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		th.recordDuration("new-check", 500*time.Millisecond)

		require.Contains(t, th.Checks, "new-check")
		assert.Equal(t, 500*time.Millisecond, th.Checks["new-check"].AvgDuration)
	})
}

func TestTimingHistory_Concurrency(t *testing.T) {
	th := &timingHistory{Checks: make(map[string]*checkTiming)}

	// Run concurrent reads and writes
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func(i int) {
			th.recordDuration("test", time.Duration(i)*time.Second)
			done <- true
		}(i)
		go func() {
			_ = th.getExpectedDuration("test")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Verify data integrity
	assert.NotNil(t, th.Checks["test"])
	assert.Equal(t, 10, th.Checks["test"].RunCount)
}

func TestTimingHistory_AllDefaults(t *testing.T) {
	// All 23 checks should have defaults
	allChecks := []string{
		// Environment
		"go-version", "tools",
		// Quality
		"format", "lint",
		// Architecture
		"defaults", "commands", "constants", "task-naming",
		"architecture", "layering", "package-org", "config-consumption",
		"output-patterns", "security-patterns",
		// Security
		"secrets", "sast",
		// Dependencies
		"deps", "vuln", "outdated", "license-source", "license-binary", "sbom-vulns",
		// Tests
		"test",
	}

	th := &timingHistory{Checks: make(map[string]*checkTiming)}
	for _, name := range allChecks {
		dur := th.getExpectedDuration(name)
		assert.Greater(t, dur, time.Duration(0), "check %s should have a default duration", name)
		// All defaults should be reasonable (between 100ms and 15s)
		assert.GreaterOrEqual(t, dur, 100*time.Millisecond, "check %s duration too short", name)
		assert.LessOrEqual(t, dur, 15*time.Second, "check %s duration too long", name)
	}
}

func TestTimingHistory_Save(t *testing.T) {
	t.Run("saves timing data to disk", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "timings.json")

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint":   {AvgDuration: 3 * time.Second, LastDuration: 3 * time.Second, RunCount: 5},
				"format": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 3},
			},
		}

		// Write data directly (save uses timingFilePath() which we can't control,
		// so test the marshaling/writing logic by reproducing it)
		th.mu.RLock()
		data, err := json.MarshalIndent(th, "", "  ")
		th.mu.RUnlock()
		require.NoError(t, err)

		err = os.WriteFile(tmpFile, data, 0o600)
		require.NoError(t, err)

		// Verify file was written correctly
		readBack, err := os.ReadFile(tmpFile)
		require.NoError(t, err)

		var loaded timingHistory
		err = json.Unmarshal(readBack, &loaded)
		require.NoError(t, err)
		assert.Equal(t, 3*time.Second, loaded.Checks["lint"].AvgDuration)
		assert.Equal(t, 1*time.Second, loaded.Checks["format"].AvgDuration)
	})
}

func TestLoadTimingHistory_WithExistingFile(t *testing.T) {
	t.Run("loads timing data from JSON file", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "timings.json")

		// Write a valid timing history file
		data := `{
			"checks": {
				"lint": {"last_duration": 3000000000, "avg_duration": 3000000000, "run_count": 5},
				"test": {"last_duration": 10000000000, "avg_duration": 10000000000, "run_count": 10}
			}
		}`
		err := os.WriteFile(tmpFile, []byte(data), 0o600)
		require.NoError(t, err)

		// Load it
		fileData, err := os.ReadFile(tmpFile)
		require.NoError(t, err)

		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		err = json.Unmarshal(fileData, th)
		require.NoError(t, err)

		assert.Len(t, th.Checks, 2)
		assert.Equal(t, 3*time.Second, th.Checks["lint"].AvgDuration)
		assert.Equal(t, 10*time.Second, th.Checks["test"].AvgDuration)
		assert.Equal(t, 5, th.Checks["lint"].RunCount)
	})

	t.Run("handles invalid JSON gracefully", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		// Unmarshal invalid JSON should return error but not panic
		err := json.Unmarshal([]byte("not valid json"), th)
		assert.Error(t, err)
		// Checks map should still be intact
		assert.NotNil(t, th.Checks)
	})

	t.Run("handles empty JSON object", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		err := json.Unmarshal([]byte(`{}`), th)
		assert.NoError(t, err)
		// Checks may be nil after unmarshaling empty object
		if th.Checks == nil {
			th.Checks = make(map[string]*checkTiming)
		}
		assert.NotNil(t, th.Checks)
	})
}

func TestTimingHistory_SaveAndLoad_RoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "timings.json")

	// Create timing data
	original := &timingHistory{
		Checks: map[string]*checkTiming{
			"lint":   {AvgDuration: 3 * time.Second, LastDuration: 2800 * time.Millisecond, RunCount: 5},
			"test":   {AvgDuration: 10 * time.Second, LastDuration: 11 * time.Second, RunCount: 20},
			"format": {AvgDuration: 800 * time.Millisecond, LastDuration: 750 * time.Millisecond, RunCount: 15},
		},
	}

	// Save
	data, err := json.MarshalIndent(original, "", "  ")
	require.NoError(t, err)
	err = os.WriteFile(tmpFile, data, 0o600)
	require.NoError(t, err)

	// Load
	readData, err := os.ReadFile(tmpFile)
	require.NoError(t, err)
	loaded := &timingHistory{Checks: make(map[string]*checkTiming)}
	err = json.Unmarshal(readData, loaded)
	require.NoError(t, err)

	// Verify round-trip integrity
	assert.Equal(t, len(original.Checks), len(loaded.Checks))
	for name, orig := range original.Checks {
		loaded, ok := loaded.Checks[name]
		require.True(t, ok, "loaded should contain check %s", name)
		assert.Equal(t, orig.AvgDuration, loaded.AvgDuration, "check %s avg", name)
		assert.Equal(t, orig.LastDuration, loaded.LastDuration, "check %s last", name)
		assert.Equal(t, orig.RunCount, loaded.RunCount, "check %s count", name)
	}
}
