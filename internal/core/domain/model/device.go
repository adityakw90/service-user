package model

import (
	"time"
)

// Device represents a device.
type Device struct {
	ID                int64
	UID               string
	DeviceFingerprint string
	DeviceName        string
	CreatedAt         time.Time
}

// Devices contains the list of devices and metadata for pagination.
type Devices struct {
	Items []Device
	Meta  *Meta
}
