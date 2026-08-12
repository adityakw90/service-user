package event

import (
	"strings"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// EntityType identifies the kind of business resource affected by an event.
type EntityType string

const (
	EntityTypeUser        EntityType = "user"
	EntityTypeDevice      EntityType = "device"
	EntityTypeUserFile    EntityType = "user_file"
	EntityTypeUserProfile EntityType = "user_profile"
	EntityTypeUserPIN     EntityType = "user_pin"
	EntityTypeSession     EntityType = "session"
)

// Entity identifies the business resource affected by an event.
type Entity struct {
	ID   string
	Type EntityType
	Name *string
}

func (e *Entity) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return domainError.ErrEntityIDRequired
	}
	if strings.TrimSpace(string(e.Type)) == "" {
		return domainError.ErrEntityTypeRequired
	}

	if err := e.checkType(); err != nil {
		return err
	}

	return nil
}

func (e *Entity) checkType() error {
	switch e.Type {
	case EntityTypeUser,
		EntityTypeDevice,
		EntityTypeUserFile,
		EntityTypeUserProfile,
		EntityTypeUserPIN,
		EntityTypeSession:
		return nil
	default:
		return domainError.ErrEntityInvalidType
	}
}

// NewUserEntity identifies a user affected by an event.
func NewUserEntity(user *model.User) Entity {
	return Entity{ID: user.UID, Type: EntityTypeUser, Name: &user.Username}
}

// NewDeviceEntity identifies a device affected by an event.
func NewDeviceEntity(device *model.Device) Entity {
	return Entity{ID: device.UID, Type: EntityTypeDevice, Name: &device.DeviceName}
}

// NewUserFileEntity identifies a user file affected by an event.
func NewUserFileEntity(file *model.UserFile) Entity {
	return Entity{ID: file.UID, Type: EntityTypeUserFile, Name: &file.FileName}
}

// NewUserProfileEntity identifies a user profile affected by an event.
func NewUserProfileEntity(profile *model.UserProfile) Entity {
	name := profile.FullName()
	return Entity{ID: profile.UserUID, Type: EntityTypeUserProfile, Name: &name}
}

// NewUserPinEntity identifies a user PIN affected by an event.
func NewUserPinEntity(pin *model.UserPin) Entity {
	return Entity{ID: pin.UserUID, Type: EntityTypeUserPIN}
}

// NewTokenEntity identifies the session affected by a token event.
func NewTokenEntity(claims *model.TokenClaims) Entity {
	return Entity{ID: claims.Sid, Type: EntityTypeSession}
}
