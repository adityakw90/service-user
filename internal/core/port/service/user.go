package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserService interface {
	// get user by uid
	Get(ctx context.Context, uid string) (*model.User, error)

	// get list user
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserListFilterParam) (*model.Users, error)

	// create user
	Create(ctx context.Context, param *param.UserCreateParam) (*model.User, error)

	// update user
	Update(ctx context.Context, uid string, param *param.UserUpdateParam) error

	// delete user
	Delete(ctx context.Context, uid string) error

	// Profile operations
	GetProfile(ctx context.Context, userUID string) (*model.UserProfile, error)
	UpdateProfile(ctx context.Context, userUID string, opts param.UserProfileUpdateParam) error

	// PIN operations
	SetPin(ctx context.Context, userUID, pin string) error

	// Device operations
	ListDevice(ctx context.Context, userUID string, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) (*model.Devices, error)
	RevokeDevice(ctx context.Context, userUID, deviceUID string) error

	// Password operations
	ChangePassword(ctx context.Context, userUID string, param *param.UserChangePasswordParam) error
}
