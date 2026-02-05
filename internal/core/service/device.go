package service

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
)

type deviceService struct {
	deviceRepo     repository.DeviceRepository
	userDeviceRepo repository.UserDeviceRepository
}

func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	userDeviceRepo repository.UserDeviceRepository,
) portSvc.DeviceService {
	return &deviceService{
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
	}
}

func (s *deviceService) Get(ctx context.Context, uid string) (*model.Device, error) {
	device, err := s.deviceRepo.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}
	return device, nil
}

func (s *deviceService) List(ctx context.Context, pagination *params.PaginationParam, filter *params.DeviceListFilterParam) (*model.Devices, error) {
	// Set defaults for pagination
	if pagination == nil {
		pagination = params.NewPaginationParam(1, 10, "created_at", "desc")
	}

	if filter == nil {
		filter = &params.DeviceListFilterParam{}
	}

	devices, err := s.deviceRepo.List(ctx, pagination, filter)
	if err != nil {
		return nil, err
	}

	return devices, nil
}

func (s *deviceService) Delete(ctx context.Context, uid string) error {
	// Get device
	device, err := s.deviceRepo.GetByUID(ctx, uid)
	if err != nil {
		return err
	}

	// TODO: Check if there are active user devices linked to this device
	// For now, just delete the device
	return s.deviceRepo.Delete(ctx, device)
}
