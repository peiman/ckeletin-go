# ADR-007: Bubble Tea for Interactive UI

## Status
Accepted

## Context

CLI applications can benefit from interactive UIs for:
- Better user experience
- Visual feedback
- Interactive selections
- Real-time updates

Requirements:
- Modern terminal UI
- Good developer experience
- Cross-platform support
- Easy testing

## Decision

Use **Bubble Tea** (Charm) for interactive terminal UIs:

### Architecture

```go
// internal/ui/ui.go
type UIRunner interface {
    RunUI(message, color string) error
}

type DefaultUIRunner struct {
    newProgram programFactory
}

func (d *DefaultUIRunner) RunUI(message, col string) error {
    m := model{message: message, colorStyle: getStyle(col)}
    p := tea.NewProgram(m)
    _, err := p.Run()
    return err
}
```

### Model Pattern

```go
type model struct {
    message    string
    colorStyle lipgloss.Style
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Handle keyboard input
}

func (m model) View() string {
    return m.colorStyle.Render(m.message)
}
```

## Consequences

### Positive
- Beautiful, modern terminal UIs
- Elm-architecture makes UIs predictable
- Excellent documentation and examples
- Active community and ecosystem
- Lipgloss for styling (colors, borders, etc.)
- Built-in testing support

### Negative
- Learning curve for Elm architecture
- Overkill for simple use cases
- Additional dependency

### Mitigations
- Interface abstraction (UIRunner) allows alternatives
- Optional UI mode (--ui flag)
- Simple mock for testing
- Clear examples in ping command

## Testing Strategy

```go
// Use interface for easy testing
type mockUIRunner struct {
    CalledWithMessage string
    ReturnError       error
}

func (m *mockUIRunner) RunUI(message, col string) error {
    m.CalledWithMessage = message
    return m.ReturnError
}
```

## Alternative Considered

**Survey** - Simpler but less flexible
**Rejected because**: Bubble Tea offers more control and better UX

## References
- `internal/ui/ui.go` - UIRunner interface and implementation
- `internal/ui/mock.go` - Test mock
- `cmd/ping.go` - Usage example
- [Bubble Tea docs](https://github.com/charmbracelet/bubbletea)
