package event

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

// EventPublisher is the secondary port for publishing domain events.
// Supports pluggable implementations (CloudEvents, Kafka, Pub/Sub).
type EventPublisher interface {
	// Publish publishes a domain event with the given event type and data.
	Publish(ctx context.Context, eventType event.EventType, eventData any) error

	// Close closes the publisher connection.
	Close() error
}
