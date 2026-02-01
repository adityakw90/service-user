package domain

import "time"

// UserProfile represents extended user information.
// 1:1 relationship with User.
type UserProfile struct {
	UserID       int64
	FirstName    string
	LastName     string
	Bio          string
	AvatarFileID *int64
	Attributes   map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// HasAvatar returns true if the profile has an avatar set.
func (p *UserProfile) HasAvatar() bool {
	return p.AvatarFileID != nil && *p.AvatarFileID != 0
}

// FullName returns the user's full name.
func (p *UserProfile) FullName() string {
	switch {
	case p.FirstName == "" && p.LastName == "":
		return ""
	case p.FirstName == "":
		return p.LastName
	case p.LastName == "":
		return p.FirstName
	default:
		return p.FirstName + " " + p.LastName
	}
}
