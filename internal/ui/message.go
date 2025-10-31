// internal/ui/message.go

package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

// PrintColoredMessage prints a message to the console with a specific color
func PrintColoredMessage(out io.Writer, message, col string) error {
	log.Debug().
		Str("component", "ui").
		Str("message", message).
		Str("color", col).
		Msg("PrintColoredMessage called")

	colorStyle, err := GetLipglossColor(col)
	if err != nil {
		log.Error().Err(err).Str("component", "ui").Str("color", col).Msg("Invalid color")
		return fmt.Errorf("invalid color: %w", err)
	}

	style := lipgloss.NewStyle().Foreground(colorStyle).Bold(true)

	log.Debug().Str("component", "ui").Msg("Attempting to write styled message")
	_, err = fmt.Fprintln(out, style.Render(message))
	if err != nil {
		log.Error().
			Err(err).
			Str("component", "ui").
			Str("message", message).
			Msg("Failed to write message")
		return fmt.Errorf("failed to write message: %w", err)
	}

	log.Debug().Str("component", "ui").Msg("PrintColoredMessage completed successfully")
	return nil
}
