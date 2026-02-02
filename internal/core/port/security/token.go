package security

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// TokenGenerator is a port for generating authentication tokens.
type TokenGenerator interface {
	GenerateAccessToken(claims model.TokenClaims) (string, error)
	GenerateRefreshToken(claims model.TokenClaims) (string, error)
	ValidateToken(token string) (*model.TokenClaims, error)
}

// TokenStore is a port for manage Token whitelist or blacklist
type TokenStore interface {
	Add(ctx context.Context, user_uid string, token string) error
	Remove(ctx context.Context, user_uid string, token string) error
	IsAllowed(ctx context.Context, user_uid string, token string) (bool, error)
}
