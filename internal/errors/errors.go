// Package errors provides custom error types and handling for the application.
package errors

import "fmt"

// AppError represents a custom application error.
type AppError struct {
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new AppError.
func NewAppError(code, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// Some predefined error codes.
const (
	ErrConfigNotFound  = "CONFIG_NOT_FOUND"
	ErrInvalidConfig   = "INVALID_CONFIG"
	ErrDatabaseConnect = "DATABASE_CONNECT_ERROR"
	// Add more error codes as needed.
)
