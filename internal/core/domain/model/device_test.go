package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

/* test completed by jojo */

func TestCore_Domain_Device(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() *Device
		checkFn func(*Device)
	}{
		{
			name: "set correctly",
			setupFn: func() *Device {
				return &Device{
					ID:                1,
					UID:               "test-uid",
					DeviceFingerprint: "fingerprint-123",
					DeviceName:        "Test Device",
					CreatedAt:         time.Now().UTC(),
				}
			},
			checkFn: func(d *Device) {
				assert.Equal(t, int64(1), d.ID)
				assert.Equal(t, "test-uid", d.UID)
				assert.Equal(t, "fingerprint-123", d.DeviceFingerprint)
				assert.Equal(t, "Test Device", d.DeviceName)
				assert.NotZero(t, d.CreatedAt)
			},
		},
		{
			name: "Zero Value",
			setupFn: func() *Device {
				var d Device
				return &d
			},
			checkFn: func(d *Device) {
				assert.Equal(t, int64(0), d.ID)
				assert.Equal(t, "", d.UID)
				assert.Equal(t, "", d.DeviceFingerprint)
				assert.Equal(t, "", d.DeviceName)
				assert.Zero(t, d.CreatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := tt.setupFn()
			tt.checkFn(device)
		})
	}
}
