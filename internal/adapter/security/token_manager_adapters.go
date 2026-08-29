package security

import (
	"context"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// ─── StatelessTokenManager ───────────────────────────────────────────────────

// StatelessTokenManager wraps TokenGenerator with no storage.
// Revoke/RevokeSession/RevokeAllSessions are no-ops.
type StatelessTokenManager struct {
	gen portsec.TokenGenerator
}

// NewStatelessTokenManager creates a TokenManager with no persistence.
func NewStatelessTokenManager(gen portsec.TokenGenerator) portsec.TokenManager {
	if gen == nil {
		panic("token generator is required")
	}
	return &StatelessTokenManager{gen: gen}
}

func (m *StatelessTokenManager) Generate(_ context.Context, claims *model.TokenClaims) (string, error) {
	return m.gen.GenerateToken(claims)
}

func (m *StatelessTokenManager) Validate(_ context.Context, tokenStr string) (*model.TokenClaims, error) {
	return m.gen.ValidateToken(tokenStr)
}

func (m *StatelessTokenManager) Revoke(_ context.Context, _ string) error {
	return nil
}

func (m *StatelessTokenManager) RevokeSession(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *StatelessTokenManager) RevokeAllSessions(_ context.Context, _ string) error {
	return nil
}

// ─── WhitelistTokenManager ───────────────────────────────────────────────────

// WhitelistTokenManager wraps TokenGenerator + TokenStore (whitelist strategy).
// Generate adds the session to the store. Validate checks the store. Revoke removes the session.
type WhitelistTokenManager struct {
	gen   portsec.TokenGenerator
	store portsec.TokenStore
}

// NewWhitelistTokenManager creates a TokenManager backed by a whitelist store.
func NewWhitelistTokenManager(gen portsec.TokenGenerator, store portsec.TokenStore) portsec.TokenManager {
	if gen == nil {
		panic("token generator is required")
	}
	if store == nil {
		panic("token store is required")
	}
	return &WhitelistTokenManager{gen: gen, store: store}
}

func (m *WhitelistTokenManager) Generate(ctx context.Context, claims *model.TokenClaims) (string, error) {
	tokenStr, err := m.gen.GenerateToken(claims)
	if err != nil {
		return "", err
	}
	if err := m.store.Add(ctx, claims.Uid, claims.Sid); err != nil {
		return "", err
	}
	return tokenStr, nil
}

func (m *WhitelistTokenManager) Validate(ctx context.Context, tokenStr string) (*model.TokenClaims, error) {
	claims, err := m.gen.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	allowed, err := m.store.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domainerrors.ErrTokenRevoked
	}
	return claims, nil
}

func (m *WhitelistTokenManager) Revoke(ctx context.Context, tokenStr string) error {
	claims, err := m.gen.ValidateToken(tokenStr)
	if err != nil {
		return err
	}
	return m.store.Remove(ctx, claims.Uid, claims.Sid)
}

func (m *WhitelistTokenManager) RevokeSession(ctx context.Context, userUID string, sid string) error {
	return m.store.Remove(ctx, userUID, sid)
}

func (m *WhitelistTokenManager) RevokeAllSessions(ctx context.Context, userUID string) error {
	return m.store.RemoveAll(ctx, userUID)
}

// ─── BlacklistTokenManager ───────────────────────────────────────────────────

// BlacklistTokenManager wraps TokenGenerator + TokenStore (blacklist strategy).
// Generate issues a pure JWT. Revoke adds the session to the blacklist. RevokeAllSessions adds wildcard "*".
type BlacklistTokenManager struct {
	gen   portsec.TokenGenerator
	store portsec.TokenStore
}

// NewBlacklistTokenManager creates a TokenManager backed by a blacklist store.
func NewBlacklistTokenManager(gen portsec.TokenGenerator, store portsec.TokenStore) portsec.TokenManager {
	if gen == nil {
		panic("token generator is required")
	}
	if store == nil {
		panic("token store is required")
	}
	return &BlacklistTokenManager{gen: gen, store: store}
}

func (m *BlacklistTokenManager) Generate(_ context.Context, claims *model.TokenClaims) (string, error) {
	return m.gen.GenerateToken(claims)
}

func (m *BlacklistTokenManager) Validate(ctx context.Context, tokenStr string) (*model.TokenClaims, error) {
	claims, err := m.gen.ValidateToken(tokenStr)
	if err != nil {
		return nil, err
	}
	// IsAllowed returns false when the key EXISTS in the blacklist
	allowed, err := m.store.IsAllowed(ctx, claims.Uid, claims.Sid)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, domainerrors.ErrTokenRevoked
	}
	return claims, nil
}

func (m *BlacklistTokenManager) Revoke(ctx context.Context, tokenStr string) error {
	claims, err := m.gen.ValidateToken(tokenStr)
	if err != nil {
		return err
	}
	return m.store.Add(ctx, claims.Uid, claims.Sid)
}

func (m *BlacklistTokenManager) RevokeSession(ctx context.Context, userUID string, sid string) error {
	return m.store.Add(ctx, userUID, sid)
}

func (m *BlacklistTokenManager) RevokeAllSessions(ctx context.Context, userUID string) error {
	return m.store.Add(ctx, userUID, "*")
}
