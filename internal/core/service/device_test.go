package service

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/signal"
	domainerrors "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/internal/core/domain/params"
	"github.com/adityakw90/service-user/internal/adapter/publisher"
	repomocks "github.com/adityakw90/service-user/test/mocks/repository"
	observermocks "github.com/adityakw90/service-user/test/mocks/observer"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// setupObserverAny allows any OnSignal calls on the observer (useful when not testing signal behavior)
func setupDeviceObserverAny(t *testing.T, observer *observermocks.MockServiceObserver[signal.DeviceSignal]) {
	// Allow any OnSignal call without checking parameters
	// Use Maybe() to make the expectation optional (can be called 0 or more times)
	// Note: Using EXPECT().OnSignal() pattern for better type safety
	observer.EXPECT().OnSignal(mock.Anything, mock.Anything, mock.AnythingOfType("signal.DeviceSignal"), mock.Anything).Maybe()
}

func TestDeviceService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*repomocks.MockDeviceRepository)
		uid        string
		want       *model.Device
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(dr *repomocks.MockDeviceRepository) {
				dr.EXPECT().GetByUID(mock.Anything, "device-uid").Return(createTestDevice(1, "device-uid", "iPhone", "fp123"), nil).Once()
			},
			uid:  "device-uid",
			want: createTestDevice(1, "device-uid", "iPhone", "fp123"),
		},
		{
			name: "Error - device not found",
			setupMocks: func(dr *repomocks.MockDeviceRepository) {
				dr.EXPECT().GetByUID(mock.Anything, "nonexistent-device").Return(nil, domainerrors.ErrDeviceNotFound).Once()
			},
			uid:     "nonexistent-device",
			wantErr: domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				func() *observermocks.MockServiceObserver[signal.DeviceSignal] {
				obs := observermocks.NewMockServiceObserver[signal.DeviceSignal](t)
				setupDeviceObserverAny(t, obs)
				return obs
			}(),
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
		setupMocks func(*repomocks.MockDeviceRepository)
		pagination *params.PaginationParam
		filter     *params.DeviceListFilterParam
		want       *model.Devices
		wantErr    error
	}{
		{
			name: "Happy Path - default pagination",
			setupMocks: func(dr *repomocks.MockDeviceRepository) {
				dr.EXPECT().List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.DeviceListFilterParam")).Return(&model.Devices{
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
				}, nil).Once()
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
			setupMocks: func(dr *repomocks.MockDeviceRepository) {
				dr.EXPECT().List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.DeviceListFilterParam")).Return(&model.Devices{
					Items: []model.Device{},
					Meta:  model.Meta{Page: 2, Limit: 20},
				}, nil).Once()
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
			setupMocks: func(dr *repomocks.MockDeviceRepository) {
				dr.EXPECT().List(mock.Anything, mock.AnythingOfType("*params.PaginationParam"), mock.AnythingOfType("*params.DeviceListFilterParam")).Return(&model.Devices{Items: []model.Device{}}, nil).Once()
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
			// Setup mocks using generated mocks
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				func() *observermocks.MockServiceObserver[signal.DeviceSignal] {
				obs := observermocks.NewMockServiceObserver[signal.DeviceSignal](t)
				setupDeviceObserverAny(t, obs)
				return obs
			}(),
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
		setupMocks func(*repomocks.MockDeviceRepository, *repomocks.MockUserDeviceRepository)
		uid        string
		wantErr    error
	}{
		{
			name: "Happy Path",
			setupMocks: func(dr *repomocks.MockDeviceRepository, udr *repomocks.MockUserDeviceRepository) {
				dr.EXPECT().GetByUID(mock.Anything, "device-uid").Return(createTestDevice(1, "device-uid", "iPhone", "fp123"), nil).Once()
				// Check for active user devices - return empty list
				udr.EXPECT().List(mock.Anything, mock.MatchedBy(func(p *params.PaginationParam) bool {
					return p != nil && *p.Page == 1 && *p.Limit == 1
				}), mock.MatchedBy(func(f *params.UserDeviceListFilterParam) bool {
					return f != nil && len(f.DeviceUids) == 1 && f.DeviceUids[0] == "device-uid" && f.Revoked != nil && *f.Revoked == false
				})).Return(&model.UserDevices{Items: []model.UserDevice{}}, nil).Once()
				dr.EXPECT().Delete(mock.Anything, mock.AnythingOfType("*model.Device")).Return(nil).Once()
			},
			uid: "device-uid",
		},
		{
			name: "Error - device not found",
			setupMocks: func(dr *repomocks.MockDeviceRepository, udr *repomocks.MockUserDeviceRepository) {
				dr.EXPECT().GetByUID(mock.Anything, "nonexistent-device").Return(nil, domainerrors.ErrDeviceNotFound).Once()
			},
			uid:     "nonexistent-device",
			wantErr: domainerrors.ErrDeviceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks using generated mocks
			mockDeviceRepo := repomocks.NewMockDeviceRepository(t)
			mockUserDeviceRepo := repomocks.NewMockUserDeviceRepository(t)

			// Setup expectations
			if tt.setupMocks != nil {
				tt.setupMocks(mockDeviceRepo, mockUserDeviceRepo)
			}

			// Create service
			svc := NewDeviceService(
				mockDeviceRepo,
				mockUserDeviceRepo,
				func() *observermocks.MockServiceObserver[signal.DeviceSignal] {
				obs := observermocks.NewMockServiceObserver[signal.DeviceSignal](t)
				setupDeviceObserverAny(t, obs)
				return obs
			}(),
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
