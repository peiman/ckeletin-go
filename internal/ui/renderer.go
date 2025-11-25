package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

// RenderSuccess prints a success message to the writer (stdout) and shadows it to the log file.
//
// Stream 1 (Data): Prints formatted message to out
// Stream 3 (Audit): Logs raw data to file
func RenderSuccess(out io.Writer, message string, data interface{}) error {
	// 1. Shadow Log (Audit Stream)
	// We log the raw data so the log file has the full context of what was returned
	event := log.Info().Str("user_message", message)

	if data != nil {
		event.Interface("data", data)
	}

	event.Msg("Command executed successfully")

	// 2. User Output (Data Stream)
	// We use a checkmark and bold text for better UX
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true) // Green
	if _, err := fmt.Fprintf(out, "%s %s\n", style.Render("✔"), message); err != nil {
		return err
	}

	// If the data is a string and distinct from message, print it too
	if str, ok := data.(string); ok && str != "" && str != message {
		if _, err := fmt.Fprintln(out, str); err != nil {
			return err
		}
	}

	return nil
}

// RenderError prints an error message to the writer (stderr usually) and logs the full error.
//
// Stream 2 (Status): Prints friendly error to out
// Stream 3 (Audit): Logs full error with stack trace to file
func RenderError(out io.Writer, friendlyMessage string, err error) error {
	// 1. Shadow Log (Audit Stream)
	// Capture the full technical error
	log.Error().
		Err(err).
		Str("user_error", friendlyMessage).
		Msg("Command failed")

	// 2. User Output (Status Stream)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true) // Red
	_, writeErr := fmt.Fprintf(out, "%s %s\n", style.Render("✘ Error:"), friendlyMessage)
	return writeErr
}

// RenderTable would go here in the future for tabular data
