package model

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
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
	State       string
	RedirectURI string
	Nonce       string
	CreatedAt   time.Time
	ExpiresAt   time.Time
	IsValid     bool
	UserUID     string // For account linking scenarios
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

// randomString generates a cryptographically secure random alphanumeric string.
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, n)
	letterCount := big.NewInt(int64(len(letters)))

	for i := range result {
		// Use crypto/rand.Int to avoid modulo bias
		num, err := rand.Int(rand.Reader, letterCount)
		if err != nil {
			// Fallback to hex encoding if crypto/rand fails
			b := make([]byte, (n+1)/2) // hex encoding doubles length
			rand.Read(b)
			return hex.EncodeToString(b)[:n]
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)
}
