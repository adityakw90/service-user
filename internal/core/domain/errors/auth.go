package errors

import "errors"

var (
	// token errors
	ErrTokenRevoked     = errors.New("token: token has been revoked")
	ErrTokenBlacklisted = errors.New("token: token is blacklisted")

	// oauth errors
	ErrOAuthInvalidState   = errors.New("oauth: invalid state parameter")
	ErrOAuthExchangeFailed = errors.New("oauth: token exchange failed")
	ErrOAuthUserInfoFailed = errors.New("oauth: failed to get user info")
)
