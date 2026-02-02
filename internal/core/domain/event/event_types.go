package event

// EventType defines the type of authentication domain event.
type EventType string

const (
	EventLogin          EventType = "auth.login"
	EventLoginFail      EventType = "auth.login_fail"
	EventLogout         EventType = "auth.logout"
	EventTokenRefresh   EventType = "auth.token_refresh"
	EventPasswordChange EventType = "auth.password_change"
	EventAccountLockout EventType = "auth.account_lockout"
	EventAccountUnlock  EventType = "auth.account_unlock"
	EventOAuthLogin     EventType = "auth.oauth_login"
	EventPINVerify      EventType = "auth.pin_verify"
	EventPINFail        EventType = "auth.pin_fail"
)
