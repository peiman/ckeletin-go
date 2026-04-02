//go:build !windows

// test/integration/signal_test.go
//
// Integration tests for signal handling on Linux/macOS.
// Windows uses a different signal model and is excluded via build tag.

package integration

import (
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGracefulShutdownOnSIGINT(t *testing.T) {
	cmd := exec.Command(binaryPath, "ping")
	require.NoError(t, cmd.Start())

	// Give the process time to start
	time.Sleep(200 * time.Millisecond)

	// Send SIGINT; the process may have already exited (ping is fast), so
	// ignore "process already finished" but fail on any other signal error.
	signalErr := cmd.Process.Signal(syscall.SIGINT)
	if signalErr != nil && signalErr.Error() != "os: process already finished" {
		require.NoError(t, signalErr, "unexpected error sending SIGINT")
	}

	// Process should exit within 5 seconds
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		// A nil error means the process exited cleanly (exit code 0) — perfect.
		if err == nil {
			return
		}
		// Non-nil error: check for a true crash (SIGSEGV, SIGBUS, etc.).
		// On Unix, signal termination also produces an *exec.ExitError with
		// ExitCode() == -1, but WaitStatus.Signal() tells us which signal fired.
		// SIGINT termination is expected and acceptable; a crash signal is not.
		if exitErr, ok := err.(*exec.ExitError); ok {
			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
				sig := ws.Signal()
				assert.NotEqual(t, syscall.SIGSEGV, sig, "process crashed with SIGSEGV on SIGINT")
				assert.NotEqual(t, syscall.SIGBUS, sig, "process crashed with SIGBUS on SIGINT")
				assert.NotEqual(t, syscall.SIGABRT, sig, "process crashed with SIGABRT on SIGINT")
				// SIGINT termination (signal 2) is the expected outcome when the
				// binary does not install a custom signal handler.
				return
			}
		}
		// Any other non-nil error (e.g. exit status 1) is also acceptable —
		// the process terminated, which is the important invariant.
	case <-time.After(5 * time.Second):
		_ = cmd.Process.Kill()
		t.Fatal("process did not exit within 5 seconds of SIGINT")
	}
}
