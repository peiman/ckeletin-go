package check

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/peiman/ckeletin-go/pkg/checkmate"
)

// printFinalSummary prints the final summary box with all check results.
// Uses styled output for TTY, plain ASCII for CI/pipes.
func (e *Executor) printFinalSummary(results []allCheckResult, passed, failed int, totalDuration time.Duration) {
	// Only clear screen in TUI mode
	if e.useTUI {
		_, _ = fmt.Fprint(e.writer, "\033[2J\033[H")
	} else {
		_, _ = fmt.Fprintln(e.writer) // Just add a blank line in CI mode
	}

	allPassed := failed == 0

	// Group results by category
	categoryOrder := []string{
		"Development Environment",
		"Code Quality",
		"Architecture Validation",
		"Security Scanning",
		"Dependencies",
		"Tests",
	}

	resultsByCategory := make(map[string][]allCheckResult)
	for _, r := range results {
		resultsByCategory[r.category] = append(resultsByCategory[r.category], r)
	}

	// Box characters - use ASCII for CI mode
	boxChars := summaryBoxChars(e.useTUI)

	boxWidth := 60
	contentWidth := boxWidth - 2

	var sb strings.Builder

	// Define styles - only use colors in TUI mode
	styles := summaryStyles(e.useTUI, allPassed)

	// Top border
	sb.WriteString(styles.border.Render(boxChars.topLeft+strings.Repeat(boxChars.horizontal, boxWidth-2)+boxChars.topRight) + "\n")

	// Empty line
	sb.WriteString(styles.border.Render(boxChars.vertical) + strings.Repeat(" ", contentWidth) + styles.border.Render(boxChars.vertical) + "\n")

	// Title
	var titleText string
	if allPassed {
		titleText = fmt.Sprintf(" %s All %d Checks Passed ", boxChars.successIcon, passed)
	} else {
		titleText = fmt.Sprintf(" %s %d/%d Checks Failed ", boxChars.failIcon, failed, passed+failed)
	}
	titleRendered := styles.title.Render(titleText)
	// Calculate padding (lipgloss.Width handles unicode properly)
	titleWidth := lipgloss.Width(titleRendered)
	leftPad := (contentWidth - titleWidth) / 2
	rightPad := contentWidth - leftPad - titleWidth
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}
	sb.WriteString(styles.border.Render(boxChars.vertical))
	sb.WriteString(strings.Repeat(" ", leftPad))
	sb.WriteString(titleRendered)
	sb.WriteString(strings.Repeat(" ", rightPad))
	sb.WriteString(styles.border.Render(boxChars.vertical) + "\n")

	// Empty line
	sb.WriteString(styles.border.Render(boxChars.vertical) + strings.Repeat(" ", contentWidth) + styles.border.Render(boxChars.vertical) + "\n")

	// Results grouped by category
	for _, catName := range categoryOrder {
		catResults, ok := resultsByCategory[catName]
		if !ok || len(catResults) == 0 {
			continue
		}

		// Category header
		catHeader := "  " + styles.dim.Render(boxChars.catSeparator+" "+catName)
		catHeaderWidth := 2 + len(boxChars.catSeparator) + 1 + len(catName) // "  " + separator + " " + name
		padding := contentWidth - catHeaderWidth
		sb.WriteString(styles.border.Render(boxChars.vertical) + catHeader)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(styles.border.Render(boxChars.vertical) + "\n")

		// Check results
		for i, r := range catResults {
			var iconStyle lipgloss.Style
			var icon string
			if r.passed {
				icon = boxChars.successIcon
				iconStyle = styles.success
			} else {
				icon = boxChars.failIcon
				iconStyle = styles.fail
			}

			// Tree connector
			connector := boxChars.treeConnector
			if i == len(catResults)-1 {
				connector = boxChars.treeLastConnector
			}

			// Format duration
			durStr := ""
			durLen := 0
			if r.duration > 0 {
				durText := fmt.Sprintf("(%s)", r.duration.Round(time.Millisecond))
				durStr = styles.dim.Render(durText)
				durLen = len(durText)
			}

			// Build line: "  ├── ✓ name              (duration)"
			line := "  " + styles.dim.Render(connector) + " " + iconStyle.Render(icon) + " " + fmt.Sprintf("%-18s", r.name) + " " + durStr
			visibleLen := 2 + len(connector) + 1 + len(icon) + 1 + 18 + 1 + durLen
			padding := contentWidth - visibleLen
			sb.WriteString(styles.border.Render(boxChars.vertical) + line)
			if padding > 0 {
				sb.WriteString(strings.Repeat(" ", padding))
			}
			sb.WriteString(styles.border.Render(boxChars.vertical) + "\n")
		}

		// Empty line after category
		sb.WriteString(styles.border.Render(boxChars.vertical) + strings.Repeat(" ", contentWidth) + styles.border.Render(boxChars.vertical) + "\n")
	}

	// Coverage
	if e.coverage > 0 {
		covText := fmt.Sprintf("%.1f%%", e.coverage)
		covLine := "  " + styles.bold.Render("Coverage:") + " " + covText
		covVisibleLen := 2 + 9 + 1 + len(covText)
		padding := contentWidth - covVisibleLen
		sb.WriteString(styles.border.Render(boxChars.vertical) + covLine)
		if padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
		sb.WriteString(styles.border.Render(boxChars.vertical) + "\n")
	}

	// Duration
	durText := totalDuration.Round(time.Millisecond).String()
	durLine := "  " + styles.bold.Render("Duration:") + " " + durText
	durVisibleLen := 2 + 9 + 1 + len(durText)
	padding := contentWidth - durVisibleLen
	sb.WriteString(styles.border.Render(boxChars.vertical) + durLine)
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString(styles.border.Render(boxChars.vertical) + "\n")

	// Empty line
	sb.WriteString(styles.border.Render(boxChars.vertical) + strings.Repeat(" ", contentWidth) + styles.border.Render(boxChars.vertical) + "\n")

	// Bottom border
	sb.WriteString(styles.border.Render(boxChars.bottomLeft+strings.Repeat(boxChars.horizontal, boxWidth-2)+boxChars.bottomRight) + "\n")

	_, _ = fmt.Fprint(e.writer, sb.String())

	// Print errors below the box if any
	if !allPassed {
		_, _ = fmt.Fprint(e.writer, "\n")
		var printer *checkmate.Printer
		if e.useTUI {
			printer = checkmate.New(checkmate.WithWriter(e.writer))
		} else {
			printer = checkmate.New(checkmate.WithWriter(e.writer), checkmate.WithTheme(checkmate.CITheme()))
		}
		for _, r := range results {
			if !r.passed && r.err != nil {
				printer.CheckFailure(r.name, r.err.Error(), r.remediation)
			}
		}
	}
}

