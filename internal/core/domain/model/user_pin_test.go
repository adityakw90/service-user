package model

import (
	"testing"
	"time"
)

func TestUserPin_IsSet(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{"empty code", "", false},
		{"non-empty code", "hashed_pin_value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &UserPin{Code: tt.code}
			if got := p.IsSet(); got != tt.want {
				t.Errorf("UserPin.IsSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserPin_Timestamps(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		code       string
		checkField string
	}{
		{"UserID is set correctly", 1, "hashed_pin", "UserID"},
		{"Code is set correctly", 1, "hashed_pin", "Code"},
		{"CreatedAt is set", 1, "hashed_pin", "CreatedAt"},
		{"UpdatedAt is set", 1, "hashed_pin", "UpdatedAt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			now := time.Now().UTC()
			p := &UserPin{
				UserID:    tt.userID,
				Code:      tt.code,
				CreatedAt: now,
				UpdatedAt: now,
			}

			switch tt.checkField {
			case "UserID":
				if p.UserID != tt.userID {
					t.Errorf("UserPin.UserID = %v, want %v", p.UserID, tt.userID)
				}
			case "Code":
				if p.Code != tt.code {
					t.Errorf("UserPin.Code = %v, want %v", p.Code, tt.code)
				}
			case "CreatedAt":
				if p.CreatedAt.IsZero() {
					t.Error("UserPin.CreatedAt is zero")
				}
			case "UpdatedAt":
				if p.UpdatedAt.IsZero() {
					t.Error("UserPin.UpdatedAt is zero")
				}
			}
		})
	}
}
