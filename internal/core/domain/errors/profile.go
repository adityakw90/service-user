package errors

var (
	ErrProfileNotFound = NewCustomError(50001, "profile not found", nil)
)
