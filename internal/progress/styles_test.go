package progress

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultStyle(t *testing.T) {
	style := DefaultStyle()

	// ASSERTION
	assert.NotEmpty(t, style.SpinnerFrames, "SpinnerFrames should not be empty")
	assert.Greater(t, style.SpinnerInterval.Milliseconds(), int64(0), "SpinnerInterval should be positive")
	assert.Greater(t, style.BarWidth, 0, "BarWidth should be positive")
	assert.NotEmpty(t, style.BarChar, "BarChar should not be empty")
	assert.NotEmpty(t, style.BarEmptyChar, "BarEmptyChar should not be empty")
}

func TestMinimalStyle(t *testing.T) {
	style := MinimalStyle()

	// ASSERTION
	assert.NotEmpty(t, style.SpinnerFrames, "SpinnerFrames should not be empty")
	assert.Greater(t, style.SpinnerInterval.Milliseconds(), int64(0), "SpinnerInterval should be positive")
	assert.Greater(t, style.BarWidth, 0, "BarWidth should be positive")
	assert.NotEmpty(t, style.BarChar, "BarChar should not be empty")
	assert.NotEmpty(t, style.BarEmptyChar, "BarEmptyChar should not be empty")
}

func TestDefaultStyle_HasAllFields(t *testing.T) {
	style := DefaultStyle()

	// Test all style fields are initialized
	assert.NotNil(t, style.SpinnerStyle)
	assert.NotNil(t, style.BarStyle)
	assert.NotNil(t, style.BarEmptyStyle)
	assert.NotNil(t, style.SuccessStyle)
	assert.NotNil(t, style.ErrorStyle)
	assert.NotNil(t, style.WarningStyle)
	assert.NotNil(t, style.PhaseStyle)
	assert.NotNil(t, style.TaskStyle)
	assert.NotNil(t, style.MessageStyle)
	assert.NotNil(t, style.CounterStyle)
}
