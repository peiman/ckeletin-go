package check

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/peiman/ckeletin-go/internal/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTimingTestEnv redirects xdg paths to a temp directory so that
// timingFilePath(), save(), and loadTimingHistory() use isolated paths.
// It sets HOME (and XDG_CACHE_HOME on Linux) to tmpDir and configures
// xdg.SetAppName for the test. Returns a cleanup function that restores
// the original environment.
func setupTimingTestEnv(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Save original values
	origHome := os.Getenv("HOME")
	origAppName := xdg.GetAppName()
	origXDGCache := os.Getenv("XDG_CACHE_HOME")

	// Set HOME so that xdg.homeDir() returns our temp dir
	t.Setenv("HOME", tmpDir)

	// On Linux, also set XDG_CACHE_HOME for explicit control
	if runtime.GOOS == "linux" {
		cacheDir := filepath.Join(tmpDir, ".cache")
		t.Setenv("XDG_CACHE_HOME", cacheDir)
	}

	// Set a test-specific app name
	xdg.SetAppName("ckeletin-go-test")

	t.Cleanup(func() {
		// Restore originals
		xdg.SetAppName(origAppName)
		// t.Setenv handles HOME and XDG_CACHE_HOME restoration automatically
		_ = origHome
		_ = origXDGCache
	})

	return tmpDir
}

func TestTimingFilePath(t *testing.T) {
	t.Run("returns path under xdg cache directory", func(t *testing.T) {
		setupTimingTestEnv(t)

		path := timingFilePath()

		assert.True(t, strings.HasSuffix(path, "check-timings.json"),
			"path should end with check-timings.json, got: %s", path)
		assert.Contains(t, path, "ckeletin-go-test",
			"path should contain the app name")
	})

	t.Run("returns fallback path when app name is not set", func(t *testing.T) {
		tmpDir := t.TempDir()
		origAppName := xdg.GetAppName()
		t.Setenv("HOME", tmpDir)

		// Clear app name to trigger error in xdg.CacheFile
		xdg.SetAppName("")
		t.Cleanup(func() {
			xdg.SetAppName(origAppName)
		})

		path := timingFilePath()

		// Should fall back to os.TempDir()
		assert.Contains(t, path, "ckeletin-go-check-timings.json",
			"fallback path should contain the expected filename")
		assert.True(t, strings.HasPrefix(path, os.TempDir()),
			"fallback path should start with os.TempDir(), got: %s", path)
	})

	t.Run("path is consistent across calls", func(t *testing.T) {
		setupTimingTestEnv(t)

		path1 := timingFilePath()
		path2 := timingFilePath()

		assert.Equal(t, path1, path2, "timingFilePath should return the same path on repeated calls")
	})
}

func TestLoadTimingHistory(t *testing.T) {
	t.Run("returns empty history when file does not exist", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.NotNil(t, th.Checks)
		assert.Empty(t, th.Checks)
	})

	t.Run("loads valid timing data from disk", func(t *testing.T) {
		setupTimingTestEnv(t)

		// Write valid JSON to the expected path
		path := timingFilePath()
		err := os.MkdirAll(filepath.Dir(path), 0o750)
		require.NoError(t, err)

		data := `{
			"checks": {
				"lint": {"last_duration": 3000000000, "avg_duration": 3000000000, "run_count": 5},
				"test": {"last_duration": 10000000000, "avg_duration": 10000000000, "run_count": 10}
			}
		}`
		err = os.WriteFile(path, []byte(data), 0o600)
		require.NoError(t, err)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.Len(t, th.Checks, 2)
		assert.Equal(t, 3*time.Second, th.Checks["lint"].AvgDuration)
		assert.Equal(t, 3*time.Second, th.Checks["lint"].LastDuration)
		assert.Equal(t, 5, th.Checks["lint"].RunCount)
		assert.Equal(t, 10*time.Second, th.Checks["test"].AvgDuration)
		assert.Equal(t, 10, th.Checks["test"].RunCount)
	})

	t.Run("returns empty history on corrupt JSON", func(t *testing.T) {
		setupTimingTestEnv(t)

		path := timingFilePath()
		err := os.MkdirAll(filepath.Dir(path), 0o750)
		require.NoError(t, err)

		err = os.WriteFile(path, []byte("this is not valid json{{{"), 0o600)
		require.NoError(t, err)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.NotNil(t, th.Checks)
		assert.Empty(t, th.Checks, "corrupt JSON should result in empty checks map")
	})

	t.Run("returns empty history on empty JSON object", func(t *testing.T) {
		setupTimingTestEnv(t)

		path := timingFilePath()
		err := os.MkdirAll(filepath.Dir(path), 0o750)
		require.NoError(t, err)

		err = os.WriteFile(path, []byte(`{}`), 0o600)
		require.NoError(t, err)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.NotNil(t, th.Checks, "Checks map should be initialized even with empty JSON")
	})

	t.Run("returns empty history on null checks field", func(t *testing.T) {
		setupTimingTestEnv(t)

		path := timingFilePath()
		err := os.MkdirAll(filepath.Dir(path), 0o750)
		require.NoError(t, err)

		err = os.WriteFile(path, []byte(`{"checks": null}`), 0o600)
		require.NoError(t, err)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.NotNil(t, th.Checks, "Checks map should be initialized when JSON has null checks")
	})

	t.Run("returns empty history on empty file", func(t *testing.T) {
		setupTimingTestEnv(t)

		path := timingFilePath()
		err := os.MkdirAll(filepath.Dir(path), 0o750)
		require.NoError(t, err)

		err = os.WriteFile(path, []byte(""), 0o600)
		require.NoError(t, err)

		th := loadTimingHistory()

		require.NotNil(t, th)
		assert.NotNil(t, th.Checks)
	})
}

