package domain

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
	now := time.Now().UTC()
	p := &UserPin{
		UserID:    1,
		Code:      "hashed_pin",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if p.UserID != 1 {
		t.Errorf("UserPin.UserID = %v, want 1", p.UserID)
	}
	if p.Code != "hashed_pin" {
		t.Errorf("UserPin.Code = %v, want 'hashed_pin'", p.Code)
	}
	if p.CreatedAt.IsZero() {
		t.Error("UserPin.CreatedAt is zero")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("UserPin.UpdatedAt is zero")
	}
}
