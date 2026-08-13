package security

import (
	"context"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/infra"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenWhitelistAdapter_Add(t *testing.T) {
	tests := []struct {
		name    string
		userUID string
		tid     string
		wantErr bool
	}{
		{name: "Successfully add token", userUID: "user-123", tid: "token-abc", wantErr: false},
		{name: "Add duplicate token", userUID: "user-123", tid: "token-abc", wantErr: false},
		{name: "Add token with different user", userUID: "user-456", tid: "token-xyz", wantErr: false},
		{name: "Add token with special chars", userUID: "user-789", tid: "token-with-special-chars-123", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := miniredis.Run()
			require.NoError(t, err)
			defer s.Close()

			client := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer client.Close()
			tracer := infra.NewNoopTracer()
			logger := infra.NewNoopLogger()

			adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)

			err = adapter.Add(context.Background(), tt.userUID, tt.tid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				key := adapter.buildKey(tt.userUID, tt.tid)
				exists, _ := client.Exists(context.Background(), key).Result()
				assert.Greater(t, exists, int64(0), "token should exist in Redis")
			}
		})
	}
}

func TestTokenWhitelistAdapter_Remove(t *testing.T) {
	tests := []struct {
		name         string
		userUID      string
		tid          string
		preAddToken  bool
		wantErr      bool
		expectExists bool
	}{
		{name: "Remove existing token", userUID: "user-123", tid: "token-abc", preAddToken: true, wantErr: false, expectExists: false},
		{name: "Remove non-existent token", userUID: "user-123", tid: "token-xyz", preAddToken: false, wantErr: false, expectExists: false},
		{name: "Remove token from user with multiple tokens", userUID: "user-multi", tid: "token-1", preAddToken: true, wantErr: false, expectExists: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := miniredis.Run()
			require.NoError(t, err)
			defer s.Close()

			client := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer client.Close()
			tracer := infra.NewNoopTracer()
			logger := infra.NewNoopLogger()

			adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)
			ctx := context.Background()

			if tt.preAddToken {
				err = adapter.Add(ctx, tt.userUID, tt.tid)
				require.NoError(t, err)
			}

			err = adapter.Remove(ctx, tt.userUID, tt.tid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				key := adapter.buildKey(tt.userUID, tt.tid)
				exists, _ := client.Exists(ctx, key).Result()
				assert.Equal(t, int64(0), exists, "token should not exist in Redis")
			}
		})
	}
}

func TestTokenWhitelistAdapter_RemoveAll(t *testing.T) {
	tests := []struct {
		name              string
		userUID           string
		tokensToAdd       []string
		wantErr           bool
		expectTokensExist bool
	}{
		{
			name:              "Remove all tokens",
			userUID:           "user-123",
			tokensToAdd:       []string{"token-1", "token-2", "token-3"},
			wantErr:           false,
			expectTokensExist: false,
		},
		{
			name:              "Remove all with no tokens",
			userUID:           "user-456",
			tokensToAdd:       nil,
			wantErr:           false,
			expectTokensExist: false,
		},
		{
			name:              "Remove all with single token",
			userUID:           "user-789",
			tokensToAdd:       []string{"only-token"},
			wantErr:           false,
			expectTokensExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := miniredis.Run()
			require.NoError(t, err)
			defer s.Close()

			client := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer client.Close()
			tracer := infra.NewNoopTracer()
			logger := infra.NewNoopLogger()

			adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)
			ctx := context.Background()

			for _, token := range tt.tokensToAdd {
				err = adapter.Add(ctx, tt.userUID, token)
				require.NoError(t, err)
			}

			err = adapter.RemoveAll(ctx, tt.userUID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				for _, token := range tt.tokensToAdd {
					key := adapter.buildKey(tt.userUID, token)
					exists, _ := client.Exists(ctx, key).Result()
					assert.Equal(t, int64(0), exists, "token should not exist in Redis")
				}
			}
		})
	}
}

func TestTokenWhitelistAdapter_IsAllowed(t *testing.T) {
	tests := []struct {
		name        string
		userUID     string
		tid         string
		preAddToken bool
		wantAllowed bool
	}{
		{name: "Token exists", userUID: "user-123", tid: "token-abc", preAddToken: true, wantAllowed: true},
		{name: "Token doesn't exist", userUID: "user-123", tid: "token-xyz", preAddToken: false, wantAllowed: false},
		{name: "Different user token", userUID: "user-456", tid: "token-abc", preAddToken: false, wantAllowed: false},
		{name: "Empty token ID", userUID: "user-789", tid: "", preAddToken: false, wantAllowed: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := miniredis.Run()
			require.NoError(t, err)
			defer s.Close()

			client := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer client.Close()
			tracer := infra.NewNoopTracer()
			logger := infra.NewNoopLogger()

			adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)
			ctx := context.Background()

			if tt.preAddToken {
				err = adapter.Add(ctx, tt.userUID, tt.tid)
				require.NoError(t, err)
			}

			allowed, err := adapter.IsAllowed(ctx, tt.userUID, tt.tid)

			assert.NoError(t, err)
			assert.Equal(t, tt.wantAllowed, allowed)
		})
	}
}

