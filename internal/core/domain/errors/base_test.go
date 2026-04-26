package errors

import (
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
