package output

import (
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// afterRun runs after error rendering and before Execute returns.
// Setup registers logger cleanup here. output used to document that
// logger.Init installed the failure log and logger.Cleanup cleared it.
// Neither did. Setup installs the failure log after the logger starts,
// and the cleanup hook clears it so a later failure cannot write to stderr.
// prepare runs at the start of Execute, before cobra parses arguments.
// The reference root hides optional flags here.
var prepare func()

// SetPrepare registers the function Execute calls before it parses arguments.
func SetPrepare(fn func()) {
	prepare = fn
}

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
	renderedErrorMessage = ""
	pinned := modePinned
	modePinned = false
	resolvedPhase = phaseUnset
	// A panic in RunE unwinds before ExecuteC returns, so the recovered
	// command is the one PersistentPreRunE stored. Clear the previous run
	// first. Null is only for a failure before any command resolves.
	currentCmd = nil
	var executed *cobra.Command
	defer func() {
		if rec := recover(); rec != nil {
			cmd := currentCmd
			if cmd == nil {
				cmd = root
			}
			// The user-facing line stays "panic: <value>" and the exit is 1.
			// The stack is only in the log file (debug, so the default console
			// level does not print it again).
			log.Debug().Str("stack", string(debug.Stack())).Interface("panic", rec).Msg("recovered panic")
			code = renderFailure(cmd, root, fmt.Errorf("panic: %v", rec))
		}
		afterRun()
	}()
	if root == nil {
		return 1
	}
	if prepare != nil {
		prepare()
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
		msg := renderedErrorMessage
		if msg == "" {
			msg = "command failed"
		}
		noteFailure(executed, errors.New(msg), 1)
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
		// The process still exits 0. The error must not disappear: the audit
		// record is the only place it goes.
		noteFailure(executed, err, 0)
		return 0
	}
	code := ExitCodeOf(err)
	if rendered && renderedStatus == "error" {
		if !StatusAgrees("error", code) {
			code = 1
		}
		noteFailure(executed, err, code)
		return code
	}
	if !StatusAgrees("error", code) {
		code = 1
	}
	if errors.Is(err, ErrRendered) {
		noteFailure(executed, err, code)
		return code
	}

	noteFailure(executed, err, code)
	name := commandLabel(executed)

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

func noteFailure(cmd *cobra.Command, err error, code int) {
	if failureLog == nil || err == nil {
		return
	}
	failureLog(err, commandLabel(cmd), code)
}

// absorbUnresolvedOutputFlag applies --output or the alias when cobra never
// ran PersistentPreRunE, which is every parse and usage error. With no flag,
// the early mode still comes from the environment. The probe defines only
// the output flag and the alias, so an alias that lives below the root cannot
// replace the real usage error with "unknown flag". An invalid mode is the
// one error this function may return.
func absorbUnresolvedOutputFlag(cmd *cobra.Command) error {
	if cmd == nil {
		return nil
	}
	root := cmd
	if cmd.Root() != nil {
		root = cmd.Root()
	}
	currentCmd = cmd
	if changedOutputFlag(cmd) {
		return UseOutputMode(PhaseEarly)
	}
	picked := pickOutputArgs(commandArgs(root))
	if value, ok := probeOutputValue(picked); ok {
		return applyProbedMode(value)
	}
	return UseOutputMode(PhaseEarly)
}

// probeOutputValue reads --output or the alias from args. A parse problem
// returns ok false and must not replace the original command error.
func probeOutputValue(args []string) (string, bool) {
	if len(args) == 0 {
		return "", false
	}
	fs := pflag.NewFlagSet("output-probe", pflag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.ParseErrorsWhitelist.UnknownFlags = true
	fs.String(outputFlagName, "", "")
	if settings.alias != "" {
		fs.String(settings.alias, "", "")
	}
	if err := fs.Parse(args); err != nil {
		return "", false
	}
	if fs.Changed(outputFlagName) {
		value, err := fs.GetString(outputFlagName)
		return value, err == nil
	}
	if settings.alias != "" && fs.Changed(settings.alias) {
		value, err := fs.GetString(settings.alias)
		return value, err == nil
	}
	return "", false
}

func applyProbedMode(value string) error {
	mode, err := validMode(value)
	if err != nil {
		return err
	}
	currentOutputMode = mode
	modePinned = false
	if PhaseEarly > resolvedPhase {
		resolvedPhase = PhaseEarly
	}
	viper.Set(config.KeyAppOutputFormat, mode)
	return nil
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
	// A missing field means this cobra changed shape. Do not guess os.Args:
	// tests pass arguments with SetArgs, and os.Args is the test binary.
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return nil
	}
	if v.Len() > 0 {
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
