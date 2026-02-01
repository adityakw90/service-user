package domain

import (
	"testing"
	"time"
)

func TestUserProfile_HasAvatar(t *testing.T) {
	tests := []struct {
		name         string
		avatarFileID *int64
		want         bool
	}{
		{"no avatar", nil, false},
		{"zero avatar file id", int64Ptr(0), false},
		{"valid avatar file id", int64Ptr(123), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserProfile{AvatarFileID: tt.avatarFileID}
			if got := p.HasAvatar(); got != tt.want {
				t.Errorf("UserProfile.HasAvatar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserProfile_FullName(t *testing.T) {
	tests := []struct {
		name     string
		firstName string
		lastName  string
		want     string
	}{
		{"both empty", "", "", ""},
		{"only first name", "John", "", "John"},
		{"only last name", "", "Doe", "Doe"},
		{"both names", "John", "Doe", "John Doe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserProfile{FirstName: tt.firstName, LastName: tt.lastName}
			if got := p.FullName(); got != tt.want {
				t.Errorf("UserProfile.FullName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserProfile_Timestamps(t *testing.T) {
	now := time.Now().UTC()
	p := &UserProfile{
		UserID:    1,
		CreatedAt: now,
		UpdatedAt: now,
		Attributes: make(map[string]any),
	}

	if p.UserID != 1 {
		t.Errorf("UserProfile.UserID = %v, want 1", p.UserID)
	}
	if p.CreatedAt.IsZero() {
		t.Error("UserProfile.CreatedAt is zero")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UserProfile.UpdatedAt is zero")
	}
}

func TestUserProfile_Attributes(t *testing.T) {
	attrs := map[string]any{
		"key1": "value1",
		"key2": 123,
	}
	p := &UserProfile{Attributes: attrs}

	if p.Attributes["key1"] != "value1" {
		t.Error("UserProfile.Attributes key1 mismatch")
	}
	if p.Attributes["key2"] != 123 {
		t.Error("UserProfile.Attributes key2 mismatch")
	}
}

// Helper function
func int64Ptr(v int64) *int64 {
	return &v
}
