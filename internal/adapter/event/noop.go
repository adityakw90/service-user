package event

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// NoOpPublisher is a no-op publisher for testing or when events are disabled.
type NoOpPublisher struct{}

// NewNoOpPublisher creates a no-op publisher.
func NewNoOpPublisher() portEvent.EventPublisher {
	return &NoOpPublisher{}
}

func (p *NoOpPublisher) Name() string {
	return "NoOpPublisher"
}

// Publish is a no-op.
func (p *NoOpPublisher) Publish(ctx context.Context, message event.Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	return nil
}

// Close is a no-op.
func (p *NoOpPublisher) Close() error {
	return nil
}
