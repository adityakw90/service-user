package event

import (
	"testing"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
	"github.com/adityakw90/service-user/internal/core/domain/model"
	"github.com/adityakw90/service-user/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestEntity_Validation(t *testing.T) {
	tests := []struct {
		name    string
		entity  Entity
		wantErr error
	}{
		{name: "valid entity", entity: Entity{ID: "1", Type: EntityTypeUser}, wantErr: nil},
		{name: "valid entity with name", entity: Entity{ID: "1", Type: EntityTypeUser, Name: util.Ptr("name")}, wantErr: nil},
		{name: "entity type", entity: Entity{ID: "1"}, wantErr: domainError.ErrEntityTypeRequired},
		{name: "entity id", entity: Entity{Type: "user"}, wantErr: domainError.ErrEntityIDRequired},
		{name: "entity type invalid", entity: Entity{ID: "1", Type: "invalid"}, wantErr: domainError.ErrEntityInvalidType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.Validate()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)

			}
		})
	}
}

func TestEntity_checkType(t *testing.T) {
	tests := []struct {
		name    string
		entity  Entity
		wantErr error
	}{
		{name: "valid entity user", entity: Entity{ID: "1", Type: EntityTypeUser}, wantErr: nil},
		{name: "valid entity device", entity: Entity{ID: "1", Type: EntityTypeDevice}, wantErr: nil},
		{name: "valid entity user file", entity: Entity{ID: "1", Type: EntityTypeUserFile}, wantErr: nil},
		{name: "valid entity user profile", entity: Entity{ID: "1", Type: EntityTypeUserProfile}, wantErr: nil},
		{name: "valid entity user pin", entity: Entity{ID: "1", Type: EntityTypeUserPIN}, wantErr: nil},
		{name: "valid entity session", entity: Entity{ID: "1", Type: EntityTypeSession}, wantErr: nil},
		{name: "empty entity type", entity: Entity{ID: "1"}, wantErr: domainError.ErrEntityInvalidType},
		{name: "invalid entity type whitespace", entity: Entity{ID: "1", Type: " "}, wantErr: domainError.ErrEntityInvalidType},
		{name: "invalid entity type invalid", entity: Entity{ID: "1", Type: "invalid"}, wantErr: domainError.ErrEntityInvalidType},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entity.checkType()
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)

			}
		})
	}
}

func TestEntity_NewUserEntity(t *testing.T) {
	tests := []struct {
		name     string
		user     *model.User
		wantID   string
		wantType EntityType
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

func TestEntity_NewTokenEntity(t *testing.T) {
	tests := []struct {
		name     string
		claims   *model.TokenClaims
		wantID   string
		wantType EntityType
	}{
		{
			name:     "token entity",
			claims:   &model.TokenClaims{Sid: "session-1", Identifier: "ada@example.com", IdentifierType: "email"},
			wantID:   "session-1",
			wantType: EntityTypeSession,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewTokenEntity(tt.claims)

			if entity.ID != tt.wantID {
				t.Errorf("NewTokenEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewTokenEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name != nil {
				t.Errorf("NewTokenEntity() Name = %q, want nil", *entity.Name)
			}
		})
	}
}

func TestEntity_NewDeviceEntity(t *testing.T) {
	tests := []struct {
		name     string
		device   *model.Device
		wantID   string
		wantType EntityType
		wantName string
	}{
		{
			name:     "device entity",
			device:   &model.Device{UID: "device-1", DeviceName: "Phone"},
			wantID:   "device-1",
			wantType: EntityTypeDevice,
			wantName: "Phone",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewDeviceEntity(tt.device)

			if entity.ID != tt.wantID {
				t.Errorf("NewDeviceEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewDeviceEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name == nil {
				t.Fatal("NewDeviceEntity() Name = nil, want device name")
			}
			if *entity.Name != tt.wantName {
				t.Errorf("NewDeviceEntity() Name = %q, want %q", *entity.Name, tt.wantName)
			}
		})
	}
}

func TestEntity_NewUserFileEntity(t *testing.T) {
	tests := []struct {
		name     string
		file     *model.UserFile
		wantID   string
		wantType EntityType
		wantName string
	}{
		{
			name:     "user file entity",
			file:     &model.UserFile{UID: "file-1", FileName: "avatar.png"},
			wantID:   "file-1",
			wantType: EntityTypeUserFile,
			wantName: "avatar.png",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewUserFileEntity(tt.file)

			if entity.ID != tt.wantID {
				t.Errorf("NewUserFileEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewUserFileEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name == nil {
				t.Fatal("NewUserFileEntity() Name = nil, want file name")
			}
			if *entity.Name != tt.wantName {
				t.Errorf("NewUserFileEntity() Name = %q, want %q", *entity.Name, tt.wantName)
			}
		})
	}
}

func TestEntity_NewUserProfileEntity(t *testing.T) {
	tests := []struct {
		name     string
		profile  *model.UserProfile
		wantID   string
		wantType EntityType
		wantName string
	}{
		{
			name:     "user profile entity",
			profile:  &model.UserProfile{UserUID: "user-1", FirstName: "Ada", LastName: "Lovelace"},
			wantID:   "user-1",
			wantType: EntityTypeUserProfile,
			wantName: "Ada Lovelace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewUserProfileEntity(tt.profile)

			if entity.ID != tt.wantID {
				t.Errorf("NewUserProfileEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewUserProfileEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name == nil {
				t.Fatal("NewUserProfileEntity() Name = nil, want profile full name")
			}
			if *entity.Name != tt.wantName {
				t.Errorf("NewUserProfileEntity() Name = %q, want %q", *entity.Name, tt.wantName)
			}
		})
	}
}

func TestEntity_NewUserPinEntity(t *testing.T) {
	tests := []struct {
		name     string
		pin      *model.UserPin
		wantID   string
		wantType EntityType
	}{
		{
			name:     "user PIN entity",
			pin:      &model.UserPin{UserUID: "user-1"},
			wantID:   "user-1",
			wantType: EntityTypeUserPIN,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewUserPinEntity(tt.pin)

			if entity.ID != tt.wantID {
				t.Errorf("NewUserPinEntity() ID = %q, want %q", entity.ID, tt.wantID)
			}
			if entity.Type != tt.wantType {
				t.Errorf("NewUserPinEntity() Type = %q, want %q", entity.Type, tt.wantType)
			}
			if entity.Name != nil {
				t.Errorf("NewUserPinEntity() Name = %q, want nil", *entity.Name)
			}
		})
	}
}
