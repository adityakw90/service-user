package event

// EventDeviceCreatedData is emitted when a device is created.
type EventDeviceCreatedData struct {
	UserUID string
}

// EventDeviceUpdatedData is emitted when a device is updated.
type EventDeviceUpdatedData struct {
	UserUID string
}

// EventDeviceDeletedData is emitted when a device is deleted.
type EventDeviceDeletedData struct {
	UserUID string
}

// EventDeviceRevokeData is emitted when a device is revoked.
type EventDeviceRevokeData struct {
	UserUID string
}