// boxCharSet holds the characters used for summary box rendering
type boxCharSet struct {
	topLeft           string
	topRight          string
	bottomLeft        string
	bottomRight       string
	horizontal        string
	vertical          string
	catSeparator      string
	treeConnector     string
	treeLastConnector string
	successIcon       string
	failIcon          string
}

// summaryBoxChars returns box-drawing characters appropriate for the output mode
func summaryBoxChars(useTUI bool) boxCharSet {
	if useTUI {
		return boxCharSet{
			topLeft:           "╭",
			topRight:          "╮",
			bottomLeft:        "╰",
			bottomRight:       "╯",
			horizontal:        "─",
			vertical:          "│",
			catSeparator:      "───",
			treeConnector:     "├──",
			treeLastConnector: "└──",
			successIcon:       "✓",
			failIcon:          "✗",
		}
	}
	return boxCharSet{
		topLeft:           "+",
		topRight:          "+",
		bottomLeft:        "+",
		bottomRight:       "+",
		horizontal:        "-",
		vertical:          "|",
		catSeparator:      "---",
		treeConnector:     "|--",
		treeLastConnector: "`--",
		successIcon:       "[OK]",
		failIcon:          "[FAIL]",
	}
}

// summaryStyleSet holds all lipgloss styles for summary rendering
type summaryStyleSet struct {
	border  lipgloss.Style
	dim     lipgloss.Style
	bold    lipgloss.Style
	success lipgloss.Style
	fail    lipgloss.Style
	title   lipgloss.Style
}

// summaryStyles returns styles appropriate for the output mode
func summaryStyles(useTUI bool, allPassed bool) summaryStyleSet {
	if !useTUI {
		plain := lipgloss.NewStyle()
		return summaryStyleSet{
			border:  plain,
			dim:     plain,
			bold:    plain,
			success: plain,
			fail:    plain,
			title:   plain,
		}
	}

	accentColor := lipgloss.Color("#78B0E7")
	failColor := lipgloss.Color("#FF5555")
	successColor := lipgloss.Color("#50FA7B")
	dimColor := lipgloss.Color("#6272A4")

	var borderColor lipgloss.Color
	var titleStyle lipgloss.Style
	if allPassed {
		borderColor = accentColor
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Background(accentColor).
			Foreground(lipgloss.Color("#000000"))
	} else {
		borderColor = failColor
		titleStyle = lipgloss.NewStyle().
			Bold(true).
			Background(failColor).
			Foreground(lipgloss.Color("#000000"))
	}

	return summaryStyleSet{
		border:  lipgloss.NewStyle().Foreground(borderColor),
		dim:     lipgloss.NewStyle().Foreground(dimColor),
		bold:    lipgloss.NewStyle().Bold(true),
		success: lipgloss.NewStyle().Foreground(successColor),
		fail:    lipgloss.NewStyle().Foreground(failColor),
		title:   titleStyle,
	}
}
