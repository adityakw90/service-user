package errors

var (
	ErrDeviceNotFound     = NewCustomError(80001, "device not found", nil)
	ErrDeviceRevoked      = NewCustomError(80002, "device access has been revoked", nil)
	ErrUserDeviceNotFound = NewCustomError(80003, "user-device relationship not found", nil)
)
