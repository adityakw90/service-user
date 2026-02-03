package repository

import (
	"context"
	"sync"
	"time"

	monitoring "github.com/adityakw90/go-monitoring"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func mapperID[T any, KS comparable, KT comparable](
	ctx context.Context,
	rTracer monitoring.Tracer,
	rLogger monitoring.Logger,
	redisClient *redis.Client,
	keys []KS,
	convertResult func(string) KT,
	cacheKeyFunc func(KS) string,
	dbFetchFunc func(KS) (*T, error),
	getValueFromStruct func(*T) KT,
	cacheDuration time.Duration,
) (map[KS]KT, error) {
	// Tracer and logger setup
	ctx, span := rTracer.StartSpan(ctx, "repository.mapperID")
	defer span.End()

	logger := rLogger.WithSpanContext(span.SpanContext())

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
	span.AddEvent("Execute Redis pipeline")
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		span.AddEvent("Redis pipeline failed", trace.WithAttributes(
			attribute.String("error.message", err.Error()),
		))
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
		span.AddEvent("All keys cached")
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

			newCtx, bgSpan := rTracer.StartChildSpan(
				ctx,
				"repository.mapperID.Child.Async",
				span,
			)
			defer bgSpan.End()

			// get logger with spancontext
			loggerBg := rLogger.WithSpanContext(bgSpan.SpanContext())

			// Fetch from DB
			bgSpan.AddEvent("Fetch from DB")
			value, err := dbFetchFunc(key)
			if err != nil {
				errChan <- err
				return
			}

			// Cache the value
			cacheKey := cacheKeyFunc(key)
			cacheValue := getValueFromStruct(value)
			bgSpan.AddEvent("Set cache", trace.WithAttributes(
				attribute.String("key", cacheKey),
			))
			if err := redisClient.Set(newCtx, cacheKey, cacheValue, cacheDuration).Err(); err != nil {
				loggerBg.Error("Failed to set cache", map[string]interface{}{
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
