package checkmate

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultTheme(t *testing.T) {
	theme := DefaultTheme()
	require.NotNil(t, theme)

	// Check lipgloss-style icons
	assert.Equal(t, "○", theme.IconPending)
	assert.Equal(t, "✓", theme.IconSuccess)
	assert.Equal(t, "✗", theme.IconFailure)
	assert.Equal(t, "•", theme.IconBullet)
	assert.Equal(t, "!", theme.IconWarning)

	// Check tree connectors
	assert.Equal(t, "├──", theme.TreeBranch)
	assert.Equal(t, "└──", theme.TreeLast)
	assert.Equal(t, "│", theme.TreeLine)

	// Check separators
	assert.Equal(t, "─", theme.CategoryChar)
	assert.Equal(t, "─", theme.SummaryChar)

	// Check widths
	assert.Equal(t, 50, theme.CategoryWidth)
	assert.Equal(t, 50, theme.SummaryWidth)

	// Check ForceColors default
	assert.False(t, theme.ForceColors)
}

func TestMinimalTheme(t *testing.T) {
	theme := MinimalTheme()
	require.NotNil(t, theme)

	// Check icons are ASCII
	assert.Equal(t, "[-]", theme.IconPending)
	assert.Equal(t, "[OK]", theme.IconSuccess)
	assert.Equal(t, "[FAIL]", theme.IconFailure)
	assert.Equal(t, "*", theme.IconBullet)
	assert.Equal(t, "[WARN]", theme.IconWarning)

	// Check tree connectors are ASCII
	assert.Equal(t, "|--", theme.TreeBranch)
	assert.Equal(t, "`--", theme.TreeLast)
	assert.Equal(t, "|", theme.TreeLine)

	// Check separators are ASCII
	assert.Equal(t, "-", theme.CategoryChar)
	assert.Equal(t, "=", theme.SummaryChar)

	// Check widths match default
	assert.Equal(t, 48, theme.CategoryWidth)
	assert.Equal(t, 45, theme.SummaryWidth)

	// Check ForceColors default
	assert.False(t, theme.ForceColors)
}

func TestCITheme(t *testing.T) {
	ciTheme := CITheme()
	minimalTheme := MinimalTheme()

	// CITheme should be equivalent to MinimalTheme
	assert.Equal(t, minimalTheme.IconPending, ciTheme.IconPending)
	assert.Equal(t, minimalTheme.IconSuccess, ciTheme.IconSuccess)
	assert.Equal(t, minimalTheme.IconFailure, ciTheme.IconFailure)
	assert.Equal(t, minimalTheme.CategoryChar, ciTheme.CategoryChar)
	assert.Equal(t, minimalTheme.SummaryChar, ciTheme.SummaryChar)
}

func TestTheme_Styles(t *testing.T) {
	// Verify styles can render content
	theme := DefaultTheme()

	// Styles should be able to render content
	rendered := theme.SuccessStyle.Render("test")
	assert.NotEmpty(t, rendered)
	assert.Contains(t, rendered, "test")

	rendered = theme.FailureStyle.Render("error")
	assert.NotEmpty(t, rendered)
	assert.Contains(t, rendered, "error")
}

func TestTheme_ForceColors(t *testing.T) {
	theme := DefaultTheme()
	theme.ForceColors = true

	assert.True(t, theme.ForceColors)
}