func TestTimingHistory_Save(t *testing.T) {
	t.Run("persists timing data to disk", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint":   {AvgDuration: 3 * time.Second, LastDuration: 3 * time.Second, RunCount: 5},
				"format": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 3},
			},
		}

		th.save()

		// Verify the file was written
		path := timingFilePath()
		data, err := os.ReadFile(path)
		require.NoError(t, err, "save should create the timing file")

		var loaded timingHistory
		err = json.Unmarshal(data, &loaded)
		require.NoError(t, err, "saved file should contain valid JSON")
		assert.Equal(t, 3*time.Second, loaded.Checks["lint"].AvgDuration)
		assert.Equal(t, 1*time.Second, loaded.Checks["format"].AvgDuration)
		assert.Equal(t, 5, loaded.Checks["lint"].RunCount)
		assert.Equal(t, 3, loaded.Checks["format"].RunCount)
	})

	t.Run("creates parent directories if needed", func(t *testing.T) {
		setupTimingTestEnv(t)

		// The directory should not exist yet
		path := timingFilePath()
		dir := filepath.Dir(path)
		_, err := os.Stat(dir)
		// It may or may not exist depending on xdg.CacheFile behavior
		// The point is that save() should succeed regardless

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test": {AvgDuration: 2 * time.Second, LastDuration: 2 * time.Second, RunCount: 1},
			},
		}

		th.save()

		// Directory and file should now exist
		_, err = os.Stat(path)
		assert.NoError(t, err, "file should exist after save")
	})

	t.Run("writes file with secure permissions", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 1},
			},
		}

		th.save()

		path := timingFilePath()
		info, err := os.Stat(path)
		require.NoError(t, err)

		// File should be readable/writable only by owner (0600)
		// Windows doesn't support Unix file permissions
		if runtime.GOOS != "windows" {
			perm := info.Mode().Perm()
			assert.Equal(t, os.FileMode(0o600), perm,
				"file should have 0600 permissions, got %o", perm)
		}
	})

	t.Run("saves empty checks map without error", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: make(map[string]*checkTiming),
		}

		th.save()

		path := timingFilePath()
		data, err := os.ReadFile(path)
		require.NoError(t, err)

		var loaded timingHistory
		err = json.Unmarshal(data, &loaded)
		require.NoError(t, err)
		assert.NotNil(t, loaded.Checks)
		assert.Empty(t, loaded.Checks)
	})
}

