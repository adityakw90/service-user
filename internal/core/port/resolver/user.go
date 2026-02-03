package resolver

import (
	"context"
)

type UserResolver interface {
	IDsByUIDs(ctx context.Context, userUIDs []string) (map[string]int64, error)
	UIDsByIDs(ctx context.Context, userIDs []int64) (map[int64]string, error)
}
