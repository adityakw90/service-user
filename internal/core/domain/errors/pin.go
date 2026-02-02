package errors

import "errors"

var (
	ErrPinNotSet               = errors.New("pin: PIN not set for user")
	ErrPinInvalid              = errors.New("pin: invalid PIN")
	ErrInvalidPINLength        = errors.New("pin: must be exactly 6 digits")
	ErrInvalidPINFormat        = errors.New("pin: must contain only digits")
	ErrPINTooManyAttempts      = errors.New("pin: too many failed attempts")
	ErrPINContainsCommonPattern = errors.New("pin: contains common pattern")
)