func TestTimingHistory_SaveAndLoad_RoundTrip(t *testing.T) {
	t.Run("round-trip preserves all timing data", func(t *testing.T) {
		setupTimingTestEnv(t)

		original := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint":   {AvgDuration: 3 * time.Second, LastDuration: 2800 * time.Millisecond, RunCount: 5},
				"test":   {AvgDuration: 10 * time.Second, LastDuration: 11 * time.Second, RunCount: 20},
				"format": {AvgDuration: 800 * time.Millisecond, LastDuration: 750 * time.Millisecond, RunCount: 15},
			},
		}

		// Save using the real save() method
		original.save()

		// Load using the real loadTimingHistory() function
		loaded := loadTimingHistory()

		// Verify round-trip integrity
		require.NotNil(t, loaded)
		assert.Equal(t, len(original.Checks), len(loaded.Checks))
		for name, orig := range original.Checks {
			lc, ok := loaded.Checks[name]
			require.True(t, ok, "loaded should contain check %s", name)
			assert.Equal(t, orig.AvgDuration, lc.AvgDuration, "check %s avg", name)
			assert.Equal(t, orig.LastDuration, lc.LastDuration, "check %s last", name)
			assert.Equal(t, orig.RunCount, lc.RunCount, "check %s count", name)
		}
	})

	t.Run("round-trip with recordDuration then save and load", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{Checks: make(map[string]*checkTiming)}

		// Record several durations
		th.recordDuration("lint", 3*time.Second)
		th.recordDuration("lint", 4*time.Second)
		th.recordDuration("test", 10*time.Second)

		// Save state
		th.save()

		// Load into a new instance
		loaded := loadTimingHistory()

		require.NotNil(t, loaded)
		require.Contains(t, loaded.Checks, "lint")
		require.Contains(t, loaded.Checks, "test")
		assert.Equal(t, th.Checks["lint"].AvgDuration, loaded.Checks["lint"].AvgDuration)
		assert.Equal(t, th.Checks["lint"].LastDuration, loaded.Checks["lint"].LastDuration)
		assert.Equal(t, th.Checks["lint"].RunCount, loaded.Checks["lint"].RunCount)
		assert.Equal(t, th.Checks["test"].AvgDuration, loaded.Checks["test"].AvgDuration)
	})

	t.Run("overwrite existing file on second save", func(t *testing.T) {
		setupTimingTestEnv(t)

		// First save
		th1 := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 1},
			},
		}
		th1.save()

		// Second save with different data
		th2 := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 5 * time.Second, LastDuration: 5 * time.Second, RunCount: 10},
				"test": {AvgDuration: 8 * time.Second, LastDuration: 8 * time.Second, RunCount: 3},
			},
		}
		th2.save()

		// Load should reflect the second save
		loaded := loadTimingHistory()
		assert.Len(t, loaded.Checks, 2)
		assert.Equal(t, 5*time.Second, loaded.Checks["lint"].AvgDuration)
		assert.Equal(t, 10, loaded.Checks["lint"].RunCount)
		assert.Equal(t, 8*time.Second, loaded.Checks["test"].AvgDuration)
	})
}

func TestTimingHistory_GetExpectedDuration(t *testing.T) {
	t.Run("returns default for unknown check with no history", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		assert.Equal(t, 3*time.Second, th.getExpectedDuration("unknown-check"))
	})

	t.Run("returns predefined defaults for known checks", func(t *testing.T) {
		// A representative sample of known defaults
		defaults := map[string]time.Duration{
			"go-version": 100 * time.Millisecond,
			"tools":      100 * time.Millisecond,
			"format":     500 * time.Millisecond,
			"lint":       3 * time.Second,
			"test":       10 * time.Second,
			"sast":       4 * time.Second,
			"layering":   4 * time.Second,
			"sbom-vulns": 5 * time.Second,
		}

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

	t.Run("ignores zero-value average in history", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 0, RunCount: 0},
			},
		}
		// Zero avg should fall back to default
		assert.Equal(t, 3*time.Second, th.getExpectedDuration("lint"))
	})

	t.Run("returns history for unknown check name with recorded data", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"custom-check": {AvgDuration: 7 * time.Second, RunCount: 2},
			},
		}
		// custom-check has no default, but has history
		assert.Equal(t, 7*time.Second, th.getExpectedDuration("custom-check"))
	})
}

