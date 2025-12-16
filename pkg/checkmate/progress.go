package checkmate

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CheckProgress represents a check with progress tracking.
type CheckProgress struct {
	Name        string
	Status      CheckStatus
	Progress    float64 // 0.0 to 1.0
	Duration    time.Duration
	Error       error
	Remediation string // How to fix the error
}

// CheckStatus represents the status of a check.
type CheckStatus int

const (
	CheckPending CheckStatus = iota
	CheckRunning
	CheckPassed
	CheckFailed
)

// ProgressModel is the Bubble Tea model for progress display.
type ProgressModel struct {
	checks      []CheckProgress
	spinner     spinner.Model
	progress    progress.Model
	done        bool
	title       string
	width       int
	startTime   time.Time
	coverage    float64 // Code coverage percentage (0.0 to 100.0)
	styles      progressStyles
	skipSummary bool // If true, don't show summary box when done
}

type progressStyles struct {
	title     lipgloss.Style
	checkName lipgloss.Style
	pending   lipgloss.Style
	running   lipgloss.Style
	passed    lipgloss.Style
	failed    lipgloss.Style
	box       lipgloss.Style
}

// ProgressModelOption configures a ProgressModel.
type ProgressModelOption func(*ProgressModel)

// WithSkipSummary disables the summary box at the end.
// Use this when showing multiple categories and you want one final summary.
func WithSkipSummary() ProgressModelOption {
	return func(m *ProgressModel) { m.skipSummary = true }
}

// NewProgressModel creates a new progress display model.
func NewProgressModel(title string, checkNames []string, opts ...ProgressModelOption) ProgressModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#78B0E7"))

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(30),
		progress.WithoutPercentage(),
	)

	checks := make([]CheckProgress, len(checkNames))
	for i, name := range checkNames {
		checks[i] = CheckProgress{
			Name:   name,
			Status: CheckPending,
		}
	}

	m := ProgressModel{
		checks:    checks,
		spinner:   s,
		progress:  p,
		title:     title,
		width:     60,
		startTime: time.Now(),
		styles: progressStyles{
			title: lipgloss.NewStyle().
				Background(lipgloss.Color("#78B0E7")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true).
				Padding(0, 2),
			checkName: lipgloss.NewStyle().
				Width(20), // Wide enough for "config-consumption"
			pending: lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")),
			running: lipgloss.NewStyle().
				Foreground(lipgloss.Color("#78B0E7")),
			passed: lipgloss.NewStyle().
				Foreground(lipgloss.Color("42")).
				Bold(true),
			failed: lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true),
			box: lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#78B0E7")).
				Padding(1, 2),
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(&m)
	}

	return m
}

// Init initializes the model.
func (m ProgressModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickCmd())
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// Update handles messages.
func (m ProgressModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.Width = msg.Width - 30

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		if !m.done {
			return m, tickCmd()
		}

	case CheckUpdateMsg:
		if msg.Index >= 0 && msg.Index < len(m.checks) {
			m.checks[msg.Index].Status = msg.Status
			m.checks[msg.Index].Progress = msg.Progress
			m.checks[msg.Index].Duration = msg.Duration
			m.checks[msg.Index].Error = msg.Error
			m.checks[msg.Index].Remediation = msg.Remediation
		}
		return m, nil

	case DoneMsg:
		m.done = true
		return m, tea.Quit

	case CoverageMsg:
		m.coverage = msg.Coverage
		return m, nil
	}

	return m, nil
}

// CheckUpdateMsg updates a check's status.
type CheckUpdateMsg struct {
	Index       int
	Status      CheckStatus
	Progress    float64
	Duration    time.Duration
	Error       error
	Remediation string // How to fix the error (shown on failure)
}

// DoneMsg signals completion.
type DoneMsg struct{}

// CoverageMsg updates the code coverage percentage.
type CoverageMsg struct {
	Coverage float64 // 0.0 to 100.0
}

// View renders the progress display.
func (m ProgressModel) View() string {
	var b strings.Builder

	// Title
	b.WriteString("\n")
	b.WriteString(m.styles.title.Render(m.title))
	b.WriteString("\n\n")

	// Progress bars for each check
	for _, check := range m.checks {
		name := m.styles.checkName.Render(check.Name)

		var status string
		var bar string

		switch check.Status {
		case CheckPending:
			bar = m.renderProgressBar(0, lipgloss.Color("241"))
			status = m.styles.pending.Render("waiting")

		case CheckRunning:
			// Animated spinner effect
			bar = m.renderProgressBar(check.Progress, lipgloss.Color("#78B0E7"))
			status = m.spinner.View()

		case CheckPassed:
			bar = m.renderProgressBar(1.0, lipgloss.Color("42"))
			status = m.styles.passed.Render("✓")

		case CheckFailed:
			bar = m.renderProgressBar(check.Progress, lipgloss.Color("196"))
			status = m.styles.failed.Render("✗")
		}

		// Duration if available
		dur := ""
		if check.Duration > 0 {
			dur = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(fmt.Sprintf(" %s", check.Duration.Round(time.Millisecond)))
		}

		b.WriteString(fmt.Sprintf("  %s %s %s%s\n", name, bar, status, dur))

		// Show error summary if failed (first line only, details in summary box)
		if check.Status == CheckFailed && check.Error != nil {
			errSummary := getErrorSummary(check.Error.Error())
			errStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196"))
			treeStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
			b.WriteString(treeStyle.Render("              └─ "))
			b.WriteString(errStyle.Render(errSummary))
			b.WriteString("\n")
		}
	}

	// If done and summary not skipped, show summary box
	if m.done && !m.skipSummary {
		b.WriteString("\n")
		b.WriteString(m.renderSummaryBox())
	}

	return b.String()
}

