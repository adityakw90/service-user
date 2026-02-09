package service

import (
	"context"
	"testing"

	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockDeviceRepository)
		uid        string
		want       *model.Device
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return createTestDevice(1, "device-uid", "iPhone", "fp123"), nil
				}
			},
			uid:  "device-uid",
			want: createTestDevice(1, "device-uid", "iPhone", "fp123"),
		},
		{
			name: "Error - device not found",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return nil, domainerrors.ErrDeviceNotFound
				}
			},
			uid:     "nonexistent-device",
			wantErr: domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				NewMockDeviceObserver(),
				publisher.NewNoOpPublisher(),
			)

			// Execute
			got, err := svc.Get(context.Background(), tt.uid)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.UID, got.UID)
			assert.Equal(t, tt.want.DeviceName, got.DeviceName)
		})
	}
}

func TestDeviceService_List(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockDeviceRepository)
		pagination *params.PaginationParam
		filter     *params.DeviceListFilterParam
		want       *model.Devices
		wantErr    error
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.DeviceListFilterParam) (*model.Devices, error) {
					return &model.Devices{
						Items: []model.Device{
							*createTestDevice(1, "device1", "iPhone", "fp1"),
							*createTestDevice(2, "device2", "Android", "fp2"),
						},
						Meta: model.Meta{
							Page:  1,
							Limit: 10,
							Total: 2,
							Pages: 1,
						},
					}, nil
				}
			},
			pagination: nil,
			filter:     nil,
			want: &model.Devices{
				Items: []model.Device{
					*createTestDevice(1, "device1", "iPhone", "fp1"),
					*createTestDevice(2, "device2", "Android", "fp2"),
				},
			},
		},
		{
			name: "Happy Path - custom pagination",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.DeviceListFilterParam) (*model.Devices, error) {
					return &model.Devices{
						Items: []model.Device{},
						Meta:  model.Meta{Page: 2, Limit: 20},
					}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(2),
				Limit:   util.Ptr(20),
				Sort:    util.Ptr("desc"),
				OrderBy: util.Ptr("created_at"),
			},
			filter: nil,
			want: &model.Devices{
				Items: []model.Device{},
			},
		},
		{
			name: "Happy Path - with filters",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.ListFunc = func(ctx context.Context, p *params.PaginationParam, f *params.DeviceListFilterParam) (*model.Devices, error) {
					return &model.Devices{Items: []model.Device{}}, nil
				}
			},
			pagination: &params.PaginationParam{
				Page:    util.Ptr(1),
				Limit:   util.Ptr(10),
				Sort:    util.Ptr("asc"),
				OrderBy: util.Ptr("created_at"),
			},
			filter: &params.DeviceListFilterParam{},
			want: &model.Devices{
				Items: []model.Device{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				NewMockDeviceObserver(),
				publisher.NewNoOpPublisher(),
			)

			// Execute
			got, err := svc.List(context.Background(), tt.pagination, tt.filter)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}

func TestDeviceService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockDeviceRepository)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return createTestDevice(1, "device-uid", "iPhone", "fp123"), nil
				}
				dr.DeleteFunc = func(ctx context.Context, device *model.Device) error {
					return nil
				}
			},
			uid: "device-uid",
		},
		{
			name: "Error - device not found",
			setupMocks: func(dr *MockDeviceRepository) {
				dr.GetByUIDFunc = func(ctx context.Context, uid string) (*model.Device, error) {
					return nil, domainerrors.ErrDeviceNotFound
				}
			},
			uid:     "nonexistent-device",
			wantErr: domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockDeviceRepo := NewMockDeviceRepository()
			mockUserDeviceRepo := NewMockUserDeviceRepository()

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				NewMockDeviceObserver(),
				publisher.NewNoOpPublisher(),
			)

			// Execute
			err := svc.Delete(context.Background(), tt.uid)

			// Assert
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			require.NoError(t, err)
		})
	}
}
