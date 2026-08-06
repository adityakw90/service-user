package service

import (
	"context"
	"fmt"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/param"
	"github.com/adityakw90/service-user/internal/core/domain/signal"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/internal/core/port/observer"
	"github.com/adityakw90/service-user/internal/core/port/repository"
	portSvc "github.com/adityakw90/service-user/internal/core/port/service"
	"github.com/adityakw90/service-user/pkg/util"
)

type deviceService struct {
	deviceRepo     repository.DeviceRepository
	userDeviceRepo repository.UserDeviceRepository
	deviceObserver observer.ServiceObserver[signal.DeviceSignal]
	eventPublisher portEvent.EventPublisher
}

func NewDeviceService(
	deviceRepo repository.DeviceRepository,
	userDeviceRepo repository.UserDeviceRepository,
	deviceObserver observer.ServiceObserver[signal.DeviceSignal],
	eventPublisher portEvent.EventPublisher,
) portSvc.DeviceService {
	if deviceObserver == nil {
		panic("deviceObserver is required")
	}
	return &deviceService{
		deviceRepo:     deviceRepo,
		userDeviceRepo: userDeviceRepo,
		deviceObserver: deviceObserver,
		eventPublisher: eventPublisher,
	}
}

func (s *deviceService) Get(ctx context.Context, uid string) (*model.Device, error) {
	s.deviceObserver.OnSignal(ctx, signal.SignalStart, signal.DeviceSignal{
		UID:       &uid,
		Operation: "get",
	}, nil)

	device, err := s.deviceRepo.GetByUID(ctx, uid)
	if err != nil {
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			UID:       &uid,
			Operation: "get",
		}, err)
		return nil, err
	}

	s.deviceObserver.OnSignal(ctx, signal.SignalSuccess, signal.DeviceSignal{
		UID:         &uid,
		DeviceName:  &device.DeviceName,
		Fingerprint: &device.DeviceFingerprint,
		Operation:   "get",
	}, nil)

	return device, nil
}

func (s *deviceService) List(ctx context.Context, pagination *param.PaginationParam, filter *param.DeviceListFilterParam) (*model.Devices, error) {
	s.deviceObserver.OnSignal(ctx, signal.SignalStart, signal.DeviceSignal{
		Operation: "list",
	}, nil)

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
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			Operation: "list",
		}, err)
		return nil, err
	}

	s.deviceObserver.OnSignal(ctx, signal.SignalSuccess, signal.DeviceSignal{
		Operation: "list",
	}, nil)

	return devices, nil
}

func (s *deviceService) Delete(ctx context.Context, uid string) error {
	s.deviceObserver.OnSignal(ctx, signal.SignalStart, signal.DeviceSignal{
		UID:       &uid,
		Operation: "delete",
	}, nil)

	// Get device
	device, err := s.deviceRepo.GetByUID(ctx, uid)
	if err != nil {
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Check if there are active user devices linked to this device
	revoked := false
	userDevices, err := s.userDeviceRepo.List(ctx, &param.PaginationParam{Page: util.Ptr(1), Limit: util.Ptr(1)}, &param.UserDeviceListFilterParam{
		DeviceUids: []string{uid},
		Revoked:    &revoked,
	})
	if err != nil {
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}
	if userDevices != nil && len(userDevices.Items) > 0 {
		s.deviceObserver.OnSignal(ctx, signal.SignalReject, signal.DeviceSignal{
			UID:         &uid,
			DeviceName:  &device.DeviceName,
			Fingerprint: &device.DeviceFingerprint,
			Operation:   "delete",
		}, fmt.Errorf("device has active user devices"))
		return fmt.Errorf("cannot delete device with %d active user devices linked", userDevices.Meta.Total)
	}

	err = s.deviceRepo.Delete(ctx, device)
	if err != nil {
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	// Publish device deleted event
	err = s.eventPublisher.Publish(ctx, event.Message{Type: event.EventDeviceDeleted, Entity: event.Entity{ID: uid, Type: "device", Name: &device.DeviceName}, Metadata: event.EventDeviceDeletedData{
		DeviceUID: uid,
	}})
	if err != nil {
		s.deviceObserver.OnSignal(ctx, signal.SignalFail, signal.DeviceSignal{
			UID:       &uid,
			Operation: "delete",
		}, err)
		return err
	}

	s.deviceObserver.OnSignal(ctx, signal.SignalSuccess, signal.DeviceSignal{
		UID:         &uid,
		DeviceName:  &device.DeviceName,
		Fingerprint: &device.DeviceFingerprint,
		Operation:   "delete",
	}, nil)

	return nil
}
