package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserPinRepository interface {
	Create(ctx context.Context, userPin *model.UserPin) (*model.UserPin, error)
	Update(ctx context.Context, userPin *model.UserPin) error
	Delete(ctx context.Context, userPin *model.UserPin) error
	GetByUserID(ctx context.Context, id int64) (*model.UserPin, error)
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserPinListFilterParam) (*model.UserPins, error)
}
