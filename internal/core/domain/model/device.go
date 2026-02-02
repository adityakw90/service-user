package model

import "time"

// Device represents a user device.
// This is the device identity table with no tracking columns.
type Device struct {
	ID                int64
	UID               string
	DeviceFingerprint string
	DeviceName        string
	CreatedAt         time.Time
}
