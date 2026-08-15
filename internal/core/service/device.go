package service

import (
	"context"
	"fmt"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
)

type deviceService struct {
	deviceRepo     repository.DeviceRepository
	userDeviceRepo repository.UserDeviceRepository
	eventPublisher portEvent.EventPublisher
}

func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	userDeviceRepo repository.UserDeviceRepository,
	eventPublisher portEvent.EventPublisher,
) portSvc.DeviceService {
	return &deviceService{
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
		eventPublisher: eventPublisher,
	}
}

func (s *deviceService) Get(ctx context.Context, uid string) (*model.Device, error) {
	device, err := s.deviceRepo.GetByUID(ctx, uid)
	if err != nil {
		return nil, err
	}

	return device, nil
}

func (s *deviceService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) (*model.Devices, error) {
	// Set defaults for pagination
	if pagination == nil {
		pagination = &param.PaginationParam{
			Page:    util.Ptr(1),
			Limit:   util.Ptr(10),
			Sort:    util.Ptr("asc"),
			OrderBy: util.Ptr("created_at"),
		}
	}

	if filter == nil {
		filter = &param.DeviceListFilterParam{}
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

	// Check if there are active user devices linked to this device
	revoked := false
	userDevices, err := s.userDeviceRepo.List(ctx, &param.PaginationParam{Page: util.Ptr(1), Limit: util.Ptr(1)}, &param.UserDeviceListFilterParam{
		DeviceUids: []string{uid},
		Revoked:    &revoked,
	})
	if err != nil {
		return err
	}
	if userDevices != nil && len(userDevices.Items) > 0 {
		return fmt.Errorf("cannot delete device with %d active user devices linked", userDevices.Meta.Total)
	}

	err = s.deviceRepo.Delete(ctx, device)
	if err != nil {
		return err
	}

	// Publish device deleted event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventDeviceDeleted, Entity: event.NewDeviceEntity(device), Metadata: event.EventDeviceDeletedData{}})
	if err != nil {
		return err
	}

	return nil
}
