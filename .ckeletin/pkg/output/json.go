// .ckeletin/pkg/output/json.go
//
// Framework-level JSON output types and mode state.
// This file is managed by ckeletin-go and auto-updates via `task ckeletin:update`.

package output

import (
	"encoding/json"
	"errors"
	"io"
)

// ErrRendered is a sentinel returned by a command that has already written its
// single JSON envelope to stdout but still needs to signal a non-zero exit.
// Execute recognizes it and exits non-zero WITHOUT rendering a second envelope.
// This keeps the "exactly one envelope per command" contract while preserving a
// meaningful process exit code for JSON consumers that gate on $?.
//
// Rendered(code) wraps this sentinel with the exit code the command chose.
// errors.Is(Rendered(code), ErrRendered) stays true. A plain ErrRendered exits 1.
var ErrRendered = errors.New("output: result already rendered")

// ExitInvalidOutput is the exit code for an output mode that is not text or json.
// Every command uses this code. The default for any other error is 1.
const ExitInvalidOutput = 2

// ExitCoder is implemented by errors that choose their own process exit code.
// Execute honours it for rendered and unrendered errors. Code 0 is rejected:
// a failure cannot exit 0, and a success cannot exit non-zero.
type ExitCoder interface {
	ExitCode() int
}

// renderedError carries the exit code for an envelope the command already wrote.
type renderedError struct {
	code int
}

func (e *renderedError) Error() string { return ErrRendered.Error() }

func (e *renderedError) Unwrap() error { return ErrRendered }

func (e *renderedError) ExitCode() int { return e.code }

// Rendered reports that the command already wrote its one envelope and that
// the process should exit with code. Code 0 is stored as 1 so an error status
// is never paired with a successful exit.
func Rendered(code int) error {
	return &renderedError{code: normalizeExitCode(code)}
}

// ExitCodeOf returns 0 for a nil error and the error's ExitCode otherwise.
// A zero code on a non-nil error becomes 1.
func ExitCodeOf(err error) int {
	if err == nil {
		return 0
	}
	var coder ExitCoder
	if errors.As(err, &coder) {
		return normalizeExitCode(coder.ExitCode())
	}
	return 1
}

func normalizeExitCode(code int) int {
	if code == 0 {
		return 1
	}
	return code
}

// StatusAgrees reports whether a machine-readable status matches the exit code.
// status "error" requires a non-zero code. status "success" requires 0.
func StatusAgrees(status string, code int) bool {
	switch status {
	case "error":
		return code != 0
	case "success":
		return code == 0
	default:
		return false
	}
}

// JSONEnvelope is the standard response wrapper for --output json mode.
// Every command emits exactly one envelope to stdout.
type JSONEnvelope struct {
	Status  string `json:"status"`  // "success" or "error"
	Command string `json:"command"` // full command path without the program name ("check model", "version")
	// unresolvedCommand marshals Command as JSON null. Used when the command
	// was never resolved (the parser boundary). An empty Command otherwise
	// stays an empty string.
	unresolvedCommand bool        `json:"-"`
	Data              interface{} `json:"data"`  // command-specific payload; nil on error
	Error             *JSONError  `json:"error"` // nil on success
}

// JSONError represents a structured error in the JSON envelope.
// Per CKSPEC-OUT-003 the error object is a TOTAL schema {code, message}: both
// keys are always present. Code is a *string so an absent classification
// serializes as JSON null (present-and-null), never omitted — `omitempty` would
// drop the key and break the fixed-key-set contract.
type JSONError struct {
	Message string  `json:"message"`
	Code    *string `json:"code"` // error classification; nil → JSON null
}

// JSONResponder can be implemented by data types passed to RenderSuccess
// to customize their JSON representation. If the data implements this interface,
// JSONResponse() is used instead of marshaling the data directly.
type JSONResponder interface {
	JSONResponse() interface{}
}

// Package-level state for the current output mode and command name.
// Safe because CLI commands execute sequentially (Cobra's execution model).
var (
	currentOutputMode  string
	currentCommandName string
	// commandNameSet is true when SetCommandName was called with a non-empty
	// override. SetCommandFrom leaves that override in place.
	commandNameSet bool
)

// SetOutputMode sets the current output mode ("text" or "json").
// Called from PersistentPreRunE after config initialization.
func SetOutputMode(mode string) {
	currentOutputMode = mode
}

// OutputMode returns the current output mode. Defaults to "text".
func OutputMode() string {
	if currentOutputMode == "" {
		return "text"
	}
	return currentOutputMode
}

// IsJSONMode returns true when --output json is active.
func IsJSONMode() bool {
	return OutputMode() == "json"
}

// SetCommandName overrides the command id in the JSON envelope.
// An empty name clears the override so the next SetCommandFrom uses the path.
// The framework sets the path itself; call this only to replace it.
func SetCommandName(name string) {
	currentCommandName = name
	commandNameSet = name != ""
}

// CommandName returns the stored command name.
func CommandName() string {
	return currentCommandName
}

// ResolveJSONData checks if data implements JSONResponder and returns
// the appropriate value for the envelope's Data field.
func ResolveJSONData(data interface{}) interface{} {
	if responder, ok := data.(JSONResponder); ok {
		return responder.JSONResponse()
	}
	return data
}

// WithUnresolvedCommand marks the envelope so Command marshals as JSON null.
// Pre-resolution failures use this when not even a parent command resolved.
func (e JSONEnvelope) WithUnresolvedCommand() JSONEnvelope {
	e.unresolvedCommand = true
	e.Command = ""
	return e
}

// MarshalJSON writes command as null only when the failure is unresolved.
func (e JSONEnvelope) MarshalJSON() ([]byte, error) {
	var command any
	if e.unresolvedCommand {
		command = nil
	} else {
		command = e.Command
	}
	payload := struct {
		Status  string      `json:"status"`
		Command any         `json:"command"`
		Data    interface{} `json:"data"`
		Error   *JSONError  `json:"error"`
	}{
		Status:  e.Status,
		Command: command,
		Data:    e.Data,
		Error:   e.Error,
	}
	return json.Marshal(payload)
}

// RenderJSON marshals a JSONEnvelope to the writer as indented JSON.
func RenderJSON(out io.Writer, envelope JSONEnvelope) error {
	encoded, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return err
	}
	encoded = append(encoded, '\n')
	_, err = out.Write(encoded)
	return err
}
