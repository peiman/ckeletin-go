// internal/ui/colors.go

package ui

import "github.com/charmbracelet/lipgloss"

// ColorMap maps color names to their lipgloss.Color values
var ColorMap = map[string]lipgloss.Color{
	"black":   lipgloss.Color("#000000"),
	"red":     lipgloss.Color("#FF0000"),
	"green":   lipgloss.Color("#00FF00"),
	"yellow":  lipgloss.Color("#FFFF00"),
	"blue":    lipgloss.Color("#0000FF"),
	"magenta": lipgloss.Color("#FF00FF"),
	"cyan":    lipgloss.Color("#00FFFF"),
	"white":   lipgloss.Color("#FFFFFF"),
}
