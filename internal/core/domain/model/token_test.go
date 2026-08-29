package model_test

import (
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestTokenClaims_ExpiresAt(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt time.Time
		wantZero  bool
	}{
		{
			name:      "zero value is zero",
			expiresAt: time.Time{},
			wantZero:  true,
		},
		{
			name:      "set value is preserved",
			expiresAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			wantZero:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims := model.TokenClaims{ExpiresAt: tt.expiresAt}
			if tt.wantZero && !claims.ExpiresAt.IsZero() {
				t.Errorf("expected zero ExpiresAt, got %v", claims.ExpiresAt)
			}
			if !tt.wantZero && claims.ExpiresAt.IsZero() {
				t.Errorf("expected non-zero ExpiresAt")
			}
		})
	}
}
