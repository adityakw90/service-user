package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

// SetupTestRedis flushes all test keys from Redis.
// Use this to clean Redis between tests.
func SetupTestRedis(ctx context.Context, client *redis.Client) error {
	// Flush all keys with the test prefix
	iter := client.Scan(ctx, 0, "test:*", 100).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// FlushTestKeys deletes keys matching a specific pattern.
func FlushTestKeys(ctx context.Context, client *redis.Client, pattern string) error {
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		if err := client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("failed to delete keys: %w", err)
		}
	}

	return nil
}

// AssertKeyExists verifies that a key exists in Redis.
func AssertKeyExists(t require.TestingT, ctx context.Context, client *redis.Client, key string) {
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.True(t, exists > 0, "key %s should exist", key)
}

// AssertKeyNotExists verifies that a key does not exist in Redis.
func AssertKeyNotExists(t require.TestingT, ctx context.Context, client *redis.Client, key string) {
	exists, err := client.Exists(ctx, key).Result()
	require.NoError(t, err)
	require.Zero(t, exists, "key %s should not exist", key)
}

// AssertKeyValue verifies that a key has a specific value.
func AssertKeyValue(t require.TestingT, ctx context.Context, client *redis.Client, key, expectedValue string) {
	value, err := client.Get(ctx, key).Result()
	require.NoError(t, err)
	require.Equal(t, expectedValue, value)
}

// GetTokenWhitelistCount counts the number of entries in the token whitelist.
func GetTokenWhitelistCount(ctx context.Context, client *redis.Client, prefix string) (int, error) {
	iter := client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	count := 0

	for iter.Next(ctx) {
		count++
	}

	if err := iter.Err(); err != nil {
		return 0, fmt.Errorf("failed to scan keys: %w", err)
	}

	return count, nil
}

// WaitForRedis polls for Redis readiness.
func WaitForRedis(ctx context.Context, client *redis.Client, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		if err := client.Ping(ctx).Err(); err == nil {
			return nil
		}

		// Wait before retrying
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}

	return fmt.Errorf("redis not ready after %d attempts", maxAttempts)
}

// GetKeysByPattern returns all keys matching a pattern.
func GetKeysByPattern(ctx context.Context, client *redis.Client, pattern string) ([]string, error) {
	iter := client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan keys: %w", err)
	}

	return keys, nil
}

// FlushAllTestKeys flushes all keys with test prefix (wildcard).
// Use with caution - this will delete all keys matching "test:*".
func FlushAllTestKeys(ctx context.Context, client *redis.Client) error {
	iter := client.Scan(ctx, 0, "test:*", 1000).Iterator()
	keys := make([]string, 0)

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		return fmt.Errorf("failed to scan keys: %w", err)
	}

	if len(keys) > 0 {
		// Delete in batches of 100 to avoid command size limits
		for i := 0; i < len(keys); i += 100 {
			end := i + 100
			if end > len(keys) {
				end = len(keys)
			}
			if err := client.Del(ctx, keys[i:end]...).Err(); err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
		}
	}

	return nil
}
