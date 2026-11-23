package output

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/progress"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewReporter adapts the status and result writers to the business-logic
// port. Step, Status, and Warn go to the status stream. Done goes to the
// result stream. JSON mode drops both, so the envelope stays the only
// stdout document. A nil command uses the process streams.
func NewReporter(cmd *cobra.Command) progress.Reporter {
	return statusReporter{cmd: cmd}
}

type statusReporter struct {
	cmd *cobra.Command
}

func (s statusReporter) Step(message string)   { _ = Progressf(s.cmd, "%s", message) }
func (s statusReporter) Status(message string) { _ = Statusf(s.cmd, "%s", message) }
func (s statusReporter) Warn(message string)   { _ = Statusf(s.cmd, "warning: %s", message) }
func (s statusReporter) Done(message string)   { _ = Resultf(s.cmd, "%s", message) }

// ansiCSI matches a color or cursor sequence. Progress strips these when
// stderr is not a terminal.
var ansiCSI = regexp.MustCompile("\x1b\\[[0-9;]*[A-Za-z]")

// Result is the data stream (stdout). It honours SetOut.
// In JSON mode the write is a no-op that returns success: the envelope is the
// only stdout document, so a text result must not land beside it.
func Result(cmd *cobra.Command) io.Writer {
	if IsJSONMode() {
		return noopWriter{}
	}
	return &shadowWriter{out: stdOut(cmd), stream: "stdout", command: commandLabel(cmd)}
}

// Status is the status stream (stderr). It honours SetErr.
// JSON mode silences it. The console logger is already silent in that mode.
func Status(cmd *cobra.Command) io.Writer {
	if IsJSONMode() {
		return noopWriter{}
	}
	return &shadowWriter{out: stdErr(cmd), stream: "stderr", command: commandLabel(cmd)}
}

// Resultf writes a formatted result to stdout. JSON mode writes nothing.
func Resultf(cmd *cobra.Command, format string, args ...any) error {
	_, err := fmt.Fprintf(Result(cmd), format, args...)
	return err
}

// Statusf writes a formatted status line to stderr. JSON mode writes nothing.
func Statusf(cmd *cobra.Command, format string, args ...any) error {
	_, err := fmt.Fprintf(Status(cmd), format, args...)
	return err
}

// Success writes msg as a result line. JSON mode writes nothing.
func Success(msg string) error {
	if !strings.HasSuffix(msg, "\n") {
		msg += "\n"
	}
	_, err := io.WriteString(Result(currentCmd), msg)
	return err
}

// Progressf writes a status line. When stderr is not a terminal the line has
// no ANSI control codes, even if format included them.
func Progressf(cmd *cobra.Command, format string, args ...any) error {
	line := fmt.Sprintf(format, args...)
	if !isTerminal(stdErr(cmd)) {
		line = stripANSI(line)
	}
	if !strings.HasSuffix(line, "\n") {
		line += "\n"
	}
	_, err := io.WriteString(Status(cmd), line)
	return err
}

func stdOut(cmd *cobra.Command) io.Writer {
	if cmd == nil {
		return os.Stdout
	}
	return cmd.OutOrStdout()
}

func stdErr(cmd *cobra.Command) io.Writer {
	if cmd == nil {
		return os.Stderr
	}
	return cmd.ErrOrStderr()
}

func isTerminal(w io.Writer) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	return isatty.IsTerminal(file.Fd())
}

func stripANSI(s string) string {
	return ansiCSI.ReplaceAllString(s, "")
}

type noopWriter struct{}

func (noopWriter) Write(p []byte) (int, error) { return len(p), nil }

type shadowWriter struct {
	out     io.Writer
	stream  string
	command string
}

func (w *shadowWriter) Write(p []byte) (int, error) {
	if w.out == nil {
		return 0, io.ErrClosedPipe
	}
	n, err := w.out.Write(p)
	if err != nil || n != len(p) {
		return n, err
	}
	log.Debug().
		Str("command", w.command).
		Str("stream", w.stream).
		Str("output", string(p)).
		Msg("user output")
	return n, nil
}
