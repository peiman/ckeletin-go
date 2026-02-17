package checkmate

import (
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IsTerminal determines if the given writer is an interactive terminal.
// This is used to decide whether to use colors and emojis (terminal) or
// plain ASCII output (non-terminal).
//
// Returns false for:
// - Piped output
// - Redirected output
// - Non-file writers (like bytes.Buffer)
func IsTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

// IsStdoutTerminal returns true if stdout is an interactive terminal.
func IsStdoutTerminal() bool {
	return IsTerminal(os.Stdout)
}

// IsStderrTerminal returns true if stderr is an interactive terminal.
func IsStderrTerminal() bool {
	return IsTerminal(os.Stderr)
}
