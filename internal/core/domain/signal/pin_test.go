package signal

import "testing"

func TestDomain_Signal_PinSignal_Creation(t *testing.T) {
	userUID := "user-123"
	success := true

	sig := PinSignal{
		UserUID:   userUID,
		Operation: "verify",
		Success:   &success,
	}

	if sig.UserUID != userUID {
		t.Errorf("UserUID = %s, want %s", sig.UserUID, userUID)
	}
	if sig.Success == nil {
		t.Error("Success should not be nil")
	}
}

func TestDomain_Signal_PinSignal_OptionalFields(t *testing.T) {
	sig := PinSignal{
		UserUID:   "user-456",
		Operation: "set",
	}

	if sig.Success != nil {
		t.Error("Success should be nil for set operation")
	}
}
