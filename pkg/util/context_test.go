package util

import (
	"context"
	"testing"
)

func TestPkg_Util_SetClientName_GetClientName(t *testing.T) {
	tests := []struct {
		name         string
		ctx          context.Context
		expectedName string
	}{
		{
			name:         "returns set client name",
			ctx:          SetClientName(context.Background(), "test-client"),
			expectedName: "test-client",
		},
		{
			name:         "returns unknown when client name not set",
			ctx:          context.Background(),
			expectedName: "unknown",
		},
		{
			name:         "preserves client name through context chain",
			ctx:          SetClientName(SetClientName(context.Background(), "first"), "second"),
			expectedName: "second",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetClientName(tt.ctx)
			if got != tt.expectedName {
				t.Errorf("GetClientName() = %v, want %v", got, tt.expectedName)
			}
		})
	}
}

func TestPkg_Util_SetActor_GetActor(t *testing.T) {
	tests := []struct {
		name              string
		ctx               context.Context
		expectedActorId   string
		expectedActorType string
		expectedActorName string
	}{
		{
			name:              "returns set actor id and type",
			ctx:               SetActor(context.Background(), "user-123", "user", "Alice"),
			expectedActorId:   "user-123",
			expectedActorType: "user",
			expectedActorName: "Alice",
		},
		{
			name:              "returns unknown for all when not set",
			ctx:               context.Background(),
			expectedActorId:   "unknown",
			expectedActorType: "unknown",
			expectedActorName: "unknown",
		},
		{
			name:              "handles empty actor id with valid type",
			ctx:               SetActor(context.Background(), "", "system", "System User"),
			expectedActorId:   "",
			expectedActorType: "system",
			expectedActorName: "System User",
		},
		{
			name:              "handles valid actor id with empty type",
			ctx:               SetActor(context.Background(), "admin-456", "", "Admin"),
			expectedActorId:   "admin-456",
			expectedActorType: "",
			expectedActorName: "Admin",
		},
		{
			name:              "handles empty actor name with valid id and type",
			ctx:               SetActor(context.Background(), "user-789", "user", ""),
			expectedActorId:   "user-789",
			expectedActorType: "user",
			expectedActorName: "",
		},
		{
			name:              "preserves actor through context chain",
			ctx:               SetActor(SetActor(context.Background(), "old-id", "old-type", "Old Name"), "new-id", "new-type", "New Name"),
			expectedActorId:   "new-id",
			expectedActorType: "new-type",
			expectedActorName: "New Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actorId, actorType, actorName := GetActor(tt.ctx)
			if actorId != tt.expectedActorId {
				t.Errorf("GetActor() actorId = %v, want %v", actorId, tt.expectedActorId)
			}
			if actorType != tt.expectedActorType {
				t.Errorf("GetActor() actorType = %v, want %v", actorType, tt.expectedActorType)
			}
			if actorName != tt.expectedActorName {
				t.Errorf("GetActor() actorName = %v, want %v", actorName, tt.expectedActorName)
			}
		})
	}
}

func TestPkg_Util_ContextKeyIndependence(t *testing.T) {
	tests := []struct {
		name       string
		clientName string
		actorId    string
		actorType  string
		actorName  string
	}{
		{
			name:       "all values set",
			clientName: "my-client",
			actorId:    "user-789",
			actorType:  "user",
			actorName:  "Alice",
		},
		{
			name:       "only client name set",
			clientName: "test-client",
			actorId:    "",
			actorType:  "",
			actorName:  "",
		},
		{
			name:       "only actor set",
			clientName: "",
			actorId:    "system-001",
			actorType:  "system",
			actorName:  "System User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			if tt.clientName != "" {
				ctx = SetClientName(ctx, tt.clientName)
			}

			if tt.actorId != "" || tt.actorType != "" {
				ctx = SetActor(ctx, tt.actorId, tt.actorType, tt.actorName)
			}

			gotClientName := GetClientName(ctx)
			if tt.clientName != "" {
				if gotClientName != tt.clientName {
					t.Errorf("GetClientName() = %v, want %v", gotClientName, tt.clientName)
				}
			} else {
				if gotClientName != "unknown" {
					t.Errorf("GetClientName() = %v, want unknown", gotClientName)
				}
			}

			gotActorId, gotActorType, gotActorName := GetActor(ctx)
			if tt.actorId != "" {
				if gotActorId != tt.actorId {
					t.Errorf("GetActor() actorId = %v, want %v", gotActorId, tt.actorId)
				}
			} else {
				if gotActorId != "unknown" {
					t.Errorf("GetActor() actorId = %v, want unknown", gotActorId)
				}
			}

			if tt.actorType != "" {
				if gotActorType != tt.actorType {
					t.Errorf("GetActor() actorType = %v, want %v", gotActorType, tt.actorType)
				}
			} else {
				if gotActorType != "unknown" {
					t.Errorf("GetActor() actorType = %v, want unknown", gotActorType)
				}
			}

			if tt.actorName != "" {
				if gotActorName != tt.actorName {
					t.Errorf("GetActor() actorName = %v, want %v", gotActorName, tt.actorName)
				}
			} else {
				if gotActorName != "unknown" {
					t.Errorf("GetActor() actorName = %v, want unknown", gotActorName)
				}
			}
		})
	}
}
