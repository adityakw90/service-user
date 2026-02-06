package errors

import "errors"

var (
	ErrInvalidUID        = errors.New("user: UID is required")
	ErrInvalidUsername   = errors.New("user: username is required")
	ErrInvalidEmail      = errors.New("user: email is required")
	ErrInvalidPassword   = errors.New("user: password must be at least 8 characters")
	ErrUserNotFound      = errors.New("user: user not found")
	ErrUserAlreadyExists = errors.New("user: user already exists")
	ErrDuplicateEmail    = errors.New("user: email already exists")
	ErrDuplicateUsername = errors.New("user: username already exists")
	ErrInvalidStatus     = errors.New("user: invalid status")
	ErrUserDeleted       = errors.New("user: user has been deleted")
	ErrUserInactive      = errors.New("user: user account is inactive")
	ErrInvalidCurrentPassword = errors.New("user: current password is incorrect")
	ErrPasswordMismatch  = errors.New("user: new password and confirm password do not match")
)
