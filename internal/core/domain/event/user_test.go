package event

import (
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/model"
)

func TestNewUserEntity(t *testing.T) {
	tests := []struct {
		name     string
		user     *model.User
		wantID   string
		wantType string
		wantName string
	}{
		{
			name:     "user entity",
			user:     &model.User{UID: "user-1", Username: "ada"},
			wantID:   "user-1",
			wantType: "user",
			wantName: "ada",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewUserEntity(tt.user)

			if entity.ID != tt.wantID {
				t.Errorf("NewUserEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewUserEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name == nil {
				t.Fatal("NewUserEntity() Name = nil, want username")
			}
			if *entity.Name != tt.wantName {
				t.Errorf("NewUserEntity() Name = %q, want %q", *entity.Name, tt.wantName)
			}
		})
	}
}
