package signal

import "github.com/adityakw90/service-user/internal/core/domain/model"

type AuthSignal struct {
	Identifier        string
	IdentifierType    string
	DeviceFingerprint *string
	DeviceIP          *string
	DeviceName        *string
	Extra             *map[string]any
	UID               *string
	Username          *string
	Email             *string
	Status            *model.UserStatus
	Active            *bool
	Deleted           *bool
}
