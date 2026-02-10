package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/adityakw90/service-user/internal/core/domain/event"
	"github.com/adityakw90/service-user/internal/infra"
	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher publishes events to RabbitMQ using the infra layer connection manager.
type RabbitMQPublisher struct {
	conn             *infra.RabbitMQConnection
	exchange         string
	exchangeType     string
	routingKeyPrefix string
	durable          bool
	source           string
}

// RabbitMQPublisherConfig holds configuration for the RabbitMQ event publisher.
type RabbitMQPublisherConfig struct {
	Source           string
	Exchange         string
	ExchangeType     string
	RoutingKeyPrefix string
	Durable          bool
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher.
// The exchange must be declared before calling this method, or by calling DeclareExchange().
func NewRabbitMQPublisher(conn *infra.RabbitMQConnection, config RabbitMQPublisherConfig) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		conn:             conn,
		exchange:         config.Exchange,
		exchangeType:     config.ExchangeType,
		routingKeyPrefix: config.RoutingKeyPrefix,
		durable:          config.Durable,
		source:           config.Source,
	}
}

// DeclareExchange declares the exchange on the RabbitMQ server.
func (r *RabbitMQPublisher) DeclareExchange() error {
	return r.conn.DeclareExchange(r.exchange, r.exchangeType, r.durable)
}

// Publish publishes an event to RabbitMQ with automatic reconnection handling.
func (r *RabbitMQPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, r.source)

	body, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	routingKey := r.getRoutingKey(string(eventType))

	err = r.conn.PublishWithContext(
		ctx,
		r.exchange,
		routingKey,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
			Headers: amqp.Table{
				"ce_type":        ce.Type,
				"ce_source":      ce.Source,
				"ce_id":          ce.ID,
				"ce_specversion": ce.SpecVersion,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to RabbitMQ: %w", err)
	}

	return nil
}

// getRoutingKey returns the full routing key for an event type.
func (r *RabbitMQPublisher) getRoutingKey(eventType string) string {
	return r.routingKeyPrefix + eventType
}

// Close closes the RabbitMQ connection.
func (r *RabbitMQPublisher) Close() error {
	return r.conn.Close()
}

// IsConnected returns true if the RabbitMQ connection is active.
func (r *RabbitMQPublisher) IsConnected() bool {
	return r.conn.IsConnected()
}
