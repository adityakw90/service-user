package event

import (
	"context"
	"fmt"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type MultiEventPublisher struct {
	publishers []portEvent.EventPublisher
	logger     monitoring.Logger
	tracer     monitoring.Tracer
}

func NewMultiEventPublisher(
	logger monitoring.Logger,
	tracer monitoring.Tracer,
	publishers ...portEvent.EventPublisher,
) *MultiEventPublisher {
	// Defensively copy and sanitize publishers.
	safePublishers := make([]portEvent.EventPublisher, 0, len(publishers))
	for _, publisher := range publishers {
		if publisher == nil {
			continue
		}
		safePublishers = append(safePublishers, publisher)
	}
	return &MultiEventPublisher{
		logger:     logger,
		tracer:     tracer,
		publishers: safePublishers,
	}
}

func (p *MultiEventPublisher) Name() string {
	return "MultiEventPublisher"
}

func (p *MultiEventPublisher) Publish(ctx context.Context, message event.Message) error {
	if err := message.Validate(); err != nil {
		return err
	}
	newCtx, span := p.tracer.StartSpan(ctx, "MultiEventPublisher.Publish")
	defer span.End()
	logger := p.logger.WithSpanContext(span.SpanContext())

	var firstErr error
	for _, publisher := range p.publishers {
		err := publisher.Publish(newCtx, message)
		if err != nil {
			logger.Error("failed to publish event", map[string]any{
				"publisher.name": publisher.Name(),
				"event.type":     message.Type,
				"error.type":     fmt.Sprintf("%T", err),
				"error.message":  err.Error(),
			})
			span.AddEvent("event published failed", trace.WithAttributes(
				attribute.String("publisher.name", publisher.Name()),
				attribute.String("event.type", string(message.Type)),
				attribute.String("error.type", fmt.Sprintf("%T", err)),
				attribute.String("error.message", err.Error()),
			))
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func (p *MultiEventPublisher) Close() error {
	var firstErr error
	for _, publisher := range p.publishers {
		err := publisher.Close()
		if err != nil {
			p.logger.Error("failed to close publisher", map[string]any{
				"publisher.name": publisher.Name(),
				"error.type":     fmt.Sprintf("%T", err),
				"error.message":  err.Error(),
			})
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}
