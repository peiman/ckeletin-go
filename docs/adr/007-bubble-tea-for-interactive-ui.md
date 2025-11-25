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

## Enforcement

**Status: N/A - Optional Pattern**

Bubble Tea usage is intentionally optional and not enforced:

**1. Opt-In Design**
- Commands use `--ui` flag to enable interactive mode
- Default is non-interactive (standard output)
- User chooses when interactive UI is appropriate

**2. Interface Abstraction**
```go
type UIRunner interface {
    RunUI(message, color string) error
}
```
- Interface allows alternative implementations
- Commands depend on interface, not concrete type
- Easy to swap UI frameworks if needed

**3. Testing Support**
- `internal/ui/mock.go` provides test implementation
- No Bubble Tea dependency needed in tests
- Interface-based testing per ADR-003

**4. Why No Enforcement**
- Interactive UI is feature enhancement, not requirement
- Some commands don't need UI (batch processing, scripts)
- Forcing UI would violate Unix philosophy
- Pattern is "available" not "mandatory"

**5. Related Validation**
While Bubble Tea itself isn't enforced, output patterns are:
```bash
task validate:output  # Ensures proper output stream usage
```
This validates that commands use proper output methods (via ui package or stdout), but doesn't require interactive features.

## References
- `internal/ui/ui.go` - UIRunner interface and implementation
- `internal/ui/mock.go` - Test mock
- `cmd/ping.go` - Usage example
- [Bubble Tea docs](https://github.com/charmbracelet/bubbletea)
