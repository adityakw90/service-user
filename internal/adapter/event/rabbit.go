package event

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/adityakw90/go-monitoring"
	"github.com/adityakw90/service-user/internal/core/domain/event"
	portEvent "github.com/adityakw90/service-user/internal/core/port/event"
	"github.com/adityakw90/service-user/pkg/util"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// RabbitConn defines the interface for RabbitMQ connection operations.
// This interface allows for mocking in tests while maintaining compatibility with *infra.Rabbit.
type RabbitConn interface {
	PublishWithConfirm(
		ctx context.Context,
		exchange string,
		routingKey string,
		contentType string,
		headers map[string]string,
		body []byte,
		deliveryMode *uint8,
		priority *uint8,
		timestamp *time.Time,
		expiration *time.Duration,
		maxRetries int,
		retryInterval time.Duration,
		confirmTimeout time.Duration,
	) error
	Close() error
}

// RabbitmqPublisher publishes events via RabbitMQ.
type RabbitmqPublisher struct {
	conn             RabbitConn
	exchange         string
	routingKeyPrefix string
	confirmTimeout   time.Duration
	maxRetries       int
	retryInterval    time.Duration
	logger           monitoring.Logger
	tracer           monitoring.Tracer
}

// RabbitmqPublisherConfig holds configuration for the RabbitMQ event publisher.
type RabbitmqPublisherConfig struct {
	Exchange         string
	RoutingKeyPrefix string
	ConfirmTimeout   time.Duration
	MaxRetries       int
	RetryInterval    time.Duration
}

func NewRabbitmqPublisher(
	conn RabbitConn,
	config RabbitmqPublisherConfig,
	logger monitoring.Logger,
	tracer monitoring.Tracer,
) portEvent.EventPublisher {
	routingKeyPrefix := config.RoutingKeyPrefix
	if routingKeyPrefix != "" && !strings.HasSuffix(routingKeyPrefix, ".") {
		routingKeyPrefix += "."
	}
	return &RabbitmqPublisher{
		conn:             conn,
		exchange:         config.Exchange,
		routingKeyPrefix: routingKeyPrefix,
		confirmTimeout:   config.ConfirmTimeout,
		maxRetries:       config.MaxRetries,
		retryInterval:    config.RetryInterval,
		logger:           logger,
		tracer:           tracer,
	}
}

func (p *RabbitmqPublisher) Name() string {
	return "RabbitmqPublisher"
}

func (p *RabbitmqPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	newCtx, span := p.tracer.StartSpan(ctx, "RabbitmqPublisher.Publish")
	defer span.End()

	// Convert to CloudEvent format
	ce := NewCloudEvent(ctx, eventType, eventData)
	span.AddEvent(string(eventType),
		trace.WithAttributes(
			attribute.String("type", ce.Type),
			attribute.String("source", ce.Source),
			attribute.String("specversion", ce.SpecVersion),
			attribute.String("id", ce.ID),
		),
	)

	body, err := json.Marshal(ce.Data)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	// build headers
	headers := map[string]string{
		"ce_type":        ce.Type,
		"ce_source":      ce.Source,
		"ce_id":          ce.ID,
		"ce_specversion": ce.SpecVersion,
		"ce_timestamp":   strconv.FormatInt(ce.Time.UnixMilli(), 10),
		"client":         ce.Data.Client,
		"actor_id":       ce.Data.ActorId,
		"actor_type":     ce.Data.ActorType,
	}

	// inject trace header
	md := p.tracer.InjectContext(newCtx)
	for k, v := range md {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	err = p.conn.PublishWithConfirm(
		newCtx,
		p.exchange,
		p.getRoutingKey(string(ce.Type)),
		"application/cloudevents+json",
		headers,
		body,
		util.Ptr[uint8](2),  // amqp persistent
		util.Ptr[uint8](10), // amqp priority
		nil,                 // timestamp
		nil,                 // expiration
		p.maxRetries,
		p.retryInterval,
		p.confirmTimeout,
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	return nil
}

func (p *RabbitmqPublisher) Close() error {
	return p.conn.Close()
}

// getRoutingKey returns the full routing key for an event type.
func (p *RabbitmqPublisher) getRoutingKey(eventType string) string {
	return p.routingKeyPrefix + strings.TrimPrefix(eventType, ".")
}
