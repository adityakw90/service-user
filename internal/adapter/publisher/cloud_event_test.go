package publisher

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

func TestToCloudEventData(t *testing.T) {
	tests := []struct {
		name      string
		eventType event.EventType
		eventData any
		source    string
		wantType  string
	}{
		{
			name:      "Login event with simple data",
			eventType: event.EventLogin,
			eventData: event.EventLoginData{
				Identifier:     "test@example.com",
				IdentifierType: "email",
			},
			source:   "service-user",
			wantType: "auth.login",
		},
		{
			name:      "Failed login event with simple data",
			eventType: event.EventLoginFailed,
			eventData: event.EventLoginFailedData{
				Identifier:     "baduser@example.com",
				IdentifierType: "email",
				FailureReason:  "invalid_credentials",
			},
			source:   "service-user",
			wantType: "auth.login_failed",
		},
		{
			name:      "User created event",
			eventType: event.EventUserCreated,
			eventData: event.EventUserCreatedData{
				UserUID:  "user-123",
				ActorUID: "system",
				Username: "testuser",
				Email:    "test@example.com",
				Status:   "active",
			},
			source:   "service-user",
			wantType: "user.created",
		},
		{
			name:      "PIN create event",
			eventType: event.EventUserCreatePin,
			eventData: map[string]interface{}{
				"user_uid": "user-123",
			},
			source:   "service-user",
			wantType: "user.create_pin",
		},
		{
			name:      "Device created event",
			eventType: event.EventDeviceCreated,
			eventData: event.EventDeviceCreatedData{
				UserUID:   "user-123",
				DeviceUID: "device-456",
			},
			source:   "service-user",
			wantType: "device.created",
		},
		{
			name:      "File created event",
			eventType: event.EventUserFileCreated,
			eventData: event.EventUserFileCreatedData{
				UserUID:  "user-123",
				FileUID:  "file-789",
				FileName: "document.pdf",
			},
			source:   "service-user",
			wantType: "user_file.created",
		},
		{
			name:      "OAuth login event",
			eventType: "auth.oauth_login",
			eventData: event.EventOAuthLoginData{
				UserUID:  "user-123",
				Provider: "google",
			},
			source:   "service-user",
			wantType: "auth.oauth_login",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toCloudEventData(tt.eventType, tt.eventData, tt.source)

			fmt.Println(got)

			if got.Type != tt.wantType {
				t.Errorf("toCloudEventData() Type = %v, want %v", got.Type, tt.wantType)
			}

			if got.Source != tt.source {
				t.Errorf("toCloudEventData() Source = %v, want %v", got.Source, tt.source)
			}

			if got.SpecVersion != "1.0" {
				t.Errorf("toCloudEventData() SpecVersion = %v, want %v", got.SpecVersion, "1.0")
			}

			if got.ID == "" {
				t.Errorf("toCloudEventData() ID should not be empty")
			}

			if got.Time == "" {
				t.Errorf("toCloudEventData() Time should not be empty")
			}

			// Verify data can be unmarshaled
			var data map[string]interface{}
			if err := json.Unmarshal(got.Data, &data); err != nil {
				t.Errorf("toCloudEventData() Data is not valid JSON: %v", err)
			}
		})
	}
}

func TestToCloudEvent(t *testing.T) {
	tests := []struct {
		name     string
		event    any
		source   string
		wantType string
		wantErr  bool
	}{
		{
			name: "Login event data",
			event: event.EventLoginData{
				Identifier:     "test@example.com",
				IdentifierType: "email",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "User created event data",
			event: event.EventUserCreatedData{
				UserUID:  "user-123",
				ActorUID: "system",
				Username: "testuser",
				Email:    "test@example.com",
				Status:   "active",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "PIN create event data",
			event: map[string]interface{}{
				"user_uid": "user-123",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "Device created event data",
			event: event.EventDeviceCreatedData{
				UserUID:   "user-123",
				DeviceUID: "device-456",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "File created event data",
			event: event.EventUserFileCreatedData{
				UserUID:  "user-123",
				FileUID:  "file-789",
				FileName: "document.pdf",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
		{
			name: "Simple map",
			event: map[string]any{
				"user_uid": "user-123",
				"email":    "test@example.com",
			},
			source:   "service-user",
			wantType: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ToCloudEvent(tt.event, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToCloudEvent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Source != tt.source {
				t.Errorf("ToCloudEvent() Source = %v, want %v", got.Source, tt.source)
			}

			if got.SpecVersion != "1.0" {
				t.Errorf("ToCloudEvent() SpecVersion = %v, want %v", got.SpecVersion, "1.0")
			}

			if got.ID == "" {
				t.Errorf("ToCloudEvent() ID should not be empty")
			}

			if got.Time == "" {
				t.Errorf("ToCloudEvent() Time should not be empty")
			}

			// Verify data can be unmarshaled
			var data map[string]interface{}
			if err := json.Unmarshal(got.Data, &data); err != nil {
				t.Errorf("ToCloudEvent() Data is not valid JSON: %v", err)
			}
		})
	}
}
