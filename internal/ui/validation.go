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

// validationData is the machine-readable config-validation payload for
// --output json. Errors and Warnings always marshal as JSON arrays (never
// null) so consumers can iterate them without a null check.
type validationData struct {
	Valid      bool     `json:"valid"`
	ConfigFile string   `json:"config_file"`
	Errors     []string `json:"errors"`
	Warnings   []string `json:"warnings"`
}

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
	data := validationData{
		Valid:      result.Valid,
		ConfigFile: result.ConfigFile,
		Errors:     make([]string, 0, len(result.Errors)),
		Warnings:   result.Warnings,
	}
	for _, e := range result.Errors {
		data.Errors = append(data.Errors, e.Error())
	}
	if data.Warnings == nil {
		data.Warnings = []string{}
	}
	if err := output.RenderJSON(out, output.JSONEnvelope{
		Status:  status,
		Command: output.CommandName(),
		Data:    data,
		Error:   jsonErr,
	}); err != nil {
		return fmt.Errorf("failed to write JSON output: %w", err)
	}
	return nil
}
