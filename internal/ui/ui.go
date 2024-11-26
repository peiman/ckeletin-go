// internal/ui/ui.go
package ui

import (
    "fmt"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
)

// model defines the Bubble Tea model
type model struct {
    message    string
    colorStyle lipgloss.Style
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

// RunUI runs the Bubble Tea UI
func RunUI(message, col string) error {
    colorStyle, err := GetLipglossColor(col)
    if err != nil {
        return err
    }

    m := model{
        message:    message,
        colorStyle: lipgloss.NewStyle().Foreground(colorStyle).Bold(true),
    }

    p := tea.NewProgram(m)
    _, err = p.Run()
    return err
}

// GetLipglossColor converts a color string to a lipgloss.Color
func GetLipglossColor(col string) (lipgloss.Color, error) {
    switch col {
    case "black":
        return lipgloss.Color("#000000"), nil
    case "red":
        return lipgloss.Color("#FF0000"), nil
    case "green":
        return lipgloss.Color("#00FF00"), nil
    case "yellow":
        return lipgloss.Color("#FFFF00"), nil
    case "blue":
        return lipgloss.Color("#0000FF"), nil
    case "magenta":
        return lipgloss.Color("#FF00FF"), nil
    case "cyan":
        return lipgloss.Color("#00FFFF"), nil
    case "white":
        return lipgloss.Color("#FFFFFF"), nil
    default:
        return "", fmt.Errorf("invalid color: %s", col)
    }
}
