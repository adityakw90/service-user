package security

import (
	"context"

	portsec "github.com/adityakw90/service-user/internal/core/port/security"
)

// NoopAttemptTracker is a no-operation implementation of AttemptTracker.
// Useful for tests or when attempt tracking is disabled.
type NoopAttemptTracker struct{}

// NewNoopAttemptTracker creates a new no-operation attempt tracker.
func NewNoopAttemptTracker() portsec.AttemptTracker {
	return &NoopAttemptTracker{}
}

// Track does nothing.
func (n *NoopAttemptTracker) Track(ctx context.Context, userUID string) error {
	return nil
}

// IsLocked always returns false (never locked).
func (n *NoopAttemptTracker) IsLocked(ctx context.Context, userUID string) (bool, error) {
	return false, nil
}

// Reset does nothing.
func (n *NoopAttemptTracker) Reset(ctx context.Context, userUID string) error {
	return nil
}
