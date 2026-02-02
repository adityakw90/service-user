package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

type UserProfileRepository interface {
	Create(ctx context.Context, profile *model.UserProfile) (*model.UserProfile, error)
	Update(ctx context.Context, profile *model.UserProfile) error
	Delete(ctx context.Context, profile *model.UserProfile) error
	GetByUserID(ctx context.Context, id int64) (*model.UserProfile, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserProfileListFilterParam) (*model.UserProfiles, error)
}
