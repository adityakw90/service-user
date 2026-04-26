package signal

import "testing"

func TestDomain_Signal_DeviceSignal_Creation(t *testing.T) {
	uid := "device-123"
	deviceName := "iPhone 15"

	sig := DeviceSignal{
		UID:        &uid,
		DeviceName: &deviceName,
		Operation:  "get",
	}

	if sig.UID == nil {
		t.Error("UID should not be nil")
	}
	if *sig.UID != uid {
		t.Errorf("UID = %s, want %s", *sig.UID, uid)
	}
}

func TestDomain_Signal_DeviceSignal_AllFieldsNil(t *testing.T) {
	sig := DeviceSignal{
		Operation: "list",
	}

	if sig.UID != nil {
		t.Error("UID should be nil")
	}
}
