package check

import (
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