// getErrorSummary returns a short summary of an error message (first line, truncated).
func getErrorSummary(errMsg string) string {
	// Get first line
	firstLine := errMsg
	if idx := strings.Index(errMsg, "\n"); idx != -1 {
		firstLine = errMsg[:idx]
	}

	// Truncate if too long
	const maxLen = 50
	if len(firstLine) > maxLen {
		firstLine = firstLine[:maxLen-3] + "..."
	}

	return firstLine
}

// formatErrorDetails splits error message into lines, limiting width and count.
func formatErrorDetails(errMsg string, maxWidth int) []string {
	const maxLines = 10 // Limit lines to keep output manageable

	// Split by newlines first
	rawLines := strings.Split(errMsg, "\n")
	var result []string

	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Wrap long lines
		for len(line) > maxWidth {
			// Find last space before maxWidth
			breakPoint := maxWidth
			for i := maxWidth; i > maxWidth/2; i-- {
				if line[i] == ' ' {
					breakPoint = i
					break
				}
			}
			result = append(result, line[:breakPoint])
			line = strings.TrimSpace(line[breakPoint:])
		}
		if line != "" {
			result = append(result, line)
		}

		if len(result) >= maxLines {
			result = append(result, "... (truncated)")
			break
		}
	}

	return result
}

func (m ProgressModel) renderProgressBar(percent float64, color lipgloss.Color) string {
	width := 25
	filled := int(percent * float64(width))
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(color)
	emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	bar := filledStyle.Render(strings.Repeat("█", filled))
	bar += emptyStyle.Render(strings.Repeat("░", empty))

	pct := lipgloss.NewStyle().
		Width(4).
		Align(lipgloss.Right).
		Foreground(lipgloss.Color("252")).
		Render(fmt.Sprintf("%d%%", int(percent*100)))

	return bar + " " + pct
}

func (m ProgressModel) renderSummaryBox() string {
	passed := 0
	failed := 0
	for _, c := range m.checks {
		switch c.Status {
		case CheckPassed:
			passed++
		case CheckFailed:
			failed++
		}
	}

	total := len(m.checks)
	allPassed := failed == 0

	// Box styling
	var borderColor lipgloss.Color
	var titleStyle lipgloss.Style
	if allPassed {
		borderColor = lipgloss.Color("42")
		titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("42")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 2)
	} else {
		borderColor = lipgloss.Color("196")
		titleStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("196")).
			Foreground(lipgloss.Color("230")).
			Bold(true).
			Padding(0, 2)
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		MarginBottom(1)

	var content strings.Builder

	// Header inside box
	if allPassed {
		content.WriteString(titleStyle.Render("✓ All Checks Passed"))
	} else {
		content.WriteString(titleStyle.Render(fmt.Sprintf("✗ %d/%d Checks Failed", failed, total)))
	}
	content.WriteString("\n\n")

	// Status items
	for _, check := range m.checks {
		var icon, style string
		switch check.Status {
		case CheckPassed:
			icon = "✓"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(icon)
		case CheckFailed:
			icon = "✗"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(icon)
		default:
			icon = "○"
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(icon)
		}

		name := check.Name
		switch check.Status {
		case CheckPassed:
			name = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(name + " passed")
		case CheckFailed:
			name = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(name + " failed")
		}

		dur := ""
		if check.Duration > 0 {
			dur = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Render(fmt.Sprintf(" (%s)", check.Duration.Round(time.Millisecond)))
		}

		content.WriteString(fmt.Sprintf("  %s %s%s\n", style, name, dur))
	}

	// Error details section for failed checks
	hasErrors := false
	for _, check := range m.checks {
		if check.Status == CheckFailed && check.Error != nil {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		content.WriteString("\n")
		errorHeader := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true).
			Render("─── Errors ───")
		content.WriteString("  " + errorHeader + "\n")

		for _, check := range m.checks {
			if check.Status != CheckFailed || check.Error == nil {
				continue
			}

			// Check name
			nameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true)
			content.WriteString("\n  " + nameStyle.Render(check.Name) + ":\n")

			// Error details (word-wrapped, indented)
			errLines := formatErrorDetails(check.Error.Error(), 60)
			errStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("252"))
			for _, line := range errLines {
				content.WriteString("    " + errStyle.Render(line) + "\n")
			}

			// Remediation if available
			if check.Remediation != "" {
				remStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("214")).
					Bold(true)
				content.WriteString("\n    " + remStyle.Render("Fix: ") + check.Remediation + "\n")
			}
		}
	}

	// Code coverage bar (only show when all checks passed)
	if allPassed {
		content.WriteString("\n")
		coveragePercent := m.coverage / 100.0 // Convert to 0.0-1.0 range
		var coverageColor lipgloss.Color
		switch {
		case m.coverage >= 80:
			coverageColor = lipgloss.Color("42") // Green
		case m.coverage >= 60:
			coverageColor = lipgloss.Color("214") // Yellow/Orange
		default:
			coverageColor = lipgloss.Color("196") // Red
		}
		bar := m.renderProgressBar(coveragePercent, coverageColor)
		content.WriteString(fmt.Sprintf("  Coverage: %s", bar))
	}

	// Duration
	duration := time.Since(m.startTime).Round(time.Millisecond)
	content.WriteString(fmt.Sprintf("\n  Duration: %s", duration))

	return box.Render(content.String())
}
