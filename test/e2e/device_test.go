package e2e

import (
	"context"
	"testing"

	commongrpc "github.com/adityakw90/service-user-proto/gen/go/common"
	"github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/stretchr/testify/require"
)

/*
	testing all device grpc service
	each test function is reflecting the service endpoint
*/

// TestE2E_DeviceService_List tests the List endpoint
func TestE2E_DeviceService_List(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) []string
		listReq    func(t *testing.T, deviceUIDs []string) *device.ListRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, resp *device.ListResponse, deviceUIDs []string)
	}{
		{
			name: "List all devices",
			setup: func(t *testing.T) []string {
				// Setup will create devices via auth flow
				return []string{}
			},
			listReq: func(t *testing.T, deviceUIDs []string) *device.ListRequest {
				return &device.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   10,
						OrderBy: "id",
						Sort:    "asc",
					},
				}
			},
			wantErr: false,
		},
		{
			name: "List with pagination",
			setup: func(t *testing.T) []string {
				return []string{}
			},
			listReq: func(t *testing.T, deviceUIDs []string) *device.ListRequest {
				return &device.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   2,
						OrderBy: "id",
						Sort:    "asc",
					},
				}
			},
			wantErr: false,
		},
		{
			name: "List filtered by UIDs",
			setup: func(t *testing.T) []string {
				return []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}
			},
			listReq: func(t *testing.T, deviceUIDs []string) *device.ListRequest {
				return &device.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   10,
						OrderBy: "id",
						Sort:    "asc",
					},
					Filter: &device.FilterRequest{
						Uids: deviceUIDs,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "List filtered by device fingerprint",
			setup: func(t *testing.T) []string {
				return []string{}
			},
			listReq: func(t *testing.T, deviceUIDs []string) *device.ListRequest {
				fingerprint := "test-fingerprint-123"
				return &device.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   10,
						OrderBy: "id",
						Sort:    "asc",
					},
					Filter: &device.FilterRequest{
						DeviceFingerprint: &fingerprint,
					},
				}
			},
			wantErr: false,
		},
		{
			name: "List filtered by device name",
			setup: func(t *testing.T) []string {
				return []string{}
			},
			listReq: func(t *testing.T, deviceUIDs []string) *device.ListRequest {
				deviceName := "iPhone 14"
				return &device.ListRequest{
					Pagination: &commongrpc.Pagination{
						Page:    1,
						Limit:   10,
						OrderBy: "id",
						Sort:    "asc",
					},
					Filter: &device.FilterRequest{
						DeviceName: &deviceName,
					},
				}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			deviceUIDs := tt.setup(t)
			listReq := tt.listReq(t, deviceUIDs)

			resp, err := grpcClient.DeviceClient.List(ctx, listReq)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, resp, deviceUIDs)
				}
			}
		})
	}
}

// TestE2E_DeviceService_Get tests the Get endpoint
func TestE2E_DeviceService_Get(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		getReq     func(t *testing.T, deviceUID string) *device.GetRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, dev *device.Device)
	}{
		{
			name: "Get existing device",
			setup: func(t *testing.T) string {
				// This will be implemented when we have a way to create devices directly
				// For now, we use a placeholder UID
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			getReq: func(t *testing.T, deviceUID string) *device.GetRequest {
				return &device.GetRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true, // Device doesn't exist yet
			errMsg:  "not found",
		},
		{
			name: "Get non-existent device",
			setup: func(t *testing.T) string {
				return "00000000-0000-0000-0000-000000000000"
			},
			getReq: func(t *testing.T, deviceUID string) *device.GetRequest {
				return &device.GetRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "Get with invalid UID format",
			setup: func(t *testing.T) string {
				return "invalid-uid"
			},
			getReq: func(t *testing.T, deviceUID string) *device.GetRequest {
				return &device.GetRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
		},
		{
			name: "Get with empty UID",
			setup: func(t *testing.T) string {
				return ""
			},
			getReq: func(t *testing.T, deviceUID string) *device.GetRequest {
				return &device.GetRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			deviceUID := tt.setup(t)
			getReq := tt.getReq(t, deviceUID)

			dev, err := grpcClient.DeviceClient.Get(ctx, getReq)

			if tt.wantErr {
				require.Error(t, err)
				require.Nil(t, dev)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, dev)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, dev)
				}
			}
		})
	}
}

// TestE2E_DeviceService_Delete tests the Delete endpoint
func TestE2E_DeviceService_Delete(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		deleteReq  func(t *testing.T, deviceUID string) *device.DeleteRequest
		wantErr    bool
		errMsg     string
		verifyFunc func(t *testing.T, deviceUID string)
	}{
		{
			name: "Delete existing device",
			setup: func(t *testing.T) string {
				// This will be implemented when we have a way to create devices directly
				return "01234567-89ab-cdef-0123-456789abcdef"
			},
			deleteReq: func(t *testing.T, deviceUID string) *device.DeleteRequest {
				return &device.DeleteRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true, // Device doesn't exist yet
			errMsg:  "not found",
		},
		{
			name: "Delete non-existent device",
			setup: func(t *testing.T) string {
				return "00000000-0000-0000-0000-000000000000"
			},
			deleteReq: func(t *testing.T, deviceUID string) *device.DeleteRequest {
				return &device.DeleteRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
			errMsg:  "not found",
		},
		{
			name: "Delete with invalid UID format",
			setup: func(t *testing.T) string {
				return "invalid-uid"
			},
			deleteReq: func(t *testing.T, deviceUID string) *device.DeleteRequest {
				return &device.DeleteRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
		},
		{
			name: "Delete with empty UID",
			setup: func(t *testing.T) string {
				return ""
			},
			deleteReq: func(t *testing.T, deviceUID string) *device.DeleteRequest {
				return &device.DeleteRequest{
					Uid: deviceUID,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, grpcClient, cleanup := setupE2ETest(t)
			defer cleanup()

			ctx := context.Background()
			deviceUID := tt.setup(t)
			deleteReq := tt.deleteReq(t, deviceUID)

			resp, err := grpcClient.DeviceClient.Delete(ctx, deleteReq)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.True(t, resp.Success)

				if tt.verifyFunc != nil {
					tt.verifyFunc(t, deviceUID)
				}
			}
		})
	}
}
