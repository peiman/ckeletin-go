// internal/ping/ping.go

package ping

import (
	"fmt"
	"io"

	"github.com/peiman/ckeletin-go/internal/ui"
	"github.com/rs/zerolog/log"
)

// Config holds configuration for the ping command
type Config struct {
	Message string
	Color   string
	UI      bool
}

// Executor handles the execution of the ping command
type Executor struct {
	cfg      Config
	uiRunner ui.UIRunner
	writer   io.Writer
}

// NewExecutor creates a new ping command executor
func NewExecutor(cfg Config, uiRunner ui.UIRunner, writer io.Writer) *Executor {
	return &Executor{
		cfg:      cfg,
		uiRunner: uiRunner,
		writer:   writer,
	}
}

// PingResponse represents the structured data returned by the command
// This demonstrates the "Shadow Log" pattern: this struct is logged to the file
// while the user sees a formatted message.
type PingResponse struct {
	Message   string `json:"message"`
	Color     string `json:"color"`
	Timestamp string `json:"timestamp"`
}

// Execute runs the ping command logic
func (e *Executor) Execute() error {
	log.Debug().Str("component", "ping").Msg("Starting ping execution")

	log.Debug().
		Str("component", "ping").
		Str("message", e.cfg.Message).
		Str("color", e.cfg.Color).
		Bool("ui_enabled", e.cfg.UI).
		Msg("Configuration loaded")

	if e.cfg.UI {
		// For UI mode, we log that we are handing over control
		log.Info().Msg("Starting interactive UI mode")

		if err := e.uiRunner.RunUI(e.cfg.Message, e.cfg.Color); err != nil {
			// Use RenderError for consistent error reporting
			if renderErr := ui.RenderError(e.writer, "Failed to run UI", err); renderErr != nil {
				log.Warn().Err(renderErr).Msg("Failed to render error message")
			}
			return fmt.Errorf("failed to run UI: %w", err)
		}
		return nil
	}

	// Non-UI mode: Structured Output + Shadow Logging

	// 1. Status Stream (Stderr) - Tell the operator what we are doing
	log.Info().Msg("Preparing output...")

	// 2. Prepare the Data

	// Validate color before proceeding (Business Rule)
	if _, err := ui.GetLipglossColor(e.cfg.Color); err != nil {
		log.Error().Err(err).Str("color", e.cfg.Color).Msg("Invalid color configuration")
		return fmt.Errorf("invalid color: %w", err)
	}

	response := PingResponse{
		Message:   e.cfg.Message,
		Color:     e.cfg.Color,
		Timestamp: "now", // In a real app, use time.Now().Format(time.RFC3339)
	}

	// 3. Data Stream (Stdout) & Audit Stream (Log File)
	// Use the UI adapter to render for the user AND log the struct for the auditor
	if err := ui.RenderSuccess(e.writer, fmt.Sprintf("Pong! %s", response.Message), response); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}
