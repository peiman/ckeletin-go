package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError(t *testing.T) {
	err := fmt.Errorf("test error")
	appErr := NewAppError(ErrConfigNotFound, "Config not found", err)

	assert.Equal(t, ErrConfigNotFound, appErr.Code)
	assert.Equal(t, "Config not found", appErr.Message)
	assert.Equal(t, err, appErr.Err)

	expectedStr := "CONFIG_NOT_FOUND: Config not found (test error)"
	assert.Equal(t, expectedStr, appErr.Error())

	appErrNoInner := NewAppError(ErrInvalidConfig, "Invalid config", nil)
	expectedStrNoInner := "INVALID_CONFIG: Invalid config"
	assert.Equal(t, expectedStrNoInner, appErrNoInner.Error())
}
