package security

import (
	"context"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/redis/go-redis/v9"
)

// TokenWhitelistAdapter implements TokenStore using Redis for token whitelist.
type TokenWhitelistAdapter struct {
	client     *redis.Client
	prefix     string
	defaultTTL time.Duration
	tracer     gomon.Tracer
	logger     gomon.Logger
}

// NewTokenWhitelistAdapter creates a new token whitelist adapter.
func NewTokenWhitelistAdapter(
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
	return &TokenWhitelistAdapter{
		client:     redisClient,
		prefix:     redisPrefix,
		defaultTTL: defaultTTL,
		tracer:     tracer,
		logger:     logger,
	}
}

// buildKey builds the Redis key for a user token.
func (a *TokenWhitelistAdapter) buildKey(userUID, tid string) string {
	return a.prefix + userUID + ":" + tid
}

// Add adds a token to the whitelist.
func (a *TokenWhitelistAdapter) Add(ctx context.Context, userUID, tid string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenWhitelist.Add")
	defer span.End()

	key := a.buildKey(userUID, tid)
	// Use a long TTL (30 days) for refresh token whitelist
	err := a.client.Set(newCtx, key, "1", a.defaultTTL).Err()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token whitelist add failed", map[string]any{"error": err, "userUID": userUID, "tid": tid})
		}
		return err
	}
	return nil
}

// Remove removes a token from the whitelist.
func (a *TokenWhitelistAdapter) Remove(ctx context.Context, userUID, tid string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenWhitelist.Remove")
	defer span.End()

	key := a.buildKey(userUID, tid)
	err := a.client.Del(newCtx, key).Err()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token whitelist remove failed", map[string]any{"error": err, "userUID": userUID, "tid": tid})
		}
		return err
	}
	return nil
}

// RemoveAll removes all tokens for a user.
func (a *TokenWhitelistAdapter) RemoveAll(ctx context.Context, userUID string) error {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenWhitelist.RemoveAll")
	defer span.End()

	pattern := a.buildKey(userUID, "*")
	iter := a.client.Scan(newCtx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(newCtx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		if a.logger != nil {
			a.logger.Error("redis token whitelist scan failed", map[string]any{"error": err, "userUID": userUID})
		}
		return err
	}

	if len(keys) > 0 {
		err := a.client.Del(newCtx, keys...).Err()
		if err != nil {
			if a.logger != nil {
				a.logger.Error("redis token whitelist removeAll failed", map[string]any{"error": err, "userUID": userUID, "keysCount": len(keys)})
			}
			return err
		}
	}

	return nil
}

// IsAllowed checks if a token is in the whitelist.
func (a *TokenWhitelistAdapter) IsAllowed(ctx context.Context, userUID, tid string) (bool, error) {
	newCtx, span := a.tracer.StartSpan(ctx, "redis.TokenWhitelist.IsAllowed")
	defer span.End()

	key := a.buildKey(userUID, tid)
	result, err := a.client.Exists(newCtx, key).Result()
	if err != nil {
		if a.logger != nil {
			a.logger.Error("redis token whitelist isallowed failed", map[string]any{"error": err, "userUID": userUID, "tid": tid})
		}
		return false, fmt.Errorf("failed to check token whitelist: %w", err)
	}
	return result > 0, nil
}

// TokenWhitelistNoOpAdapter implements port.TokenStore using Redis.
type TokenWhitelistNoOpAdapter struct{}

// NewTokenWhitelistNoOpAdapter creates a new token whitelist adapter.
func NewTokenWhitelistNoOpAdapter() portsec.TokenStore {
	return &TokenWhitelistNoOpAdapter{}
}

// Add adds a token to the whitelist.
func (a *TokenWhitelistNoOpAdapter) Add(ctx context.Context, userUID, tid string) error {
	return nil
}

// Remove removes a token from the whitelist.
func (a *TokenWhitelistNoOpAdapter) Remove(ctx context.Context, userUID, tid string) error {
	return nil
}

// RemoveAll removes all tokens for a user.
func (a *TokenWhitelistNoOpAdapter) RemoveAll(ctx context.Context, userUID string) error {
	return nil
}

// IsAllowed checks if a token is in the whitelist.
func (a *TokenWhitelistNoOpAdapter) IsAllowed(ctx context.Context, userUID, tid string) (bool, error) {
	return true, nil
}
