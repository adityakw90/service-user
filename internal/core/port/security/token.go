package security

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// TokenGenerator is a port for generating authentication tokens.
type TokenGenerator interface {
	GenerateToken(claims *model.TokenClaims) (string, error)
	ValidateToken(token string) (*model.TokenClaims, error)
}

// TokenStore is a port for manage Token whitelist or blacklist
type TokenStore interface {
	Add(ctx context.Context, user_uid string, tid string) error
	Remove(ctx context.Context, user_uid string, tid string) error
	RemoveAll(ctx context.Context, user_uid string) error
	IsAllowed(ctx context.Context, user_uid string, tid string) (bool, error)
}
