package errors

var (
	// auth errors
	ErrInvalidCredentials    = NewCustomError(20001, "invalid credentials", nil)
	ErrInvalidIdentifierType = NewCustomError(20002, "invalid identifier type", nil)

	// rate limiting errors
	ErrAuthTooManyAttempts = NewCustomError(20003, "too many failed attempts", nil)
	ErrRateLimitExceeded   = NewCustomError(20004, "rate limit exceeded", nil)

	// token errors
	ErrTokenRevoked      = NewCustomError(20005, "token: token has been revoked", nil)
	ErrTokenBlacklisted  = NewCustomError(20006, "token: token is blacklisted", nil)
	ErrTokenInvalid      = NewCustomError(20007, "token: invalid token", nil)
	ErrTokenInvalidClaim = NewCustomError(20008, "token: invalid token claim", nil)
	ErrTokenExpired      = NewCustomError(20009, "token: token expired", nil)
	ErrInvalidTokenType  = NewCustomError(20010, "token: invalid token type", nil)

	// oauth errors
	ErrOAuthClientIDRequired           = NewCustomError(20011, "oauth: client id is required", nil)
	ErrOAuthClientSecretRequired       = NewCustomError(20012, "oauth: client secret is required", nil)
	ErrOAuthFailedGenerateCodeVerifier = NewCustomError(20013, "oauth: failed to generate code verifier", nil)
	ErrOAuthInvalidMinVerifierLength   = NewCustomError(20014, "oauth: invalid minimum verifier length", nil)
	ErrOAuthInvalidMaxVerifierLength   = NewCustomError(20015, "oauth: invalid maximum verifier length", nil)
	ErrOAuthInvalidState               = NewCustomError(20016, "oauth: invalid state parameter", nil)
	ErrOAuthExchangeFailed             = NewCustomError(20017, "oauth: token exchange failed", nil)
	ErrOAuthUserInfoFailed             = NewCustomError(20018, "oauth: failed to get user info", nil)
	ErrOAuthAccessDenied               = NewCustomError(20019, "oauth: access denied", nil)
	ErrOAuthInvalidCode                = NewCustomError(20020, "oauth: invalid code", nil)
	ErrOAuthServerError                = NewCustomError(20021, "oauth: server error", nil)
	ErrOAuthTemporarilyUnavailable     = NewCustomError(20022, "oauth: temporarily unavailable", nil)
	// PKCE-specific errors
	ErrOAuthCodeVerifierMissing = NewCustomError(20023, "oauth: code verifier not found", nil)
)
