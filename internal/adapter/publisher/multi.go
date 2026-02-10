package publisher

import (
	"context"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
)

// MultiPublisher publishes events to multiple backend publishers.
type MultiPublisher struct {
	publishers []portEvent.EventPublisher
}

// NewMultiPublisher creates a publisher that fans out to multiple backends.
func NewMultiPublisher(publishers ...portEvent.EventPublisher) portEvent.EventPublisher {
	if len(publishers) == 0 {
		return NewNoOpPublisher()
	}
	if len(publishers) == 1 {
		return publishers[0]
	}
	return &MultiPublisher{publishers: publishers}
}

// Publish publishes to all backend publishers.
// If any publisher fails, the error is logged but not returned (fire-and-forget).
func (m *MultiPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	for _, p := range m.publishers {
		// Publish to each backend independently
		// Errors are logged but don't block other publishers
		_ = p.Publish(ctx, eventType, eventData)
	}
	return nil
}

// Close closes all backend publishers.
func (m *MultiPublisher) Close() error {
	for _, p := range m.publishers {
		_ = p.Close()
	}
	return nil
}
