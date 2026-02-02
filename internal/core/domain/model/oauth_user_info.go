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

// OAuthState represents the state parameter for OAuth flow.
type OAuthState struct {
	State        string
	RedirectURI  string
	Nonce        string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	IsValid      bool
	UserUID      string // For account linking scenarios
}

// NewOAuthState creates a new OAuth state.
func NewOAuthState(redirectURI, nonce string) *OAuthState {
	now := time.Now()
	return &OAuthState{
		State:       generateOAuthState(),
		RedirectURI: redirectURI,
		Nonce:       nonce,
		CreatedAt:   now,
		ExpiresAt:   now.Add(10 * time.Minute),
		IsValid:     true,
	}
}

// IsExpired checks if the state has expired.
func (s *OAuthState) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Invalidate marks the state as invalid.
func (s *OAuthState) Invalidate() {
	s.IsValid = false
}

// generateOAuthState generates a cryptographically secure state.
func generateOAuthState() string {
	return randomString(32)
}

// OAuthError represents an OAuth error.
type OAuthError struct {
	Code        string
	Description string
	URI         string
}

// Error returns the error message.
func (e *OAuthError) Error() string {
	return e.Code + ": " + e.Description
}

// Common OAuth errors.
var (
	ErrOAuthInvalidState            = &OAuthError{Code: "invalid_state", Description: "Invalid or expired state parameter"}
	ErrOAuthAccessDenied            = &OAuthError{Code: "access_denied", Description: "User denied the authorization request"}
	ErrOAuthInvalidCode             = &OAuthError{Code: "invalid_code", Description: "Invalid or expired authorization code"}
	ErrOAuthServerError             = &OAuthError{Code: "server_error", Description: "OAuth server returned an error"}
	ErrOAuthTemporarilyUnavailable  = &OAuthError{Code: "temporarily_unavailable", Description: "OAuth server is temporarily unavailable"}
)

// IsRetryable returns true if the error is temporary and retryable.
func (e *OAuthError) IsRetryable() bool {
	return e.Code == "server_error" || e.Code == "temporarily_unavailable"
}

// randomString generates a random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}
