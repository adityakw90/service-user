package signal

import (
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestDomain_Signal_UserSignal_Creation(t *testing.T) {
	uid := "user-123"
	username := "testuser"
	status := model.UserStatusActive
	active := true

	sig := UserSignal{
		UID:      &uid,
		Username: &username,
		Status:   &status,
		Active:   &active,
		Operation: "get",
	}

	if sig.UID == nil {
		t.Error("UID should not be nil")
	}
	if *sig.UID != uid {
		t.Errorf("UID = %s, want %s", *sig.UID, uid)
	}
}

func TestDomain_Signal_UserSignal_AllFieldsNil(t *testing.T) {
	sig := UserSignal{
		Operation: "list",
	}

	if sig.UID != nil {
		t.Error("UID should be nil")
	}
	if sig.Username != nil {
		t.Error("Username should be nil")
	}
}
