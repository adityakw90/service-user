package security_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/security"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func newTestTokenGenerator(t *testing.T) portsec.TokenGenerator {
	t.Helper()
	return security.NewJWTGenerator("test-secret-key-256bits-long!!!!", 15*time.Minute, 7*24*time.Hour)
}

func newTestWhitelistStore(t *testing.T, client *redis.Client) portsec.TokenStore {
	t.Helper()
	return security.NewTokenWhitelistAdapter(client, "test-wl:", 15*time.Minute, infra.NewNoopTracer(), nil)
}

func newTestBlacklistStore(t *testing.T, client *redis.Client) portsec.TokenStore {
	t.Helper()
	return security.NewTokenBlacklistAdapter(client, "test-bl:", 24*time.Hour, infra.NewNoopTracer(), nil)
}

var testClaims = &model.TokenClaims{
	Uid:            "user-uid-abc",
	Sid:            "session-sid-xyz",
	Type:           model.TokenTypeAccess,
	Identifier:     "user@example.com",
	IdentifierType: "email",
}

// helper: check if error wraps ErrTokenRevoked
func isTokenRevoked(err error) bool {
	return errors.Is(err, domainerrors.ErrTokenRevoked)
}

// ===== StatelessTokenManager =====

func TestStatelessTokenManager_Generate(t *testing.T) {
	gen := newTestTokenGenerator(t)
	mgr := security.NewStatelessTokenManager(gen)

	tokenStr, err := mgr.Generate(context.Background(), testClaims)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if tokenStr == "" {
		t.Error("expected non-empty token string")
	}
}

