package event

import (
	"context"
	"encoding/json"
	"regexp"
	"testing"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	domainevent "github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/pkg/util"
)

func isValidUUID(id string) bool {
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	return pattern.MatchString(id)
}

func setupContext(ctx context.Context, clientName, actorId, actorType string, actorNames ...string) context.Context {
	actorName := "test-actor"
	if len(actorNames) > 0 {
		actorName = actorNames[0]
	}
	ctx = util.SetClientName(ctx, clientName)
	ctx = util.SetActor(ctx, actorId, actorType, actorName)
	return ctx
}

func TestNewCloudEvent(t *testing.T) {
	tests := []struct {
		name        string
		setupCtx    func() context.Context
		eventType   event.EventType
		eventData   any
		wantSource  string
		wantSpecVer string
		validate    func(t *testing.T, ce CloudEvent)
	}{
		{
			name: "Happy Path - EventLoginData with full context",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test-client", "user-123", "user", "test-actor")
			},
			eventType: event.EventLogin,
			eventData: event.EventLoginData{
				Identifier:     "test@example.com",
				IdentifierType: "email",
				DeviceUID:      util.Ptr("device-123"),
				DeviceName:     util.Ptr("iPhone"),
				IPAddress:      util.Ptr("192.168.1.1"),
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventLogin) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventLogin)
				}
				if ce.Data.Client != "test-client" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "test-client")
				}
				if ce.Data.ActorId != "user-123" {
					t.Errorf("ActorId = %v, want %v", ce.Data.ActorId, "user-123")
				}
				if ce.Data.ActorType != "user" {
					t.Errorf("ActorType = %v, want %v", ce.Data.ActorType, "user")
				}
				if ce.Data.ActorName != "test-actor" {
					t.Errorf("ActorName = %v, want %v", ce.Data.ActorName, "test-actor")
				}
				if !isValidUUID(ce.ID) {
					t.Errorf("ID = %v is not a valid UUID", ce.ID)
				}
				var loginData event.EventLoginData
				if err := json.Unmarshal(ce.Data.MetaData, &loginData); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if loginData.Identifier != "test@example.com" {
					t.Errorf("Identifier = %v, want %v", loginData.Identifier, "test@example.com")
				}
			},
		},
		{
			name: "Context Defaults - Empty context uses unknown values",
			setupCtx: func() context.Context {
				return context.Background()
			},
			eventType:   event.EventLogin,
			eventData:   event.EventLoginData{Identifier: "test@example.com"},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Data.Client != "unknown" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "unknown")
				}
				if ce.Data.ActorId != "unknown" {
					t.Errorf("ActorId = %v, want %v", ce.Data.ActorId, "unknown")
				}
				if ce.Data.ActorType != "unknown" {
					t.Errorf("ActorType = %v, want %v", ce.Data.ActorType, "unknown")
				}
				if ce.Data.ActorName != "unknown" {
					t.Errorf("ActorName = %v, want %v", ce.Data.ActorName, "unknown")
				}
			},
		},
		{
			name: "EventLoginFailedData",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "mobile-app", "user-456", "user")
			},
			eventType: event.EventLoginFailed,
			eventData: event.EventLoginFailedData{
				Identifier:     "baduser@example.com",
				IdentifierType: "email",
				FailureReason:  "invalid_password",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventLoginFailed) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventLoginFailed)
				}
				var data event.EventLoginFailedData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.FailureReason != "invalid_password" {
					t.Errorf("FailureReason = %v, want %v", data.FailureReason, "invalid_password")
				}
			},
		},
		{
			name: "EventPinVerifyData - Success case",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "web", "user-789", "user")
			},
			eventType: event.EventPINVerify,
			eventData: event.EventPinVerifyData{
				Success: true,
				Reason:  "valid_pin",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventPINVerify) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventPINVerify)
				}
				var data event.EventPinVerifyData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if !data.Success {
					t.Errorf("Success = %v, want %v", data.Success, true)
				}
			},
		},
		{
			name: "EventPinVerifyData - Failure case",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "web", "user-789", "user")
			},
			eventType: event.EventPINVerify,
			eventData: event.EventPinVerifyData{
				Success: false,
				Reason:  "invalid_pin",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				var data event.EventPinVerifyData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.Success {
					t.Errorf("Success = %v, want %v", data.Success, false)
				}
				if data.Reason != "invalid_pin" {
					t.Errorf("Reason = %v, want %v", data.Reason, "invalid_pin")
				}
			},
		},
		{
			name: "EventLoginLocked",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "auth-service", "system-1", "system")
			},
			eventType: event.EventLoginLocked,
			eventData: event.EventLoginLockedData{
				Identifier:     "locked@example.com",
				IdentifierType: "email",
				FailureReason:  "too_many_attempts",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventLoginLocked) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventLoginLocked)
				}
				var data event.EventLoginLockedData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.FailureReason != "too_many_attempts" {
					t.Errorf("FailureReason = %v, want %v", data.FailureReason, "too_many_attempts")
				}
			},
		},
		{
			name: "Marshal Error - Non-serializable data",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "test", "actor-1", "system")
			},
			eventType:   event.EventLogin,
			eventData:   make(chan int),
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				var errData map[string]any
				if err := json.Unmarshal(ce.Data.MetaData, &errData); err != nil {
					t.Fatalf("Failed to unmarshal error data: %v", err)
				}
				if _, ok := errData["error"]; !ok {
					t.Errorf("MetaData should contain 'error' key, got %v", errData)
				}
			},
		},
		{
			name: "User Created Event",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "admin-panel", "admin-1", "admin")
			},
			eventType: event.EventUserCreated,
			eventData: struct {
				UserUID string `json:"user_uid"`
				Email   string `json:"email"`
			}{
				UserUID: "new-user-123",
				Email:   "newuser@example.com",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventUserCreated) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventUserCreated)
				}
				if ce.Data.Client != "admin-panel" {
					t.Errorf("Client = %v, want %v", ce.Data.Client, "admin-panel")
				}
			},
		},
		{
			name: "Token Refresh Event",
			setupCtx: func() context.Context {
				return setupContext(context.Background(), "mobile-app", "user-123", "user")
			},
			eventType: event.EventTokenRefresh,
			eventData: event.EventTokenRefreshData{
				Identifier:     "test@example.com",
				IdentifierType: "email",
			},
			wantSource:  Source,
			wantSpecVer: SpecVersion,
			validate: func(t *testing.T, ce CloudEvent) {
				if ce.Type != string(event.EventTokenRefresh) {
					t.Errorf("Type = %v, want %v", ce.Type, event.EventTokenRefresh)
				}
				var data event.EventTokenRefreshData
				if err := json.Unmarshal(ce.Data.MetaData, &data); err != nil {
					t.Errorf("Failed to unmarshal MetaData: %v", err)
				}
				if data.Identifier != "test@example.com" {
					t.Errorf("Identifier = %v, want %v", data.Identifier, "test@example.com")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := tt.setupCtx()
			ce := NewCloudEvent(ctx, event.Message{Type: tt.eventType, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: tt.eventData})

			if ce.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", ce.Source, tt.wantSource)
			}
			if ce.SpecVersion != tt.wantSpecVer {
				t.Errorf("SpecVersion = %v, want %v", ce.SpecVersion, tt.wantSpecVer)
			}
			if tt.validate != nil {
				tt.validate(t, ce)
			}
		})
	}
}

