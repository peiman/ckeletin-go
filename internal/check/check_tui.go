package check

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
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

// TUIExecutor runs checks with a Bubble Tea progress UI.
type TUIExecutor struct {
	cfg     Config
	writer  io.Writer
	checks  []checkItem
	program *tea.Program   // Reference to the Bubble Tea program for sending messages
	timings *timingHistory // Historical timing data for progress estimation
}

type checkItem struct {
	name        string
	fn          func(ctx context.Context) error
	remediation string
}

// timingFilePath returns the path to the timing history file
func timingFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = os.TempDir()
	}
	// Use Clean to sanitize the path and prevent traversal
	return filepath.Clean(filepath.Join(configDir, "ckeletin-go", "check-timings.json"))
}

// loadTimingHistory loads timing data from disk
func loadTimingHistory() *timingHistory {
	th := &timingHistory{Checks: make(map[string]*checkTiming)}

	data, err := os.ReadFile(timingFilePath())
	if err != nil {
		return th // Return empty history if file doesn't exist
	}

	// Ignore JSON errors, just use empty history
	_ = json.Unmarshal(data, th)
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
		return
	}

	// Ensure directory exists
	dir := filepath.Dir(timingFilePath())
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return
	}

	// Write atomically with secure permissions
	_ = os.WriteFile(timingFilePath(), data, 0o600)
}

// getExpectedDuration returns the expected duration for a check
func (th *timingHistory) getExpectedDuration(name string) time.Duration {
	th.mu.RLock()
	defer th.mu.RUnlock()

	if t, ok := th.Checks[name]; ok && t.AvgDuration > 0 {
		return t.AvgDuration
	}
	// Default estimates for first run (in seconds)
	defaults := map[string]time.Duration{
		"format": 2 * time.Second,
		"lint":   5 * time.Second,
		"test":   10 * time.Second,
		"deps":   1 * time.Second,
		"vuln":   3 * time.Second,
	}
	if d, ok := defaults[name]; ok {
		return d
	}
	return 5 * time.Second // Generic default
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

// NewTUIExecutor creates a new TUI-based executor.
func NewTUIExecutor(cfg Config, writer io.Writer) *TUIExecutor {
	e := &TUIExecutor{
		cfg:     cfg,
		writer:  writer,
		timings: loadTimingHistory(),
	}

	// Register checks
	e.checks = []checkItem{
		{name: "format", fn: e.checkFormat, remediation: "Run: task format"},
		{name: "lint", fn: e.checkLint, remediation: "Run: task lint"},
		{name: "test", fn: e.checkTest, remediation: "Fix failing tests"},
		{name: "deps", fn: e.checkDeps, remediation: "Run: go mod tidy"},
		{name: "vuln", fn: e.checkVuln, remediation: "Update vulnerable dependencies"},
	}

	return e
}

// Execute runs all checks with the TUI progress display.
func (e *TUIExecutor) Execute(ctx context.Context) error {
	// Get check names
	names := make([]string, len(e.checks))
	for i, c := range e.checks {
		names[i] = c.name
	}

	// Create the progress model
	model := checkmate.NewProgressModel("Code Quality", names)

	// Create the program with options for non-TTY compatibility
	p := tea.NewProgram(model,
		tea.WithOutput(e.writer),
		tea.WithInput(nil), // Don't try to open /dev/tty for input
	)
	e.program = p // Store reference for coverage callback

	// Run checks in a goroutine
	go func() {
		for i, check := range e.checks {
			// Start progress animation with expected duration from history
			done := make(chan struct{})
			go e.animateProgress(p, i, check.name, done)

			// Run the check
			start := time.Now()
			err := check.fn(ctx)
			duration := time.Since(start)

			// Stop progress animation
			close(done)

			// Record timing for future runs
			e.timings.recordDuration(check.name, duration)

			if err != nil {
				p.Send(checkmate.CheckUpdateMsg{
					Index:    i,
					Status:   checkmate.CheckFailed,
					Progress: 1.0,
					Duration: duration,
					Error:    err,
				})

				if e.cfg.FailFast {
					break
				}
			} else {
				p.Send(checkmate.CheckUpdateMsg{
					Index:    i,
					Status:   checkmate.CheckPassed,
					Progress: 1.0,
					Duration: duration,
				})
			}

			// Small delay to see the progress
			time.Sleep(100 * time.Millisecond)
		}

		// Save timing history for next run
		e.timings.save()

		// Signal done
		time.Sleep(200 * time.Millisecond)
		p.Send(checkmate.DoneMsg{})
	}()

	// Run the TUI
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

// Check functions (reuse from checks.go via embedding or copy)
func (e *TUIExecutor) checkFormat(ctx context.Context) error {
	executor := &Executor{cfg: e.cfg}
	return executor.checkFormat(ctx)
}

func (e *TUIExecutor) checkLint(ctx context.Context) error {
	executor := &Executor{cfg: e.cfg}
	return executor.checkLint(ctx)
}

func (e *TUIExecutor) checkTest(ctx context.Context) error {
	executor := &Executor{
		cfg: e.cfg,
		onCoverage: func(coverage float64) {
			if e.program != nil {
				e.program.Send(checkmate.CoverageMsg{Coverage: coverage})
			}
		},
	}
	return executor.checkTest(ctx)
}

// animateProgress sends progress updates based on historical timing data
func (e *TUIExecutor) animateProgress(p *tea.Program, idx int, checkName string, done <-chan struct{}) {
	expectedDuration := e.timings.getExpectedDuration(checkName)
	startTime := time.Now()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			// Calculate progress as percentage of expected duration
			// Cap at 95% to leave room for completion
			progress := float64(elapsed) / float64(expectedDuration)
			if progress > 0.95 {
				progress = 0.95
			}
			p.Send(checkmate.CheckUpdateMsg{
				Index:    idx,
				Status:   checkmate.CheckRunning,
				Progress: progress,
			})
		}
	}
}

func (e *TUIExecutor) checkDeps(ctx context.Context) error {
	executor := &Executor{cfg: e.cfg}
	return executor.checkDeps(ctx)
}

func (e *TUIExecutor) checkVuln(ctx context.Context) error {
	executor := &Executor{cfg: e.cfg}
	return executor.checkVuln(ctx)
}
