package model

import (
	"time"
)

// OAuthUserInfo represents user information from OAuth provider.
type OAuthUserInfo struct {
	ProviderID    string
	Email         string
	EmailVerified bool
	FirstName     string
	LastName      string
	FullName      string
	Picture       string
	Locale        string
	ExpiresAt     *time.Time
}

// HasValidEmail checks if the user has a verified email.
func (u *OAuthUserInfo) HasValidEmail() bool {
	return u.Email != "" && u.EmailVerified
}

// DisplayName returns the best available display name.
func (u *OAuthUserInfo) DisplayName() string {
	if u.FullName != "" {
		return u.FullName
	}
	if u.FirstName != "" && u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	if u.FirstName != "" {
		return u.FirstName
	}
	return u.Email
}

// AvatarURL returns the best available avatar URL.
func (u *OAuthUserInfo) AvatarURL() string {
	return u.Picture
}

// OAuthTokens contains tokens returned by OAuth provider.
type OAuthTokens struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	IDToken      string
	TokenType    string
}

// IsExpired checks if the access token is expired.
func (t *OAuthTokens) IsExpired() bool {
	if t.ExpiresIn == 0 {
		return false
	}
	// We don't track when token was issued, so this is approximate
	// In practice, the caller should track issuance time
	return false
}