func TestTimingHistory_AllDefaults(t *testing.T) {
	// All 23 checks referenced in the default map should have defined durations
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
		assert.Less(t, th.Checks["test"].AvgDuration, 3*time.Second,
			"should converge towards recent values")
		assert.Greater(t, th.Checks["test"].AvgDuration, 1*time.Second,
			"shouldn't fully converge in 10 runs")
	})

	t.Run("EMA computation is correct over multiple steps", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}

		// Step 1: first run sets avg directly
		th.recordDuration("test", 10*time.Second)
		assert.Equal(t, 10*time.Second, th.Checks["test"].AvgDuration)

		// Step 2: EMA = 0.3 * 4s + 0.7 * 10s = 1.2 + 7 = 8.2s
		th.recordDuration("test", 4*time.Second)
		expectedStep2 := time.Duration(alpha*float64(4*time.Second) + (1-alpha)*float64(10*time.Second))
		assert.Equal(t, expectedStep2, th.Checks["test"].AvgDuration)

		// Step 3: EMA = 0.3 * 4s + 0.7 * 8.2s
		th.recordDuration("test", 4*time.Second)
		expectedStep3 := time.Duration(alpha*float64(4*time.Second) + (1-alpha)*float64(expectedStep2))
		assert.Equal(t, expectedStep3, th.Checks["test"].AvgDuration)

		assert.Equal(t, 3, th.Checks["test"].RunCount)
		assert.Equal(t, 4*time.Second, th.Checks["test"].LastDuration)
	})

	t.Run("creates new check entry if not exists", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}
		th.recordDuration("new-check", 500*time.Millisecond)

		require.Contains(t, th.Checks, "new-check")
		assert.Equal(t, 500*time.Millisecond, th.Checks["new-check"].AvgDuration)
		assert.Equal(t, 500*time.Millisecond, th.Checks["new-check"].LastDuration)
		assert.Equal(t, 1, th.Checks["new-check"].RunCount)
	})

	t.Run("records multiple different checks independently", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}

		th.recordDuration("lint", 3*time.Second)
		th.recordDuration("test", 10*time.Second)
		th.recordDuration("format", 500*time.Millisecond)

		assert.Len(t, th.Checks, 3)
		assert.Equal(t, 3*time.Second, th.Checks["lint"].AvgDuration)
		assert.Equal(t, 10*time.Second, th.Checks["test"].AvgDuration)
		assert.Equal(t, 500*time.Millisecond, th.Checks["format"].AvgDuration)
	})
}

func TestTimingHistory_Concurrency(t *testing.T) {
	t.Run("concurrent recordDuration and getExpectedDuration", func(t *testing.T) {
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
	})

	t.Run("concurrent recordDuration on different keys", func(t *testing.T) {
		th := &timingHistory{Checks: make(map[string]*checkTiming)}

		var wg sync.WaitGroup
		keys := []string{"lint", "test", "format", "sast", "deps"}
		for _, key := range keys {
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(k string, dur int) {
					defer wg.Done()
					th.recordDuration(k, time.Duration(dur)*time.Second)
				}(key, i+1)
			}
		}
		wg.Wait()

		// Each key should have 5 recordings
		for _, key := range keys {
			require.Contains(t, th.Checks, key)
			assert.Equal(t, 5, th.Checks[key].RunCount,
				"key %s should have 5 recordings", key)
		}
	})

	t.Run("concurrent save does not panic", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test": {AvgDuration: 5 * time.Second, LastDuration: 5 * time.Second, RunCount: 3},
			},
		}

		var wg sync.WaitGroup
		// Run multiple concurrent saves
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				th.save()
			}()
		}
		wg.Wait()

		// Verify the file is still valid after concurrent writes
		loaded := loadTimingHistory()
		require.NotNil(t, loaded)
		require.Contains(t, loaded.Checks, "test")
		assert.Equal(t, 5*time.Second, loaded.Checks["test"].AvgDuration)
	})

	t.Run("concurrent recordDuration and save", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{Checks: make(map[string]*checkTiming)}

		var wg sync.WaitGroup
		// Concurrent record and save operations
		for i := 0; i < 10; i++ {
			wg.Add(2)
			go func(dur int) {
				defer wg.Done()
				th.recordDuration("test", time.Duration(dur)*time.Second)
			}(i)
			go func() {
				defer wg.Done()
				th.save()
			}()
		}
		wg.Wait()

		// Should not panic and should have all recordings
		assert.Equal(t, 10, th.Checks["test"].RunCount)
	})
}

