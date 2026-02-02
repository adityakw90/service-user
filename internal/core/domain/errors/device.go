package errors

import "errors"

var (
	ErrDeviceNotFound     = errors.New("device: device not found")
	ErrDeviceRevoked      = errors.New("device: device access has been revoked")
	ErrUserDeviceNotFound = errors.New("device: user-device relationship not found")
)
