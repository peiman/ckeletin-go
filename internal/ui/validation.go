// internal/ui/validation.go
//
// JSON-mode rendering for config-validation results (ADR-001: presentation
// logic lives outside cmd/).

package ui

import (
	"fmt"
	"io"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/config/validator"
	"github.com/peiman/ckeletin-go/.ckeletin/pkg/output"
)

// RenderValidationJSON emits the single JSON envelope for a config-validation
// result. exitErr (from validator.ExitCodeForResult) determines the envelope
// status and error message; the caller keeps ownership of the process exit
// signal (output.ErrRendered). A non-nil return means the envelope could not
// be written, not that validation failed.
func RenderValidationJSON(out io.Writer, result *validator.Result, exitErr error) error {
	status := "success"
	var jsonErr *output.JSONError
	if exitErr != nil {
		status = "error"
		jsonErr = &output.JSONError{Message: exitErr.Error()}
	}
	errMsgs := make([]string, len(result.Errors))
	for i, e := range result.Errors {
		errMsgs[i] = e.Error()
	}
	if err := output.RenderJSON(out, output.JSONEnvelope{
		Status:  status,
		Command: output.CommandName(),
		Data: map[string]any{
			"valid":       result.Valid,
			"config_file": result.ConfigFile,
			"errors":      errMsgs,
			"warnings":    result.Warnings,
		},
		Error: jsonErr,
	}); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}
