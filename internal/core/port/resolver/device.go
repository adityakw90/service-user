package resolver

import (
	"context"
)

type DeviceResolver interface {
	IDsByUIDs(ctx context.Context, deviceUIDs []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, deviceIDs []int64) (map[int64]string, error)
}
