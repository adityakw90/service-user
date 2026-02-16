package port

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// OAuthProvider is a port for OAuth authentication providers.
type OAuthProvider interface {
	// GetAuthorizationURL returns the OAuth authorization URL with state parameter.
	GetAuthorizationURL(ctx context.Context, redirectURI, state string) (string, error)

	// ExchangeCode exchanges the authorization code for tokens.
	ExchangeCode(ctx context.Context, code, state, redirectURI string) (*model.OAuthTokens, error)

	// GetUserInfo retrieves user information using the access token.
	GetUserInfo(ctx context.Context, token *model.OAuthTokens) (*model.OAuthUserInfo, error)
}
