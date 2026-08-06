package security

import (
	"context"
	"fmt"
	"time"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	rateLimitKeyPrefix = "auth:rate_limit"
)

// RedisRateLimiter implements RateLimiter using Redis sorted sets.
// Suitable for production environments with distributed systems.
type RedisRateLimiter struct {
	client     *redis.Client
	limit      int
	windowSize time.Duration
	keyPrefix  string // prefix for Redis keys (e.g., "auth:rate_limit:ip:")
}

// NewRedisRateLimiter creates a new Redis-based rate limiter with configured limit and window.
func NewRedisRateLimiter(client *redis.Client, limit int, windowSize time.Duration) portsec.RateLimiter {
	return &RedisRateLimiter{
		client:     client,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  rateLimitKeyPrefix + ":ip:", // default prefix for IP-based rate limiting
	}
}

// Acquire attempts to acquire a rate limit slot for the given device IP.
// Uses a sliding window algorithm with Redis sorted sets.
func (r *RedisRateLimiter) Acquire(ctx context.Context, deviceIp string) (bool, error) {
	rateLimitKey := r.keyPrefix + deviceIp
	now := float64(time.Now().UnixMilli())
	windowStart := now - float64(r.windowSize.Milliseconds())
	member := uuid.New().String() // Unique member for each request

	// Use a Lua script for atomic operations
	script := `
		local key = KEYS[1]
		local now = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local limit = tonumber(ARGV[3])
		local member = ARGV[4]

		-- Remove entries outside the window
		redis.call("ZREMRANGEBYSCORE", key, 0, window_start)

		-- Count current entries
		local count = redis.call("ZCARD", key)

		-- Check if limit exceeded
		if count >= limit then
			return 0
		end

		-- Add current request with timestamp as score and unique member
		redis.call("ZADD", key, now, member)

		-- Set expiry to window duration
		redis.call("EXPIRE", key, tonumber(ARGV[5]))

		-- Return allowed
		return 1
	`

	result, err := r.client.Eval(ctx, script, []string{rateLimitKey}, now, windowStart, r.limit, member, int(r.windowSize.Seconds())).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Parse result - returns 1 for allowed, 0 for denied
	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected result format from rate limit script")
	}

	return allowed == 1, nil
}

// Reset resets the rate limit counter for a device IP.
func (r *RedisRateLimiter) Reset(ctx context.Context, deviceIp string) error {
	rateLimitKey := r.keyPrefix + deviceIp

	err := r.client.Del(ctx, rateLimitKey).Err()
	if err != nil {
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}
