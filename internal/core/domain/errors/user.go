package errors

var (
	ErrInvalidUID             = NewCustomError(40001, "UID is required", nil)
	ErrInvalidUsername        = NewCustomError(40002, "username is required", nil)
	ErrInvalidEmail           = NewCustomError(40003, "email is required", nil)
	ErrInvalidPassword        = NewCustomError(40004, "password must be at least 8 characters", nil)
	ErrUserNotFound           = NewCustomError(40005, "user not found", nil)
	ErrUserAlreadyExists      = NewCustomError(40006, "user already exists", nil)
	ErrDuplicateEmail         = NewCustomError(40007, "email already exists", nil)
	ErrDuplicateUsername      = NewCustomError(40008, "username already exists", nil)
	ErrInvalidStatus          = NewCustomError(40009, "invalid status", nil)
	ErrUserDeleted            = NewCustomError(40010, "user has been deleted", nil)
	ErrUserInactive           = NewCustomError(40011, "user account is inactive", nil)
	ErrInvalidCurrentPassword = NewCustomError(40012, "current password is incorrect", nil)
	ErrPasswordMismatch       = NewCustomError(40013, "new password and confirm password do not match", nil)
)
