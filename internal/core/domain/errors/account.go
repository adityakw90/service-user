package errors

import "errors"

var (
	ErrAccountLockedOut = errors.New("account: account is locked out")
	ErrTooManyAttempts  = errors.New("auth: too many failed attempts")
)
