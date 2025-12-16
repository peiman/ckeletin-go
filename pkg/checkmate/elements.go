package checkmate

import (
	"fmt"
	"strings"
	"time"
)

// Note: lipgloss import removed - we use p.style() helper for rendering

// renderCategoryHeader renders a category header with lipgloss styling.
// Example: " Code Quality " (with colored background)
func (p *Printer) renderCategoryHeader(title string) {
	// Render title with styled background (lipgloss style)
	styledTitle := p.style(p.theme.CategoryStyle, title)

	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintln(p.writer, styledTitle)
	_, _ = fmt.Fprintln(p.writer)
}

// renderCheckHeader renders a check-in-progress message.
// Example: "├── ○ format..."
// In TTY mode: prints without newline so the result can overwrite this line.
// In non-TTY mode: skips output (only shows final result).
func (p *Printer) renderCheckHeader(message string) {
	// In non-TTY mode, skip the running indicator - only show final result
	if !p.isTerminal {
		return
	}
	tree := p.style(p.theme.TreeStyle, p.theme.TreeBranch)
	icon := p.style(p.theme.PendingStyle, p.theme.IconPending)
	text := p.style(p.theme.PendingStyle, message)
	// No newline - result will overwrite this line
	_, _ = fmt.Fprintf(p.writer, "%s %s %s", tree, icon, text)
}

// renderCheckSuccess renders a success message.
// Example: "├── ✓ format"
// In TTY mode: clears the current line first (to overwrite the check header).
// In non-TTY mode: prints a clean line without escape codes.
func (p *Printer) renderCheckSuccess(message string) {
	tree := p.style(p.theme.TreeStyle, p.theme.TreeBranch)
	icon := p.style(p.theme.SuccessStyle, p.theme.IconSuccess)
	text := p.style(p.theme.SuccessStyle, message)

	if p.isTerminal {
		// \r = carriage return, \033[K = clear to end of line
		_, _ = fmt.Fprintf(p.writer, "\r\033[K%s %s %s\n", tree, icon, text)
	} else {
		// Clean output for CI/pipes - no escape codes
		_, _ = fmt.Fprintf(p.writer, "%s %s %s\n", tree, icon, text)
	}
}

// renderCheckFailure renders a failure with details and remediation.
// In TTY mode: clears the current line first (to overwrite the check header).
// In non-TTY mode: prints a clean line without escape codes.
func (p *Printer) renderCheckFailure(title, details, remediation string) {
	tree := p.style(p.theme.TreeStyle, p.theme.TreeBranch)
	icon := p.style(p.theme.FailureStyle, p.theme.IconFailure)
	text := p.style(p.theme.FailureStyle, title)
	treeLine := p.style(p.theme.TreeStyle, p.theme.TreeLine)

	if p.isTerminal {
		// \r = carriage return, \033[K = clear to end of line
		_, _ = fmt.Fprintf(p.writer, "\r\033[K%s %s %s\n", tree, icon, text)
	} else {
		// Clean output for CI/pipes - no escape codes
		_, _ = fmt.Fprintf(p.writer, "%s %s %s\n", tree, icon, text)
	}

	if details != "" {
		detailsHeader := p.style(p.theme.NoteStyle, "Details:")
		_, _ = fmt.Fprintf(p.writer, "%s   %s\n", treeLine, detailsHeader)
		// Indent each line of details
		for _, line := range strings.Split(details, "\n") {
			styled := p.style(p.theme.InfoStyle, line)
			_, _ = fmt.Fprintf(p.writer, "%s     %s\n", treeLine, styled)
		}
	}

	if remediation != "" {
		howToFix := p.style(p.theme.NoteStyle, "How to fix:")
		_, _ = fmt.Fprintf(p.writer, "%s   %s\n", treeLine, howToFix)
		// Each remediation line gets a bullet
		for _, line := range strings.Split(remediation, "\n") {
			if line != "" {
				bullet := p.style(p.theme.WarningStyle, p.theme.IconBullet)
				_, _ = fmt.Fprintf(p.writer, "%s     %s %s\n", treeLine, bullet, line)
			}
		}
	}

	// Add blank line between failures for readability
	_, _ = fmt.Fprintf(p.writer, "%s\n", treeLine)
}

