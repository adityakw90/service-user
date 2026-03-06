package repository

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id int64) (*model.Device, error)
	GetByUID(ctx context.Context, uid string) (*model.Device, error)
	GetByFingerprint(ctx context.Context, fingerprint string) (*model.Device, error)
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error)
	ListByUserID(ctx context.Context, userId int64, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error)
}
