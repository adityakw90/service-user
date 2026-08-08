package event

// EventUserCreatedData is emitted when a user is created.
type EventUserCreatedData struct {
	UserUID  string
	ActorUID string
	Username string
	Email    string
	Status   string
}

// EventUserUpdatedData is emitted when a user is updated.
type EventUserUpdatedData struct {
	UserUID      string
	ActorUID     string
	ChangesCount int
}

// EventUserDeletedData is emitted when a user is deleted.
type EventUserDeletedData struct {
	UserUID  string
	ActorUID string
}

// EventUserUpdatePasswordData is emitted when a user's password is updated.
type EventUserUpdatePasswordData struct {
	UserUID  string
	ActorUID string
}

// EventUserCreatePinData is emitted when a user's PIN is created.
type EventUserCreatePinData struct {
	UserUID  string
	ActorUID string
}

// EventUserUpdatePinData is emitted when a user's PIN is updated.
type EventUserUpdatePinData struct {
	UserUID  string
	ActorUID string
}

// EventUserUpdateProfileData is emitted when a user's profile is updated.
type EventUserUpdateProfileData struct {
	UserUID  string
	ActorUID string
}

// EventUserRevokeDeviceData is emitted when a user's device is revoked.
type EventUserRevokeDeviceData struct {
	UserUID   string
	ActorUID  string
	DeviceUID string
}
