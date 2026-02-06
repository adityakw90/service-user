package request

import (
	"strings"

	device "github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user/internal/core/domain/params"
)

// DeviceGetRequest represents validated device get request.
type DeviceGetRequest struct {
	Uid string `validate:"required"`
}

// DeviceGetRequestFromPb creates a DeviceGetRequest from protobuf.
func DeviceGetRequestFromPb(req *device.GetRequest) *DeviceGetRequest {
	return &DeviceGetRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}

type DeviceFilterRequest struct {
	Uids              []string `validate:"omitempty"`
	DeviceName        *string  `validate:"omitempty"`
	DeviceFingerprint *string  `validate:"omitempty"`
}

func (r *DeviceFilterRequest) ToDeviceFilterParams() *params.DeviceListFilterParam {
	return &params.DeviceListFilterParam{
		Uids:              r.Uids,
		DeviceName:        r.DeviceName,
		DeviceFingerprint: r.DeviceFingerprint,
	}
}

func DeviceFilterRequestFromPb(req *device.FilterRequest) *DeviceFilterRequest {
	payload := &DeviceFilterRequest{
		Uids: req.GetUids(),
	}

	if req.DeviceName != nil {
		fieldDeviceName := strings.TrimSpace(req.GetDeviceName())
		if fieldDeviceName != "" {
			payload.DeviceName = &fieldDeviceName
		}
	}

	if req.DeviceFingerprint != nil {
		fieldDeviceFingerprint := strings.TrimSpace(req.GetDeviceFingerprint())
		if fieldDeviceFingerprint != "" {
			payload.DeviceFingerprint = &fieldDeviceFingerprint
		}
	}

	return payload
}

// DeviceListRequest represents validated list request.
type DeviceListRequest struct {
	Pagination *PaginationRequest
	Filter     *DeviceFilterRequest
}

func (r *DeviceListRequest) ToDeviceListParams() *params.DeviceListParam {
	return &params.DeviceListParam{
		Pagination: r.Pagination.ToPaginationParams(),
		Filter:     r.Filter.ToDeviceFilterParams(),
	}
}

// DeviceListRequestFromPb creates a DeviceListRequest from protobuf.
func DeviceListRequestFromPb(req *device.ListRequest) *DeviceListRequest {
	payload := &DeviceListRequest{}

	if req.Pagination != nil {
		payload.Pagination = PaginationRequestFromPb(req.GetPagination())
	}

	if req.Filter != nil {
		payload.Filter = DeviceFilterRequestFromPb(req.GetFilter())
	}

	return payload
}

// DeviceDeleteRequest represents validated device delete request.
type DeviceDeleteRequest struct {
	Uid string `validate:"required"`
}

// DeviceDeleteRequestFromPb creates a DeviceDeleteRequest from protobuf.
func DeviceDeleteRequestFromPb(req *device.DeleteRequest) *DeviceDeleteRequest {
	return &DeviceDeleteRequest{
		Uid: strings.TrimSpace(req.Uid),
	}
}
