package event

// EventUserCreatedData is emitted when a user is created.
type EventUserCreatedData struct {
	Username string
	Email    string
	Status   string
}

// EventUserUpdatedData is emitted when a user is updated.
type EventUserUpdatedData struct {
	ChangesCount int
}

// EventUserDeletedData is emitted when a user is deleted.
type EventUserDeletedData struct {
}

// EventUserUpdatePasswordData is emitted when a user's password is updated.
type EventUserUpdatePasswordData struct {
}

// EventUserCreatePinData is emitted when a user's PIN is created.
type EventUserCreatePinData struct {
}

// EventUserUpdatePinData is emitted when a user's PIN is updated.
type EventUserUpdatePinData struct {
}

// EventUserUpdateProfileData is emitted when a user's profile is updated.
type EventUserUpdateProfileData struct {
}

// EventUserRevokeDeviceData is emitted when a user's device is revoked.
type EventUserRevokeDeviceData struct {
	UserUID string
}
