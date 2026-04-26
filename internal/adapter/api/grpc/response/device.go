package response

import (
	device "github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

// ToProtoDevice converts domain Device to proto Device (user package).
func ToProtoDevice(d *model.Device) *user.Device {
	if d == nil {
		return nil
	}
	return &user.Device{
		DeviceUid:  d.UID,
		DeviceName: d.DeviceName,
		CreatedAt:  toProtoTimestampPB(d.CreatedAt),
	}
}

// ToProtoDeviceFull converts domain Device to proto Device (device package).
func ToProtoDeviceFull(d *model.Device) *device.Device {
	if d == nil {
		return nil
	}
	return &device.Device{
		Uid:               d.UID,
		DeviceFingerprint: d.DeviceFingerprint,
		DeviceName:        d.DeviceName,
		CreatedAt:         toProtoTimestampPB(d.CreatedAt),
	}
}

// ToProtoDeviceList converts domain Devices to proto ListResponse.
func ToProtoDeviceList(devices *model.Devices, meta *model.Meta) *device.ListResponse {
	if devices == nil {
		return &device.ListResponse{Meta: ToProtoMeta(meta)}
	}
	items := make([]*device.Device, len(devices.Items))
	for i, d := range devices.Items {
		items[i] = ToProtoDeviceFull(&d)
	}

	return &device.ListResponse{
		Items: items,
		Meta:  ToProtoMeta(meta),
	}
}
