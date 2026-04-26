package security

import "context"

// AttemptTracker tracks failed attempts and manages lockout.
type AttemptTracker interface {
	// Track records a failed attempt for the user.
	Track(ctx context.Context, userUID string) error

	// IsLocked checks if the user is currently locked out.
	IsLocked(ctx context.Context, userUID string) (bool, error)

	// Reset clears the failed attempt counter for the user.
	Reset(ctx context.Context, userUID string) error
}
