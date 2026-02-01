package domain

import "errors"

// Domain errors for user operations.
var (
	// Validation errors
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

	// Profile errors
	ErrProfileNotFound = errors.New("profile: profile not found")

	// Device errors
	ErrDeviceNotFound = errors.New("device: device not found")
	ErrDeviceRevoked  = errors.New("device: device access has been revoked")

	// Pin errors
	ErrPinNotSet        = errors.New("pin: PIN not set for user")
	ErrPinInvalid       = errors.New("pin: invalid PIN")
	ErrPinAlreadyExists = errors.New("pin: PIN already exists for user")

	// File errors
	ErrFileNotFound = errors.New("file: file not found")
)
