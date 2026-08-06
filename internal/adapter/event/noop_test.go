package event

import (
	"context"
	"testing"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

func TestNoOpPublisher(t *testing.T) {
	pub := NewNoOpPublisher()

	ctx := context.Background()

	// Should not return error
	err := pub.Publish(ctx, event.Message{Type: event.EventLogin, Entity: event.Entity{ID: "entity-1", Type: "user"}, Metadata: event.EventLoginData{
		Identifier:     "test@example.com",
		IdentifierType: "email",
	}})
	if err != nil {
		t.Errorf("NoOpPublisher.Publish() error = %v", err)
	}

	// Close should not return error
	err = pub.Close()
	if err != nil {
		t.Errorf("NoOpPublisher.Close() error = %v", err)
	}
}
