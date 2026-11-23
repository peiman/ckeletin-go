// internal/ping/ping.go

package ping

import (
	"fmt"

	"github.com/peiman/ckeletin-go/.ckeletin/pkg/progress"
	"github.com/rs/zerolog/log"
)

// UIRunner is the interactive presentation port. The command layer injects it.
// Business logic does not import the UI package.
type UIRunner interface {
	RunUI(message, color string) error
}

// Config holds configuration for the ping command.
type Config struct {
	Message string
	Color   string
	UI      bool
}

// Executor handles the execution of the ping command.
type Executor struct {
	cfg      Config
	uiRunner UIRunner
}

// NewExecutor creates a new ping command executor.
func NewExecutor(cfg Config, uiRunner UIRunner) *Executor {
	return &Executor{cfg: cfg, uiRunner: uiRunner}
}

// PingResponse is the data the command layer renders.
type PingResponse struct {
	Message   string `json:"message"`
	Color     string `json:"color"`
	Timestamp string `json:"timestamp"`
}

var knownColors = map[string]struct{}{
	"black": {}, "red": {}, "green": {}, "yellow": {},
	"blue": {}, "magenta": {}, "cyan": {}, "white": {},
}

// Execute runs the ping command logic. It reports through rep and does not
// write stdout or stderr. A nil reporter discards the reports.
func (e *Executor) Execute(rep progress.Reporter) (*PingResponse, error) {
	if rep == nil {
		rep = progress.Nop{}
	}
	log.Debug().Str("component", "ping").Msg("Starting ping execution")
	log.Debug().
		Str("component", "ping").
		Str("message", e.cfg.Message).
		Str("color", e.cfg.Color).
		Bool("ui_enabled", e.cfg.UI).
		Msg("Configuration loaded")

	if e.cfg.UI {
		log.Debug().Msg("Starting interactive UI mode")
		rep.Step("starting interactive UI")
		if err := e.uiRunner.RunUI(e.cfg.Message, e.cfg.Color); err != nil {
			rep.Warn("failed to run UI")
			return nil, fmt.Errorf("failed to run UI: %w", err)
		}
		rep.Done("interactive UI finished")
		return &PingResponse{
			Message:   e.cfg.Message,
			Color:     e.cfg.Color,
			Timestamp: "now",
		}, nil
	}

	rep.Step("preparing output")
	if _, ok := knownColors[e.cfg.Color]; !ok {
		log.Debug().Str("color", e.cfg.Color).Msg("Invalid color configuration")
		rep.Warn("invalid color")
		return nil, fmt.Errorf("invalid color: %s", e.cfg.Color)
	}

	response := &PingResponse{
		Message:   e.cfg.Message,
		Color:     e.cfg.Color,
		Timestamp: "now",
	}
	return response, nil
}
