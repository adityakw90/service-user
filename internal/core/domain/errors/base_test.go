package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCore_Domain_CustomError(t *testing.T) {
	// table driven test for type custom error
	tests := []struct {
		name      string
		code      int
		message   string
		errors    ErrorMap
		wantError string
	}{
		{
			name:      "success",
			code:      1,
			message:   "success",
			errors:    nil,
			wantError: "[1] success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CustomError{
				Code:    tt.code,
				Message: tt.message,
				Errors:  tt.errors,
			}
			assert.Equal(t, tt.wantError, err.Error())
		})
	}
}

func TestCore_Domain_NewCustomError(t *testing.T) {
	// table driven test for func NewCustomError
	tests := []struct {
		name       string
		code       int
		message    string
		errors     ErrorMap
		wantObject *CustomError
		wantError  string
	}{
		{
			name:    "success",
			code:    1,
			message: "success",
			errors:  nil,
			wantObject: &CustomError{
				Code:    1,
				Message: "success",
				Errors:  nil,
			},
			wantError: "[1] success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewCustomError(tt.code, tt.message, tt.errors)
			assert.Equal(t, tt.wantError, err.Error())
			assert.Equal(t, tt.wantObject, err)
		})
	}
}

func TestCore_Domain_CustomError_Cause(t *testing.T) {
	causeErr := errors.New("underlying DB error")
	baseErr := NewCustomError(999, "something went wrong", nil)
	otherErr := NewCustomError(888, "different error", nil)

	tests := []struct {
		name          string
		err           *CustomError
		cause         error
		wantCode      int
		wantMessage   string
		wantErrorMsg  string
		checkIsTarget error
		wantIsResult  bool
	}{
		{
			name:          "Without Cause",
			err:           baseErr,
			cause:         nil,
			wantCode:      999,
			wantMessage:   "something went wrong",
			wantErrorMsg:  "[999] something went wrong",
			checkIsTarget: baseErr,
			wantIsResult:  true,
		},
		{
			name:          "With Cause",
			err:           baseErr.WithCause(causeErr),
			cause:         causeErr,
			wantCode:      999,
			wantMessage:   "something went wrong",
			wantErrorMsg:  "[999] something went wrong: underlying DB error",
			checkIsTarget: baseErr,
			wantIsResult:  true,
		},
		{
			name:          "With Cause check against self",
			err:           baseErr.WithCause(causeErr),
			cause:         causeErr,
			wantCode:      999,
			wantMessage:   "something went wrong",
			wantErrorMsg:  "[999] something went wrong: underlying DB error",
			checkIsTarget: baseErr.WithCause(causeErr),
			wantIsResult:  true,
		},
		{
			name:          "Check against different error",
			err:           baseErr.WithCause(causeErr),
			cause:         causeErr,
			wantCode:      999,
			wantMessage:   "something went wrong",
			wantErrorMsg:  "[999] something went wrong: underlying DB error",
			checkIsTarget: otherErr,
			wantIsResult:  false,
		},
		{
			name:          "Check against standard error",
			err:           baseErr.WithCause(causeErr),
			cause:         causeErr,
			wantCode:      999,
			wantMessage:   "something went wrong",
			wantErrorMsg:  "[999] something went wrong: underlying DB error",
			checkIsTarget: errors.New("other error"),
			wantIsResult:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantCode, tt.err.Code)
			assert.Equal(t, tt.wantMessage, tt.err.Message)
			assert.Equal(t, tt.cause, tt.err.Cause)
			assert.Equal(t, tt.cause, tt.err.Unwrap())
			assert.Equal(t, tt.wantErrorMsg, tt.err.Error())

			if tt.cause != nil {
				assert.True(t, errors.Is(tt.err, tt.cause))
			}

			assert.Equal(t, tt.wantIsResult, errors.Is(tt.err, tt.checkIsTarget))
		})
	}
}

func TestCore_Domain_CustomError_ErrorIsAndAs(t *testing.T) {
	defaultErr := NewCustomError(100, "default error", nil)
	tests := []struct {
		name         string
		baseErr      func() error
		checkErr     func() error
		wantIsResult bool
		wantAsResult bool
	}{
		{
			name:         "same error, same cause",
			baseErr:      func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr:     func() error { return NewCustomError(999, "something went wrong", nil) },
			wantIsResult: true,
			wantAsResult: true,
		},
		{
			name:         "same error, different message",
			baseErr:      func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr:     func() error { return NewCustomError(999, "different message", nil) },
			wantIsResult: true,
			wantAsResult: true,
		},
		{
			name:    "same error, different cause",
			baseErr: func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr: func() error {
				return NewCustomError(999, "something went wrong", nil).WithCause(errors.New("different cause"))
			},
			wantIsResult: true,
			wantAsResult: true,
		},
		{
			name:         "different error",
			baseErr:      func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr:     func() error { return NewCustomError(888, "different error", nil) },
			wantIsResult: false,
			wantAsResult: true,
		},
		{
			name:         "standard error",
			baseErr:      func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr:     func() error { return errors.New("different error") },
			wantIsResult: false,
			wantAsResult: true,
		},
		{
			name:         "nil error",
			baseErr:      func() error { return NewCustomError(999, "something went wrong", nil) },
			checkErr:     func() error { return nil },
			wantIsResult: false,
			wantAsResult: true,
		},
		{
			name:    "same error, different cause",
			baseErr: func() error { return defaultErr },
			checkErr: func() error {
				return defaultErr.WithCause(errors.New("different cause"))
			},
			wantIsResult: true,
			wantAsResult: true,
		},
		{
			name:         "standard error has no CustomError",
			baseErr:      func() error { return errors.New("some standard error") },
			checkErr:     func() error { return NewCustomError(999, "something", nil) },
			wantIsResult: false,
			wantAsResult: false,
		},
		{
			name:         "nil base error",
			baseErr:      func() error { return nil },
			checkErr:     func() error { return NewCustomError(999, "something", nil) },
			wantIsResult: false,
			wantAsResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseErr := tt.baseErr()
			checkErr := tt.checkErr()

			// test errors.Is
			result := errors.Is(baseErr, checkErr)
			assert.Equal(t,
				tt.wantIsResult,
				result,
				"'errors.Is' %s should be %t",
				tt.name,
				tt.wantIsResult,
			)

			// test errors.As
			var customErr *CustomError
			result = errors.As(baseErr, &customErr)
			if tt.wantAsResult {
				assert.True(t, result)
				assert.Equal(t, baseErr, customErr)
			} else {
				assert.False(t, result)
				assert.Nil(t, customErr)
			}
		})
	}
}
