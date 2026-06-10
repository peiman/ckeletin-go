package checkmate

import "github.com/charmbracelet/lipgloss"

// Theme defines the visual appearance of check output.
// Create custom themes by copying DefaultTheme() and modifying values.
type Theme struct {
	// Icons (lipgloss style)
	IconPending string // For CheckHeader (default: ○)
	IconSuccess string // For CheckSuccess (default: ✓)
	IconFailure string // For CheckFailure (default: ✗)
	IconBullet  string // For list items (default: •)
	IconWarning string // For warnings (default: !)

	// Tree connectors
	TreeBranch string // Middle item connector (default: ├──)
	TreeLast   string // Last item connector (default: └──)
	TreeLine   string // Vertical line (default: │)

	// Separators
	SummaryChar string // Horizontal rule character for the summary box border (default: ─)

	// Widths
	SummaryWidth int // Width of the summary box (default: 50)

	// Styles (lipgloss)
	SuccessStyle  lipgloss.Style
	FailureStyle  lipgloss.Style
	WarningStyle  lipgloss.Style
	CategoryStyle lipgloss.Style
	NoteStyle     lipgloss.Style
	InfoStyle     lipgloss.Style
	PendingStyle  lipgloss.Style // For in-progress checks
	TreeStyle     lipgloss.Style // For tree connectors

	// Behavior

	// ForceColors retains the configured/default theme on non-TTY writers
	// instead of degrading to MinimalTheme. It does not force ANSI emission:
	// color emission is decided by the lipgloss renderer for the writer.
	// Themes passed via WithTheme are always retained, so this only matters
	// when the theme was not explicitly chosen.
	ForceColors bool
}

// DefaultTheme returns the default lipgloss-style theme.
// Uses clean Unicode icons and tree connectors.
func DefaultTheme() *Theme {
	return &Theme{
		// Lipgloss-style icons
		IconPending: "○",
		IconSuccess: "✓",
		IconFailure: "✗",
		IconBullet:  "•",
		IconWarning: "!",

		// Tree connectors
		TreeBranch: "├──",
		TreeLast:   "└──",
		TreeLine:   "│",

		SummaryChar: "─",

		SummaryWidth: 50,

		// Bold green for success
		SuccessStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true),
		// Bold red for failure
		FailureStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true),
		// Bold orange for warnings
		WarningStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Bold(true),
		// Lipgloss-style category headers with background
		CategoryStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("63")).
			Bold(true).
			Padding(0, 1),
		// Dim italic for notes
		NoteStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")).
			Italic(true),
		// Dim for info
		InfoStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")),
		// Dim for pending/in-progress
		PendingStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("245")),
		// Dim for tree connectors
		TreeStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")),

		ForceColors: false,
	}
}

// MinimalTheme returns a theme without colors or Unicode icons.
// Suitable for CI environments, piped output, or accessibility needs.
func MinimalTheme() *Theme {
	return &Theme{
		IconPending: "[-]",
		IconSuccess: "[OK]",
		IconFailure: "[FAIL]",
		IconBullet:  "*",
		IconWarning: "[WARN]",

		TreeBranch: "|--",
		TreeLast:   "`--",
		TreeLine:   "|",

		SummaryChar: "=",

		SummaryWidth: 45,

		SuccessStyle:  lipgloss.NewStyle(),
		FailureStyle:  lipgloss.NewStyle(),
		WarningStyle:  lipgloss.NewStyle(),
		CategoryStyle: lipgloss.NewStyle(),
		NoteStyle:     lipgloss.NewStyle(),
		InfoStyle:     lipgloss.NewStyle(),
		PendingStyle:  lipgloss.NewStyle(),
		TreeStyle:     lipgloss.NewStyle(),

		ForceColors: false,
	}
}

// CITheme is an alias for MinimalTheme, optimized for CI/CD pipelines.
func CITheme() *Theme {
	return MinimalTheme()
}
