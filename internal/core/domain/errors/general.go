package errors

const (
	// error code
	ErrCodeInternalServerError     = 10001
	ErrCodeTraceInformationMissing = 10002
	ErrCodeRequestCanceled         = 10003
	ErrCodeRequestTimeout          = 10004
	ErrCodeRequestAborted          = 10005
	ErrCodeUnimplemented           = 10006
	ErrCodeNotFound                = 10007
	ErrCodeInvalidArgument         = 10008
	ErrCodeValidation              = 10009
	ErrCodePermissionDenied        = 10010
	ErrCodeResourceConflict        = 10011
	ErrCodeInvalidEntity           = 10012

	// error message
	ErrMessageInternalServerError     = "internal server error"
	ErrMessageTraceInformationMissing = "trace information missing in request"
	ErrMessageRequestCanceled         = "request canceled"
	ErrMessageRequestTimeout          = "request timeout"
	ErrMessageRequestAborted          = "request aborted"
	ErrMessageUnimplemented           = "unimplemented"
	ErrMessageNotFound                = "resource not found"
	ErrMessageInvalidArgument         = "invalid argument"
	ErrMessageValidation              = "validation error"
	ErrMessagePermissionDenied        = "permission denied"
	ErrMessageResourceConflict        = "resource conflict"
	ErrMessageInvalidEntity           = "entity validation failed: missing required fields"
)

// general errors
var (
	ErrInternalServerError     = NewCustomError(ErrCodeInternalServerError, ErrMessageInternalServerError, nil)
	ErrTraceInformationMissing = NewCustomError(ErrCodeTraceInformationMissing, ErrMessageTraceInformationMissing, nil)
	ErrRequestCanceled         = NewCustomError(ErrCodeRequestCanceled, ErrMessageRequestCanceled, nil)
	ErrRequestTimeout          = NewCustomError(ErrCodeRequestTimeout, ErrMessageRequestTimeout, nil)
	ErrRequestAborted          = NewCustomError(ErrCodeRequestAborted, ErrMessageRequestAborted, nil)
	ErrUnimplemented           = NewCustomError(ErrCodeUnimplemented, ErrMessageUnimplemented, nil)
	ErrNotFound                = NewCustomError(ErrCodeNotFound, ErrMessageNotFound, nil)
	ErrInvalidArgument         = NewCustomError(ErrCodeInvalidArgument, ErrMessageInvalidArgument, nil)
	ErrValidation              = NewCustomError(ErrCodeValidation, ErrMessageValidation, nil)
	ErrPermissionDenied        = NewCustomError(ErrCodePermissionDenied, ErrMessagePermissionDenied, nil)
	ErrResourceConflict        = NewCustomError(ErrCodeResourceConflict, ErrMessageResourceConflict, nil)
	ErrInvalidEntity           = NewCustomError(ErrCodeInvalidEntity, ErrMessageInvalidEntity, nil)
)
