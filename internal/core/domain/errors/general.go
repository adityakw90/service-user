package errors

// general errors
var (
	ErrInternalServerError     = NewCustomError(10001, "internal server error", nil)
	ErrTraceInformationMissing = NewCustomError(10002, "trace information missing in request", nil)
	ErrRequestCanceled         = NewCustomError(10003, "request canceled", nil)
	ErrRequestTimeout          = NewCustomError(10004, "request timeout", nil)
	ErrRequestAborted          = NewCustomError(10005, "request aborted", nil)
	ErrUnimplemented           = NewCustomError(10006, "unimplemented", nil)
	ErrNotFound                = NewCustomError(10007, "resource not found", nil)
	ErrInvalidArgument         = NewCustomError(10008, "invalid argument", nil)
	ErrValidation              = NewCustomError(10009, "validation error", nil)
	ErrPermissionDenied        = NewCustomError(10010, "permission denied", nil)
	ErrResourceConflict        = NewCustomError(10011, "resource conflict", nil)
)
