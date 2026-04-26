package event

// EventType defines the type of authentication domain event.
type EventType string

const (
	// Auth Events
	EventLogin        EventType = "auth.login"
	EventLoginFailed  EventType = "auth.login_failed"
	EventLoginLocked  EventType = "auth.login_locked"
	EventTokenRefresh EventType = "auth.token_refresh"
	EventRevokeToken  EventType = "auth.revoke_token"
	EventPINVerify    EventType = "auth.pin_verify"
	EventPINFail      EventType = "auth.pin_fail"

	// User Events
	EventUserCreated        EventType = "user.created"
	EventUserUpdated        EventType = "user.updated"
	EventUserDeleted        EventType = "user.deleted"
	EventUserUpdatePassword EventType = "user.update_password"
	EventUserCreatePin      EventType = "user.create_pin"
	EventUserUpdatePin      EventType = "user.update_pin"
	EventUserUpdateProfile  EventType = "user.update_profile"
	EventUserRevokeDevice   EventType = "user.revoke_device"

	// File Events
	EventUserFileCreated EventType = "user_file.created"
	EventUserFileUpdated EventType = "user_file.updated"
	EventUserFileDeleted EventType = "user_file.deleted"

	// Device Events
	EventDeviceCreated EventType = "device.created"
	EventDeviceUpdated EventType = "device.updated"
	EventDeviceDeleted EventType = "device.deleted"
)
