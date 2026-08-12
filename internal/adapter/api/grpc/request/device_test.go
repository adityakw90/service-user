package request

import (
	"testing"

	commonpb "github.com/adityakw90/service-user-proto/gen/go/common"
	devicepb "github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeviceGetRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  some-device-uid  ", wantUid: "some-device-uid"},
		{name: "already trimmed value unchanged", raw: "device-uid-abc", wantUid: "device-uid-abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeviceGetRequestFromPb(&devicepb.GetRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestDeviceDeleteRequestFromPb(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantUid string
	}{
		{name: "trims surrounding whitespace", raw: "  deleted-device-uid  ", wantUid: "deleted-device-uid"},
		{name: "already trimmed value unchanged", raw: "device-uid-xyz", wantUid: "device-uid-xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeviceDeleteRequestFromPb(&devicepb.DeleteRequest{Uid: tt.raw})
			assert.Equal(t, tt.wantUid, got.Uid)
		})
	}
}

func TestDeviceFilterRequestFromPb(t *testing.T) {
	name := " My Phone "
	fingerprint := " fp12345 "
	emptyName := "   "
	emptyFingerprint := ""

	tests := []struct {
		name                  string
		req                   *devicepb.FilterRequest
		wantUids              []string
		wantDeviceName        *string
		wantDeviceFingerprint *string
	}{
		{
			name: "all fields populated and trimmed",
			req: &devicepb.FilterRequest{
				Uids:              []string{"uid1", "uid2"},
				DeviceName:        &name,
				DeviceFingerprint: &fingerprint,
			},
			wantUids:              []string{"uid1", "uid2"},
			wantDeviceName:        strPtr("My Phone"),
			wantDeviceFingerprint: strPtr("fp12345"),
		},
		{
			name: "whitespace-only optional fields become nil",
			req: &devicepb.FilterRequest{
				DeviceName:        &emptyName,
				DeviceFingerprint: &emptyFingerprint,
			},
			wantUids:              nil,
			wantDeviceName:        nil,
			wantDeviceFingerprint: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeviceFilterRequestFromPb(tt.req)
			assert.Equal(t, tt.wantUids, got.Uids)
			assertOptionalString(t, tt.wantDeviceName, got.DeviceName, "DeviceName")
			assertOptionalString(t, tt.wantDeviceFingerprint, got.DeviceFingerprint, "DeviceFingerprint")

			// Verify ToDeviceFilterParams maps fields unchanged.
			params := got.ToDeviceFilterParams()
			require.NotNil(t, params)
			assert.Equal(t, got.Uids, params.Uids)
			assert.Equal(t, got.DeviceName, params.DeviceName)
			assert.Equal(t, got.DeviceFingerprint, params.DeviceFingerprint)
		})
	}
}

func TestDeviceListRequestFromPb(t *testing.T) {
	deviceName := "device-a"

	tests := []struct {
		name           string
		req            *devicepb.ListRequest
		wantPage       int
		wantLimit      int
		wantSort       string
		wantOrderBy    string
		wantDeviceName *string
	}{
		{
			name: "with pagination and filter",
			req: &devicepb.ListRequest{
				Pagination: &commonpb.Pagination{
					Page:  3,
					Limit: 25,
					Sort:  "asc",
				},
				Filter: &devicepb.FilterRequest{
					DeviceName: &deviceName,
				},
			},
			wantPage:       3,
			wantLimit:      25,
			wantSort:       "asc",
			wantOrderBy:    "",
			wantDeviceName: strPtr("device-a"),
		},
		{
			name:        "nil pagination and filter use defaults",
			req:         &devicepb.ListRequest{},
			wantPage:    1,
			wantLimit:   10,
			wantSort:    "desc",
			wantOrderBy: "created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeviceListRequestFromPb(tt.req)

			params := got.ToDeviceListParams()
			require.NotNil(t, params)
			require.NotNil(t, params.Pagination)
			assert.Equal(t, tt.wantPage, *params.Pagination.Page)
			assert.Equal(t, tt.wantLimit, *params.Pagination.Limit)
			assert.Equal(t, tt.wantSort, *params.Pagination.Sort)
			assert.Equal(t, tt.wantOrderBy, *params.Pagination.OrderBy)

			require.NotNil(t, params.Filter)
			assertOptionalString(t, tt.wantDeviceName, params.Filter.DeviceName, "DeviceName")
		})
	}
}

func TestDeviceListRequestToDeviceListParams_DefaultPagination(t *testing.T) {
	tests := []struct {
		name        string
		req         *DeviceListRequest
		wantPage    int
		wantLimit   int
		wantSort    string
		wantOrderBy string
	}{
		{
			name:        "nil pagination and nil filter apply defaults",
			req:         &DeviceListRequest{},
			wantPage:    1,
			wantLimit:   10,
			wantSort:    "desc",
			wantOrderBy: "created_at",
		},
		{
			name: "explicit pagination is forwarded",
			req: &DeviceListRequest{
				Pagination: &PaginationRequest{Page: 5, Limit: 50, Sort: "asc", OrderBy: "uid"},
			},
			wantPage:    5,
			wantLimit:   50,
			wantSort:    "asc",
			wantOrderBy: "uid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := tt.req.ToDeviceListParams()
			require.NotNil(t, params)
			require.NotNil(t, params.Pagination)
			assert.Equal(t, tt.wantPage, *params.Pagination.Page)
			assert.Equal(t, tt.wantLimit, *params.Pagination.Limit)
			assert.Equal(t, tt.wantSort, *params.Pagination.Sort)
			assert.Equal(t, tt.wantOrderBy, *params.Pagination.OrderBy)
			require.NotNil(t, params.Filter)
		})
	}
}

// assertOptionalString and strPtr helpers are defined in auth_test.go (same package).
