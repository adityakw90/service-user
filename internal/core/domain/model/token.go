package model

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type Token struct {
	Access  string
	Refresh string
}

type TokenClaims struct {
	Uid            string         // user uid
	Sid            string         // session id
	Type           string         // token type (access, refresh)
	Identifier     string         // identifier (username, email, phone)
	IdentifierType string         // identifier type (username, email, phone)
	Extra          map[string]any // extra data
}

func (t *TokenClaims) IsAccess() bool {
	return t.Type == TokenTypeAccess
}

func (t *TokenClaims) IsRefresh() bool {
	return t.Type == TokenTypeRefresh
}
