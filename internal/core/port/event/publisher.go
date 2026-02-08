package event

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/event"
)

// EventPublisher is the secondary port for publishing authentication events.
// Supports pluggable implementations (CloudEvents, Kafka, Pub/Sub).
type EventPublisher interface {
	// Publish publishes an authentication event.
	Publish(ctx context.Context, event *event.AuthEvent) error

	// Close closes the publisher connection.
	Close() error
}
