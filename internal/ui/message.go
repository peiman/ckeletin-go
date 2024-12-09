package ui

import (
	"fmt"
	"io"

	"github.com/charmbracelet/lipgloss"
)

// PrintColoredMessage prints a message to the console with a specific color
func PrintColoredMessage(out io.Writer, message, col string) error {
	colorStyle := lipgloss.Color(col)
	style := lipgloss.NewStyle().Foreground(colorStyle).Bold(true)
	_, err := fmt.Fprintln(out, style.Render(message))
	return err
}
