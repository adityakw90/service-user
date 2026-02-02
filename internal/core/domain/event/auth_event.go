package event

import (
	"time"
)

// AuthEvent represents an authentication domain event.
type AuthEvent struct {
	ID               string
	Type             EventType
	UserUID          string
	Identifier       string
	IdentifierType   string
	Success          bool
	FailureReason    string
	DeviceFingerprint string
	IPAddress        string
	Metadata         map[string]any
	Timestamp        time.Time
}

// NewAuthEvent creates a new authentication event.
func NewAuthEvent(
	eventType EventType,
	userUID, identifier, identifierType string,
	success bool,
	opts ...AuthEventOption,
) *AuthEvent {
	event := &AuthEvent{
		ID:             generateEventID(),
		Type:           eventType,
		UserUID:        userUID,
		Identifier:     identifier,
		IdentifierType: identifierType,
		Success:        success,
		Timestamp:      time.Now().UTC(),
	}

	for _, opt := range opts {
		opt(event)
	}

	return event
}

// AuthEventOption configures an AuthEvent.
type AuthEventOption func(*AuthEvent)

// WithFailureReason sets the failure reason.
func WithFailureReason(reason string) AuthEventOption {
	return func(e *AuthEvent) {
		e.FailureReason = reason
	}
}

// WithDevice sets device information.
func WithDevice(fingerprint, ip string) AuthEventOption {
	return func(e *AuthEvent) {
		e.DeviceFingerprint = fingerprint
		e.IPAddress = ip
	}
}

// WithMetadata sets additional metadata.
func WithMetadata(m map[string]any) AuthEventOption {
	return func(e *AuthEvent) {
		e.Metadata = m
	}
}

// generateEventID generates a unique event ID.
func generateEventID() string {
	// Using UUID v7 format for time-ordered events
	return time.Now().UTC().Format("20060102150405") + "-" + randomString(8)
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

// LoginEvent creates a login success event.
func LoginEvent(userUID, identifier, identifierType, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventLogin,
		userUID, identifier, identifierType,
		true,
		WithDevice(fingerprint, ip),
	)
}

// LoginFailedEvent creates a login failure event.
func LoginFailedEvent(identifier, identifierType, reason, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventLoginFail,
		"", identifier, identifierType,
		false,
		WithFailureReason(reason),
		WithDevice(fingerprint, ip),
	)
}

// LogoutEvent creates a logout event.
func LogoutEvent(userUID, identifier, identifierType, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventLogout,
		userUID, identifier, identifierType,
		true,
		WithDevice(fingerprint, ip),
	)
}

// TokenRefreshEvent creates a token refresh event.
func TokenRefreshEvent(userUID, identifier, identifierType, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventTokenRefresh,
		userUID, identifier, identifierType,
		true,
		WithDevice(fingerprint, ip),
	)
}

// PasswordChangeEvent creates a password change event.
func PasswordChangeEvent(userUID, identifier, identifierType, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventPasswordChange,
		userUID, identifier, identifierType,
		true,
		WithDevice(fingerprint, ip),
	)
}

// AccountLockoutEvent creates an account lockout event.
func AccountLockoutEvent(userUID, identifier, identifierType, reason string) *AuthEvent {
	return NewAuthEvent(
		EventAccountLockout,
		userUID, identifier, identifierType,
		false,
		WithFailureReason(reason),
	)
}

// AccountUnlockEvent creates an account unlock event.
func AccountUnlockEvent(userUID, identifier, identifierType string) *AuthEvent {
	return NewAuthEvent(
		EventAccountUnlock,
		userUID, identifier, identifierType,
		true,
	)
}

// OAuthLoginEvent creates an OAuth login event.
func OAuthLoginEvent(userUID, identifier, provider, fingerprint, ip string) *AuthEvent {
	return NewAuthEvent(
		EventOAuthLogin,
		userUID, identifier, provider,
		true,
		WithDevice(fingerprint, ip),
		WithMetadata(map[string]any{"oauth_provider": provider}),
	)
}

// PINVerifyEvent creates a PIN verification event.
func PINVerifyEvent(userUID string, success bool, reason string) *AuthEvent {
	eventType := EventPINVerify
	if !success {
		eventType = EventPINFail
	}
	return NewAuthEvent(
		eventType,
		userUID, "", "pin",
		success,
		WithFailureReason(reason),
	)
}
