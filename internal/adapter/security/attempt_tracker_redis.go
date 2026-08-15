package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key patterns
	failedAttemptsKey = "auth:failed_attempts:%s"
	lockoutKey        = "auth:lockout:%s"
)

// RedisAttemptTracker implements AttemptTracker using Redis.
// Suitable for production environments with distributed systems.
type RedisAttemptTracker struct {
	client          *redis.Client
	threshold       int
	lockoutDuration time.Duration
	counterTTL      time.Duration
	tracer          gomon.Tracer
	logger          gomon.Logger
}

// NewRedisAttemptTracker creates a new Redis-based attempt tracker.
func NewRedisAttemptTracker(client *redis.Client, threshold int, lockoutDuration, counterTTL time.Duration, tracer gomon.Tracer, logger gomon.Logger) portsec.AttemptTracker {
	if client == nil {
		panic("redis client is required")
	}
	if tracer == nil {
		panic("tracer is required")
	}
	return &RedisAttemptTracker{
		client:          client,
		threshold:       threshold,
		lockoutDuration: lockoutDuration,
		counterTTL:      counterTTL,
		tracer:          tracer,
		logger:          logger,
	}
}

// Track records a failed attempt for the user.
func (r *RedisAttemptTracker) Track(ctx context.Context, userUID string) error {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.AttemptTracker.Track")
	defer span.End()

	key := fmt.Sprintf(failedAttemptsKey, userUID)

	// Increment counter atomically
	count, err := r.client.Incr(newCtx, key).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker incr failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return fmt.Errorf("failed to track attempt: %w", err)
	}

	// Set TTL on first attempt or refresh TTL on subsequent attempts
	// This ensures the counter expires after the configured TTL
	if count == 1 {
		// First attempt, set TTL
		err = r.client.Expire(newCtx, key, r.counterTTL).Err()
	} else {
		// Refresh TTL on subsequent attempts
		err = r.client.Expire(newCtx, key, r.counterTTL).Err()
	}

	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker expire failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return fmt.Errorf("failed to set counter TTL: %w", err)
	}

	// Check if threshold reached
	if count >= int64(r.threshold) {
		// Set lockout
		lockoutKey := fmt.Sprintf(lockoutKey, userUID)
		err = r.client.Set(newCtx, lockoutKey, 1, r.lockoutDuration).Err()
		if err != nil {
			if r.logger != nil {
				r.logger.Error("redis attempt tracker lockout set failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
			}
			return fmt.Errorf("failed to set lockout: %w", err)
		}
		return fmt.Errorf("account locked after %d failed attempts", count)
	}

	return nil
}

// IsLocked checks if a user account is currently locked.
func (r *RedisAttemptTracker) IsLocked(ctx context.Context, userUID string) (bool, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.AttemptTracker.IsLocked")
	defer span.End()

	lockoutKey := fmt.Sprintf(lockoutKey, userUID)

	// Check if lockout key exists and has not expired
	exists, err := r.client.Exists(newCtx, lockoutKey).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker islocked failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return false, fmt.Errorf("failed to check lockout status: %w", err)
	}

	return exists > 0, nil
}

// Reset clears the failed attempt counter for the user.
func (r *RedisAttemptTracker) Reset(ctx context.Context, userUID string) error {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.AttemptTracker.Reset")
	defer span.End()

	// Delete both the counter and lockout keys
	keys := []string{
		fmt.Sprintf(failedAttemptsKey, userUID),
		fmt.Sprintf(lockoutKey, userUID),
	}

	err := r.client.Del(newCtx, keys...).Err()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker reset failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return fmt.Errorf("failed to reset attempt tracker: %w", err)
	}

	return nil
}

// GetFailedAttempts returns the current failed attempt count for a user.
// This is a helper method for monitoring/debugging.
func (r *RedisAttemptTracker) GetFailedAttempts(ctx context.Context, userUID string) (int, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.AttemptTracker.GetFailedAttempts")
	defer span.End()

	key := fmt.Sprintf(failedAttemptsKey, userUID)

	val, err := r.client.Get(newCtx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker getfailedattempts failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return 0, fmt.Errorf("failed to get failed attempts: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker parse failed attempts failed", map[string]any{"error": err, "userUID": truncateID(userUID), "value": val})
		}
		return 0, fmt.Errorf("failed to parse failed attempts: %w", err)
	}

	return count, nil
}

// GetLockoutRemaining returns the remaining lockout duration for a user.
// This is a helper method for monitoring/debugging.
func (r *RedisAttemptTracker) GetLockoutRemaining(ctx context.Context, userUID string) (time.Duration, error) {
	newCtx, span := r.tracer.StartSpan(ctx, "redis.AttemptTracker.GetLockoutRemaining")
	defer span.End()

	lockoutKey := fmt.Sprintf(lockoutKey, userUID)

	ttl, err := r.client.TTL(newCtx, lockoutKey).Result()
	if err != nil {
		if r.logger != nil {
			r.logger.Error("redis attempt tracker getlockoutremaining failed", map[string]any{"error": err, "userUID": truncateID(userUID)})
		}
		return 0, fmt.Errorf("failed to get lockout remaining: %w", err)
	}

	// TTL returns -2 if key doesn't exist, -1 if key has no expiry
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}
