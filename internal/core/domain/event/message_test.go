package event

import (
	"errors"
	"testing"
)

func TestNewMessage(t *testing.T) {
	entityName := "Ada"
	entity := Entity{ID: "user-1", Type: "user", Name: &entityName}
	metadata := EventUserUpdatedData{UserUID: "user-1"}

	message, err := NewMessage(EventUserUpdated, entity, metadata)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	if message.Type != EventUserUpdated || message.Entity != entity {
		t.Fatalf("NewMessage() = %#v", message)
	}
}

func TestNewMessageValidation(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		entity    Entity
		want      error
	}{
		{name: "event type", entity: Entity{ID: "1", Type: "user"}, want: ErrEventTypeRequired},
		{name: "entity type", eventType: EventUserUpdated, entity: Entity{ID: "1"}, want: ErrEntityTypeRequired},
		{name: "entity id", eventType: EventUserUpdated, entity: Entity{Type: "user"}, want: ErrEntityIDRequired},
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
