package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type deviceService struct {
}

func NewDeviceService() portSvc.DeviceService {
	return &deviceService{}
}

func (s *deviceService) Get(ctx context.Context, uid string) (*model.Device, error) {
	return nil, nil
}

func (s *deviceService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	return nil, nil
}

func (s *deviceService) Delete(ctx context.Context, uid string) error {
	return nil
}
