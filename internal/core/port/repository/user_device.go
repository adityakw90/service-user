package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

type UserDeviceRepository interface {
	Create(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error)
	Update(ctx context.Context, device *model.UserDevice) error
	Delete(ctx context.Context, device *model.UserDevice) error
	GetByUserIDAndDeviceID(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.UserDeviceListFilterParam) (*model.UserDevices, error)
	Revoke(ctx context.Context, userID, deviceID int64) error
}