func TestStatelessTokenManager_Validate(t *testing.T) {
	gen := newTestTokenGenerator(t)
	mgr := security.NewStatelessTokenManager(gen)

	tokenStr, _ := mgr.Generate(context.Background(), testClaims)
	claims, err := mgr.Validate(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if claims.Uid != testClaims.Uid {
		t.Errorf("got Uid %q, want %q", claims.Uid, testClaims.Uid)
	}
}

func TestStatelessTokenManager_Revoke_IsNoop(t *testing.T) {
	gen := newTestTokenGenerator(t)
	mgr := security.NewStatelessTokenManager(gen)

	tokenStr, _ := mgr.Generate(context.Background(), testClaims)

	if err := mgr.Revoke(context.Background(), tokenStr); err != nil {
		t.Errorf("Revoke returned unexpected error: %v", err)
	}
	if err := mgr.RevokeSession(context.Background(), "uid", "sid"); err != nil {
		t.Errorf("RevokeSession returned unexpected error: %v", err)
	}
	if err := mgr.RevokeAllSessions(context.Background(), "uid"); err != nil {
		t.Errorf("RevokeAllSessions returned unexpected error: %v", err)
	}
}

// ===== WhitelistTokenManager =====

func TestWhitelistTokenManager_Generate_AddsToStore(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestWhitelistStore(t, client)
	mgr := security.NewWhitelistTokenManager(gen, store)

	_, err := mgr.Generate(context.Background(), testClaims)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	allowed, err := store.IsAllowed(context.Background(), testClaims.Uid, testClaims.Sid)
	if err != nil {
		t.Fatalf("IsAllowed failed: %v", err)
	}
	if !allowed {
		t.Error("expected token to be in whitelist after Generate")
	}
}

func TestWhitelistTokenManager_Validate_NotInStore_ReturnsRevoked(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestWhitelistStore(t, client)
	mgr := security.NewWhitelistTokenManager(gen, store)

	// Generate raw token WITHOUT adding to store
	rawGen := newTestTokenGenerator(t)
	tokenStr, _ := rawGen.GenerateToken(testClaims)

	_, err := mgr.Validate(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("expected error for token not in whitelist")
	}
	if !isTokenRevoked(err) {
		t.Errorf("expected ErrTokenRevoked, got: %v", err)
	}
}

func TestWhitelistTokenManager_Revoke_RemovesFromStore(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestWhitelistStore(t, client)
	mgr := security.NewWhitelistTokenManager(gen, store)

	tokenStr, _ := mgr.Generate(context.Background(), testClaims)

	if err := mgr.Revoke(context.Background(), tokenStr); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	allowed, _ := store.IsAllowed(context.Background(), testClaims.Uid, testClaims.Sid)
	if allowed {
		t.Error("expected token to be removed from whitelist after Revoke")
	}
}

func TestWhitelistTokenManager_RevokeSession(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestWhitelistStore(t, client)
	mgr := security.NewWhitelistTokenManager(gen, store)

	_ = store.Add(context.Background(), "uid-1", "sid-1")

	if err := mgr.RevokeSession(context.Background(), "uid-1", "sid-1"); err != nil {
		t.Fatalf("RevokeSession failed: %v", err)
	}

	allowed, _ := store.IsAllowed(context.Background(), "uid-1", "sid-1")
	if allowed {
		t.Error("expected session to be removed from whitelist after RevokeSession")
	}
}

func TestWhitelistTokenManager_RevokeAllSessions(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestWhitelistStore(t, client)
	mgr := security.NewWhitelistTokenManager(gen, store)

	_ = store.Add(context.Background(), "uid-2", "sid-a")
	_ = store.Add(context.Background(), "uid-2", "sid-b")

	if err := mgr.RevokeAllSessions(context.Background(), "uid-2"); err != nil {
		t.Fatalf("RevokeAllSessions failed: %v", err)
	}

	for _, sid := range []string{"sid-a", "sid-b"} {
		allowed, _ := store.IsAllowed(context.Background(), "uid-2", sid)
		if allowed {
			t.Errorf("expected session %q to be removed after RevokeAllSessions", sid)
		}
	}
}

// ===== BlacklistTokenManager =====

func TestBlacklistTokenManager_Generate_DoesNotAddToStore(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestBlacklistStore(t, client)
	mgr := security.NewBlacklistTokenManager(gen, store)

	_, err := mgr.Generate(context.Background(), testClaims)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// IsAllowed=true means NOT in blacklist — correct for a fresh token
	allowed, err := store.IsAllowed(context.Background(), testClaims.Uid, testClaims.Sid)
	if err != nil {
		t.Fatalf("IsAllowed failed: %v", err)
	}
	if !allowed {
		t.Error("expected fresh token to not be blacklisted")
	}
}

func TestBlacklistTokenManager_Validate_ValidToken(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestBlacklistStore(t, client)
	mgr := security.NewBlacklistTokenManager(gen, store)

	tokenStr, _ := mgr.Generate(context.Background(), testClaims)
	claims, err := mgr.Validate(context.Background(), tokenStr)
	if err != nil {
		t.Fatalf("Validate failed: %v", err)
	}
	if claims.Uid != testClaims.Uid {
		t.Errorf("got Uid %q, want %q", claims.Uid, testClaims.Uid)
	}
}

func TestBlacklistTokenManager_Revoke_BlocksValidation(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestBlacklistStore(t, client)
	mgr := security.NewBlacklistTokenManager(gen, store)

	tokenStr, _ := mgr.Generate(context.Background(), testClaims)

	if err := mgr.Revoke(context.Background(), tokenStr); err != nil {
		t.Fatalf("Revoke failed: %v", err)
	}

	_, err := mgr.Validate(context.Background(), tokenStr)
	if err == nil {
		t.Fatal("expected error after Revoke")
	}
	if !isTokenRevoked(err) {
		t.Errorf("expected ErrTokenRevoked, got: %v", err)
	}
}

func TestBlacklistTokenManager_RevokeAllSessions_AddsWildcard(t *testing.T) {
	client := newTestRedis(t)
	gen := newTestTokenGenerator(t)
	store := newTestBlacklistStore(t, client)
	mgr := security.NewBlacklistTokenManager(gen, store)

	if err := mgr.RevokeAllSessions(context.Background(), "uid-3"); err != nil {
		t.Fatalf("RevokeAllSessions failed: %v", err)
	}

	// "*" entry should be in the blacklist (IsAllowed returns false = it's in the blacklist)
	allowed, _ := store.IsAllowed(context.Background(), "uid-3", "*")
	if allowed {
		t.Error("expected wildcard '*' to be in blacklist after RevokeAllSessions")
	}
}
