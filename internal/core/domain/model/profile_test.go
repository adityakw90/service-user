package model

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user/pkg/util"
)

func TestUserProfile_HasAvatar(t *testing.T) {
	tests := []struct {
		name         string
		avatarFileID *int64
		want         bool
	}{
		{"no avatar", nil, false},
		{"zero avatar file id", util.Ptr(int64(0)), false},
		{"valid avatar file id", util.Ptr(int64(123)), true},
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
		name      string
		firstName string
		lastName  string
		want      string
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
	tests := []struct {
		name       string
		userID     int64
		checkField string
	}{
		{"UserID is set correctly", 1, "UserID"},
		{"CreatedAt is set", 1, "CreatedAt"},
		{"UpdatedAt is set", 1, "UpdatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			p := &UserProfile{
				UserID:     tt.userID,
				CreatedAt:  now,
				UpdatedAt:  now,
				Attributes: make(map[string]any),
			}

			switch tt.checkField {
			case "UserID":
				if p.UserID != tt.userID {
					t.Errorf("UserProfile.UserID = %v, want %v", p.UserID, tt.userID)
				}
			case "CreatedAt":
				if p.CreatedAt.IsZero() {
					t.Error("UserProfile.CreatedAt is zero")
				}
			case "UpdatedAt":
				if p.UpdatedAt.IsZero() {
					t.Error("UserProfile.UpdatedAt is zero")
				}
			}
		})
	}
}

func TestUserProfile_Attributes(t *testing.T) {
	tests := []struct {
		name        string
		attrs       map[string]any
		key         string
		expectedVal any
	}{
		{"string value", map[string]any{"key1": "value1"}, "key1", "value1"},
		{"int value", map[string]any{"key2": 123}, "key2", 123},
		{"multiple keys", map[string]any{"key1": "value1", "key2": 123}, "key1", "value1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserProfile{Attributes: tt.attrs}
			if p.Attributes[tt.key] != tt.expectedVal {
				t.Errorf("UserProfile.Attributes[%s] = %v, want %v", tt.key, p.Attributes[tt.key], tt.expectedVal)
			}
		})
	}
}
