package model

import "time"

// UserPin represents a user's PIN entity.
type UserPin struct {
	UserID    int64
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsSet returns true if a PIN has been set for this user.
func (p *UserPin) IsSet() bool {
	return p.Code != ""
}
