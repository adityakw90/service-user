package domain

import "time"

// UserStatus represents the status of a user account.
type UserStatus int32

const (
	UserStatusInactive UserStatus = 0
	UserStatusActive   UserStatus = 1
	UserStatusBanned   UserStatus = -1
)

// User represents the core user entity in the domain layer.
// This is a pure domain object with no external dependencies.
type User struct {
	ID        int64
	UID       string
	Username  string
	Email     string
	Password  string // Never expose password
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// IsActive returns true if the user account is active.
func (u *User) IsActive() bool {
	return u.Status == UserStatusActive
}

// IsDeleted returns true if the user has been soft-deleted.
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// CanLogin returns true if the user can authenticate.
func (u *User) CanLogin() bool {
	return u.IsActive() && !u.IsDeleted()
}

// Activate sets the user status to active.
// Returns error if user is already banned (critical invariant).
func (u *User) Activate() error {
	if u.Status == UserStatusBanned {
		return ErrUserInactive
	}
	u.Status = UserStatusActive
	return nil
}

// Deactivate sets the user status to inactive.
func (u *User) Deactivate() {
	u.Status = UserStatusInactive
}

// Ban marks the user as banned.
func (u *User) Ban() {
	u.Status = UserStatusBanned
}
