package security

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
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
}

// NewRedisAttemptTracker creates a new Redis-based attempt tracker.
func NewRedisAttemptTracker(client *redis.Client, threshold int, lockoutDuration, counterTTL time.Duration) portsec.AttemptTracker {
	return &RedisAttemptTracker{
		client:          client,
		threshold:       threshold,
		lockoutDuration: lockoutDuration,
		counterTTL:      counterTTL,
	}
}

// Track records a failed attempt for the user.
func (r *RedisAttemptTracker) Track(ctx context.Context, userUID string) error {
	key := fmt.Sprintf(failedAttemptsKey, userUID)

	// Increment counter atomically
	count, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to track attempt: %w", err)
	}

	// Set TTL on first attempt or refresh TTL on subsequent attempts
	// This ensures the counter expires after the configured TTL
	if count == 1 {
		// First attempt, set TTL
		err = r.client.Expire(ctx, key, r.counterTTL).Err()
	} else {
		// Refresh TTL on subsequent attempts
		err = r.client.Expire(ctx, key, r.counterTTL).Err()
	}

	if err != nil {
		return fmt.Errorf("failed to set counter TTL: %w", err)
	}

	// Check if threshold reached
	if count >= int64(r.threshold) {
		// Set lockout
		lockoutKey := fmt.Sprintf(lockoutKey, userUID)
		err = r.client.Set(ctx, lockoutKey, 1, r.lockoutDuration).Err()
		if err != nil {
			return fmt.Errorf("failed to set lockout: %w", err)
		}
		return fmt.Errorf("account locked after %d failed attempts", count)
	}

	return nil
}

// IsLocked checks if a user account is currently locked.
func (r *RedisAttemptTracker) IsLocked(ctx context.Context, userUID string) (bool, error) {
	lockoutKey := fmt.Sprintf(lockoutKey, userUID)

	// Check if lockout key exists and has not expired
	exists, err := r.client.Exists(ctx, lockoutKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check lockout status: %w", err)
	}

	return exists > 0, nil
}

// Reset clears the failed attempt counter for the user.
func (r *RedisAttemptTracker) Reset(ctx context.Context, userUID string) error {
	// Delete both the counter and lockout keys
	keys := []string{
		fmt.Sprintf(failedAttemptsKey, userUID),
		fmt.Sprintf(lockoutKey, userUID),
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to reset attempt tracker: %w", err)
	}

	return nil
}

// GetFailedAttempts returns the current failed attempt count for a user.
// This is a helper method for monitoring/debugging.
func (r *RedisAttemptTracker) GetFailedAttempts(ctx context.Context, userUID string) (int, error) {
	key := fmt.Sprintf(failedAttemptsKey, userUID)

	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get failed attempts: %w", err)
	}

	count, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("failed to parse failed attempts: %w", err)
	}

	return count, nil
}

// GetLockoutRemaining returns the remaining lockout duration for a user.
// This is a helper method for monitoring/debugging.
func (r *RedisAttemptTracker) GetLockoutRemaining(ctx context.Context, userUID string) (time.Duration, error) {
	lockoutKey := fmt.Sprintf(lockoutKey, userUID)

	ttl, err := r.client.TTL(ctx, lockoutKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get lockout remaining: %w", err)
	}

	// TTL returns -2 if key doesn't exist, -1 if key has no expiry
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}
