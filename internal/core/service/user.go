package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type userService struct {
}

func NewUserService() portSvc.UserService {
	return &userService{}
}

func (s *userService) Get(ctx context.Context, uid string) (*model.User, error) {
	return nil, nil
}

func (s *userService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserListFilterParam) (*model.Users, error) {
	return nil, nil
}

func (s *userService) Create(ctx context.Context, param *params.UserCreateParam) (*model.User, error) {
	return nil, nil
}

func (s *userService) Update(ctx context.Context, uid string, param *params.UserUpdateParam) error {
	return nil
}

func (s *userService) Delete(ctx context.Context, uid string) error {
	return nil
}

func (s *userService) GetProfile(ctx context.Context, userUID string) (*model.UserProfile, error) {
	return nil, nil
}

func (s *userService) UpdateProfile(ctx context.Context, userUID string, opts params.UserProfileUpdateParam) error {
	return nil
}

func (s *userService) SetPin(ctx context.Context, userUID, pin string) error {
	return nil
}

func (s *userService) ListDevice(ctx context.Context, userUID string, opts params.UserDeviceListFilterParam) (*model.Devices, error) {
	return nil, nil
}

func (s *userService) RevokeDevice(ctx context.Context, userUID, deviceUID string) error {
	return nil
}
