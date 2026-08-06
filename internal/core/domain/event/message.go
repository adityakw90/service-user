package event

import (
	"strings"

	domainError "github.com/adityakw90/service-user/internal/core/domain/errors"
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
		return domainError.ErrEventTypeRequired
	}
	if strings.TrimSpace(m.Entity.Type) == "" {
		return domainError.ErrEntityTypeRequired
	}
	if strings.TrimSpace(m.Entity.ID) == "" {
		return domainError.ErrEntityIDRequired
	}
	return nil
}
