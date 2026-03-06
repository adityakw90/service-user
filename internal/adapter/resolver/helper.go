package resolver

import (
	"context"
	"fmt"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	"github.com/redis/go-redis/v9"
)

type identity struct {
	id  int64
	uid string
}

// mapperID is a generic helper for bidirectional ID/UID mapping with caching.
// Uses Redis pipeline for batch GET operations and concurrent DB fetches.
func mapperID[T comparable, U any](
	ctx context.Context,
	logger monitoring.Logger,
	redisClient *redis.Client,
	input []T,
	parseResult func(string) U,
	keyFunc func(T) string,
	fetchFunc func(context.Context, T) (*identity, error),
	extractFunc func(*identity) U,
	cacheDuration time.Duration,
) (map[T]U, error) {
	result := make(map[T]U)
	missing := make([]T, 0, len(input))
	missingIdx := make(map[T]int, len(input))

	// Try cache first using Redis pipeline for batch GET
	pipe := redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(input))

	for i, item := range input {
		key := keyFunc(item)
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Pipeline failed, try individual GETs as fallback
		for _, item := range input {
			key := keyFunc(item)
			cached, err := redisClient.Get(ctx, key).Result()
			if err == nil && cached != "" {
				result[item] = parseResult(cached)
			} else {
				missing = append(missing, item)
				missingIdx[item] = len(missing) - 1
			}
		}
	} else {
		// Pipeline succeeded
		for i, item := range input {
			cached, _ := cmds[i].Result()
			if cached != "" {
				result[item] = parseResult(cached)
			} else {
				missing = append(missing, item)
				missingIdx[item] = len(missing) - 1
			}
		}
	}

	// Fetch missing from DB concurrently
	if len(missing) > 0 {
		type fetchResult struct {
			item T
			id   *identity
			err  error
		}

		results := make(chan fetchResult, len(missing))

		for _, item := range missing {
			go func(it T) {
				// Create context with timeout for each fetch
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				id, err := fetchFunc(ctx, it)
				results <- fetchResult{item: it, id: id, err: err}
			}(item)
		}

		for range missing {
			fr := <-results
			if fr.err != nil {
				return nil, fr.err
			}

			// Cache the result
			key := keyFunc(fr.item)
			value := fmt.Sprintf("%d", fr.id.id)
			if err := redisClient.Set(ctx, key, value, cacheDuration).Err(); err != nil {
				logger.Debug("Failed to cache", map[string]interface{}{
					"key":   key,
					"error": err.Error(),
				})
			}

			result[fr.item] = extractFunc(fr.id)
		}
	}

	return result, nil
}

// mapperUID is a generic helper for ID to UID mapping with caching.
func mapperUID[T comparable, U any](
	ctx context.Context,
	logger monitoring.Logger,
	redisClient *redis.Client,
	input []T,
	parseResult func(string) U,
	keyFunc func(T) string,
	fetchFunc func(context.Context, T) (*identity, error),
	extractFunc func(*identity) U,
	cacheDuration time.Duration,
) (map[T]U, error) {
	result := make(map[T]U)
	missing := make([]T, 0, len(input))

	// Try cache first using Redis pipeline
	pipe := redisClient.Pipeline()
	cmds := make([]*redis.StringCmd, len(input))

	for i, item := range input {
		key := keyFunc(item)
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// Pipeline failed, try individual GETs
		for _, item := range input {
			key := keyFunc(item)
			cached, err := redisClient.Get(ctx, key).Result()
			if err == nil && cached != "" {
				result[item] = parseResult(cached)
			} else {
				missing = append(missing, item)
			}
		}
	} else {
		for i, item := range input {
			cached, _ := cmds[i].Result()
			if cached != "" {
				result[item] = parseResult(cached)
			} else {
				missing = append(missing, item)
			}
		}
	}

	// Fetch missing from DB concurrently
	if len(missing) > 0 {
		type fetchResult struct {
			item T
			id   *identity
			err  error
		}

		results := make(chan fetchResult, len(missing))

		for _, item := range missing {
			go func(it T) {
				ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
				defer cancel()

				id, err := fetchFunc(ctx, it)
				results <- fetchResult{item: it, id: id, err: err}
			}(item)
		}

		for range missing {
			fr := <-results
			if fr.err != nil {
				return nil, fr.err
			}

			key := keyFunc(fr.item)
			value := fr.id.uid
			if err := redisClient.Set(ctx, key, value, cacheDuration).Err(); err != nil {
				logger.Debug("Failed to cache", map[string]interface{}{
					"key":   key,
					"error": err.Error(),
				})
			}

			result[fr.item] = extractFunc(fr.id)
		}
	}

	return result, nil
}
