package security

import (
	"context"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// NoopRateLimiter is a no-operation implementation of RateLimiter.
// Useful for tests or when rate limiting is disabled.
type NoopRateLimiter struct{}

// NewNoopRateLimiter creates a new no-operation rate limiter.
func NewNoopRateLimiter() portsec.RateLimiter {
	return &NoopRateLimiter{}
}

// Acquire always returns true (always allowed).
func (n *NoopRateLimiter) Acquire(ctx context.Context, deviceIp string) (bool, error) {
	return true, nil
}

// Reset does nothing.
func (n *NoopRateLimiter) Reset(ctx context.Context, deviceIp string) error {
	return nil
}
