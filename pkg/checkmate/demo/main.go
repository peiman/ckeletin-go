package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func main() {
	fmt.Println()
	fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("  LIPGLOSS STYLE SHOWCASE"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  Demonstrating what's possible with lipgloss"))
	fmt.Println()

	// 1. GRADIENT HEADERS (like the Lip Gloss logo)
	fmt.Println(sectionTitle("1. GRADIENT-STYLE HEADERS"))
	colors := []string{"#F25D94", "#FF85B8", "#FFA3CC", "#FF6B9D", "#E84A7D"}
	for i, color := range colors {
		style := lipgloss.NewStyle().
			Background(lipgloss.Color(color)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true).
			Padding(0, 2).
			MarginLeft(i * 2)
		fmt.Println(style.Render("Lip Gloss"))
	}
	fmt.Println()

	// 2. BORDERED BOXES
	fmt.Println(sectionTitle("2. BORDERED BOXES"))

	roundedBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2)
	fmt.Println(roundedBox.Render("Rounded Border\nLooks nice and modern"))

	thickBox := lipgloss.NewStyle().
		Border(lipgloss.ThickBorder()).
		BorderForeground(lipgloss.Color("208")).
		Padding(1, 2)
	fmt.Println(thickBox.Render("Thick Border\nBold and strong"))

	doubleBox := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("42")).
		Padding(1, 2)
	fmt.Println(doubleBox.Render("Double Border\nClassic style"))

	// 3. COLORED BADGES/PILLS
	fmt.Println(sectionTitle("3. BADGES & PILLS"))
	badges := []struct {
		text string
		bg   string
		fg   string
	}{
		{"SUCCESS", "#10B981", "#FFFFFF"},
		{"WARNING", "#F59E0B", "#000000"},
		{"ERROR", "#EF4444", "#FFFFFF"},
		{"INFO", "#3B82F6", "#FFFFFF"},
		{"NEW", "#8B5CF6", "#FFFFFF"},
		{"BETA", "#EC4899", "#FFFFFF"},
	}
	badgeRow := ""
	for _, b := range badges {
		badge := lipgloss.NewStyle().
			Background(lipgloss.Color(b.bg)).
			Foreground(lipgloss.Color(b.fg)).
			Bold(true).
			Padding(0, 1).
			MarginRight(1)
		badgeRow += badge.Render(b.text) + " "
	}
	fmt.Println(badgeRow)
	fmt.Println()

	// 4. STATUS INDICATORS
	fmt.Println(sectionTitle("4. STATUS INDICATORS"))
	statuses := []struct {
		icon  string
		text  string
		color string
	}{
		{"✓", "All systems operational", "42"},
		{"●", "Database connected", "42"},
		{"○", "Cache warming up...", "214"},
		{"✗", "API rate limited", "196"},
		{"◐", "Syncing data...", "39"},
	}
	for _, s := range statuses {
		icon := lipgloss.NewStyle().Foreground(lipgloss.Color(s.color)).Bold(true).Render(s.icon)
		text := lipgloss.NewStyle().Foreground(lipgloss.Color("252")).Render(s.text)
		fmt.Printf("  %s %s\n", icon, text)
	}
	fmt.Println()

	// 5. TREE STRUCTURE
	fmt.Println(sectionTitle("5. TREE STRUCTURE"))
	treeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	failStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render("  Project"))
	fmt.Printf("  %s %s %s\n", treeStyle.Render("├──"), successStyle.Render("✓"), "src/")
	fmt.Printf("  %s   %s %s %s\n", treeStyle.Render("│"), treeStyle.Render("├──"), successStyle.Render("✓"), "main.go")
	fmt.Printf("  %s   %s %s %s\n", treeStyle.Render("│"), treeStyle.Render("└──"), successStyle.Render("✓"), "utils.go")
	fmt.Printf("  %s %s %s\n", treeStyle.Render("├──"), failStyle.Render("✗"), "tests/")
	fmt.Printf("  %s   %s %s %s\n", treeStyle.Render("│"), treeStyle.Render("└──"), failStyle.Render("✗"), dimStyle.Render("broken_test.go"))
	fmt.Printf("  %s %s %s\n", treeStyle.Render("└──"), successStyle.Render("✓"), "docs/")
	fmt.Println()

	// 6. PROGRESS INDICATORS
	fmt.Println(sectionTitle("6. PROGRESS BARS"))
	progressBars := []struct {
		label   string
		percent int
		color   string
	}{
		{"Build", 100, "#10B981"},
		{"Test", 75, "#3B82F6"},
		{"Deploy", 45, "#F59E0B"},
		{"Verify", 20, "#EF4444"},
	}
	for _, p := range progressBars {
		filled := p.percent / 5
		empty := 20 - filled
		bar := lipgloss.NewStyle().Foreground(lipgloss.Color(p.color)).Render(strings.Repeat("█", filled))
		bar += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", empty))
		label := lipgloss.NewStyle().Width(8).Render(p.label)
		pct := lipgloss.NewStyle().Width(4).Align(lipgloss.Right).Render(fmt.Sprintf("%d%%", p.percent))
		fmt.Printf("  %s %s %s\n", label, bar, pct)
	}
	fmt.Println()

	// 7. STYLED PANELS
	fmt.Println(sectionTitle("7. INFORMATION PANELS"))

	infoPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Foreground(lipgloss.Color("39")).
		Padding(0, 1).
		Width(50)
	fmt.Println(infoPanel.Render("ℹ Info: This is an informational message with a nice blue border."))

	warnPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("214")).
		Foreground(lipgloss.Color("214")).
		Padding(0, 1).
		Width(50)
	fmt.Println(warnPanel.Render("⚠ Warning: Something needs your attention!"))

	errorPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("196")).
		Foreground(lipgloss.Color("196")).
		Padding(0, 1).
		Width(50)
	fmt.Println(errorPanel.Render("✗ Error: Something went wrong!"))

	successPanel := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("42")).
		Foreground(lipgloss.Color("42")).
		Padding(0, 1).
		Width(50)
	fmt.Println(successPanel.Render("✓ Success: Operation completed successfully!"))
	fmt.Println()

	// 8. TABLE-LIKE OUTPUT
	fmt.Println(sectionTitle("8. TABLE LAYOUT"))
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("212")).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	dimCell := cellStyle.Foreground(lipgloss.Color("245"))

	fmt.Printf("  %s%s%s\n",
		headerStyle.Width(15).Render("Package"),
		headerStyle.Width(10).Render("Version"),
		headerStyle.Width(12).Render("Status"))

	rows := [][]string{
		{"lipgloss", "v0.10.0", "✓ Latest"},
		{"bubbletea", "v0.25.0", "✓ Latest"},
		{"bubbles", "v0.18.0", "↑ Update"},
		{"glamour", "v0.6.0", "✓ Latest"},
	}
	for _, row := range rows {
		status := row[2]
		statusStyle := cellStyle
		if strings.Contains(status, "✓") {
			statusStyle = statusStyle.Foreground(lipgloss.Color("42"))
		} else {
			statusStyle = statusStyle.Foreground(lipgloss.Color("214"))
		}
		fmt.Printf("  %s%s%s\n",
			cellStyle.Width(15).Render(row[0]),
			dimCell.Width(10).Render(row[1]),
			statusStyle.Width(12).Render(status))
	}
	fmt.Println()

	// 9. FANCY SUMMARY BOX
	fmt.Println(sectionTitle("9. SUMMARY BOXES"))

	summarySuccess := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("42")).
		Foreground(lipgloss.Color("42")).
		Bold(true).
		Padding(1, 3).
		Width(40).
		Align(lipgloss.Center)
	fmt.Println(summarySuccess.Render("✓ ALL CHECKS PASSED\n5/5 successful"))

	summaryFail := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("196")).
		Foreground(lipgloss.Color("196")).
		Bold(true).
		Padding(1, 3).
		Width(40).
		Align(lipgloss.Center)
	fmt.Println(summaryFail.Render("✗ CHECKS FAILED\n2/5 failed"))
	fmt.Println()

	// 10. STYLED LISTS
	fmt.Println(sectionTitle("10. STYLED LISTS"))

	bullet := lipgloss.NewStyle().Foreground(lipgloss.Color("212")).Bold(true).Render("→")
	checkBullet := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true).Render("✓")
	crossBullet := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true).Render("✗")

	items := []string{"Format code", "Run linter", "Execute tests", "Build binary", "Deploy"}
	for i, item := range items {
		var b string
		switch {
		case i < 3:
			b = checkBullet
		case i == 3:
			b = crossBullet
		default:
			b = bullet
		}
		fmt.Printf("  %s %s\n", b, item)
	}
	fmt.Println()

	// 11. COLOR PALETTE DEMO
	fmt.Println(sectionTitle("11. COLOR PALETTE"))

	// Gradient row
	gradientColors := []string{"201", "200", "199", "198", "197", "196", "202", "208", "214", "220", "226", "190", "154", "118", "82", "46", "47", "48", "49", "50", "51", "45", "39", "33", "27", "21"}
	colorRow := ""
	for _, c := range gradientColors {
		colorRow += lipgloss.NewStyle().Background(lipgloss.Color(c)).Render(" ")
	}
	fmt.Printf("  %s\n", colorRow)

	// Another gradient
	blueGradient := []string{"17", "18", "19", "20", "21", "27", "33", "39", "45", "51", "87", "123", "159", "195"}
	colorRow2 := ""
	for _, c := range blueGradient {
		colorRow2 += lipgloss.NewStyle().Background(lipgloss.Color(c)).Render("  ")
	}
	fmt.Printf("  %s\n", colorRow2)
	fmt.Println()

	// 12. FINAL SHOWCASE - COMBINED STYLES
	fmt.Println(sectionTitle("12. COMBINED SHOWCASE"))

	outerBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(1, 2).
		Width(60)

	title := lipgloss.NewStyle().
		Background(lipgloss.Color("63")).
		Foreground(lipgloss.Color("230")).
		Bold(true).
		Padding(0, 2).
		MarginBottom(1).
		Render("Code Quality Report")

	content := title + "\n\n"
	content += fmt.Sprintf("  %s %s\n", successStyle.Render("✓"), "Formatting passed")
	content += fmt.Sprintf("  %s %s\n", successStyle.Render("✓"), "Linting passed")
	content += fmt.Sprintf("  %s %s\n", successStyle.Render("✓"), "Tests passed (156/156)")
	content += fmt.Sprintf("  %s %s\n", successStyle.Render("✓"), "Coverage: 87.3%%")
	content += "\n"

	// Mini progress bar
	coverageBar := lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(strings.Repeat("█", 17))
	coverageBar += lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render(strings.Repeat("░", 3))
	content += fmt.Sprintf("  Coverage: %s 87%%", coverageBar)

	fmt.Println(outerBox.Render(content))
	fmt.Println()

	// Footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Italic(true).
		Render("  Made with lipgloss - github.com/charmbracelet/lipgloss")
	fmt.Println(footer)
	fmt.Println()
}

func sectionTitle(title string) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(true).
		MarginTop(1).
		MarginBottom(1)
	underline := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Render(strings.Repeat("─", len(title)))
	return style.Render("  "+title) + "\n  " + underline
}
