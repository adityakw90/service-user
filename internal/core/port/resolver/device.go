package resolver

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// DeviceResolver handles device UID <-> ID resolution with caching.
type DeviceResolver interface {
	IDsByUIDs(ctx context.Context, uids []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, ids []int64) (map[int64]string, error)
	Invalidate(ctx context.Context, opts ...params.InvalidateOpt) error
}
