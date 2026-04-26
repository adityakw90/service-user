package security

import (
	"context"
	"fmt"
	"time"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/redis/go-redis/v9"
)

// TokenBlacklistAdapter implements port.TokenStore using Redis.
type TokenBlacklistAdapter struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
}

// NewTokenBlacklistAdapter creates a new token blacklist adapter.
func NewTokenBlacklistAdapter(
	redisClient *redis.Client,
	redisPrefix string,
	defaultTTL time.Duration,
) portsec.TokenStore {
	return &TokenBlacklistAdapter{
		client:     redisClient,
		prefix:     redisPrefix,
		defaultTTL: defaultTTL,
	}
}

// buildKey builds the Redis key for a user token.
func (a *TokenBlacklistAdapter) buildKey(userUID, tid string) string {
	return a.prefix + userUID + ":" + tid
}

// Add adds a token to the whitelist.
func (a *TokenBlacklistAdapter) Add(ctx context.Context, userUID, tid string) error {
	key := a.buildKey(userUID, tid)
	// Use a long TTL (30 days) for refresh token whitelist
	return a.client.Set(ctx, key, "1", a.defaultTTL).Err()
}

// Remove removes a token from the whitelist.
func (a *TokenBlacklistAdapter) Remove(ctx context.Context, userUID, tid string) error {
	key := a.buildKey(userUID, tid)
	return a.client.Del(ctx, key).Err()
}

// RemoveAll removes all tokens for a user.
func (a *TokenBlacklistAdapter) RemoveAll(ctx context.Context, userUID string) error {
	pattern := a.buildKey(userUID, "*")
	iter := a.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) > 0 {
		return a.client.Del(ctx, keys...).Err()
	}

	return nil
}

// IsAllowed checks if a token is in the blacklist.
func (a *TokenBlacklistAdapter) IsAllowed(ctx context.Context, userUID, tid string) (bool, error) {
	key := a.buildKey(userUID, tid)
	result, err := a.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return result == 0, nil
}

// TokenBlacklistNoOpAdapter implements port.TokenStore using Redis.
type TokenBlacklistNoOpAdapter struct{}

// NewTokenBlacklistNoOpAdapter creates a new token blacklist adapter.
func NewTokenBlacklistNoOpAdapter() portsec.TokenStore {
	return &TokenBlacklistNoOpAdapter{}
}

// Add adds a token to the whitelist.
func (a *TokenBlacklistNoOpAdapter) Add(ctx context.Context, userUID, tid string) error {
	return nil
}

// Remove removes a token from the whitelist.
func (a *TokenBlacklistNoOpAdapter) Remove(ctx context.Context, userUID, tid string) error {
	return nil
}

// RemoveAll removes all tokens for a user.
func (a *TokenBlacklistNoOpAdapter) RemoveAll(ctx context.Context, userUID string) error {
	return nil
}

// IsAllowed checks if a token is in the blacklist.
func (a *TokenBlacklistNoOpAdapter) IsAllowed(ctx context.Context, userUID, tid string) (bool, error) {
	return true, nil
}
