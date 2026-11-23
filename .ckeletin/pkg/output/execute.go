package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
)

// afterRun runs after error rendering and before Execute returns.
// Setup registers logger cleanup here. output used to document that
// logger.Init installed the failure log and logger.Cleanup cleared it.
// Neither did. Setup installs the failure log after the logger starts,
// and the cleanup hook clears it so a later failure cannot write to stderr.
var afterRun = func() {}

// failureLog records a command failure on the audit stream.
// Setup installs it after logger.Init. While it is nil, Execute does not
// log. Usage errors happen before any logger exists, and the default
// zerolog logger would write a raw JSON line to stderr.
var failureLog func(err error, command string, code int)

// SetFailureLog registers the audit record for a command failure.
// A nil function clears it. Setup clears it after logger.Cleanup.
func SetFailureLog(fn func(err error, command string, code int)) {
	failureLog = fn
}

// SetAfterRun registers the function Execute calls after it has rendered an
// error and before it returns. Setup passes logger cleanup. A nil function
// clears it.
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
// A nil error exits 0, unless the command already rendered status "error":
// that exits 1. A rendered status "success" exits 0 even when the command
// returned Rendered(code). The exit is corrected to match the envelope.
// Any other error exits non-zero. Code 0 on an error becomes 1.
//
// A panic in the command becomes that same error document and a non-zero
// exit. The audit-log hook runs after the write, so a failure is logged
// before the file is closed. Nothing is logged when no failure hook is
// installed.
//
// A parse or usage error never reaches PersistentPreRunE. When the mode was
// not resolved and not preset, Execute reads --output (and the alias) from
// the arguments and can still print one JSON envelope. An unresolved command
// marshals command as JSON null.
func Execute(root *cobra.Command) (code int) {
	rendered = false
	renderedStatus = ""
	pinned := modePinned
	modePinned = false
	resolvedPhase = phaseUnset
	var executed *cobra.Command
	defer func() {
		if rec := recover(); rec != nil {
			// executed is still nil: the panic unwound before ExecuteC returned.
			// Write to root. A stale currentCmd belongs to an earlier run.
			code = renderFailure(root, root, fmt.Errorf("panic: %v", rec))
		}
		afterRun()
	}()
	if root == nil {
		return 1
	}
	root.SilenceErrors = true
	root.SilenceUsage = true

	var err error
	executed, err = root.ExecuteC()
	if err == nil {
		return successExit(executed)
	}
	if !pinned && resolvedPhase == phaseUnset {
		if modeErr := absorbUnresolvedOutputFlag(executed); modeErr != nil {
			err = modeErr
		}
	}
	return renderFailure(executed, root, err)
}

// successExit prints the empty success envelope when needed, then returns
// the exit code that agrees with a document the command already wrote.
func successExit(executed *cobra.Command) int {
	writeEmptySuccess(executed)
	if rendered && renderedStatus == "error" {
		return 1
	}
	return 0
}

func renderFailure(executed, root *cobra.Command, err error) int {
	if err == nil {
		return 1
	}
	if !commandNameSet {
		SetCommandFrom(executed)
	}
	if rendered && renderedStatus == "success" {
		return 0
	}
	code := ExitCodeOf(err)
	if rendered && renderedStatus == "error" {
		if !StatusAgrees("error", code) {
			return 1
		}
		return code
	}
	if !StatusAgrees("error", code) {
		code = 1
	}
	if errors.Is(err, ErrRendered) {
		return code
	}

	name := commandLabel(executed)
	if failureLog != nil {
		failureLog(err, name, code)
	}

	if IsJSONMode() {
		env := JSONEnvelope{
			Status:  "error",
			Command: name,
			Data:    nil,
			Error:   &JSONError{Message: err.Error(), Code: errorClass(err)},
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

// absorbUnresolvedOutputFlag applies --output or the alias when cobra never
// ran PersistentPreRunE, which is every parse and usage error. It returns
// the mode error when the flag value is not text or json.
func absorbUnresolvedOutputFlag(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}
	root := cmd
	if cmd.Root() != nil {
		root = cmd.Root()
	}
	currentCmd = cmd
	if !changedOutputFlag(cmd) {
		picked := pickOutputArgs(commandArgs(root))
		if len(picked) == 0 {
			return nil
		}
		if root.Flags().Parsed() {
			return nil
		}
		if err := root.ParseFlags(picked); err != nil {
			return err
		}
	}
	if !changedOutputFlag(cmd) {
		return nil
	}
	return UseOutputMode(PhaseEarly)
}

func changedOutputFlag(cmd *cobra.Command) bool {
	if _, ok := explicitFlag(cmd, outputFlagName); ok {
		return true
	}
	if settings.alias == "" {
		return false
	}
	_, ok := explicitFlag(cmd, settings.alias)
	return ok
}

func pickOutputArgs(args []string) []string {
	names := []string{"--" + outputFlagName}
	if settings.alias != "" {
		names = append(names, "--"+settings.alias)
	}
	var picked []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			break
		}
		for _, name := range names {
			if arg == name && i+1 < len(args) {
				picked = append(picked, arg, args[i+1])
				i++
				break
			}
			prefix := name + "="
			if strings.HasPrefix(arg, prefix) {
				picked = append(picked, arg)
				break
			}
		}
	}
	return picked
}

// commandArgs reads the arguments cobra stored. The field is unexported.
// A nil store means the real process arguments, except under go test.
func commandArgs(cmd *cobra.Command) []string {
	if cmd == nil {
		return nil
	}
	v := reflect.ValueOf(cmd).Elem().FieldByName("args")
	if v.IsValid() && v.Kind() == reflect.Slice && v.Len() > 0 {
		out := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			out[i] = v.Index(i).String()
		}
		return out
	}
	if len(os.Args) > 1 && !strings.HasPrefix(os.Args[1], "-test.") && !strings.HasPrefix(os.Args[1], "--test.") {
		return os.Args[1:]
	}
	return nil
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
