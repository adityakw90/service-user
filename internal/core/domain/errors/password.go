package errors

import "errors"

var (
	ErrPasswordTooShort    = errors.New("password: must be at least 12 characters")
	ErrPasswordNoUppercase = errors.New("password: must contain uppercase letter")
	ErrPasswordNoLowercase = errors.New("password: must contain lowercase letter")
	ErrPasswordNoDigit     = errors.New("password: must contain a number")
	ErrPasswordNoSpecial   = errors.New("password: must contain special character")
	ErrPasswordWeakPattern = errors.New("password: contains weak pattern")
)
