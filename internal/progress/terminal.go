package progress

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IsInteractive determines if the given writer is an interactive terminal.
// This is used to decide whether to use Bubble Tea (interactive) or
// simple console output (non-interactive).
//
// Returns false for:
// - Piped output
// - Redirected output
// - Non-file writers (like bytes.Buffer)
//
// Per ADR-012, this enables automatic switching between:
// - Interactive TTY: Bubble Tea with animations
// - Non-interactive (CI/piped): Simple line-based output
func IsInteractive(w io.Writer) bool {
	// Check if the writer is a file
	f, ok := w.(*os.File)
	if !ok {
		return false
	}

	// Check if it's a terminal
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// IsStderrInteractive checks if stderr is an interactive terminal.
// This is a convenience function since progress typically outputs to stderr.
func IsStderrInteractive() bool {
	return IsInteractive(os.Stderr)
}

// IsStdoutInteractive checks if stdout is an interactive terminal.
func IsStdoutInteractive() bool {
	return IsInteractive(os.Stdout)
}
