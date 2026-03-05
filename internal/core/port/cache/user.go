package cache

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

type UserCache interface {
	Get(ctx context.Context, uid string, fallback func(ctx context.Context, uid string) (*model.User, error)) (*model.User, error)
}
