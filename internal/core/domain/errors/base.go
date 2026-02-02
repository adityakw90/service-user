package errors

import "fmt"

// ErrorMap represents a dictionary-like structure for field errors.
type ErrorMap map[string][]string

// CustomError defines a structured error type with a code and message.
type CustomError struct {
	Code    int
	Message string
	Errors  ErrorMap // Stores field-specific validation errors
}

// Error implements the Go error interface for CustomError.
func (e *CustomError) Error() string {
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// NewCustomError creates a new CustomError instance.
func NewCustomError(code int, message string, errors ErrorMap) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}
