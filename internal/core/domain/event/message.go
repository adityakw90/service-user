package event

import (
	"errors"
	"strings"
)

var (
	ErrEventTypeRequired  = errors.New("event type is required")
	ErrEntityTypeRequired = errors.New("event entity type is required")
	ErrEntityIDRequired   = errors.New("event entity id is required")
)

// Entity identifies the business resource affected by an event.
type Entity struct {
	ID   string
	Type string
	Name *string
}

// Message is the domain-level contract passed to event publishers.
type Message struct {
	Type     EventType
	Entity   Entity
	Metadata any
}

// NewMessage creates a valid domain event message.
func NewMessage(eventType EventType, entity Entity, metadata any) (Message, error) {
	message := Message{Type: eventType, Entity: entity, Metadata: metadata}
	if err := message.Validate(); err != nil {
		return Message{}, err
	}
	return message, nil
}

// Validate checks fields required by the audit service.
func (m Message) Validate() error {
	if strings.TrimSpace(string(m.Type)) == "" {
		return ErrEventTypeRequired
	}
	if strings.TrimSpace(m.Entity.Type) == "" {
		return ErrEntityTypeRequired
	}
	if strings.TrimSpace(m.Entity.ID) == "" {
		return ErrEntityIDRequired
	}
	return nil
}
