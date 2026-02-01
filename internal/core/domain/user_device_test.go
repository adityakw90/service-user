package domain

import (
	"testing"
	"time"
)

func TestUserDevice_IsRevoked(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
		want      bool
	}{
		{"not revoked", nil, false},
		{"revoked", func() *time.Time { t := time.Now(); return &t }(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ud := &UserDevice{RevokedAt: tt.revokedAt}
			if got := ud.IsRevoked(); got != tt.want {
				t.Errorf("UserDevice.IsRevoked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserDevice_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
		want      bool
	}{
		{"not revoked", nil, true},
		{"revoked", func() *time.Time { t := time.Now(); return &t }(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ud := &UserDevice{RevokedAt: tt.revokedAt}
			if got := ud.IsActive(); got != tt.want {
				t.Errorf("UserDevice.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserDevice_Revoke(t *testing.T) {
	ud := &UserDevice{
		UserID:   1,
		DeviceID: 2,
	}

	if ud.IsRevoked() {
		t.Error("UserDevice should not be revoked before Revoke()")
	}

	ud.Revoke()

	if !ud.IsRevoked() {
		t.Error("UserDevice should be revoked after Revoke()")
	}
	if ud.RevokedAt == nil {
		t.Error("UserDevice.RevokedAt should be set after Revoke()")
	}
}

func TestUserDevice_Touch(t *testing.T) {
	oldTime := time.Now().Add(-time.Minute)
	ud := &UserDevice{
		UserID:       1,
		DeviceID:     2,
		LastActiveAt: oldTime,
	}

	ud.Touch()

	if !ud.LastActiveAt.After(oldTime) {
		t.Error("UserDevice.LastActiveAt should be updated after Touch()")
	}
}

func TestUserDevice_Fields(t *testing.T) {
	now := time.Now().UTC()
	ud := &UserDevice{
		UserID:       1,
		DeviceID:     2,
		IPAddress:    "192.168.1.1",
		LastActiveAt: now,
		CreatedAt:    now,
	}

	if ud.UserID != 1 {
		t.Errorf("UserDevice.UserID = %v, want 1", ud.UserID)
	}
	if ud.DeviceID != 2 {
		t.Errorf("UserDevice.DeviceID = %v, want 2", ud.DeviceID)
	}
	if ud.IPAddress != "192.168.1.1" {
		t.Errorf("UserDevice.IPAddress = %v, want '192.168.1.1'", ud.IPAddress)
	}
}
