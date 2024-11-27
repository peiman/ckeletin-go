// internal/ui/message.go

package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

// PrintColoredMessage prints a message to the console with a specific color style
func PrintColoredMessage(out io.Writer, message, col string) error {
	colorStyle, err := GetLipglossColor(col)
	if err != nil {
		log.Error().Err(err).Str("color", col).Msg("Invalid color")
		return fmt.Errorf("invalid color: %w", err)
	}

	style := lipgloss.NewStyle().Foreground(colorStyle).Bold(true)
	fmt.Fprintln(out, style.Render(message))
	return nil
}
