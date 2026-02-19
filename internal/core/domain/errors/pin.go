package errors

var (
	ErrPinNotSet          = NewCustomError(60001, "PIN not set for user", nil)
	ErrPinInvalid         = NewCustomError(60002, "invalid PIN", nil)
	ErrPINTooManyAttempts = NewCustomError(60003, "too many failed attempts", nil)
)
