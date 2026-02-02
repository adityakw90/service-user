package errors

import "errors"

var (
	ErrPinNotSet          = errors.New("pin: PIN not set for user")
	ErrPinInvalid         = errors.New("pin: invalid PIN")
	ErrPINTooManyAttempts = errors.New("pin: too many failed attempts")
)