func TestTokenWhitelistAdapter_TTL(t *testing.T) {
	tests := []struct {
		name    string
		ttl     time.Duration
		forward time.Duration
	}{
		{name: "Token expires after TTL", ttl: 2 * time.Hour, forward: 3 * time.Hour},
		{name: "Token still valid within TTL", ttl: 24 * time.Hour, forward: 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := miniredis.Run()
			require.NoError(t, err)
			defer s.Close()

			client := redis.NewClient(&redis.Options{Addr: s.Addr()})
			defer client.Close()
			tracer := infra.NewNoopTracer()
			logger := infra.NewNoopLogger()

			adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", tt.ttl, tracer, logger).(*TokenWhitelistAdapter)
			ctx := context.Background()

			userUID := "user-123"
			tid := "token-abc"

			err = adapter.Add(ctx, userUID, tid)
			require.NoError(t, err)

			allowed, err := adapter.IsAllowed(ctx, userUID, tid)
			require.NoError(t, err)
			assert.True(t, allowed, "token should be allowed immediately after adding")

			s.FastForward(tt.forward)

			allowed, err = adapter.IsAllowed(ctx, userUID, tid)
			require.NoError(t, err)

			if tt.forward > tt.ttl {
				assert.False(t, allowed, "token should be expired after TTL")
			} else {
				assert.True(t, allowed, "token should still be valid within TTL")
			}
		})
	}
}

func TestTokenWhitelistAdapter_MultipleUsers(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()
	tracer := infra.NewNoopTracer()
	logger := infra.NewNoopLogger()

	adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)
	ctx := context.Background()

	user1 := "user-1"
	user2 := "user-2"
	token1 := "token-1"
	token2 := "token-2"

	err = adapter.Add(ctx, user1, token1)
	require.NoError(t, err)

	err = adapter.Add(ctx, user2, token2)
	require.NoError(t, err)

	allowed1, _ := adapter.IsAllowed(ctx, user1, token1)
	assert.True(t, allowed1)

	allowed2, _ := adapter.IsAllowed(ctx, user2, token2)
	assert.True(t, allowed2)

	crossUser1, _ := adapter.IsAllowed(ctx, user1, token2)
	assert.False(t, crossUser1)

	crossUser2, _ := adapter.IsAllowed(ctx, user2, token1)
	assert.False(t, crossUser2)
}

func TestTokenWhitelistAdapter_MultipleTokensPerUser(t *testing.T) {
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	client := redis.NewClient(&redis.Options{Addr: s.Addr()})
	defer client.Close()
	tracer := infra.NewNoopTracer()
	logger := infra.NewNoopLogger()

	adapter := NewTokenWhitelistAdapter(client, "token:whitelist:", 30*24*time.Hour, tracer, logger).(*TokenWhitelistAdapter)
	ctx := context.Background()

	userUID := "user-multi"
	tokens := []string{"token-1", "token-2", "token-3"}

	for _, token := range tokens {
		err = adapter.Add(ctx, userUID, token)
		require.NoError(t, err)
	}

	for _, token := range tokens {
		allowed, err := adapter.IsAllowed(ctx, userUID, token)
		require.NoError(t, err)
		assert.True(t, allowed, "token %s should be allowed", token)
	}

	err = adapter.Remove(ctx, userUID, tokens[1])
	require.NoError(t, err)

	allowed1, _ := adapter.IsAllowed(ctx, userUID, tokens[0])
	assert.True(t, allowed1, "token-1 should still be allowed")

	allowed2, _ := adapter.IsAllowed(ctx, userUID, tokens[1])
	assert.False(t, allowed2, "token-2 should be removed")

	allowed3, _ := adapter.IsAllowed(ctx, userUID, tokens[2])
	assert.True(t, allowed3, "token-3 should still be allowed")
}

func TestTokenWhitelistNoOpAdapter(t *testing.T) {
	adapter := NewTokenWhitelistNoOpAdapter().(*TokenWhitelistNoOpAdapter)
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func() error
	}{
		{name: "Add returns no error", fn: func() error { return adapter.Add(ctx, "user-123", "token-abc") }},
		{name: "Remove returns no error", fn: func() error { return adapter.Remove(ctx, "user-123", "token-abc") }},
		{name: "RemoveAll returns no error", fn: func() error { return adapter.RemoveAll(ctx, "user-123") }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NoError(t, tt.fn())
		})
	}

	t.Run("IsAllowed always returns true", func(t *testing.T) {
		allowed, err := adapter.IsAllowed(ctx, "user-123", "token-abc")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}
