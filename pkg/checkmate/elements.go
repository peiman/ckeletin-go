package checkmate

import (
	"fmt"
	"strings"
)

// renderCategoryHeader renders a category header line.
// Example: "─── Code Quality ────────────────────────"
func (p *Printer) renderCategoryHeader(title string) {
	// Calculate separator length
	// Format: "─── Title ────────────────────"
	// Where prefix is 3 chars + space, and suffix fills to CategoryWidth
	titleLen := len(title)
	separatorLen := p.theme.CategoryWidth - titleLen - 5 // 5 = "─── " prefix + " " after title
	if separatorLen < 3 {
		separatorLen = 3
	}

	prefix := strings.Repeat(p.theme.CategoryChar, 3)
	suffix := strings.Repeat(p.theme.CategoryChar, separatorLen)

	line := fmt.Sprintf("%s %s %s", prefix, title, suffix)
	line = p.theme.CategoryStyle.Render(line)

	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintln(p.writer, line)
}

// renderCheckHeader renders a check-in-progress message.
// Example: "🔍 Checking formatting..."
func (p *Printer) renderCheckHeader(message string) {
	_, _ = fmt.Fprintf(p.writer, "%s %s...\n", p.theme.IconSearch, message)
}

// renderCheckSuccess renders a success message.
// Example: "✅ All files properly formatted"
func (p *Printer) renderCheckSuccess(message string) {
	icon := p.theme.SuccessStyle.Render(p.theme.IconSuccess)
	_, _ = fmt.Fprintf(p.writer, "%s %s\n", icon, message)
}

// renderCheckFailure renders a failure with details and remediation.
func (p *Printer) renderCheckFailure(title, details, remediation string) {
	icon := p.theme.FailureStyle.Render(p.theme.IconFailure)

	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintf(p.writer, "%s %s\n", icon, title)

	if details != "" {
		_, _ = fmt.Fprintln(p.writer)
		_, _ = fmt.Fprintln(p.writer, "Details:")
		// Indent each line of details
		for _, line := range strings.Split(details, "\n") {
			_, _ = fmt.Fprintf(p.writer, "  %s\n", line)
		}
	}

	if remediation != "" {
		_, _ = fmt.Fprintln(p.writer)
		_, _ = fmt.Fprintln(p.writer, "How to fix:")
		// Each remediation line gets a bullet
		for _, line := range strings.Split(remediation, "\n") {
			if line != "" {
				_, _ = fmt.Fprintf(p.writer, "  %s %s\n", p.theme.IconBullet, line)
			}
		}
	}

	_, _ = fmt.Fprintln(p.writer)
}

// renderCheckSummary renders a summary box.
func (p *Printer) renderCheckSummary(status Status, title string, items []string) {
	separator := strings.Repeat(p.theme.SummaryChar, p.theme.SummaryWidth)

	var icon string
	if status == StatusSuccess {
		icon = p.theme.SuccessStyle.Render(p.theme.IconSuccess)
	} else {
		icon = p.theme.FailureStyle.Render(p.theme.IconFailure)
	}

	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintln(p.writer, separator)
	_, _ = fmt.Fprintf(p.writer, "%s %s\n", icon, title)

	if len(items) > 0 {
		_, _ = fmt.Fprintln(p.writer)
		for _, item := range items {
			_, _ = fmt.Fprintln(p.writer, item)
		}
	}

	_, _ = fmt.Fprintln(p.writer, separator)
}

// renderCheckInfo renders indented info lines.
func (p *Printer) renderCheckInfo(lines []string) {
	for _, line := range lines {
		styled := p.theme.InfoStyle.Render(line)
		_, _ = fmt.Fprintf(p.writer, "   %s\n", styled)
	}
}

// renderCheckNote renders a note.
func (p *Printer) renderCheckNote(message string) {
	note := p.theme.NoteStyle.Render("Note:")
	_, _ = fmt.Fprintln(p.writer)
	_, _ = fmt.Fprintf(p.writer, "%s %s\n", note, message)
}
