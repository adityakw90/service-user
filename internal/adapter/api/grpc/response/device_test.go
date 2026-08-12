package response

import (
	"testing"
	"time"

	device "github.com/adityakw90/service-user-proto/gen/go/device"
	"github.com/adityakw90/service-user-proto/gen/go/user"
	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestToProtoDevice(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input *model.Device
		want  *user.Device
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid device",
			input: &model.Device{
				UID:        "device-123",
				DeviceName: "iPhone 14",
				CreatedAt:  now,
			},
			want: &user.Device{
				DeviceUid:  "device-123",
				DeviceName: "iPhone 14",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoDevice(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoDevice() = %v, want nil", got)
				}
				return
			}

			if got.DeviceUid != tt.want.DeviceUid {
				t.Errorf("ToProtoDevice().DeviceUid = %v, want %v", got.DeviceUid, tt.want.DeviceUid)
			}
			if got.DeviceName != tt.want.DeviceName {
				t.Errorf("ToProtoDevice().DeviceName = %v, want %v", got.DeviceName, tt.want.DeviceName)
			}
		})
	}
}

func TestToProtoDeviceFull(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		input *model.Device
		want  *device.Device
	}{
		{
			name:  "Nil input",
			input: nil,
			want:  nil,
		},
		{
			name: "Valid full device",
			input: &model.Device{
				UID:               "device-456",
				DeviceFingerprint: "fp-abc123",
				DeviceName:        "MacBook Pro",
				CreatedAt:         now,
			},
			want: &device.Device{
				Uid:               "device-456",
				DeviceFingerprint: "fp-abc123",
				DeviceName:        "MacBook Pro",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoDeviceFull(tt.input)

			if tt.want == nil {
				if got != nil {
					t.Errorf("ToProtoDeviceFull() = %v, want nil", got)
				}
				return
			}

			if got.Uid != tt.want.Uid {
				t.Errorf("ToProtoDeviceFull().Uid = %v, want %v", got.Uid, tt.want.Uid)
			}
			if got.DeviceFingerprint != tt.want.DeviceFingerprint {
				t.Errorf("ToProtoDeviceFull().DeviceFingerprint = %v, want %v", got.DeviceFingerprint, tt.want.DeviceFingerprint)
			}
			if got.DeviceName != tt.want.DeviceName {
				t.Errorf("ToProtoDeviceFull().DeviceName = %v, want %v", got.DeviceName, tt.want.DeviceName)
			}
		})
	}
}

func TestToProtoDeviceList(t *testing.T) {
	tests := []struct {
		name    string
		devices *model.Devices
		meta    *model.Meta
		wantLen int
	}{
		{
			name:    "Nil devices",
			devices: nil,
			meta:    &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
			wantLen: 0,
		},
		{
			name: "Empty list",
			devices: &model.Devices{
				Items: []model.Device{},
			},
			meta:    &model.Meta{Page: 1, Limit: 10, Total: 0, Pages: 0},
			wantLen: 0,
		},
		{
			name: "List with devices",
			devices: &model.Devices{
				Items: []model.Device{
					{UID: "device-1", DeviceName: "iPhone"},
					{UID: "device-2", DeviceName: "iPad"},
				},
			},
			meta:    &model.Meta{Page: 1, Limit: 10, Total: 2, Pages: 1},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToProtoDeviceList(tt.devices, tt.meta)

			if len(got.Items) != tt.wantLen {
				t.Errorf("ToProtoDeviceList() len = %d, want %d", len(got.Items), tt.wantLen)
			}
		})
	}
}
