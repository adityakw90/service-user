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
	// Publisher confirm settings
	confirmTimeout time.Duration
	maxRetries     int
	retryInterval  time.Duration
	// Queue declaration settings
	queueName          string
	queueDurable       bool
	queueAutoDelete    bool
	queueExclusive     bool
	queueEnabled       bool // Whether to declare and bind a queue
}

// RabbitMQPublisherConfig holds configuration for the RabbitMQ event publisher.
type RabbitMQPublisherConfig struct {
	Source           string
	Exchange         string
	ExchangeType     string
	RoutingKeyPrefix string
	Durable          bool
	// Publisher confirms (at-least-once delivery)
	ConfirmTimeout time.Duration
	MaxRetries     int
	RetryInterval  time.Duration
	// Queue declaration (stores messages when no consumers are running)
	QueueName       string
	QueueDurable    bool
	QueueAutoDelete bool
	QueueExclusive  bool
	QueueEnabled    bool // Set to true to declare and bind a queue
}

// NewRabbitMQPublisher creates a new RabbitMQ event publisher.
// The exchange must be declared before calling this method, or by calling SetupInfrastructure().
func NewRabbitMQPublisher(conn *infra.RabbitMQConnection, config RabbitMQPublisherConfig) *RabbitMQPublisher {
	// Set defaults for publisher confirms
	if config.ConfirmTimeout <= 0 {
		config.ConfirmTimeout = 30 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 5
	}
	if config.RetryInterval <= 0 {
		config.RetryInterval = 1 * time.Second
	}

	// Set defaults for queue
	if config.QueueName == "" {
		config.QueueName = config.Exchange + ".queue"
	}

	return &RabbitMQPublisher{
		conn:             conn,
		exchange:         config.Exchange,
		exchangeType:     config.ExchangeType,
		routingKeyPrefix: config.RoutingKeyPrefix,
		durable:          config.Durable,
		source:           config.Source,
		confirmTimeout:   config.ConfirmTimeout,
		maxRetries:       config.MaxRetries,
		retryInterval:    config.RetryInterval,
		queueName:        config.QueueName,
		queueDurable:     config.QueueDurable,
		queueAutoDelete:  config.QueueAutoDelete,
		queueExclusive:   config.QueueExclusive,
		queueEnabled:     config.QueueEnabled,
	}
}

// SetupInfrastructure declares the exchange (and optionally a queue) on the RabbitMQ server.
// When QueueEnabled is true, it also declares and binds a durable queue to store messages
// even when no consumers are running.
func (r *RabbitMQPublisher) SetupInfrastructure() error {
	// Declare the exchange
	if err := r.conn.DeclareExchange(r.exchange, r.exchangeType, r.durable); err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Optionally declare and bind queue
	if r.queueEnabled {
		// Declare the queue
		if err := r.conn.DeclareQueue(infra.QueueConfig{
			Name:       r.queueName,
			Durable:    r.queueDurable,
			AutoDelete: r.queueAutoDelete,
			Exclusive:  r.queueExclusive,
			Args:       nil,
		}); err != nil {
			return fmt.Errorf("failed to declare queue: %w", err)
		}

		// Bind queue to exchange with routing key pattern
		// For topic exchanges, use wildcards to match all event types
		bindingKey := r.routingKeyPrefix + "*"
		if err := r.conn.BindQueue(r.queueName, r.exchange, bindingKey, nil); err != nil {
			return fmt.Errorf("failed to bind queue to exchange: %w", err)
		}
	}

	return nil
}

// DeclareExchange is a legacy method that calls SetupInfrastructure.
// Deprecated: Use SetupInfrastructure() instead.
func (r *RabbitMQPublisher) DeclareExchange() error {
	return r.SetupInfrastructure()
}

// Publish publishes an event to RabbitMQ with at-least-once delivery semantics.
// Uses publisher confirms to ensure the broker has acknowledged the message.
func (r *RabbitMQPublisher) Publish(ctx context.Context, eventType event.EventType, eventData any) error {
	// Convert to CloudEvents format
	ce := toCloudEventData(eventType, eventData, r.source)

	body, err := json.Marshal(ce)
	if err != nil {
		return fmt.Errorf("failed to marshal cloud event: %w", err)
	}

	routingKey := r.getRoutingKey(string(eventType))

	err = r.conn.PublishWithConfirm(
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
		r.maxRetries,
		r.retryInterval,
		r.confirmTimeout,
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
