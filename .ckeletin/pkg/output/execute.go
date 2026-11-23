package output

import (
	"errors"
	"fmt"
	"io"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// afterRun runs after error rendering and before Execute returns.
// The reference CLI registers logger.Cleanup so the audit file stays open
// through the error envelope. output cannot import logger: logger imports output.
var afterRun = func() {}

// failureLog records a command failure on the audit stream. logger.Init
// installs it once the file writer exists. While it is nil, JSON mode must
// not log: the default zerolog logger writes to stderr, which JSON mode keeps
// silent. Text mode falls back to log.Debug on that default logger.
var failureLog func(err error, command string, code int)

// SetFailureLog registers the audit record for a command failure.
// A nil function clears it. logger.Cleanup clears it so a later init failure
// cannot write the event to stderr.
func SetFailureLog(fn func(err error, command string, code int)) {
	failureLog = fn
}

// SetAfterRun registers the function Execute calls after it has rendered an
// error and before it returns. Pass logger.Cleanup. A nil function clears it.
func SetAfterRun(fn func()) {
	if fn == nil {
		afterRun = func() {}
		return
	}
	afterRun = fn
}

// Execute runs root and returns the process exit code.
//
// JSON mode writes exactly one envelope to stdout for an error the command did
// not already render. Text mode writes "Error: …" to stderr and leaves stdout
// empty. ErrRendered and Rendered(code) suppress the second document.
// A nil error exits 0. Any error exits non-zero. Code 0 on an error becomes 1.
//
// The audit-log hook runs after that write, so a failure is logged before the
// file is closed.
func Execute(root *cobra.Command) int {
	rendered = false
	defer afterRun()
	if root == nil {
		return 1
	}
	root.SilenceErrors = true
	root.SilenceUsage = true

	executed, err := root.ExecuteC()
	if err == nil {
		writeEmptySuccess(executed)
		return 0
	}
	if !commandNameSet {
		SetCommandFrom(executed)
	}
	code := ExitCodeOf(err)
	if !StatusAgrees("error", code) {
		code = 1
	}
	if errors.Is(err, ErrRendered) {
		return code
	}

	name := commandLabel(executed)
	if failureLog != nil {
		failureLog(err, name, code)
	} else if !IsJSONMode() {
		log.Debug().
			Err(err).
			Str("command", name).
			Int("exit_code", code).
			Msg("command failed")
	}

	if IsJSONMode() {
		env := JSONEnvelope{
			Status:  "error",
			Command: name,
			Data:    nil,
			Error:   &JSONError{Message: err.Error()},
		}
		if name == "" {
			env = env.WithUnresolvedCommand()
		}
		_ = RenderJSON(writerOr(executed, root, true), env)
	} else {
		fmt.Fprintf(writerOr(executed, root, false), "Error: %v\n", err)
	}
	return code
}

// writeEmptySuccess prints one success envelope when JSON mode produced no
// document. An OUT-002 exemption keeps its native output. Text mode is unchanged.
func writeEmptySuccess(executed *cobra.Command) {
	if !IsJSONMode() || rendered {
		return
	}
	if _, ok := exemptionReason(executed); ok {
		return
	}
	if !commandNameSet {
		SetCommandFrom(executed)
	}
	name := commandLabel(executed)
	_ = RenderJSON(writerOr(executed, nil, true), JSONEnvelope{
		Status:  "success",
		Command: name,
		Data:    nil,
		Error:   nil,
	})
}

func writerOr(executed, root *cobra.Command, stdout bool) io.Writer {
	cmd := executed
	if cmd == nil {
		cmd = root
	}
	if cmd == nil {
		return io.Discard
	}
	if stdout {
		return cmd.OutOrStdout()
	}
	return cmd.ErrOrStderr()
}
