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

// Execute runs the ping command logic
func (e *Executor) Execute() error {
	log.Debug().Msg("Starting ping execution")

	log.Debug().
		Str("message", e.cfg.Message).
		Str("color", e.cfg.Color).
		Bool("ui_enabled", e.cfg.UI).
		Msg("Configuration loaded")

	log.Debug().
		Str("writer_type", fmt.Sprintf("%T", e.writer)).
		Msg("Using writer")

	if e.cfg.UI {
		log.Info().Str("message", e.cfg.Message).Str("color", e.cfg.Color).Msg("Starting UI")
		if err := e.uiRunner.RunUI(e.cfg.Message, e.cfg.Color); err != nil {
			log.Error().Err(err).Msg("Failed to run UI")
			return err
		}
		return nil
	}

	// Non-UI mode: print the message
	err := ui.PrintColoredMessage(e.writer, e.cfg.Message, e.cfg.Color)
	if err != nil {
		log.Error().
			Err(err).
			Str("message", e.cfg.Message).
			Str("color", e.cfg.Color).
			Msg("Failed to print colored message")
		// Wrap the error to provide context
		return fmt.Errorf("failed to print colored message: %w", err)
	}

	log.Debug().Msg("Ping execution completed successfully")
	return nil
}
