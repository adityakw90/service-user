package security

import (
	"context"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/redis/go-redis/v9"
)

// TokenBlacklistAdapter implements port.TokenStore using Redis.
type TokenBlacklistAdapter struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	tracer     gomon.Tracer
	logger     gomon.Logger
}

// NewTokenBlacklistAdapter creates a new token blacklist adapter.
func NewTokenBlacklistAdapter(
	redisClient *redis.Client,
	redisPrefix string,
	defaultTTL time.Duration,
	tracer gomon.Tracer,
	logger gomon.Logger,
) portsec.TokenStore {
	if redisClient == nil {
		panic("redis client is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &TokenBlacklistAdapter{
		client:     redisClient,
		prefix:     redisPrefix,
		defaultTTL: defaultTTL,
		tracer:     tracer,
		logger:     logger,
	}
}

// buildKey builds the Redis key for a user token.
func (a *TokenBlacklistAdapter) buildKey(userUID, tid string) string {
	return a.prefix + userUID + ":" + tid
}

// Add adds a token to the whitelist.
func (a *TokenBlacklistAdapter) Add(ctx context.Context, userUID, tid string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenBlacklist.Add")
	defer span.End()

	key := a.buildKey(userUID, tid)
	// Use a long TTL (30 days) for refresh token whitelist
	err := a.client.Set(newCtx, key, "1", a.defaultTTL).Err()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token blacklist add failed", map[string]any{"error": err, "userUID": truncateID(userUID), "tid": truncateID(tid)})
		}
		return err
	}
	return nil
}

// Remove removes a token from the whitelist.
func (a *TokenBlacklistAdapter) Remove(ctx context.Context, userUID, tid string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenBlacklist.Remove")
	defer span.End()

	key := a.buildKey(userUID, tid)
	err := a.client.Del(newCtx, key).Err()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token blacklist remove failed", map[string]any{"error": err, "userUID": truncateID(userUID), "tid": truncateID(tid)})
		}
		return err
	}
	return nil
}

// RemoveAll removes all tokens for a user.
func (a *TokenBlacklistAdapter) RemoveAll(ctx context.Context, userUID string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenBlacklist.RemoveAll")
	defer span.End()

	pattern := a.buildKey(userUID, "*")
	iter := a.client.Scan(newCtx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(newCtx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		if a.logger != nil {
			a.logger.Error("redis token blacklist scan failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return err
	}

	if len(keys) > 0 {
		err := a.client.Del(newCtx, keys...).Err()
		if err != nil {
			if a.logger != nil {
				a.logger.Error("redis token blacklist removeAll failed", map[string]any{"error": err, "userUID": truncateID(userUID), "keysCount": len(keys)})
			}
			return err
		}
	}

	return nil
}

// IsAllowed checks if a token is in the blacklist.
func (a *TokenBlacklistAdapter) IsAllowed(ctx context.Context, userUID, tid string) (bool, error) {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenBlacklist.IsAllowed")
	defer span.End()

	key := a.buildKey(userUID, tid)
	result, err := a.client.Exists(newCtx, key).Result()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token blacklist isallowed failed", map[string]any{"error": err, "userUID": truncateID(userUID), "tid": truncateID(tid)})
		}
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
