package logger

import (
	"io"
	"sync"
)

// consoleHold keeps formatted console lines while an interactive display
// owns the status stream. The log file is a different writer, so it still
// receives each line when it is logged. JSON mode disables the console
// before this writer, so a held display adds nothing to stderr.
type consoleHold struct {
	mu      sync.Mutex
	held    bool
	out     io.Writer
	pending [][]byte
}

var consoleHoldGate consoleHold

func (h *consoleHold) bind(out io.Writer) {
	h.mu.Lock()
	h.out = out
	h.held = false
	h.pending = nil
	h.mu.Unlock()
}

func (h *consoleHold) Write(p []byte) (int, error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.held {
		h.pending = append(h.pending, append([]byte(nil), p...))
		return len(p), nil
	}
	if h.out == nil {
		return len(p), nil
	}
	return h.out.Write(p)
}

// HoldConsole stops console lines from reaching the status stream. Lines
// logged while it is held are kept in order. The log file is unchanged.
func HoldConsole() {
	consoleHoldGate.mu.Lock()
	consoleHoldGate.held = true
	consoleHoldGate.mu.Unlock()
}

// ReleaseConsole writes the held console lines, whole and in order, and
// lets later lines through.
func ReleaseConsole() {
	consoleHoldGate.mu.Lock()
	lines := consoleHoldGate.pending
	consoleHoldGate.pending = nil
	consoleHoldGate.held = false
	out := consoleHoldGate.out
	consoleHoldGate.mu.Unlock()
	if out == nil {
		return
	}
	for _, line := range lines {
		_, _ = out.Write(line)
	}
}
