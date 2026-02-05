package errors

import "errors"

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

// IsRetryable returns true if the error is temporary and retryable.
func (e *OAuthError) IsRetryable() bool {
	return e.Code == "server_error" || e.Code == "temporarily_unavailable"
}

var (
	// token errors
	ErrTokenRevoked      = errors.New("token: token has been revoked")
	ErrTokenBlacklisted  = errors.New("token: token is blacklisted")
	ErrTokenInvalid      = errors.New("token: invalid token")
	ErrTokenInvalidClaim = errors.New("token: invalid token claim")
	ErrTokenExpired      = errors.New("token: token expired")
	ErrInvalidTokenType  = errors.New("token: invalid token type")

	// auth errors
	ErrInvalidCredentials = errors.New("invalid credentials")

	// oauth errors
	ErrOAuthInvalidState           = &OAuthError{Code: "invalid_state", Description: "Invalid or expired state parameter"}
	ErrOAuthExchangeFailed         = errors.New("oauth: token exchange failed")
	ErrOAuthUserInfoFailed         = errors.New("oauth: failed to get user info")
	ErrOAuthAccessDenied           = &OAuthError{Code: "access_denied", Description: "User denied the authorization request"}
	ErrOAuthInvalidCode            = &OAuthError{Code: "invalid_code", Description: "Invalid or expired authorization code"}
	ErrOAuthServerError            = &OAuthError{Code: "server_error", Description: "OAuth server returned an error"}
	ErrOAuthTemporarilyUnavailable = &OAuthError{Code: "temporarily_unavailable", Description: "OAuth server is temporarily unavailable"}
)
