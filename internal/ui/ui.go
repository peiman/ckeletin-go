// internal/ui/ui.go

package ui

import (
	"fmt"

	"github.com/rs/zerolog/log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// UIRunner defines an interface for running a UI
type UIRunner interface {
	RunUI(message, col string) error
}

// programFactory is a function type that creates a new tea program
type programFactory func(m tea.Model) *tea.Program

// DefaultUIRunner is the default implementation of UIRunner
type DefaultUIRunner struct {
	// Allow for testing by passing a custom program factory
	newProgram programFactory
}

// NewDefaultUIRunner creates a new DefaultUIRunner with the standard program factory
func NewDefaultUIRunner() *DefaultUIRunner {
	return &DefaultUIRunner{
		newProgram: func(m tea.Model) *tea.Program {
			return tea.NewProgram(m)
		},
	}
}

// For testing only - allows creating a DefaultUIRunner with a nil program factory
// This simulates a successful UI run without actually creating a tea.Program
func NewTestUIRunner() *DefaultUIRunner {
	return &DefaultUIRunner{
		newProgram: nil,
	}
}

// RunUI runs the Bubble Tea UI
func (d *DefaultUIRunner) RunUI(message, col string) error {
	colorStyle, err := GetLipglossColor(col)
	if err != nil {
		log.Error().
			Err(err).
			Str("message", message).
			Str("color", col).
			Msg("Failed to get color style")
		return err
	}

	m := model{
		message:    message,
		colorStyle: lipgloss.NewStyle().Foreground(colorStyle).Bold(true),
	}

	// For testing - if newProgram is nil, just log success and return
	if d.newProgram == nil {
		log.Info().
			Str("message", message).
			Str("color", col).
			Msg("UI ran successfully (test mode)")
		return nil
	}

	// Use the program factory
	p := d.newProgram(m)
	_, err = p.Run()
	if err != nil {
		log.Error().
			Err(err).
			Str("message", message).
			Str("color", col).
			Msg("Failed to run UI")
		return err
	}

	log.Info().
		Str("message", message).
		Str("color", col).
		Msg("UI ran successfully")

	return nil
}

// GetLipglossColor converts a color string to a lipgloss.Color
func GetLipglossColor(col string) (lipgloss.Color, error) {
	if color, ok := ColorMap[col]; ok {
		return color, nil
	}
	err := fmt.Errorf("invalid color: %s", col)
	log.Error().
		Err(err).
		Str("color", col).
		Msg("Failed to get lipgloss color")
	return "", err
}

// model defines the Bubble Tea model
type model struct {
	message    string
	colorStyle lipgloss.Style
	done       bool
}

// Init initializes the model (no-op)
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles messages and updates the model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		case msg.Type == tea.KeyEsc:
			return m, tea.Quit
		case msg.String() == "q":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the model's view
func (m model) View() string {
	return m.colorStyle.Render(m.message) + "\n\nPress 'q' or 'CTRL-C' to exit."
}
