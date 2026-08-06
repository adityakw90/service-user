package event

import (
	"errors"
	"testing"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
)

func TestNewMessage(t *testing.T) {
	entityName := "Ada"
	tests := []struct {
		name      string
		eventType EventType
		entity    Entity
		metadata  any
	}{
		{
			name:      "valid message",
			eventType: EventUserUpdated,
			entity:    Entity{ID: "user-1", Type: "user", Name: &entityName},
			metadata:  EventUserUpdatedData{UserUID: "user-1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message, err := NewMessage(tt.eventType, tt.entity, tt.metadata)
			if err != nil {
				t.Fatalf("NewMessage() error = %v", err)
			}
			if message.Type != tt.eventType || message.Entity != tt.entity {
				t.Fatalf("NewMessage() = %#v", message)
			}
		})
	}
}

func TestNewMessageValidation(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		entity    Entity
		want      error
	}{
		{name: "event type", entity: Entity{ID: "1", Type: "user"}, want: domainError.ErrEventTypeRequired},
		{name: "entity type", eventType: EventUserUpdated, entity: Entity{ID: "1"}, want: domainError.ErrEntityTypeRequired},
		{name: "entity id", eventType: EventUserUpdated, entity: Entity{Type: "user"}, want: domainError.ErrEntityIDRequired},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMessage(tt.eventType, tt.entity, nil)
			if !errors.Is(err, tt.want) {
				t.Fatalf("NewMessage() error = %v, want %v", err, tt.want)
			}
		})
	}
}
