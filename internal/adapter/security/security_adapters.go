package security

import (
	"context"
	"fmt"
	"time"

	gomon "github.com/adityakw90/go-monitoring"
	portsec "github.com/adityakw90/service-user/internal/core/port/security"
	"github.com/redis/go-redis/v9"
)

// truncateID returns a safe log-correlation prefix of an identifier.
// Only the first 8 characters are retained so that raw security identifiers
// (userUID, tid, device IP) are never written to logs verbatim.
func truncateID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "…"
}

// AttemptTrackerConfig holds configuration for an attempt tracker.
type AttemptTrackerConfig struct {
	Backend           string        // "redis" or "memory"
	LockoutThreshold  int           // Number of failed attempts before lockout
	LockoutDuration   time.Duration // How long to lock out the account
	LockoutCounterTTL time.Duration // How long to keep failed attempt counters
}

// RateLimiterConfig holds configuration for the rate limiter.
type RateLimiterConfig struct {
	Backend    string        // "redis" or "memory"
	Limit      int           // Maximum number of requests
	WindowSize time.Duration // Time window for rate limiting
}

// SecurityConfig holds configuration for security adapters.
type SecurityConfig struct {
	// Login attempt tracker settings
	LoginTracker AttemptTrackerConfig

	// PIN attempt tracker settings
	PINTracker AttemptTrackerConfig

	// Rate limiting settings
	RateLimiter RateLimiterConfig
}

// SecurityAdapters contains the security adapter instances.
type SecurityAdapters struct {
	LoginTracker portsec.AttemptTracker
	PINTracker   portsec.AttemptTracker
	RateLimiter  portsec.RateLimiter
	closeFunc    func() error
}

// Close releases any resources held by the security adapters.
func (s *SecurityAdapters) Close() error {
	if s.closeFunc != nil {
		return s.closeFunc()
	}
	return nil
}

// NewSecurityAdapters creates security adapters based on the provided configuration.
func NewSecurityAdapters(ctx context.Context, cfg SecurityConfig, redisClient *redis.Client, tracer gomon.Tracer, logger gomon.Logger) (*SecurityAdapters, error) {
	// Determine if we need Redis connection for any adapter
	needsRedis := cfg.LoginTracker.Backend == "redis" ||
		cfg.PINTracker.Backend == "redis" ||
		cfg.RateLimiter.Backend == "redis"

	if needsRedis && redisClient == nil {
		return nil, fmt.Errorf("redis client is required for redis backend")
	}

	if needsRedis && tracer == nil {
		return nil, fmt.Errorf("tracer is required for redis backend")
	}

	// Create login tracker
	loginTracker, err := newAttemptTracker(ctx, redisClient, cfg.LoginTracker, tracer, logger)
	if err != nil {
		if redisClient != nil {
			redisClient.Close()
		}
		return nil, fmt.Errorf("failed to create login tracker: %w", err)
	}

	// Create PIN tracker
	pinTracker, err := newAttemptTracker(ctx, redisClient, cfg.PINTracker, tracer, logger)
	if err != nil {
		if redisClient != nil {
			redisClient.Close()
		}
		return nil, fmt.Errorf("failed to create PIN tracker: %w", err)
	}

	// Create rate limiter
	rateLimiter, err := newRateLimiter(ctx, redisClient, cfg.RateLimiter, tracer, logger)
	if err != nil {
		if redisClient != nil {
			redisClient.Close()
		}
		return nil, fmt.Errorf("failed to create rate limiter: %w", err)
	}

	// Build close function
	closeFunc := func() error {
		// Close memory trackers to stop cleanup goroutines
		if tracker, ok := loginTracker.(*MemoryAttemptTracker); ok {
			tracker.Close()
		}
		if tracker, ok := pinTracker.(*MemoryAttemptTracker); ok {
			tracker.Close()
		}
		// Close Redis connection if it was created
		if redisClient != nil {
			return redisClient.Close()
		}
		return nil
	}

	return &SecurityAdapters{
		LoginTracker: loginTracker,
		PINTracker:   pinTracker,
		RateLimiter:  rateLimiter,
		closeFunc:    closeFunc,
	}, nil
}

// newAttemptTracker creates an attempt tracker based on the backend configuration.
func newAttemptTracker(ctx context.Context, client *redis.Client, cfg AttemptTrackerConfig, tracer gomon.Tracer, logger gomon.Logger) (portsec.AttemptTracker, error) {
	switch cfg.Backend {
	case "memory":
		return NewMemoryAttemptTracker(
			cfg.LockoutThreshold,
			cfg.LockoutDuration,
			cfg.LockoutCounterTTL,
		), nil
	case "redis":
		if client == nil {
			return nil, fmt.Errorf("redis client is required for redis backend")
		}
		return NewRedisAttemptTracker(
			client,
			cfg.LockoutThreshold,
			cfg.LockoutDuration,
			cfg.LockoutCounterTTL,
			tracer,
			logger,
		), nil
	default:
		return nil, fmt.Errorf("unknown attempt tracker backend: %s (must be 'memory' or 'redis')", cfg.Backend)
	}
}

// newRateLimiter creates a rate limiter based on the backend configuration.
func newRateLimiter(ctx context.Context, client *redis.Client, cfg RateLimiterConfig, tracer gomon.Tracer, logger gomon.Logger) (portsec.RateLimiter, error) {
	switch cfg.Backend {
	case "memory":
		return NewMemoryRateLimiter(
			cfg.Limit,
			cfg.WindowSize,
		), nil
	case "redis":
		if client == nil {
			return nil, fmt.Errorf("redis client is required for redis backend")
		}
		return NewRedisRateLimiter(
			client,
			cfg.Limit,
			cfg.WindowSize,
			tracer,
			logger,
		), nil
	default:
		return nil, fmt.Errorf("unknown rate limiter backend: %s (must be 'memory' or 'redis')", cfg.Backend)
	}
}
