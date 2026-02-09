package event

type EventLoginData struct {
	Identifier     string
	IdentifierType string
}

type EventLoginFailedData struct {
	Identifier     string
	IdentifierType string
	FailureReason  string
}

type EventLoginLockedData struct {
	Identifier     string
	IdentifierType string
	FailureReason  string
}

type EventTokenRefreshData struct {
	Identifier     string
	IdentifierType string
}

type EventRevokeTokenData struct {
	Identifier     string
	IdentifierType string
}

// EventPinVerifyData is emitted when a PIN is verified.
type EventPinVerifyData struct {
	UserUID string
	Success bool
	Reason  string
}

// EventPinFailData is emitted when a PIN verification fails.
type EventPinFailData struct {
	UserUID string
	Reason  string
}

// EventOAuthLoginData is emitted when a user logs in via OAuth.
type EventOAuthLoginData struct {
	UserUID  string
	Provider string
}
