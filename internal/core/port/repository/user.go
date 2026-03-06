package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id int64) (*model.User, error)
	GetByUID(ctx context.Context, uid string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error)
	AddUserDevice(ctx context.Context, user *model.User, device *model.Device) error
}
