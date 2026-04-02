package check

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/peiman/ckeletin-go/internal/xdg"
	"github.com/rs/zerolog/log"
)

// checkTiming stores historical timing data for a check
type checkTiming struct {
	LastDuration time.Duration `json:"last_duration"`
	AvgDuration  time.Duration `json:"avg_duration"`
	RunCount     int           `json:"run_count"`
}

// timingHistory stores timing data for all checks
type timingHistory struct {
	Checks map[string]*checkTiming `json:"checks"`
	mu     sync.RWMutex
}

// timingFilePath returns the path to the timing history file.
// Uses XDG cache directory since timing data is ephemeral/regenerable.
func timingFilePath() string {
	path, err := xdg.CacheFile("check-timings.json")
	if err != nil {
		// Fallback to temp dir if XDG not configured
		return filepath.Join(os.TempDir(), "ckeletin-go-check-timings.json")
	}
	return path
}

// loadTimingHistory loads timing data from disk
func loadTimingHistory() *timingHistory {
	th := &timingHistory{Checks: make(map[string]*checkTiming)}

	data, err := os.ReadFile(timingFilePath())
	if err != nil {
		return th // Return empty history if file doesn't exist
	}

	if err := json.Unmarshal(data, th); err != nil {
		log.Debug().Err(err).Msg("Failed to parse timing history, using empty history")
	}
	if th.Checks == nil {
		th.Checks = make(map[string]*checkTiming)
	}
	return th
}

// save persists timing data to disk
func (th *timingHistory) save() {
	th.mu.RLock()
	defer th.mu.RUnlock()

	data, err := json.MarshalIndent(th, "", "  ")
	if err != nil {
		log.Debug().Err(err).Msg("Failed to marshal timing history")
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(timingFilePath())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		log.Debug().Err(err).Str("dir", dir).Msg("Failed to create timing history directory")
		return
	}

	// Atomic write: write to temp file in same directory, then rename.
	// os.Rename is atomic on most filesystems, so readers will either see
	// the old complete file or the new complete file, never a partial write.
	// Use a unique temp name to avoid Windows file locking on concurrent writes.
	path := timingFilePath()
	tmpFile := fmt.Sprintf("%s.tmp.%d", path, time.Now().UnixNano())
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		log.Debug().Err(err).Msg("Failed to write temp timing file")
		return
	}
	if err := os.Rename(tmpFile, path); err != nil {
		if removeErr := os.Remove(tmpFile); removeErr != nil {
			log.Debug().Err(removeErr).Str("file", tmpFile).Msg("Failed to clean up temp timing file")
		}
		log.Debug().Err(err).Msg("Failed to rename timing file")
	}
}

// getExpectedDuration returns the expected duration for a check
func (th *timingHistory) getExpectedDuration(name string) time.Duration {
	th.mu.RLock()
	defer th.mu.RUnlock()

	if t, ok := th.Checks[name]; ok && t.AvgDuration > 0 {
		return t.AvgDuration
	}
	// Default estimates for first run
	defaults := map[string]time.Duration{
		// Environment
		"go-version": 100 * time.Millisecond,
		"tools":      100 * time.Millisecond,
		// Quality
		"format": 500 * time.Millisecond,
		"lint":   3 * time.Second,
		// Architecture
		"defaults":           100 * time.Millisecond,
		"commands":           200 * time.Millisecond,
		"constants":          500 * time.Millisecond,
		"task-naming":        200 * time.Millisecond,
		"architecture":       500 * time.Millisecond,
		"layering":           4 * time.Second,
		"package-org":        500 * time.Millisecond,
		"config-consumption": 100 * time.Millisecond,
		"output-patterns":    100 * time.Millisecond,
		"security-patterns":  100 * time.Millisecond,
		// Security
		"secrets": 200 * time.Millisecond,
		"sast":    4 * time.Second,
		// Dependencies
		"deps":           1 * time.Second,
		"vuln":           2 * time.Second,
		"outdated":       2 * time.Second,
		"license-source": 1 * time.Second,
		"license-binary": 1 * time.Second,
		"sbom-vulns":     5 * time.Second,
		// Tests
		"test": 10 * time.Second,
	}
	if d, ok := defaults[name]; ok {
		return d
	}
	return 3 * time.Second // Generic default for unknown checks
}

// recordDuration updates timing data after a check completes
func (th *timingHistory) recordDuration(name string, duration time.Duration) {
	th.mu.Lock()
	defer th.mu.Unlock()

	t, ok := th.Checks[name]
	if !ok {
		t = &checkTiming{}
		th.Checks[name] = t
	}

	t.LastDuration = duration
	t.RunCount++

	// Update rolling average (exponential moving average with alpha=0.3)
	// This gives more weight to recent runs while considering history
	if t.AvgDuration == 0 {
		t.AvgDuration = duration
	} else {
		alpha := 0.3
		t.AvgDuration = time.Duration(alpha*float64(duration) + (1-alpha)*float64(t.AvgDuration))
	}
}
