package security

import "context"

// RateLimiter is the secondary port for rate limiting.
// Implements sliding window rate limiting for authentication attempts.
type RateLimiter interface {
	// Acquire attempts to acquire a rate limit slot for the given device IP.
	// Returns whether the request is allowed (within rate limit).
	// This operation atomically checks and increments the counter.
	Acquire(ctx context.Context, deviceIp string) (bool, error)

	// Reset resets the rate limit counter for a device IP.
	Reset(ctx context.Context, deviceIp string) error
}
