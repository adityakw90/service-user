package domain

import (
	"testing"
	"time"
)

func TestDevice_Fields(t *testing.T) {
	now := time.Now().UTC()
	d := &Device{
		ID:                1,
		UID:               "test-uid",
		DeviceFingerprint: "fingerprint-123",
		DeviceName:        "Test Device",
		CreatedAt:         now,
	}

	if d.ID != 1 {
		t.Errorf("Device.ID = %v, want 1", d.ID)
	}
	if d.UID != "test-uid" {
		t.Errorf("Device.UID = %v, want 'test-uid'", d.UID)
	}
	if d.DeviceFingerprint != "fingerprint-123" {
		t.Errorf("Device.DeviceFingerprint = %v, want 'fingerprint-123'", d.DeviceFingerprint)
	}
	if d.DeviceName != "Test Device" {
		t.Errorf("Device.DeviceName = %v, want 'Test Device'", d.DeviceName)
	}
	if d.CreatedAt.IsZero() {
		t.Error("Device.CreatedAt is zero")
	}
}

func TestDevice_ZeroValue(t *testing.T) {
	var d Device
	if d.ID != 0 {
		t.Errorf("zero Device.ID = %v, want 0", d.ID)
	}
	if d.UID != "" {
		t.Errorf("zero Device.UID = %v, want ''", d.UID)
	}
	if d.DeviceFingerprint != "" {
		t.Errorf("zero Device.DeviceFingerprint = %v, want ''", d.DeviceFingerprint)
	}
}
