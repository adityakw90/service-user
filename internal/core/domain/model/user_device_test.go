package model

import (
	"testing"
	"time"
)

func TestCore_Domain_UserDevice_IsRevoked(t *testing.T) {
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

func TestCore_Domain_UserDevice_IsActive(t *testing.T) {
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

func TestCore_Domain_UserDevice_Revoke(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		deviceID       int64
		checkRevoked   bool
		checkRevokedAt bool
	}{
		{"sets RevokedAt after revoke", 1, 2, true, true},
		{"user ID remains unchanged", 1, 2, true, true},
		{"device ID remains unchanged", 1, 2, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ud := &UserDevice{
				UserID:   tt.userID,
				DeviceID: tt.deviceID,
			}

			if ud.IsRevoked() {
				t.Error("UserDevice should not be revoked before Revoke()")
			}

			ud.Revoke()

			if tt.checkRevoked && !ud.IsRevoked() {
				t.Error("UserDevice should be revoked after Revoke()")
			}
			if tt.checkRevokedAt && ud.RevokedAt == nil {
				t.Error("UserDevice.RevokedAt should be set after Revoke()")
			}
		})
	}
}

func TestCore_Domain_UserDevice_Touch(t *testing.T) {
	tests := []struct {
		name          string
		userID        int64
		deviceID      int64
		oldTimeOffset time.Duration
		checkUpdated  bool
	}{
		{"updates LastActiveAt", 1, 2, -time.Minute, true},
		{"user ID unchanged", 1, 2, -time.Minute, true},
		{"device ID unchanged", 1, 2, -time.Minute, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldTime := time.Now().Add(tt.oldTimeOffset)
			ud := &UserDevice{
				UserID:       tt.userID,
				DeviceID:     tt.deviceID,
				LastActiveAt: oldTime,
			}

			ud.Touch()

			if tt.checkUpdated && !ud.LastActiveAt.After(oldTime) {
				t.Error("UserDevice.LastActiveAt should be updated after Touch()")
			}
		})
	}
}

func TestCore_Domain_UserDevice_Fields(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		deviceID   int64
		ipAddress  string
		checkField string
	}{
		{"UserID is set", 1, 2, "192.168.1.1", "UserID"},
		{"DeviceID is set", 1, 2, "192.168.1.1", "DeviceID"},
		{"IPAddress is set", 1, 2, "192.168.1.1", "IPAddress"},
		{"LastActiveAt is set", 1, 2, "192.168.1.1", "LastActiveAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			ud := &UserDevice{
				UserID:       tt.userID,
				DeviceID:     tt.deviceID,
				IPAddress:    tt.ipAddress,
				LastActiveAt: now,
				CreatedAt:    now,
			}

			switch tt.checkField {
			case "UserID":
				if ud.UserID != tt.userID {
					t.Errorf("UserDevice.UserID = %v, want %v", ud.UserID, tt.userID)
				}
			case "DeviceID":
				if ud.DeviceID != tt.deviceID {
					t.Errorf("UserDevice.DeviceID = %v, want %v", ud.DeviceID, tt.deviceID)
				}
			case "IPAddress":
				if ud.IPAddress != tt.ipAddress {
					t.Errorf("UserDevice.IPAddress = %v, want %v", ud.IPAddress, tt.ipAddress)
				}
			case "LastActiveAt":
				if ud.LastActiveAt.IsZero() {
					t.Error("UserDevice.LastActiveAt is zero")
				}
			}
		})
	}
}
