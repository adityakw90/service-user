package errors

var (
	ErrFileNotFound = NewCustomError(70001, "file not found", nil)
)
