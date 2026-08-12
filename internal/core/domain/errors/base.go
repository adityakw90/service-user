package errors

import "fmt"

// ErrorMap represents a dictionary-like structure for field errors.
type ErrorMap map[string][]string

// CustomError defines a structured error type with a code and message.
type CustomError struct {
	Code    int
	Message string
	Errors  ErrorMap // Stores field-specific validation errors
	Cause   error    // Field to store the underlying cause error
}

// Error implements the Go error interface for CustomError.
func (e *CustomError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause of the error.
func (e *CustomError) Unwrap() error {
	return e.Cause
}

// Is checks if the target error is a CustomError with the same error code.
func (e *CustomError) Is(target error) bool {
	if t, ok := target.(*CustomError); ok {
		return e.Code == t.Code
	}
	return false
}

// WithCause creates a new CustomError copy with the given cause.
func (e *CustomError) WithCause(cause error) *CustomError {
	return &CustomError{
		Code:    e.Code,
		Message: e.Message,
		Errors:  e.Errors,
		Cause:   cause,
	}
}

// NewCustomError creates a new CustomError instance.
func NewCustomError(code int, message string, errors ErrorMap) *CustomError {
	return &CustomError{
		Code:    code,
		Message: message,
		Errors:  errors,
	}
}
