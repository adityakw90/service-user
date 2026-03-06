package resolver

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// UserResolver handles UID <-> ID resolution with caching.
type UserResolver interface {
	// IDsByUIDs resolves user UIDs to their internal IDs.
	// Returns a map of UID to ID.
	// Returns domain error if any UID is not found.
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)

	// UIDsByIDs resolves user internal IDs to their UIDs.
	// Returns a map of ID to UID.
	// Returns domain error if any ID is not found.
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)

	// Invalidate clears cached entries for the specified UIDs/IDs.
	// Accepts functional options: WithUIDs(), WithIDs()
	Invalidate(ctx context.Context, opts ...params.InvalidateOpt) error
}