func TestNewCloudEvent_Constants(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")
	ce := NewCloudEvent(ctx, event.Message{Type: event.EventLogin, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: event.EventLoginData{}})

	if ce.Source != Source {
		t.Errorf("Source = %v, want %v", ce.Source, Source)
	}
	if ce.SpecVersion != SpecVersion {
		t.Errorf("SpecVersion = %v, want %v", ce.SpecVersion, SpecVersion)
	}
}

func TestNewCloudEvent_IDGeneration(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")

	ids := make(map[string]bool)
	for range 100 {
		ce := NewCloudEvent(ctx, event.Message{Type: event.EventLogin, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: event.EventLoginData{}})
		if !isValidUUID(ce.ID) {
			t.Errorf("ID = %v is not a valid UUID", ce.ID)
		}
		if ids[ce.ID] {
			t.Errorf("Duplicate ID generated: %v", ce.ID)
		}
		ids[ce.ID] = true
	}
}

func TestNewCloudEvent_TimeFormat(t *testing.T) {
	t.Parallel()
	ctx := setupContext(context.Background(), "test", "actor-1", "system")

	ce := NewCloudEvent(ctx, event.Message{Type: event.EventLogin, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: event.EventLoginData{}})

	now := time.Now().UTC()
	eventTime := ce.Time.UTC()
	if eventTime.After(now.Add(5 * time.Second)) {
		t.Errorf("Time = %v is too far in the future", ce.Time)
	}
	if eventTime.Before(now.Add(-5 * time.Second)) {
		t.Errorf("Time = %v is too far in the past", ce.Time)
	}
}

func TestNewCloudEventCopiesEntity(t *testing.T) {
	name := "Ada"
	ce := NewCloudEvent(context.Background(), domainevent.Message{
		Type:     domainevent.EventUserUpdated,
		Entity:   domainevent.Entity{ID: "user-1", Type: "user", Name: &name},
		Metadata: domainevent.EventUserUpdatedData{ChangesCount: 1},
	})

	if ce.Data.EntityId != "user-1" || ce.Data.EntityType != "user" {
		t.Fatalf("entity = %q/%q", ce.Data.EntityType, ce.Data.EntityId)
	}
	if ce.Data.EntityName == nil || *ce.Data.EntityName != name {
		t.Fatalf("entity name = %v", ce.Data.EntityName)
	}
}
