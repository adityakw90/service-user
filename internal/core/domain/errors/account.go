package errors

var (
	ErrAccountLockedOut = NewCustomError(30001, "account is locked out", nil)
)