// renderCheckSummary renders a beautiful summary box with borders.
func (p *Printer) renderCheckSummary(status Status, title string, items []string) {
	width := p.theme.SummaryWidth

	// Box drawing characters
	topLeft := "╭"
	topRight := "╮"
	bottomLeft := "╰"
	bottomRight := "╯"
	horizontal := "─"
	vertical := "│"

	// For minimal theme, use ASCII
	if p.theme.IconSuccess == "[OK]" {
		topLeft = "+"
		topRight = "+"
		bottomLeft = "+"
		bottomRight = "+"
		horizontal = "-"
		vertical = "|"
	}

	// Style the box based on status
	var boxStyle, iconStyled, titleStyled string
	if status == StatusSuccess {
		boxStyle = p.style(p.theme.SuccessStyle, "")
		iconStyled = p.style(p.theme.SuccessStyle, p.theme.IconSuccess)
		titleStyled = p.style(p.theme.SuccessStyle, title)
	} else {
		boxStyle = p.style(p.theme.FailureStyle, "")
		iconStyled = p.style(p.theme.FailureStyle, p.theme.IconFailure)
		titleStyled = p.style(p.theme.FailureStyle, title)
	}
	_ = boxStyle // Used for color reference

	// Build the box
	horizontalLine := strings.Repeat(horizontal, width-2)

	// Top border
	var topBorder, bottomBorder string
	if status == StatusSuccess {
		topBorder = p.style(p.theme.SuccessStyle, topLeft+horizontalLine+topRight)
		bottomBorder = p.style(p.theme.SuccessStyle, bottomLeft+horizontalLine+bottomRight)
	} else {
		topBorder = p.style(p.theme.FailureStyle, topLeft+horizontalLine+topRight)
		bottomBorder = p.style(p.theme.FailureStyle, bottomLeft+horizontalLine+bottomRight)
	}

	styledVertical := func() string {
		if status == StatusSuccess {
			return p.style(p.theme.SuccessStyle, vertical)
		}
		return p.style(p.theme.FailureStyle, vertical)
	}

	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintln(p.writer, topBorder)

	// Empty line
	_, _ = fmt.Fprintf(p.writer, "%s%s%s\n", styledVertical(), strings.Repeat(" ", width-2), styledVertical())

	// Title line centered
	titleContent := fmt.Sprintf("%s %s", iconStyled, titleStyled)
	// Calculate visible length (without ANSI codes) - approximate
	visibleLen := len(p.theme.IconSuccess) + 1 + len(title)
	padding := (width - 2 - visibleLen) / 2
	if padding < 1 {
		padding = 1
	}
	_, _ = fmt.Fprintf(p.writer, "%s%s%s%s%s\n",
		styledVertical(),
		strings.Repeat(" ", padding),
		titleContent,
		strings.Repeat(" ", width-2-padding-visibleLen),
		styledVertical())

	// Items if present
	if len(items) > 0 {
		_, _ = fmt.Fprintf(p.writer, "%s%s%s\n", styledVertical(), strings.Repeat(" ", width-2), styledVertical())
		for i, item := range items {
			var connector string
			if i == len(items)-1 {
				connector = p.theme.TreeLast
			} else {
				connector = p.theme.TreeBranch
			}
			itemIcon := p.style(p.theme.SuccessStyle, p.theme.IconSuccess)
			if status == StatusFailure {
				itemIcon = p.style(p.theme.FailureStyle, p.theme.IconFailure)
			}
			styledConnector := p.style(p.theme.TreeStyle, connector)
			itemLine := fmt.Sprintf("  %s %s %s", styledConnector, itemIcon, item)
			itemPadding := width - 2 - len(connector) - len(p.theme.IconSuccess) - len(item) - 5
			if itemPadding < 0 {
				itemPadding = 0
			}
			_, _ = fmt.Fprintf(p.writer, "%s%s%s%s\n",
				styledVertical(),
				itemLine,
				strings.Repeat(" ", itemPadding),
				styledVertical())
		}
	}

	// Empty line
	_, _ = fmt.Fprintf(p.writer, "%s%s%s\n", styledVertical(), strings.Repeat(" ", width-2), styledVertical())

	// Bottom border
	_, _ = fmt.Fprintln(p.writer, bottomBorder)
}

// renderCheckInfo renders indented info lines.
func (p *Printer) renderCheckInfo(lines []string) {
	for _, line := range lines {
		styled := p.style(p.theme.InfoStyle, line)
		_, _ = fmt.Fprintf(p.writer, "   %s\n", styled)
	}
}

// renderCheckNote renders a note.
func (p *Printer) renderCheckNote(message string) {
	note := p.style(p.theme.NoteStyle, "Note:")
	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintf(p.writer, "%s %s\n", note, message)
}

// renderCheckLine renders a single-line check result mimicking TUI output.
// Example: "format .......................... [OK] 1.451s"
// In TTY mode: does nothing (animated output handles it).
// In non-TTY mode: prints the TUI-style single line.
func (p *Printer) renderCheckLine(name string, status Status, duration time.Duration) {
	// In TTY mode, CheckHeader + CheckSuccess/CheckFailure handle the output
	if p.isTerminal {
		return
	}

	// Calculate padding for alignment (target width: 40 chars for name + dots)
	const lineWidth = 40
	nameLen := len(name)
	dotsLen := lineWidth - nameLen
	if dotsLen < 3 {
		dotsLen = 3
	}
	dots := strings.Repeat(".", dotsLen)

	// Format duration
	durationStr := formatDuration(duration)

	// Build the line
	var icon, iconStyled string
	if status == StatusSuccess {
		icon = p.theme.IconSuccess
		iconStyled = p.style(p.theme.SuccessStyle, icon)
	} else {
		icon = p.theme.IconFailure
		iconStyled = p.style(p.theme.FailureStyle, icon)
	}

	nameStyled := p.style(p.theme.InfoStyle, name)
	dotsStyled := p.style(p.theme.TreeStyle, dots)
	durationStyled := p.style(p.theme.NoteStyle, durationStr)

	_, _ = fmt.Fprintf(p.writer, "%s %s %s %s\n", nameStyled, dotsStyled, iconStyled, durationStyled)
}

// formatDuration formats a duration for display.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.3fs", d.Seconds())
}
