package event

// EventUserCreatedData is emitted when a user is created.
type EventUserCreatedData struct {
	ActorUID string
	Username string
	Email    string
	Status   string
}

// EventUserUpdatedData is emitted when a user is updated.
type EventUserUpdatedData struct {
	ActorUID     string
	ChangesCount int
}

// EventUserDeletedData is emitted when a user is deleted.
type EventUserDeletedData struct {
	ActorUID string
}

// EventUserUpdatePasswordData is emitted when a user's password is updated.
type EventUserUpdatePasswordData struct {
	ActorUID string
}

// EventUserCreatePinData is emitted when a user's PIN is created.
type EventUserCreatePinData struct {
	ActorUID string
}

// EventUserUpdatePinData is emitted when a user's PIN is updated.
type EventUserUpdatePinData struct {
	ActorUID string
}

// EventUserUpdateProfileData is emitted when a user's profile is updated.
type EventUserUpdateProfileData struct {
	ActorUID string
}

// EventUserRevokeDeviceData is emitted when a user's device is revoked.
type EventUserRevokeDeviceData struct {
	UserUID  string
	ActorUID string
}
