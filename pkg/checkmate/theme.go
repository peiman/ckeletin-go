package checkmate

import "github.com/charmbracelet/lipgloss"

// Theme defines the visual appearance of check output.
// Create custom themes by copying DefaultTheme() and modifying values.
type Theme struct {
	// Icons
	IconSearch  string // For CheckHeader (default: 🔍)
	IconSuccess string // For CheckSuccess (default: ✅)
	IconFailure string // For CheckFailure (default: ❌)
	IconBullet  string // For remediation items (default: •)
	IconWarning string // For warnings (default: ⚠️)

	// Separators
	CategoryChar string // Character for category header line (default: ─)
	SummaryChar  string // Character for summary box (default: ━)

	// Widths
	CategoryWidth int // Width of category header (default: 48)
	SummaryWidth  int // Width of summary separator (default: 45)

	// Styles (lipgloss)
	SuccessStyle  lipgloss.Style
	FailureStyle  lipgloss.Style
	WarningStyle  lipgloss.Style
	CategoryStyle lipgloss.Style
	NoteStyle     lipgloss.Style
	InfoStyle     lipgloss.Style

	// Behavior
	ForceColors bool // Force colors even in non-TTY (useful for testing)
}

// DefaultTheme returns the default colorful theme with emojis.
// Best for interactive terminal use.
func DefaultTheme() *Theme {
	return &Theme{
		IconSearch:  "🔍",
		IconSuccess: "✅",
		IconFailure: "❌",
		IconBullet:  "•",
		IconWarning: "⚠️",

		CategoryChar: "─",
		SummaryChar:  "━",

		CategoryWidth: 48,
		SummaryWidth:  45,

		SuccessStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("42")),  // Green
		FailureStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("196")), // Red
		WarningStyle:  lipgloss.NewStyle().Foreground(lipgloss.Color("214")), // Orange
		CategoryStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // Gray
		NoteStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true),
		InfoStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("250")),

		ForceColors: false,
	}
}

// MinimalTheme returns a theme without colors or emojis.
// Suitable for CI environments, piped output, or accessibility needs.
func MinimalTheme() *Theme {
	return &Theme{
		IconSearch:  "[-]",
		IconSuccess: "[OK]",
		IconFailure: "[FAIL]",
		IconBullet:  "*",
		IconWarning: "[WARN]",

		CategoryChar: "-",
		SummaryChar:  "=",

		CategoryWidth: 48,
		SummaryWidth:  45,

		SuccessStyle:  lipgloss.NewStyle(),
		FailureStyle:  lipgloss.NewStyle(),
		WarningStyle:  lipgloss.NewStyle(),
		CategoryStyle: lipgloss.NewStyle(),
		NoteStyle:     lipgloss.NewStyle(),
		InfoStyle:     lipgloss.NewStyle(),

		ForceColors: false,
	}
}

// CITheme is an alias for MinimalTheme, optimized for CI/CD pipelines.
func CITheme() *Theme {
	return MinimalTheme()
}
