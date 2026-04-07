// .ckeletin/pkg/output/json.go
//
// Framework-level JSON output types and mode state.
// This file is managed by ckeletin-go and auto-updates via `task ckeletin:update`.

package output

import (
	"encoding/json"
	"io"
)

// JSONEnvelope is the standard response wrapper for --output json mode.
// Every command emits exactly one envelope to stdout.
type JSONEnvelope struct {
	Status  string      `json:"status"`  // "success" or "error"
	Command string      `json:"command"` // cobra command name (e.g., "ping")
	Data    interface{} `json:"data"`    // command-specific payload; nil on error
	Error   *JSONError  `json:"error"`   // nil on success
}

// JSONError represents a structured error in the JSON envelope.
type JSONError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"` // optional error classification
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

// SetCommandName stores the active command name for the JSON envelope.
// Called from PersistentPreRunE.
func SetCommandName(name string) {
	currentCommandName = name
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
