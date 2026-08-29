package security_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/adapter/security"
	"github.com/adityakw90/service-user/internal/infra"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newFactoryTestRedis(t *testing.T) *redis.Client {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()})
}

func TestNewTokenManager(t *testing.T) {
	gen := security.NewJWTGenerator("test-secret-key-256bits-long!!!!", 15*time.Minute, 7*24*time.Hour)
	client := newFactoryTestRedis(t)
	store := security.NewTokenBlacklistAdapter(client, "test-factory:", 24*time.Hour, infra.NewNoopTracer(), nil)

	tests := []struct {
		name     string
		strategy string
		gen      interface{}
		store    interface{}
		wantErr  bool
		wantType string
	}{
		{
			name:     "stateless strategy",
			strategy: "stateless",
			wantErr:  false,
		},
		{
			name:     "empty string defaults to stateless",
			strategy: "",
			wantErr:  false,
		},
		{
			name:     "whitelist strategy",
			strategy: "whitelist",
			wantErr:  false,
		},
		{
			name:     "blacklist strategy",
			strategy: "blacklist",
			wantErr:  false,
		},
		{
			name:     "unknown strategy returns error",
			strategy: "invalid",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, err := security.NewTokenManager(tt.strategy, gen, store)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if mgr == nil {
				t.Error("expected non-nil TokenManager")
			}
		})
	}
}
