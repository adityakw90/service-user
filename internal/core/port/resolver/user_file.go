package resolver

import (
	"context"
)

type UserFileResolver interface {
	IDsByUIDs(ctx context.Context, userFileUIDs []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, userFileIDs []int64) (map[int64]string, error)
}
