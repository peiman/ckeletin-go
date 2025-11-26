package progress

import (
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Style defines the visual appearance of progress elements.
type Style struct {
	// Spinner configuration
	SpinnerFrames   []string
	SpinnerInterval time.Duration
	SpinnerStyle    lipgloss.Style

	// Progress bar configuration
	BarWidth      int
	BarChar       string
	BarEmptyChar  string
	BarStyle      lipgloss.Style
	BarEmptyStyle lipgloss.Style

	// Status styles
	SuccessStyle lipgloss.Style
	ErrorStyle   lipgloss.Style
	WarningStyle lipgloss.Style

	// Text styles
	PhaseStyle   lipgloss.Style
	TaskStyle    lipgloss.Style
	MessageStyle lipgloss.Style
	CounterStyle lipgloss.Style
}

// DefaultStyle returns the default progress style.
func DefaultStyle() *Style {
	return &Style{
		// Braille spinner (smooth animation)
		SpinnerFrames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		SpinnerInterval: 80 * time.Millisecond,
		SpinnerStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("205")), // Pink

		// Progress bar
		BarWidth:      30,
		BarChar:       "█",
		BarEmptyChar:  "░",
		BarStyle:      lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // Green
		BarEmptyStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("240")), // Gray

		// Status colors
		SuccessStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),  // Green
		ErrorStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true), // Red
		WarningStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true), // Orange

		// Text styles
		PhaseStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("39")).Bold(true), // Blue
		TaskStyle:    lipgloss.NewStyle().Foreground(lipgloss.Color("252")),           // Light gray
		MessageStyle: lipgloss.NewStyle(),
		CounterStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // Gray
	}
}

// MinimalStyle returns a minimal style with fewer visual elements.
func MinimalStyle() *Style {
	return &Style{
		// Simple spinner
		SpinnerFrames:   []string{"-", "\\", "|", "/"},
		SpinnerInterval: 100 * time.Millisecond,
		SpinnerStyle:    lipgloss.NewStyle(),

		// Simple progress bar
		BarWidth:      20,
		BarChar:       "#",
		BarEmptyChar:  "-",
		BarStyle:      lipgloss.NewStyle(),
		BarEmptyStyle: lipgloss.NewStyle(),

		// Plain status
		SuccessStyle: lipgloss.NewStyle(),
		ErrorStyle:   lipgloss.NewStyle(),
		WarningStyle: lipgloss.NewStyle(),

		// Plain text
		PhaseStyle:   lipgloss.NewStyle(),
		TaskStyle:    lipgloss.NewStyle(),
		MessageStyle: lipgloss.NewStyle(),
		CounterStyle: lipgloss.NewStyle(),
	}
}
