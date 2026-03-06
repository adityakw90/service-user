package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
)

type DeviceService interface {
	// get device by uid
	Get(ctx context.Context, uid string) (*model.Device, error)

	// get list device
	List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error)

	// delete device
	Delete(ctx context.Context, uid string) error
}
