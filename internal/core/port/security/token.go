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

// TokenManager is a unified port for token generation, validation, and revocation.
// Use two separate instances (access, refresh) wired to concrete strategy adapters.
type TokenManager interface {
	Generate(ctx context.Context, claims *model.TokenClaims) (string, error)
	Validate(ctx context.Context, tokenStr string) (*model.TokenClaims, error)
	Revoke(ctx context.Context, tokenStr string) error
	RevokeSession(ctx context.Context, userUID string, sid string) error
	RevokeAllSessions(ctx context.Context, userUID string) error
}
