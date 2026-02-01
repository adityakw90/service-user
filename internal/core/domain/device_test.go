package domain

import (
	"testing"
	"time"
)

func TestDevice_Fields(t *testing.T) {
	tests := []struct {
		name             string
		id               int64
		uid              string
		deviceFingerprint string
		deviceName       string
		checkField       string
	}{
		{"ID is set correctly", 1, "test-uid", "fingerprint-123", "Test Device", "ID"},
		{"UID is set correctly", 1, "test-uid", "fingerprint-123", "Test Device", "UID"},
		{"DeviceFingerprint is set", 1, "test-uid", "fingerprint-123", "Test Device", "DeviceFingerprint"},
		{"DeviceName is set", 1, "test-uid", "fingerprint-123", "Test Device", "DeviceName"},
		{"CreatedAt is set", 1, "test-uid", "fingerprint-123", "Test Device", "CreatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			d := &Device{
				ID:                tt.id,
				UID:               tt.uid,
				DeviceFingerprint: tt.deviceFingerprint,
				DeviceName:        tt.deviceName,
				CreatedAt:         now,
			}

			switch tt.checkField {
			case "ID":
				if d.ID != tt.id {
					t.Errorf("Device.ID = %v, want %v", d.ID, tt.id)
				}
			case "UID":
				if d.UID != tt.uid {
					t.Errorf("Device.UID = %v, want %v", d.UID, tt.uid)
				}
			case "DeviceFingerprint":
				if d.DeviceFingerprint != tt.deviceFingerprint {
					t.Errorf("Device.DeviceFingerprint = %v, want %v", d.DeviceFingerprint, tt.deviceFingerprint)
				}
			case "DeviceName":
				if d.DeviceName != tt.deviceName {
					t.Errorf("Device.DeviceName = %v, want %v", d.DeviceName, tt.deviceName)
				}
			case "CreatedAt":
				if d.CreatedAt.IsZero() {
					t.Error("Device.CreatedAt is zero")
				}
			}
		})
	}
}

func TestDevice_ZeroValue(t *testing.T) {
	tests := []struct {
		name       string
		got        interface{}
		want       interface{}
		fieldCheck string
	}{
		{"zero ID", int64(0), int64(0), "ID"},
		{"zero UID", "", "", "UID"},
		{"zero DeviceFingerprint", "", "", "DeviceFingerprint"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Device
			switch tt.fieldCheck {
			case "ID":
				if d.ID != tt.want.(int64) {
					t.Errorf("zero Device.ID = %v, want %v", d.ID, tt.want)
				}
			case "UID":
				if d.UID != tt.want.(string) {
					t.Errorf("zero Device.UID = %v, want %v", d.UID, tt.want)
				}
			case "DeviceFingerprint":
				if d.DeviceFingerprint != tt.want.(string) {
					t.Errorf("zero Device.DeviceFingerprint = %v, want %v", d.DeviceFingerprint, tt.want)
				}
			}
		})
	}
}
