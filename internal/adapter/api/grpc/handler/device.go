package handler

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	common "github.com/adityakw90/service-user-proto/gen/go/common"
	device "github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/request"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/response"
	"github.com/adityakw90/service-user/internal/adapter/api/grpc/validator"
	portsvc "github.com/adityakw90/service-user/internal/core/port/service"
)

// DeviceHandler implements the gRPC DeviceService.
type DeviceHandler struct {
	device.UnimplementedDeviceServiceServer
	service   portsvc.DeviceService
	validator *validator.Validator
}

// NewDeviceHandler creates a new DeviceHandler.
func NewDeviceHandler(service portsvc.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		service:   service,
		validator: validator.New(),
	}
}

// Get retrieves a single device by UID.
func (h *DeviceHandler) Get(ctx context.Context, req *device.GetRequest) (*device.Device, error) {
	dto := validator.DeviceGetRequestDTO{Uid: req.Uid}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	d, err := h.service.Get(ctx, req.Uid)
	if err != nil {
		return nil, response.MapError(err)
	}

	return response.ToProtoDeviceFull(d), nil
}

// List retrieves a list of devices.
func (h *DeviceHandler) List(ctx context.Context, req *device.ListRequest) (*device.ListResponse, error) {
	pagination := request.ToPaginationParam(req.Pagination)
	filter := request.ToDeviceListFilterParam(req.Filter)

	result, err := h.service.List(ctx, pagination, filter)
	if err != nil {
		return nil, response.MapError(err)
	}

	items := make([]*device.Device, len(result.Items))
	for i, d := range result.Items {
		items[i] = response.ToProtoDeviceFull(&d)
	}

	meta := &common.Meta{
		Total: int64(len(result.Items)),
		Limit: int32(*pagination.Limit),
	}

	return &device.ListResponse{
		Items: items,
		Meta:  meta,
	}, nil
}

// Delete deletes a device by UID.
func (h *DeviceHandler) Delete(ctx context.Context, req *device.DeleteRequest) (*common.Success, error) {
	dto := validator.DeviceDeleteRequestDTO{Uid: req.Uid}
	if err := h.validator.Struct(dto); err != nil {
		return nil, status.Error(codes.InvalidArgument, validator.ValidationErrors(err))
	}

	if err := h.service.Delete(ctx, req.Uid); err != nil {
		return nil, response.MapError(err)
	}

	return &common.Success{Success: true}, nil
}