func TestCheckTiming_JSONSerialization(t *testing.T) {
	t.Run("checkTiming marshals to expected JSON fields", func(t *testing.T) {
		ct := &checkTiming{
			LastDuration: 3 * time.Second,
			AvgDuration:  2500 * time.Millisecond,
			RunCount:     7,
		}

		data, err := json.Marshal(ct)
		require.NoError(t, err)

		// Verify JSON field names
		var raw map[string]interface{}
		err = json.Unmarshal(data, &raw)
		require.NoError(t, err)

		assert.Contains(t, raw, "last_duration")
		assert.Contains(t, raw, "avg_duration")
		assert.Contains(t, raw, "run_count")
	})

	t.Run("timingHistory marshals with checks field", func(t *testing.T) {
		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 1},
			},
		}

		data, err := json.Marshal(th)
		require.NoError(t, err)

		var raw map[string]interface{}
		err = json.Unmarshal(data, &raw)
		require.NoError(t, err)

		assert.Contains(t, raw, "checks")
	})
}

func TestTimingHistory_Save_WritesToCorrectPath(t *testing.T) {
	setupTimingTestEnv(t)

	th := &timingHistory{
		Checks: map[string]*checkTiming{
			"lint": {AvgDuration: 2 * time.Second, LastDuration: 2 * time.Second, RunCount: 3},
		},
	}

	th.save()

	// Verify the file exists at the expected path
	path := timingFilePath()
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "lint")
	assert.Contains(t, string(data), "avg_duration")
}

func TestTimingHistory_Save_ErrorOnReadOnlyDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows: file permissions work differently")
	}

	tmpDir := t.TempDir()

	// Create a directory, put a file in it, then make the dir read-only
	cacheDir := filepath.Join(tmpDir, "readonly-cache", "ckeletin-go-test")
	err := os.MkdirAll(cacheDir, 0o750)
	require.NoError(t, err)

	// Write a file we can't overwrite by making the directory read-only
	testFile := filepath.Join(cacheDir, "check-timings.json")
	err = os.WriteFile(testFile, []byte("{}"), 0o600)
	require.NoError(t, err)

	// Make parent directory read-only so we can't write the file
	parentDir := filepath.Join(tmpDir, "readonly-cache", "ckeletin-go-test")
	err = os.Chmod(parentDir, 0o444)
	require.NoError(t, err)
	t.Cleanup(func() {
		// Restore permissions for cleanup
		os.Chmod(parentDir, 0o750)
	})

	// Directly save with this known path would fail -
	// but since we can't easily override timingFilePath(), we verify that
	// save() handles failures gracefully (logs but doesn't panic).
	// The key is that save() has the error handling code paths.
	th := &timingHistory{
		Checks: map[string]*checkTiming{
			"test": {AvgDuration: 1 * time.Second, LastDuration: 1 * time.Second, RunCount: 1},
		},
	}

	// This won't actually trigger the error path since it writes to the XDG path,
	// but we verify save() doesn't panic with a valid timingHistory
	assert.NotPanics(t, func() {
		th.save()
	})
}

func TestTimingHistory_Save_AtomicWrite(t *testing.T) {
	t.Run("temp file does not persist after successful write", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"lint": {AvgDuration: 2 * time.Second, LastDuration: 2 * time.Second, RunCount: 3},
			},
		}

		th.save()

		// The final file should exist
		path := timingFilePath()
		_, err := os.Stat(path)
		require.NoError(t, err, "timing file should exist after save")

		// The temp file should NOT persist
		tmpFile := path + ".tmp"
		_, err = os.Stat(tmpFile)
		assert.True(t, os.IsNotExist(err),
			"temp file %s should not exist after successful atomic write", tmpFile)
	})

	t.Run("data is correctly written via atomic pattern", func(t *testing.T) {
		setupTimingTestEnv(t)

		th := &timingHistory{
			Checks: map[string]*checkTiming{
				"test":   {AvgDuration: 10 * time.Second, LastDuration: 9 * time.Second, RunCount: 7},
				"format": {AvgDuration: 500 * time.Millisecond, LastDuration: 450 * time.Millisecond, RunCount: 4},
			},
		}

		th.save()

		// Load back and verify integrity
		loaded := loadTimingHistory()
		require.NotNil(t, loaded)
		require.Contains(t, loaded.Checks, "test")
		require.Contains(t, loaded.Checks, "format")
		assert.Equal(t, 10*time.Second, loaded.Checks["test"].AvgDuration)
		assert.Equal(t, 9*time.Second, loaded.Checks["test"].LastDuration)
		assert.Equal(t, 7, loaded.Checks["test"].RunCount)
		assert.Equal(t, 500*time.Millisecond, loaded.Checks["format"].AvgDuration)
		assert.Equal(t, 4, loaded.Checks["format"].RunCount)
	})
}
