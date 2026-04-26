package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type UserDeviceRepository interface {
	Create(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error)
	Update(ctx context.Context, device *model.UserDevice) error
	UpdateSessionID(ctx context.Context, userID, deviceID int64, sessionID string) error
	Delete(ctx context.Context, device *model.UserDevice) error
	GetByUserIDAndDeviceID(ctx context.Context, userID int64, deviceID int64) (*model.UserDevice, error)
	List(ctx context.Context, pagination *param.PaginationParam, filter *param.UserDeviceListFilterParam) (*model.UserDevices, error)
	Revoke(ctx context.Context, userID, deviceID int64) error
}
