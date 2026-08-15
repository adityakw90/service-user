package security

import (
	"context"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
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
	tracer     gomon.Tracer
	logger     gomon.Logger
}

// NewRedisRateLimiter creates a new Redis-based rate limiter with configured limit and window.
func NewRedisRateLimiter(client *redis.Client, limit int, windowSize time.Duration, tracer gomon.Tracer, logger gomon.Logger) portsec.RateLimiter {
	if client == nil {
		panic("redis client is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &RedisRateLimiter{
		client:     client,
		limit:      limit,
		windowSize: windowSize,
		keyPrefix:  rateLimitKeyPrefix + ":ip:", // default prefix for IP-based rate limiting
		tracer:     tracer,
		logger:     logger,
	}
}

// Acquire attempts to acquire a rate limit slot for the given device IP.
// Uses a sliding window algorithm with Redis sorted sets.
func (r *RedisRateLimiter) Acquire(ctx context.Context, deviceIp string) (bool, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.RateLimiter.Acquire")
	defer span.End()

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

	result, err := r.client.Eval(newCtx, script, []string{rateLimitKey}, now, windowStart, r.limit, member, int(r.windowSize.Seconds())).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis rate limiter acquire failed", map[string]any{"error": err, "deviceIp": truncateID(deviceIp)})
		}
		return false, fmt.Errorf("failed to check rate limit: %w", err)
	}

	// Parse result - returns 1 for allowed, 0 for denied
	allowed, ok := result.(int64)
	if !ok {
		if r.logger != nil {
			r.logger.Error("redis rate limiter unexpected result format", map[string]any{"deviceIp": truncateID(deviceIp), "result": result})
		}
		return false, fmt.Errorf("unexpected result format from rate limit script")
	}

	return allowed == 1, nil
}

// Reset resets the rate limit counter for a device IP.
func (r *RedisRateLimiter) Reset(ctx context.Context, deviceIp string) error {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.RateLimiter.Reset")
	defer span.End()

	rateLimitKey := r.keyPrefix + deviceIp

	err := r.client.Del(newCtx, rateLimitKey).Err()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis rate limiter reset failed", map[string]any{"error": err, "deviceIp": truncateID(deviceIp)})
		}
		return fmt.Errorf("failed to reset rate limit: %w", err)
	}

	return nil
}
