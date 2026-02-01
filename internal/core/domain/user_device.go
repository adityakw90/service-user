package domain

import "time"

// UserDevice represents the relationship between a user and a device.
// This tracks device usage including IP, activity, and revocation status.
type UserDevice struct {
	UserID       int64
	DeviceID     int64
	IPAddress    string
	LastActiveAt time.Time
	RevokedAt    *time.Time
	CreatedAt    time.Time
}

// IsRevoked returns true if access from this device has been revoked.
func (ud *UserDevice) IsRevoked() bool {
	return ud.RevokedAt != nil
}

// IsActive returns true if the device access is not revoked.
func (ud *UserDevice) IsActive() bool {
	return !ud.IsRevoked()
}

// Revoke marks this device access as revoked.
func (ud *UserDevice) Revoke() {
	now := time.Now().UTC()
	ud.RevokedAt = &now
}

// Touch updates the last_active_at timestamp.
func (ud *UserDevice) Touch() {
	ud.LastActiveAt = time.Now().UTC()
}
