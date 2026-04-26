package model

import "time"

// UserPin represents a user's PIN entity.
type UserPin struct {
	UserID    int64
	UserUID   string
	Code      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsSet returns true if a PIN has been set for this user.
func (p *UserPin) IsSet() bool {
	return p.Code != ""
}

type UserPins struct {
	Items []UserPin
	Meta  Meta
}
