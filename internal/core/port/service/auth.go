package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

// AuthService is the primary port (inbound) for authentication use cases.
type AuthService interface {
	// Authenticate authenticates a user with credentials and returns tokens.
	Authenticate(ctx context.Context, payload *param.AuthParams) (*model.Token, error)

	// Google OAuth
	GoogleOAuth(ctx context.Context, redirectURI string) (string, string, error)

	// Handle Google Oauth redirection
	HandleGoogleOAuth(ctx context.Context, code, state, redirectURI string) (*model.Token, error)

	// RefreshToken refreshes access token using a refresh token.
	RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error)

	// ValidateToken validates an access token and returns claims.
	ValidateToken(ctx context.Context, accessToken string) (*model.TokenClaims, error)

	// Token management
	RevokeToken(ctx context.Context, token string, tokenType string) error

	// VerifyPin verifies a user's PIN.
	VerifyPin(ctx context.Context, userUid string, pin string) (bool, error)
}
