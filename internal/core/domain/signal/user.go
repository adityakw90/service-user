package signal

import "github.com/adityakw90/service-user/internal/core/domain/model"

// UserSignal represents data for user service operations.
type UserSignal struct {
	// Request context
	UID      *string // User UID being operated on
	ActorUID *string // Admin performing action (if applicable)

	// User data (populated based on operation)
	Username *string
	Email    *string
	Status   *model.UserStatus
	Active   *bool

	// Operation context
	Operation    string // "get", "list", "create", "update", "delete", "get_profile", "update_profile", "change_password", "set_pin", "list_device", "revoke_device"
	ChangesCount int    // For update operations
}
