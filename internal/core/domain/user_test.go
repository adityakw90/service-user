package domain

import (
	"testing"
	"time"
)

func TestUser_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status UserStatus
		want   bool
	}{
		{"inactive user", UserStatusInactive, false},
		{"active user", UserStatusActive, true},
		{"banned user", UserStatusBanned, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.status}
			if got := u.IsActive(); got != tt.want {
				t.Errorf("User.IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_IsDeleted(t *testing.T) {
	tests := []struct {
		name      string
		deletedAt *time.Time
		want      bool
	}{
		{"not deleted", nil, false},
		{"soft deleted", func() *time.Time { t := time.Now(); return &t }(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{DeletedAt: tt.deletedAt}
			if got := u.IsDeleted(); got != tt.want {
				t.Errorf("User.IsDeleted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_CanLogin(t *testing.T) {
	tests := []struct {
		name      string
		status    UserStatus
		deletedAt *time.Time
		want      bool
	}{
		{"active and not deleted", UserStatusActive, nil, true},
		{"inactive and not deleted", UserStatusInactive, nil, false},
		{"banned and not deleted", UserStatusBanned, nil, false},
		{"active but deleted", UserStatusActive, func() *time.Time { t := time.Now(); return &t }(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.status, DeletedAt: tt.deletedAt}
			if got := u.CanLogin(); got != tt.want {
				t.Errorf("User.CanLogin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUser_Activate(t *testing.T) {
	tests := []struct {
		name        string
		initialStatus UserStatus
		wantErr     bool
		wantStatus  UserStatus
	}{
		{"activate inactive user", UserStatusInactive, false, UserStatusActive},
		{"activate active user", UserStatusActive, false, UserStatusActive},
		{"cannot activate banned user", UserStatusBanned, true, UserStatusBanned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.initialStatus}
			err := u.Activate()
			if (err != nil) != tt.wantErr {
				t.Errorf("User.Activate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if u.Status != tt.wantStatus {
				t.Errorf("User.Status = %v, want %v", u.Status, tt.wantStatus)
			}
		})
	}
}

func TestUser_Deactivate(t *testing.T) {
	u := &User{Status: UserStatusActive}
	u.Deactivate()
	if u.Status != UserStatusInactive {
		t.Errorf("User.Deactivate() = %v, want %v", u.Status, UserStatusInactive)
	}
}

func TestUser_Ban(t *testing.T) {
	tests := []struct {
		name         string
		initialStatus UserStatus
		wantStatus   UserStatus
	}{
		{"ban active user", UserStatusActive, UserStatusBanned},
		{"ban inactive user", UserStatusInactive, UserStatusBanned},
		{"ban already banned user", UserStatusBanned, UserStatusBanned},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{Status: tt.initialStatus}
			u.Ban()
			if u.Status != tt.wantStatus {
				t.Errorf("User.Ban() = %v, want %v", u.Status, tt.wantStatus)
			}
		})
	}
}

func TestUserStatus_Values(t *testing.T) {
	if UserStatusInactive != 0 {
		t.Errorf("UserStatusInactive = %v, want 0", UserStatusInactive)
	}
	if UserStatusActive != 1 {
		t.Errorf("UserStatusActive = %v, want 1", UserStatusActive)
	}
	if UserStatusBanned != -1 {
		t.Errorf("UserStatusBanned = %v, want -1", UserStatusBanned)
	}
}
