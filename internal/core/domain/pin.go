package domain

import "time"

// UserPin represents a hashed PIN for security.
type UserPin struct {
	UserID    int64
	Code      string // Hashed PIN, never expose
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsSet returns true if a PIN has been set for this user.
func (p *UserPin) IsSet() bool {
	return p.Code != ""
}
