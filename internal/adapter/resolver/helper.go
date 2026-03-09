package resolver

import (
	"context"
	"strconv"
	"sync"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/redis/go-redis/v9"
)

// mapperID is a generic helper for bidirectional ID/UID mapping with caching.
// Uses Redis pipeline for batch GET operations and concurrent DB fetches.
// Copied from service-access for consistency across microservices.
func mapperID[T any, KS comparable, KT comparable](
	ctx context.Context,
	logger monitoring.Logger,
	redisClient *redis.Client,
	keys []KS,
	convertResult func(string) KT,
	cacheKeyFunc func(KS) string,
	dbFetchFunc func(KS) (*T, error),
	getValueFromStruct func(*T) KT,
	cacheDuration time.Duration,
) (map[KS]KT, error) {
	// Result map
	results := make(map[KS]KT)
	var uncachedKeys []KS

	// Redis pipeline to batch GET requests
	pipe := redisClient.Pipeline()
	cacheResults := make([]*redis.StringCmd, len(keys))

	for i, key := range keys {
		cacheKey := cacheKeyFunc(key)
		cacheResults[i] = pipe.Get(ctx, cacheKey)
	}

	// Execute the Redis pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// Process the results, identify cache misses
	for i, result := range cacheResults {
		cachedValue, err := result.Result()
		switch err {
		case nil:
			// Convert the cached value and store it
			results[keys[i]] = convertResult(cachedValue)
		case redis.Nil:
			// Cache miss, add the key to uncachedKeys
			uncachedKeys = append(uncachedKeys, keys[i])
		default:
			// Log Redis errors but continue
			logger.Error("Redis error", map[string]interface{}{
				"error.message": err.Error(),
				"key":           keys[i],
			})
		}
	}

	// If all keys were cached, return the result
	if len(uncachedKeys) == 0 {
		return results, nil
	}

	// Use goroutines to fetch uncached keys from the database
	errChan := make(chan error, len(uncachedKeys))
	var wg sync.WaitGroup
	mu := sync.Mutex{} // Protects shared access to the map

	for _, key := range uncachedKeys {
		wg.Add(1)
		go func(key KS) {
			defer wg.Done()

			// use new context with timeout
			newCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
			defer cancel()

			// Fetch from DB
			value, err := dbFetchFunc(key)
			if err != nil {
				errChan <- err
				return
			}

			// Cache the value
			cacheKey := cacheKeyFunc(key)
			cacheValue := getValueFromStruct(value)
			if err := redisClient.Set(newCtx, cacheKey, cacheValue, cacheDuration).Err(); err != nil {
				// log redis error but allowed
				logger.Error("Failed to set cache", map[string]interface{}{
					"error.message": err.Error(),
				})
			}

			// Store the result
			mu.Lock()
			results[key] = cacheValue
			mu.Unlock()

			errChan <- nil
		}(key)
	}

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check if any error was returned
	for err := range errChan {
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

// helper for invalidate cache
// Uses Redis pipeline for efficient batch operations.
func invalidate(
	ctx context.Context,
	redisClient *redis.Client,
	redisPrefix string,
	opts ...param.InvalidateOpt,
) error {
	// Parse options
	options := &param.InvalidateOptions{}
	for _, opt := range opts {
		opt(options)
	}

	if len(options.UIDs) == 0 && len(options.IDs) == 0 {
		return nil
	}

	// Use a map to avoid duplicate keys (e.g., if both UID and ID are passed)
	keysToDelete := make(map[string]struct{}, (len(options.UIDs)+len(options.IDs))*2)

	// Pipeline to batch all GET operations
	pipe := redisClient.Pipeline()

	// Queue GET commands for UIDs to find their corresponding IDs
	uidGetCmds := make([]*redis.StringCmd, len(options.UIDs))
	for i, uid := range options.UIDs {
		uidKey := redisPrefix + ":" + uid + ":id"
		uidGetCmds[i] = pipe.Get(ctx, uidKey)
	}

	// Queue GET commands for IDs to find their corresponding UIDs
	idGetCmds := make([]*redis.StringCmd, len(options.IDs))
	for i, id := range options.IDs {
		idStr := strconv.FormatInt(id, 10)
		idKey := redisPrefix + ":id:" + idStr + ":uid"
		idGetCmds[i] = pipe.Get(ctx, idKey)
	}

	// Execute the pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Log error but continue - some keys might still be cached
		// We'll delete the keys we know about (forward/reverse mappings)
	}

	// Process UID results - add forward mapping and lookup reverse mapping
	for i, uid := range options.UIDs {
		uidKey := redisPrefix + ":" + uid + ":id"
		keysToDelete[uidKey] = struct{}{}

		// Try to get the ID from cache to build the reverse key
		idStr, err := uidGetCmds[i].Result()
		if err == nil && idStr != "" {
			// ID exists in cache, also delete the reverse mapping key
			idKey := redisPrefix + ":id:" + idStr + ":uid"
			keysToDelete[idKey] = struct{}{}
		}
	}

	// Process ID results - add reverse mapping and lookup forward mapping
	for i, id := range options.IDs {
		idStr := strconv.FormatInt(id, 10)
		idKey := redisPrefix + ":id:" + idStr + ":uid"
		keysToDelete[idKey] = struct{}{}

		// Try to get the UID from cache to build the forward key
		uidStr, err := idGetCmds[i].Result()
		if err == nil && uidStr != "" {
			// UID exists in cache, also delete the forward mapping key
			uidKey := redisPrefix + ":" + uidStr + ":id"
			keysToDelete[uidKey] = struct{}{}
		}
	}

	if len(keysToDelete) == 0 {
		return nil
	}

	// Convert map to slice for DEL command
	keys := make([]string, 0, len(keysToDelete))
	for key := range keysToDelete {
		keys = append(keys, key)
	}

	return redisClient.Del(ctx, keys...).Err()
}
